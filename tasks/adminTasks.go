package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"handworks-api/types"
	"handworks-api/utils"
	"math"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/organizationmembership"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/jackc/pgx/v5"
)

type AdminTasks struct{}
type AccountPort interface {
	SignUpCustomer(ctx context.Context, req types.SignUpCustomerRequest) (*types.SignUpCustomerResponse, error)
	SignUpEmployee(ctx context.Context, req types.SignUpEmployeeRequest) (*types.SignUpEmployeeResponse, error)
	SignUpAdmin(ctx context.Context, req types.SignUpAdminRequest) (*types.SignUpAdminResponse, error)
}

func (t *AdminTasks) resolveDateRange(filter string) (time.Time, time.Time, error) {
	now := time.Now()

	switch filter {
	case "", "week":
		start := now.AddDate(0, 0, -7)
		return start, now, nil
	case "month":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return start, now, nil
	case "year":
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		return start, now, nil
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("invalid date filter")
	}
}

func (t *AdminTasks) resolvePreviousRange(
	start time.Time,
	end time.Time,
) (time.Time, time.Time) {
	duration := end.Sub(start)
	prevEnd := start
	prevStart := start.Add(-duration)
	return prevStart, prevEnd
}

func calcGrowth(current, previous int32) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	return math.Round((float64(current-previous)/float64(previous))*100) / 100
}

func (t *AdminTasks) fetchTodayBookings(ctx context.Context, tx pgx.Tx) (int32, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var count int32
	err := tx.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM booking.basebookings bb
		WHERE bb.createdat >= $1
		  AND bb.createdat < $2
	`, startOfDay, endOfDay).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (t *AdminTasks) fetchPendingActions(ctx context.Context, tx pgx.Tx) (int32, error) {
	var pendingBookings int32
	err := tx.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM booking.basebookings bb
		WHERE bb.reviewstatus IN ('PENDING', 'REQUESTED')
	`).Scan(&pendingBookings)
	if err != nil {
		return 0, err
	}

	var unpaidOrders int32
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM payment.orders o
		WHERE LOWER(o.payment_status) LIKE 'pending%'
		   OR LOWER(o.payment_status) IN ('awaiting_payment_method', 'failed')
	`).Scan(&unpaidOrders)
	if err != nil {
		return 0, err
	}

	var lowStockCount int32
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM inventory.items i
		WHERE i.status IN ('LOW', 'DANGER', 'OUT_OF_STOCK')
		   OR i.quantity <= 0
		   OR (i.max_quantity > 0 AND (i.quantity::numeric / i.max_quantity::numeric) <= 0.20)
	`).Scan(&lowStockCount)
	if err != nil {
		return 0, err
	}

	return pendingBookings + unpaidOrders + lowStockCount, nil
}

func (t *AdminTasks) fetchRevenueSummary(ctx context.Context, tx pgx.Tx, start, end time.Time) (*types.RevenueSummary, error) {
	var paid float64
	var unpaid float64

	err := tx.QueryRow(ctx, `
		SELECT
			COALESCE(SUM((o.total_amount - o.remaining_balance)::numeric), 0)::float8 AS paid,
			COALESCE(SUM(CASE
				WHEN LOWER(o.payment_status) IN ('pending_downpayment', 'pending_fullpayment') THEN o.remaining_balance
				ELSE 0
			END), 0)::float8 AS unpaid
		FROM payment.orders o
		WHERE o.created_at BETWEEN $1 AND $2
	`, start, end).Scan(&paid, &unpaid)
	if err != nil {
		return nil, err
	}

	return &types.RevenueSummary{
		Revenue: paid + unpaid,
		Paid:    paid,
		Unpaid:  unpaid,
	}, nil
}

