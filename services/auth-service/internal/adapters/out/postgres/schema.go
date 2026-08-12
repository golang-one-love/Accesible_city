package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR(254) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
ALTER TABLE users ADD COLUMN IF NOT EXISTS nickname VARCHAR(100) NOT NULL DEFAULT '';
`

const createRoleApplicationsTable = `
CREATE TABLE IF NOT EXISTS role_applications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    requested_role VARCHAR(50) NOT NULL,
    comment VARCHAR(500) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    reviewer_id UUID,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_role_applications_user ON role_applications(user_id);
CREATE INDEX IF NOT EXISTS idx_role_applications_status ON role_applications(status);
`

func EnsureSchema(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, createUsersTable); err != nil {
		return err
	}
	_, err := pool.Exec(ctx, createRoleApplicationsTable)
	return err
}
