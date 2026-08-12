package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/accessible-path/auth-service/internal/domain/entity"
	"github.com/accessible-path/auth-service/internal/domain/service"
	"github.com/accessible-path/auth-service/internal/domain/valueobject"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(user *entity.User) error {
	ctx := context.Background()
	query := `
		INSERT INTO users (id, email, password_hash, nickname, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query,
		user.ID, user.Email.String(), user.PasswordHash, user.Nickname, user.Role, user.IsActive, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return service.ErrUserAlreadyExists
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByEmail(email valueobject.Email) (*entity.User, error) {
	ctx := context.Background()
	query := `
		SELECT id, email, password_hash, nickname, role, is_active, created_at, updated_at
		FROM users WHERE email = $1
	`
	row := r.pool.QueryRow(ctx, query, email.String())
	return r.scanUser(row)
}

func (r *UserRepository) GetByID(id string) (*entity.User, error) {
	ctx := context.Background()
	query := `
		SELECT id, email, password_hash, nickname, role, is_active, created_at, updated_at
		FROM users WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	return r.scanUser(row)
}

func (r *UserRepository) Update(user *entity.User) error {
	ctx := context.Background()
	query := `
		UPDATE users SET nickname = $2, password_hash = $3, role = $4, is_active = $5, updated_at = $6
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, user.ID, user.Nickname, user.PasswordHash, user.Role, user.IsActive, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (r *UserRepository) List() ([]*entity.User, error) {
	ctx := context.Background()
	query := `
		SELECT id, email, password_hash, nickname, role, is_active, created_at, updated_at
		FROM users ORDER BY created_at
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]*entity.User, 0)
	for rows.Next() {
		user, err := r.scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func (r *UserRepository) scanUser(row interface {
	Scan(dest ...any) error
}) (*entity.User, error) {
	var (
		id           string
		emailStr     string
		passwordHash string
		nickname     sql.NullString
		role         string
		isActive     bool
		createdAt    sql.NullTime
		updatedAt    sql.NullTime
	)

	if err := row.Scan(&id, &emailStr, &passwordHash, &nickname, &role, &isActive, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, service.ErrUserNotFound
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}

	email, err := valueobject.NewEmail(emailStr)
	if err != nil {
		return nil, fmt.Errorf("parse email: %w", err)
	}

	user := &entity.User{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
		Nickname:     nickname.String,
		Role:         entity.Role(role),
		IsActive:     isActive,
	}
	if createdAt.Valid {
		user.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		user.UpdatedAt = updatedAt.Time
	}
	return user, nil
}