func (t *AdminTasks) fetchClientSegmentation(ctx context.Context, tx pgx.Tx, start, end time.Time) (int32, int32, int32, int32, error) {
	var activeClients int32
	err := tx.QueryRow(ctx, `
		SELECT COUNT(DISTINCT bb.custid)::int
		FROM booking.basebookings bb
		WHERE bb.createdat BETWEEN $1 AND $2
	`, start, end).Scan(&activeClients)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	var newClients int32
	err = tx.QueryRow(ctx, `
		WITH customer_history AS (
			SELECT
				bb.custid,
				MIN(bb.createdat) AS first_booking,
				BOOL_OR(bb.createdat BETWEEN $1 AND $2) AS has_booking_in_period
			FROM booking.basebookings bb
			GROUP BY bb.custid
		)
		SELECT COUNT(*)::int
		FROM customer_history ch
		WHERE ch.has_booking_in_period
		  AND ch.first_booking BETWEEN $1 AND $2
	`, start, end).Scan(&newClients)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	var returningClients int32
	err = tx.QueryRow(ctx, `
		WITH customer_history AS (
			SELECT
				bb.custid,
				MIN(bb.createdat) AS first_booking,
				BOOL_OR(bb.createdat BETWEEN $1 AND $2) AS has_booking_in_period
			FROM booking.basebookings bb
			GROUP BY bb.custid
		)
		SELECT COUNT(*)::int
		FROM customer_history ch
		WHERE ch.has_booking_in_period
		  AND ch.first_booking < $1
	`, start, end).Scan(&returningClients)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	var inactiveClients int32
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM account.customers c
		LEFT JOIN booking.basebookings bb
			ON bb.custid = c.id
		   AND bb.createdat BETWEEN $1 AND $2
		WHERE bb.id IS NULL
	`, start, end).Scan(&inactiveClients)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return activeClients, newClients, returningClients, inactiveClients, nil
}

func (t *AdminTasks) fetchEmployeeStats(ctx context.Context, tx pgx.Tx) (int32, int32, error) {
	var active int32
	err := tx.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM account.employees e
		WHERE e.status IN ('ACTIVE', 'ONDUTY')
	`).Scan(&active)
	if err != nil {
		return 0, 0, err
	}

	var total int32
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM account.employees e
	`).Scan(&total)
	if err != nil {
		return 0, 0, err
	}

	return active, total, nil
}

func (t *AdminTasks) fetchLowStockItems(ctx context.Context, tx pgx.Tx, limit int32) ([]types.InventoryAlert, error) {
	rows, err := tx.Query(ctx, `
		SELECT i.name, i.quantity
		FROM inventory.items i
		WHERE i.status IN ('LOW', 'DANGER', 'OUT_OF_STOCK')
		   OR i.quantity <= 0
		   OR (i.max_quantity > 0 AND (i.quantity::numeric / i.max_quantity::numeric) <= 0.20)
		ORDER BY i.quantity ASC, i.updated_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	alerts := make([]types.InventoryAlert, 0)
	for rows.Next() {
		var item types.InventoryAlert
		if err := rows.Scan(&item.Name, &item.Stock); err != nil {
			return nil, err
		}
		item.ID = int32(len(alerts) + 1)
		alerts = append(alerts, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return alerts, nil
}

func (t *AdminTasks) fetchTopServices(ctx context.Context, tx pgx.Tx, start, end time.Time, limit int32) ([]types.TopService, error) {
	rows, err := tx.Query(ctx, `
		SELECT s.service_type::text AS service_name, COUNT(*)::int AS booking_count
		FROM booking.bookings b
		JOIN booking.basebookings bb ON bb.id = b.base_booking_id
		JOIN booking.services s ON s.id = b.main_service_id
		WHERE bb.createdat BETWEEN $1 AND $2
		GROUP BY s.service_type
		ORDER BY booking_count DESC
		LIMIT $3
	`, start, end, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := make([]types.TopService, 0)
	for rows.Next() {
		var item types.TopService
		if err := rows.Scan(&item.Name, &item.Bookings); err != nil {
			return nil, err
		}
		item.ID = int32(len(services) + 1)
		services = append(services, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return services, nil
}

func (t *AdminTasks) FetchAdminDashboardData(ctx context.Context, tx pgx.Tx, logger *utils.Logger, dateFilter string) (*types.AdminDashboardResponse, error) {
	logger.Info("admin dashboard analytics started: dateFilter=%s", dateFilter)
	start, end, err := t.resolveDateRange(dateFilter)
	if err != nil {
		logger.Error("admin dashboard analytics failed at resolveDateRange: %v", err)
		return nil, fmt.Errorf("invalid date filter format: %s", err)
	}
	logger.Debug("admin dashboard range resolved: start=%s end=%s", start.Format(time.RFC3339), end.Format(time.RFC3339))
	prevStart, prevEnd := t.resolvePreviousRange(start, end)
	logger.Debug("admin dashboard previous range: start=%s end=%s", prevStart.Format(time.RFC3339), prevEnd.Format(time.RFC3339))
	var curRaw []byte
	if err := tx.QueryRow(
		ctx,
		`SELECT admin.get_admin_dashboard_stats($1, $2)`,
		start, end,
	).Scan(&curRaw); err != nil {
		logger.Error("admin dashboard analytics failed at current stats query: %v", err)
		return nil, fmt.Errorf("fetch current dashboard stats: %w", err)
	}

	var prevRaw []byte
	if err := tx.QueryRow(
		ctx,
		`SELECT admin.get_admin_dashboard_stats($1, $2)`,
		prevStart, prevEnd,
	).Scan(&prevRaw); err != nil {
		logger.Error("admin dashboard analytics failed at previous stats query: %v", err)
		return nil, fmt.Errorf("fetch previous dashboard stats: %w", err)
	}

	var cur, prev types.DashboardData
	if err := json.Unmarshal(curRaw, &cur); err != nil {
		logger.Error("admin dashboard analytics failed at current stats unmarshal: %v", err)
		return nil, fmt.Errorf("unmarshal current dashboard stats: %w", err)
	}
	if err := json.Unmarshal(prevRaw, &prev); err != nil {
		logger.Error("admin dashboard analytics failed at previous stats unmarshal: %v", err)
		return nil, fmt.Errorf("unmarshal previous dashboard stats: %w", err)
	}
	growthIndex := &types.GrowthIndex{
		SalesGrowthIndex:          calcGrowth(cur.Sales, prev.Sales),
		BookingsGrowthIndex:       calcGrowth(cur.Bookings, prev.Bookings),
		ActiveSessionsGrowthIndex: calcGrowth(cur.ActiveSessions, prev.ActiveSessions),
	}

	todayBookings, err := t.fetchTodayBookings(ctx, tx)
	if err != nil {
		logger.Error("admin dashboard analytics failed at fetchTodayBookings: %v", err)
		return nil, fmt.Errorf("fetchTodayBookings: %w", err)
	}
	logger.Debug("admin dashboard metric todayBookings=%d", todayBookings)

	pendingActions, err := t.fetchPendingActions(ctx, tx)
	if err != nil {
		logger.Error("admin dashboard analytics failed at fetchPendingActions: %v", err)
		return nil, fmt.Errorf("fetchPendingActions: %w", err)
	}
	logger.Debug("admin dashboard metric pendingActions=%d", pendingActions)

	revenueSummary, err := t.fetchRevenueSummary(ctx, tx, start, end)
	if err != nil {
		logger.Error("admin dashboard analytics failed at fetchRevenueSummary: %v", err)
		return nil, fmt.Errorf("fetchRevenueSummary: %w", err)
	}
	logger.Debug("admin dashboard metric revenue=%.2f paid=%.2f unpaid=%.2f", revenueSummary.Revenue, revenueSummary.Paid, revenueSummary.Unpaid)

	activeClients, newClients, returningClients, inactiveClients, err := t.fetchClientSegmentation(ctx, tx, start, end)
	if err != nil {
		logger.Error("admin dashboard analytics failed at fetchClientSegmentation: %v", err)
		return nil, fmt.Errorf("fetchClientSegmentation: %w", err)
	}
	logger.Debug("admin dashboard metric clients active=%d new=%d returning=%d inactive=%d", activeClients, newClients, returningClients, inactiveClients)

	employeesActive, employeesTotal, err := t.fetchEmployeeStats(ctx, tx)
	if err != nil {
		logger.Error("admin dashboard analytics failed at fetchEmployeeStats: %v", err)
		return nil, fmt.Errorf("fetchEmployeeStats: %w", err)
	}
	logger.Debug("admin dashboard metric employees active=%d total=%d", employeesActive, employeesTotal)

	lowStockItems, err := t.fetchLowStockItems(ctx, tx, 5)
	if err != nil {
		logger.Error("admin dashboard analytics failed at fetchLowStockItems: %v", err)
		return nil, fmt.Errorf("fetchLowStockItems: %w", err)
	}
	logger.Debug("admin dashboard metric lowStockItems=%d", len(lowStockItems))

	topServices, err := t.fetchTopServices(ctx, tx, start, end, 5)
	if err != nil {
		logger.Error("admin dashboard analytics failed at fetchTopServices: %v", err)
		return nil, fmt.Errorf("fetchTopServices: %w", err)
	}
	logger.Debug("admin dashboard metric topServices=%d", len(topServices))
	logger.Info("admin dashboard analytics completed successfully")

	return &types.AdminDashboardResponse{
		Sales:          cur.Sales,
		Bookings:       cur.Bookings,
		ActiveSessions: cur.ActiveSessions,
		Clients:        cur.Clients,
		GrowthIndex:    *growthIndex,

		TodayBookings:    todayBookings,
		PendingActions:   pendingActions,
		Revenue:          revenueSummary.Revenue,
		Paid:             revenueSummary.Paid,
		Unpaid:           revenueSummary.Unpaid,
		ActiveClients:    activeClients,
		NewClients:       newClients,
		ReturningClients: returningClients,
		InactiveClients:  inactiveClients,
		EmployeesActive:  employeesActive,
		EmployeesTotal:   employeesTotal,
		LowStockItems:    lowStockItems,
		UnreadMessages:   0,
		RecentActivities: []types.RecentActivity{},
		TopServices:      topServices,
	}, nil
}
func (t *AdminTasks) CreateClerkUser(ctx context.Context, req *types.OnboardEmployeeRequest) (*clerk.User, error) {
	params := &user.CreateParams{
		EmailAddresses:          &[]string{req.Email},
		FirstName:               &req.FirstName,
		LastName:                &req.LastName,
		SkipPasswordRequirement: clerk.Bool(true),
	}
	newUser, err := user.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create clerk user: %w", err)
	}
	return newUser, nil
}
func (t *AdminTasks) AcceptBooking(ctx context.Context, tx pgx.Tx, bookingID string) error {
	_, err := tx.Exec(ctx,
		`UPDATE booking.basebookings bb
		SET reviewstatus = 'SCHEDULED'
		FROM booking.bookings b
		WHERE bb.id = b.base_booking_id
		AND b.id = $1`,
		bookingID,
	)
	return err
}

func (t *AdminTasks) AssignResourcesToBooking(ctx context.Context, tx pgx.Tx, bookingID string, resources []types.ItemQuantity) error {
	// Remove existing resource usage records for this booking
	_, err := tx.Exec(ctx,
		`UPDATE booking.bookings
		 SET resource_ids = '{}'::uuid[]
		 WHERE id = $1`,
		bookingID,
	)
	if err != nil {
		return fmt.Errorf("failed to clear resource_ids: %w", err)
	}

	var usedIDs []string
	for _, r := range resources {
		var usedID string
		err = tx.QueryRow(ctx,
			`INSERT INTO booking.booking_inventory_used (item_id, item_type, quantity_used)
			 VALUES ($1, 'RESOURCE', $2)
			 RETURNING id`,
			r.ItemID, r.Quantity,
		).Scan(&usedID)
		if err != nil {
			return fmt.Errorf("failed to insert resource usage for item %s: %w", r.ItemID, err)
		}
		usedIDs = append(usedIDs, usedID)
	}

	_, err = tx.Exec(ctx,
		`UPDATE booking.bookings
		 SET resource_ids = $2::uuid[]
		 WHERE id = $1`,
		bookingID, usedIDs,
	)
	return err
}

func (t *AdminTasks) AssignEquipmentToBooking(ctx context.Context, tx pgx.Tx, bookingID string, equipment []types.ItemQuantity) error {
	// Remove existing equipment usage records for this booking
	_, err := tx.Exec(ctx,
		`UPDATE booking.bookings
		 SET equipment_ids = '{}'::uuid[]
		 WHERE base_booking_id = $1`,
		bookingID,
	)
	if err != nil {
		return fmt.Errorf("failed to clear equipment_ids: %w", err)
	}

	var usedIDs []string
	for _, e := range equipment {
		var usedID string
		err = tx.QueryRow(ctx,
			`INSERT INTO booking.booking_inventory_used (item_id, item_type, quantity_used)
			 VALUES ($1, 'EQUIPMENT', $2)
			 RETURNING id`,
			e.ItemID, e.Quantity,
		).Scan(&usedID)
		if err != nil {
			return fmt.Errorf("failed to insert equipment usage for item %s: %w", e.ItemID, err)
		}
		usedIDs = append(usedIDs, usedID)
	}

	_, err = tx.Exec(ctx,
		`UPDATE booking.bookings
		 SET equipment_ids = $2::uuid[]
		 WHERE base_booking_id = $1`,
		bookingID, usedIDs,
	)
	return err
}

func (t *AdminTasks) AddToClerkOrganization(ctx context.Context, clerkUserID, organizationID, role string) (*clerk.OrganizationMembership, error) {
	params := &organizationmembership.CreateParams{
		OrganizationID: *clerk.String(organizationID),
		UserID:         clerk.String(clerkUserID),
		Role:           clerk.String(role),
	}

	membership, err := organizationmembership.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to add user to organization: %w", err)
	}
	return membership, nil
}
func (t *AdminTasks) FetchCalendarBookings(ctx context.Context, tx pgx.Tx) (*types.CalendarBookingResponse, error) {
	var rawJSON []byte

	err := tx.QueryRow(ctx,
		`SELECT booking.get_calendar_bookings()`,
	).Scan(&rawJSON)
	if err != nil {
		return nil, fmt.Errorf("failed calling sproc get_calendar_bookings: %w", err)
	}

	var response types.CalendarBookingResponse
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal calendar bookings: %w", err)
	}

	return &response, nil
}
