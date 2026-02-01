package services

import (
	"context"
	"fmt"
	"handworks-api/types"

	"github.com/jackc/pgx/v5"
)

func (s *AdminService) withTx(
	ctx context.Context,
	fn func(pgx.Tx) error,
) (err error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				s.Logger.Error("rollback failed: %v", rbErr)
			}
		} else {
			err = tx.Commit(ctx)
		}
	}()
	return fn(tx)
}

func (s *AdminService) GetAdminDashboard(ctx context.Context, req *types.AdminDashboardRequest) (*types.AdminDashboardResponse, error) {
		var res *types.AdminDashboardResponse

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		res, err = s.Tasks.FetchAdminDashboardData(ctx, tx, s.Logger, req.DateFilter)
		return err
	}); err != nil {
		s.Logger.Error("Failed to fetch Quotes: %v", err)
		return nil, err
	}

	return res, nil
}