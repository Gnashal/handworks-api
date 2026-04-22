package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"handworks-api/types"
	"handworks-api/utils"
	"math"
	"strings"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/organizationmembership"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/jackc/pgx/v5"
)

type AdminTasks struct{}

var (
	ErrBookingNotFound            = errors.New("booking not found")
	ErrEmployeeNotFoundOrInactive = errors.New("employee not found or inactive")
	ErrCleanerAlreadyAssigned     = errors.New("cleaner is already assigned to booking")
	ErrCleanerNotAssigned         = errors.New("cleaner is not assigned to booking")
	ErrCleanerHasConflict         = errors.New("cleaner has conflicting booking schedule")
)

type BookingScheduleWindow struct {
	StartSched time.Time
	EndSched   time.Time
}
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

		result, updateErr := tx.Exec(ctx,
			`UPDATE inventory.items
			 SET quantity = quantity - $2
			 WHERE id = $1`,
			r.ItemID, r.Quantity,
		)
		if updateErr != nil {
			return fmt.Errorf("failed to decrement resource inventory for item %s: %w", r.ItemID, updateErr)
		}
		if result.RowsAffected() == 0 {
			return fmt.Errorf("resource inventory item not found: %s", r.ItemID)
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
		 WHERE id = $1`,
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

		result, updateErr := tx.Exec(ctx,
			`UPDATE inventory.items
			 SET quantity = quantity - $2
			 WHERE id = $1`,
			e.ItemID, e.Quantity,
		)
		if updateErr != nil {
			return fmt.Errorf("failed to decrement equipment inventory for item %s: %w", e.ItemID, updateErr)
		}
		if result.RowsAffected() == 0 {
			return fmt.Errorf("equipment inventory item not found: %s", e.ItemID)
		}

		usedIDs = append(usedIDs, usedID)
	}

	_, err = tx.Exec(ctx,
		`UPDATE booking.bookings
		 SET equipment_ids = $2::uuid[]
		 WHERE id = $1`,
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
func (t *AdminTasks) FetchCalendarBookings(ctx context.Context, tx pgx.Tx, month string) (*types.CalendarBookingResponse, error) {
	var rawJSON []byte
	start, end, err := utils.GetCurrentCalendarMonth(month)
	if err != nil {
		return nil, err
	}

	err = tx.QueryRow(ctx,
		`SELECT booking.get_calendar_bookings($1, $2)`,
		start,
		end,
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

func (t *AdminTasks) FetchBookingTrends(ctx context.Context, tx pgx.Tx) (*types.BookingTrendsResponse, error) {
	now := time.Now()
	loc := now.Location()

	weekStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, -6)
	weekEnd := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)

	weeklyRows, err := tx.Query(ctx, `
		SELECT date(bb.createdat) AS day_bucket, COUNT(*)::int AS bookings_count
		FROM booking.basebookings bb
		WHERE bb.createdat >= $1
		  AND bb.createdat < $2
		GROUP BY day_bucket
		ORDER BY day_bucket
	`, weekStart, weekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weekly booking trends: %w", err)
	}
	defer weeklyRows.Close()

	weeklyCounts := make(map[string]int32)
	for weeklyRows.Next() {
		var bucket time.Time
		var count int32
		if err := weeklyRows.Scan(&bucket, &count); err != nil {
			return nil, fmt.Errorf("failed to scan weekly booking trends: %w", err)
		}
		weeklyCounts[bucket.In(loc).Format("2006-01-02")] = count
	}
	if err := weeklyRows.Err(); err != nil {
		return nil, fmt.Errorf("weekly booking trends rows error: %w", err)
	}

	weeklyData := make([]types.BookingTrendPoint, 0, 7)
	for i := 0; i < 7; i++ {
		day := weekStart.AddDate(0, 0, i)
		key := day.Format("2006-01-02")
		weeklyData = append(weeklyData, types.BookingTrendPoint{
			Label: day.Format("Mon"),
			Value: weeklyCounts[key],
		})
	}

	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc).AddDate(0, -11, 0)
	nextMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc).AddDate(0, 1, 0)

	monthlyRows, err := tx.Query(ctx, `
		SELECT date_trunc('month', bb.createdat) AS month_bucket, COUNT(*)::int AS bookings_count
		FROM booking.basebookings bb
		WHERE bb.createdat >= $1
		  AND bb.createdat < $2
		GROUP BY month_bucket
		ORDER BY month_bucket
	`, monthStart, nextMonthStart)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch monthly booking trends: %w", err)
	}
	defer monthlyRows.Close()

	monthlyCounts := make(map[string]int32)
	for monthlyRows.Next() {
		var bucket time.Time
		var count int32
		if err := monthlyRows.Scan(&bucket, &count); err != nil {
			return nil, fmt.Errorf("failed to scan monthly booking trends: %w", err)
		}
		monthlyCounts[bucket.In(loc).Format("2006-01")] = count
	}
	if err := monthlyRows.Err(); err != nil {
		return nil, fmt.Errorf("monthly booking trends rows error: %w", err)
	}

	monthlyData := make([]types.BookingTrendPoint, 0, 12)
	for i := 0; i < 12; i++ {
		month := monthStart.AddDate(0, i, 0)
		key := month.Format("2006-01")
		monthlyData = append(monthlyData, types.BookingTrendPoint{
			Label: month.Format("Jan"),
			Value: monthlyCounts[key],
		})
	}

	return &types.BookingTrendsResponse{
		WeeklyData:  weeklyData,
		MonthlyData: monthlyData,
	}, nil
}

func (t *AdminTasks) GetBookingScheduleWindow(ctx context.Context, tx pgx.Tx, bookingID string) (*BookingScheduleWindow, error) {
	var out BookingScheduleWindow
	err := tx.QueryRow(ctx, `
		SELECT bb.startsched, bb.endsched
		FROM booking.bookings b
		JOIN booking.basebookings bb ON bb.id = b.base_booking_id
		WHERE b.id = $1
	`, bookingID).Scan(&out.StartSched, &out.EndSched)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("failed to load booking schedule window: %w", err)
	}

	return &out, nil
}

func (t *AdminTasks) ValidateActiveCleaner(ctx context.Context, tx pgx.Tx, employeeID string) error {
	var found bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM account.employees e
			WHERE e.id = $1
			  AND e.position = 'cleaner'
			  AND e.status IN ('ACTIVE', 'ONDUTY')
		)
	`, employeeID).Scan(&found)
	if err != nil {
		return fmt.Errorf("failed to validate cleaner: %w", err)
	}
	if !found {
		return ErrEmployeeNotFoundOrInactive
	}
	return nil
}

