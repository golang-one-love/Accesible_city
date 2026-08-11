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

func EnsureSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, createGraphCacheTable)
	return err
}
