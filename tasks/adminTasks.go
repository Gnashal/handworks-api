package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"handworks-api/types"
	"handworks-api/utils"
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

func (t *AdminTasks) FetchAdminDashboardData(ctx context.Context, tx pgx.Tx, logger *utils.Logger, dateFilter string) (*types.AdminDashboardResponse, error) {
	start, end, err := t.resolveDateRange(dateFilter)
	if err != nil {
		logger.Error("invalid date format")
		return nil, fmt.Errorf("invalid date filter format: %s", err)
	}
	var rawJSON []byte
	var res types.AdminDashboardResponse
	err = tx.QueryRow(ctx, `SELECT admin.get_admin_dashboard_stats($1, $2)`, 
	start, end).Scan(&rawJSON)
	if err != nil {
		return nil, fmt.Errorf("failed calling sproc get_admin_dashboard_stats: %w", err)
	}
	if err = json.Unmarshal(rawJSON, &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bookings: %w", err)
	}
	return &res, nil
}

