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
type AdminTasks struct {}
type AccountPort interface {
	SignUpCustomer(ctx context.Context, req types.SignUpCustomerRequest) (*types.SignUpCustomerResponse, error)
	SignUpEmployee(ctx context.Context, req types.SignUpEmployeeRequest) (*types.SignUpEmployeeResponse, error)
	SignUpAdmin(ctx context.Context, req types.SignUpAdminRequest) (*types.SignUpAdminResponse, error)
}

func (t *AdminTasks) resolveDateRange(filter string) (time.Time, time.Time, error) {
	now := time.Now()

	switch filter {
	case "week":
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
	return math.Round((float64(current-previous) / float64(previous)) * 100) / 100
}

func (t *AdminTasks) FetchAdminDashboardData(ctx context.Context, tx pgx.Tx, logger *utils.Logger, dateFilter string) (*types.AdminDashboardResponse, error) {
	start, end, err := t.resolveDateRange(dateFilter)
	if err != nil {
		logger.Error("invalid date format")
		return nil, fmt.Errorf("invalid date filter format: %s", err)
	}
	prevStart, prevEnd := t.resolvePreviousRange(start, end)
	var curRaw []byte
	if err := tx.QueryRow(
		ctx,
		`SELECT admin.get_admin_dashboard_stats($1, $2)`,
		start, end,
	).Scan(&curRaw); err != nil {
		return nil, err
	}

	var prevRaw []byte
	if err := tx.QueryRow(
		ctx,
		`SELECT admin.get_admin_dashboard_stats($1, $2)`,
		prevStart, prevEnd,
	).Scan(&prevRaw); err != nil {
		return nil, err
	}

	var cur, prev types.DashboardData
	if err := json.Unmarshal(curRaw, &cur); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(prevRaw, &prev); err != nil {
		return nil, err
	}
	growthIndex := &types.GrowthIndex{
		SalesGrowthIndex:          calcGrowth(cur.Sales, prev.Sales),
		BookingsGrowthIndex:       calcGrowth(cur.Bookings, prev.Bookings),
		ActiveSessionsGrowthIndex: calcGrowth(cur.ActiveSessions, prev.ActiveSessions),
	}
	
	return &types.AdminDashboardResponse{
		Sales: cur.Sales,
		Bookings: cur.Bookings,
		ActiveSessions: cur.ActiveSessions,
		Clients: cur.Clients,
		GrowthIndex: *growthIndex,
	}, nil
}
func (t *AdminTasks) CreateClerkUser(ctx context.Context, req* types.OnboardEmployeeRequest) (*clerk.User, error) {
	params := &user.CreateParams{
		EmailAddresses: &[]string{req.Email},
		FirstName: &req.FirstName,
		LastName: &req.LastName,
		SkipPasswordRequirement: clerk.Bool(true),

	}
	newUser, err := user.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create clerk user: %w", err)
	}
	return newUser, nil
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

