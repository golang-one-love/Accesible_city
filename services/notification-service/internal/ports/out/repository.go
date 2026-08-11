package out

import (
	"github.com/accessible-path/notification-service/internal/domain/entity"
	"github.com/google/uuid"
)

type NotificationRepository interface {
	Create(notification *entity.Notification) (*entity.Notification, error)
	GetByID(id uuid.UUID) (*entity.Notification, error)
	Update(notification *entity.Notification) error
	ListByUser(userID uuid.UUID, limit, offset int) ([]*entity.Notification, int, error)
	CountUnread(userID uuid.UUID) (int, error)
	MarkAllAsRead(userID uuid.UUID) error
}

type EventBus interface {
	Subscribe(stream, group, consumer string, handler func(map[string]interface{})) error
}

type WebSocketManager interface {
	Broadcast(message map[string]interface{}) error
	Register(userID uuid.UUID, conn interface{}) error
	Unregister(userID uuid.UUID) error
}