func (t *AdminTasks) CleanerHasScheduleConflict(
	ctx context.Context,
	tx pgx.Tx,
	employeeID string,
	startSched, endSched time.Time,
	excludeBookingID string,
) (bool, error) {
	var count int
	err := tx.QueryRow(ctx, `
		SELECT COUNT(1)::int
		FROM booking.bookings b
		JOIN booking.basebookings bb ON bb.id = b.base_booking_id
		WHERE $1 = ANY(b.cleaner_ids)
		  AND bb.startsched < $3
		  AND bb.endsched > $2
		  AND (
			NULLIF($4, '')::uuid IS NULL
			OR b.id <> NULLIF($4, '')::uuid
		  )
		  AND UPPER(COALESCE(bb.reviewstatus, '')) NOT IN ('CANCELLED', 'REJECTED')
		  AND UPPER(COALESCE(bb.status, '')) <> 'CANCELLED'
	`, employeeID, startSched, endSched, excludeBookingID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed checking cleaner schedule conflict: %w", err)
	}

	return count > 0, nil
}

func (t *AdminTasks) FetchAvailableCleanersByBooking(
	ctx context.Context,
	tx pgx.Tx,
	bookingID string,
	startSched, endSched time.Time,
) ([]types.AvailableCleaner, error) {
	rows, err := tx.Query(ctx, `
		SELECT
			e.id,
			a.first_name,
			a.last_name
		FROM account.employees e
		JOIN account.accounts a ON a.id = e.account_id
		WHERE e.position = 'cleaner'
		  AND e.status IN ('ACTIVE', 'ONDUTY')
		  AND NOT EXISTS (
			SELECT 1
			FROM booking.bookings b
			JOIN booking.basebookings bb ON bb.id = b.base_booking_id
			WHERE e.id = ANY(b.cleaner_ids)
			  AND bb.startsched < $2
			  AND bb.endsched > $1
			  AND b.id <> $3::uuid
			  AND UPPER(COALESCE(bb.reviewstatus, '')) NOT IN ('CANCELLED', 'REJECTED')
			  AND UPPER(COALESCE(bb.status, '')) <> 'CANCELLED'
		  )
		ORDER BY a.last_name ASC, a.first_name ASC
	`, startSched, endSched, bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch available cleaners: %w", err)
	}
	defer rows.Close()

	out := make([]types.AvailableCleaner, 0)
	for rows.Next() {
		var cleaner types.AvailableCleaner
		if scanErr := rows.Scan(&cleaner.EmployeeID, &cleaner.FirstName, &cleaner.LastName); scanErr != nil {
			return nil, fmt.Errorf("failed scanning available cleaner row: %w", scanErr)
		}
		out = append(out, cleaner)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed iterating available cleaner rows: %w", err)
	}

	return out, nil
}

func (t *AdminTasks) AssignEmployeeToBooking(
	ctx context.Context,
	tx pgx.Tx,
	bookingID, employeeID string,
	action types.AssignEmployeeAction,
) error {
	normalizedAction := strings.ToUpper(string(action))

	switch normalizedAction {
	case string(types.AssignEmployeeActionAdd):
		result, err := tx.Exec(ctx, `
			UPDATE booking.bookings b
			SET cleaner_ids = array_append(b.cleaner_ids, $2)
			WHERE b.id = $1
			  AND NOT ($2 = ANY(b.cleaner_ids))
		`, bookingID, employeeID)
		if err != nil {
			return fmt.Errorf("failed assigning cleaner to booking: %w", err)
		}
		if result.RowsAffected() == 0 {
			var exists bool
			if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM booking.bookings WHERE id = $1)`, bookingID).Scan(&exists); err != nil {
				return fmt.Errorf("failed validating booking existence after assign: %w", err)
			}
			if !exists {
				return ErrBookingNotFound
			}
			return ErrCleanerAlreadyAssigned
		}
		return nil
	case string(types.AssignEmployeeActionRemove):
		result, err := tx.Exec(ctx, `
			UPDATE booking.bookings b
			SET cleaner_ids = array_remove(b.cleaner_ids, $2)
			WHERE b.id = $1
			  AND $2 = ANY(b.cleaner_ids)
		`, bookingID, employeeID)
		if err != nil {
			return fmt.Errorf("failed unassigning cleaner from booking: %w", err)
		}
		if result.RowsAffected() == 0 {
			var exists bool
			if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM booking.bookings WHERE id = $1)`, bookingID).Scan(&exists); err != nil {
				return fmt.Errorf("failed validating booking existence after unassign: %w", err)
			}
			if !exists {
				return ErrBookingNotFound
			}
			return ErrCleanerNotAssigned
		}
		return nil
	default:
		return fmt.Errorf("invalid action: %s", action)
	}
}
