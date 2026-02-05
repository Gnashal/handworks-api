package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"handworks-api/types"
	"handworks-api/utils"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
)
type AdminTasks struct {}

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

