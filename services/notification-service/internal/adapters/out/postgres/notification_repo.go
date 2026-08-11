package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/accessible-path/notification-service/internal/domain/entity"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"
)

type PostgresNotificationRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresNotificationRepository(pool *pgxpool.Pool) *PostgresNotificationRepository {
	return &PostgresNotificationRepository{pool: pool}
}

func (r *PostgresNotificationRepository) Create(notification *entity.Notification) (*entity.Notification, error) {
	ctx := context.Background()
	
	payloadJSON, _ := json.Marshal(notification.Payload)
	
	_, err := r.pool.Exec(ctx, `
		INSERT INTO notifications (id, user_id, type, title, message, payload, is_read, created_at, read_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, notification.ID, notification.UserID, notification.Type, notification.Title, notification.Message, payloadJSON, notification.IsRead, notification.CreatedAt, notification.ReadAt)
	
	if err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}
	
	return notification, nil
}

func (r *PostgresNotificationRepository) GetByID(id uuid.UUID) (*entity.Notification, error) {
	ctx := context.Background()
	
	var payloadJSON []byte
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, type, title, message, payload, is_read, created_at, read_at
		FROM notifications WHERE id = $1
	`, id)
	
	var n entity.Notification
	var readAt *string
	err := row.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message, &payloadJSON, &n.IsRead, &n.CreatedAt, &readAt)
	if err != nil {
		return nil, err
	}
	
	json.Unmarshal(payloadJSON, &n.Payload)
	if readAt != nil {
		t, _ := time.Parse(time.RFC3339, *readAt)
		n.ReadAt = &t
	}
	
	return &n, nil
}

func (r *PostgresNotificationRepository) Update(notification *entity.Notification) error {
	ctx := context.Background()
	
	payloadJSON, _ := json.Marshal(notification.Payload)
	
	_, err := r.pool.Exec(ctx, `
		UPDATE notifications SET is_read = $2, read_at = $3, payload = $4
		WHERE id = $1
	`, notification.ID, notification.IsRead, notification.ReadAt, payloadJSON)
	
	return err
}

func (r *PostgresNotificationRepository) ListByUser(userID uuid.UUID, limit, offset int) ([]*entity.Notification, int, error) {
	ctx := context.Background()
	
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, type, title, message, payload, is_read, created_at, read_at
		FROM notifications WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	notifications := make([]*entity.Notification, 0)
	for rows.Next() {
		var n entity.Notification
		var payloadJSON []byte
		var readAt *string
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message, &payloadJSON, &n.IsRead, &n.CreatedAt, &readAt); err != nil {
			return nil, 0, err
		}
		json.Unmarshal(payloadJSON, &n.Payload)
		if readAt != nil {
			t, _ := time.Parse(time.RFC3339, *readAt)
			n.ReadAt = &t
		}
		notifications = append(notifications, &n)
	}
	
	return notifications, total, nil
}

func (r *PostgresNotificationRepository) CountUnread(userID uuid.UUID) (int, error) {
	ctx := context.Background()
	
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`, userID).Scan(&count)
	return count, err
}

func (r *PostgresNotificationRepository) MarkAllAsRead(userID uuid.UUID) error {
	ctx := context.Background()
	
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		UPDATE notifications SET is_read = true, read_at = $1
		WHERE user_id = $2 AND is_read = false
	`, now, userID)
	
	return err
}