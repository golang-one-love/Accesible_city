package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

const createGraphCacheTable = `
CREATE TABLE IF NOT EXISTS osm_graph_cache (
    id INTEGER PRIMARY KEY,
    nodes JSONB NOT NULL,
    edges JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

const createRoutesTable = `
CREATE TABLE IF NOT EXISTS routes (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    start_lat DOUBLE PRECISION NOT NULL,
    start_lon DOUBLE PRECISION NOT NULL,
    finish_lat DOUBLE PRECISION NOT NULL,
    finish_lon DOUBLE PRECISION NOT NULL,
    mobility_profile VARCHAR(20) NOT NULL,
    points JSONB NOT NULL,
    total_distance DOUBLE PRECISION NOT NULL,
    max_severity INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_routes_user ON routes(user_id);
`

func EnsureSchema(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, createGraphCacheTable); err != nil {
		return err
	}
	_, err := pool.Exec(ctx, createRoutesTable)
	return err
}