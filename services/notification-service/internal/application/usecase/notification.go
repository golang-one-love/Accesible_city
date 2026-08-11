package usecase

import (
	"github.com/accessible-path/notification-service/internal/domain/entity"
	"github.com/accessible-path/notification-service/internal/domain/service"
	"github.com/accessible-path/notification-service/internal/ports/in"
	"github.com/google/uuid"
)

type notificationUseCase struct {
	notificationService *service.NotificationService
}

func NewNotificationUseCase(notificationService *service.NotificationService) in.NotificationUseCase {
	return &notificationUseCase{notificationService: notificationService}
}

func (u *notificationUseCase) GetNotifications(userID uuid.UUID, limit, offset int) ([]*entity.Notification, int, error) {
	return u.notificationService.GetNotifications(userID, limit, offset)
}

func (u *notificationUseCase) GetUnreadCount(userID uuid.UUID) (int, error) {
	return u.notificationService.GetUnreadCount(userID)
}

func (u *notificationUseCase) MarkAsRead(userID uuid.UUID, notificationID uuid.UUID) error {
	return u.notificationService.MarkAsRead(userID, notificationID)
}

func (u *notificationUseCase) MarkAllAsRead(userID uuid.UUID) error {
	return u.notificationService.MarkAllAsRead(userID)
}