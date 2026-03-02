package config

import (
	"context"
	"handworks-api/utils"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDB(log *utils.Logger, ctx context.Context) (*pgxpool.Pool, error) {
	connStr := os.Getenv("DB_CONN")
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, err
	}

	// Ping check for db
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
