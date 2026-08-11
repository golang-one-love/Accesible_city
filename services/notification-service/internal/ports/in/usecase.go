package in

import (
	"github.com/accessible-path/notification-service/internal/domain/entity"
	"github.com/google/uuid"
)

type NotificationUseCase interface {
	GetNotifications(userID uuid.UUID, limit, offset int) ([]*entity.Notification, int, error)
	GetUnreadCount(userID uuid.UUID) (int, error)
	MarkAsRead(userID uuid.UUID, notificationID uuid.UUID) error
	MarkAllAsRead(userID uuid.UUID) error
}