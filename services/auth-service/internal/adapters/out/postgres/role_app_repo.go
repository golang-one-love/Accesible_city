package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/accessible-path/auth-service/internal/domain/entity"
	"github.com/accessible-path/auth-service/internal/domain/service"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoleApplicationRepository struct {
	pool *pgxpool.Pool
}

func NewRoleApplicationRepository(pool *pgxpool.Pool) *RoleApplicationRepository {
	return &RoleApplicationRepository{pool: pool}
}

func (r *RoleApplicationRepository) Create(app *entity.RoleApplication) error {
	ctx := context.Background()
	query := `
		INSERT INTO role_applications (id, user_id, requested_role, comment, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query,
		app.ID, app.UserID, app.RequestedRole, app.Comment, app.Status, app.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create role application: %w", err)
	}
	return nil
}

func (r *RoleApplicationRepository) GetByID(id string) (*entity.RoleApplication, error) {
	ctx := context.Background()
	query := `
		SELECT id, user_id, requested_role, comment, status, reviewer_id, created_at, reviewed_at
		FROM role_applications WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	return r.scanApplication(row)
}

func (r *RoleApplicationRepository) ListByUser(userID string) ([]*entity.RoleApplication, error) {
	ctx := context.Background()
	query := `
		SELECT id, user_id, requested_role, comment, status, reviewer_id, created_at, reviewed_at
		FROM role_applications WHERE user_id = $1 ORDER BY created_at DESC
	`
	return r.listApplications(ctx, query, userID)
}

func (r *RoleApplicationRepository) List() ([]*entity.RoleApplication, error) {
	ctx := context.Background()
	query := `
		SELECT id, user_id, requested_role, comment, status, reviewer_id, created_at, reviewed_at
		FROM role_applications ORDER BY created_at DESC
	`
	return r.listApplications(ctx, query)
}

func (r *RoleApplicationRepository) Update(app *entity.RoleApplication) error {
	ctx := context.Background()
	query := `
		UPDATE role_applications
		SET status = $2, reviewer_id = $3, reviewed_at = $4
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, app.ID, app.Status, app.ReviewerID, app.ReviewedAt)
	if err != nil {
		return fmt.Errorf("update role application: %w", err)
	}
	return nil
}

func (r *RoleApplicationRepository) listApplications(ctx context.Context, query string, args ...any) ([]*entity.RoleApplication, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list role applications: %w", err)
	}
	defer rows.Close()

	apps := make([]*entity.RoleApplication, 0)
	for rows.Next() {
		app, err := r.scanApplication(rows)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list role applications: %w", err)
	}
	return apps, nil
}

func (r *RoleApplicationRepository) scanApplication(row interface {
	Scan(dest ...any) error
}) (*entity.RoleApplication, error) {
	var (
		id            string
		userID        string
		requestedRole string
		comment       sql.NullString
		status        string
		reviewerID    sql.NullString
		createdAt     sql.NullTime
		reviewedAt    sql.NullTime
	)

	if err := row.Scan(&id, &userID, &requestedRole, &comment, &status, &reviewerID, &createdAt, &reviewedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, service.ErrApplicationNotFound
		}
		return nil, fmt.Errorf("scan role application: %w", err)
	}

	app := &entity.RoleApplication{
		ID:            id,
		UserID:        userID,
		RequestedRole: entity.Role(requestedRole),
		Comment:       comment.String,
		Status:        entity.RoleApplicationStatus(status),
		ReviewerID:    reviewerID.String,
	}
	if createdAt.Valid {
		app.CreatedAt = createdAt.Time
	}
	if reviewedAt.Valid {
		app.ReviewedAt = reviewedAt.Time
	}
	return app, nil
}