package entity

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationBarrierApproved NotificationType = "barrier_approved"
	NotificationBarrierResolved NotificationType = "barrier_resolved"
	NotificationBarrierRejected NotificationType = "barrier_rejected"
	NotificationBarrierNearby   NotificationType = "barrier_nearby"
	NotificationNewPOI          NotificationType = "new_poi"
	NotificationRouteUpdated    NotificationType = "route_updated"
)

type Notification struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      NotificationType
	Title     string
	Message   string
	Payload   map[string]string
	IsRead    bool
	CreatedAt time.Time
	ReadAt    *time.Time
}

func NewNotification(userID uuid.UUID, notifType NotificationType, title, message string, payload map[string]string) *Notification {
	return &Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      notifType,
		Title:     title,
		Message:   message,
		Payload:   payload,
		IsRead:    false,
		CreatedAt: time.Now().UTC(),
	}
}

func (n *Notification) MarkAsRead() {
	n.IsRead = true
	now := time.Now().UTC()
	n.ReadAt = &now
}