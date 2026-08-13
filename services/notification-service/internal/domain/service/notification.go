package service

import (
	"errors"

	"github.com/accessible-path/notification-service/internal/domain/entity"
	"github.com/accessible-path/notification-service/internal/ports/out"
	"github.com/google/uuid"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrUnauthorized         = errors.New("unauthorized")
)

type NotificationService struct {
	repo       out.NotificationRepository
	eventBus   out.EventBus
	wsManager  out.WebSocketManager
}

func NewNotificationService(repo out.NotificationRepository, eventBus out.EventBus, wsManager out.WebSocketManager) *NotificationService {
	return &NotificationService{
		repo:      repo,
		eventBus:  eventBus,
		wsManager: wsManager,
	}
}

func (s *NotificationService) CreateNotification(userID uuid.UUID, notifType entity.NotificationType, title, message string, payload map[string]string) (*entity.Notification, error) {
	notif := entity.NewNotification(userID, notifType, title, message, payload)
	return s.repo.Create(notif)
}

func (s *NotificationService) GetNotifications(userID uuid.UUID, limit, offset int) ([]*entity.Notification, int, error) {
	return s.repo.ListByUser(userID, limit, offset)
}

func (s *NotificationService) GetUnreadCount(userID uuid.UUID) (int, error) {
	return s.repo.CountUnread(userID)
}

func (s *NotificationService) MarkAsRead(userID uuid.UUID, notificationID uuid.UUID) error {
	notif, err := s.repo.GetByID(notificationID)
	if err != nil {
		return ErrNotificationNotFound
	}
	if notif.UserID != userID {
		return ErrUnauthorized
	}
	notif.MarkAsRead()
	return s.repo.Update(notif)
}

func (s *NotificationService) MarkAllAsRead(userID uuid.UUID) error {
	return s.repo.MarkAllAsRead(userID)
}

func (s *NotificationService) HandleBarrierApproved(event map[string]interface{}) error {
	payload, _ := event["payload"].(map[string]interface{})
	barrierID, _ := payload["barrier_id"].(string)
	reporterID, _ := payload["reporter_id"].(string)

	userID := uuid.Nil
	if parsed, err := uuid.Parse(reporterID); err == nil {
		userID = parsed
	}

	notif := entity.NewNotification(
		userID,
		entity.NotificationBarrierApproved,
		"Барьер одобрен",
		"Новый барьер был одобрен модератором",
		map[string]string{"barrier_id": barrierID},
	)
	
	_, err := s.repo.Create(notif)
	if err != nil {
		return err
	}
	
	return s.wsManager.Broadcast(map[string]interface{}{
		"type": "notification",
		"data": map[string]interface{}{
			"id":        notif.ID.String(),
			"user_id":   userID.String(),
			"type":      notif.Type,
			"title":     notif.Title,
			"message":   notif.Message,
			"payload":   notif.Payload,
			"created_at": notif.CreatedAt,
		},
	})
}

func (s *NotificationService) HandleBarrierResolved(event map[string]interface{}) error {
	payload, _ := event["payload"].(map[string]interface{})
	barrierID, _ := payload["barrier_id"].(string)
	reporterID, _ := payload["reporter_id"].(string)

	userID := uuid.Nil
	if parsed, err := uuid.Parse(reporterID); err == nil {
		userID = parsed
	}

	notif := entity.NewNotification(
		userID,
		entity.NotificationBarrierResolved,
		"Барьер устранен",
		"Барьер на вашем маршруте был устранен",
		map[string]string{"barrier_id": barrierID},
	)

	_, err := s.repo.Create(notif)
	if err != nil {
		return err
	}

	return s.wsManager.Broadcast(map[string]interface{}{
		"type": "notification",
		"data": map[string]interface{}{
			"id":         notif.ID.String(),
			"user_id":    userID.String(),
			"type":       notif.Type,
			"title":      notif.Title,
			"message":    notif.Message,
			"payload":    notif.Payload,
			"created_at": notif.CreatedAt,
		},
	})
}

func (s *NotificationService) HandleBarrierRejected(event map[string]interface{}) error {
	payload, _ := event["payload"].(map[string]interface{})
	barrierID, _ := payload["barrier_id"].(string)
	reporterID, _ := payload["reporter_id"].(string)
	moderatorID, _ := payload["moderator_id"].(string)
	comment, _ := payload["comment"].(string)

	userID := uuid.Nil
	if parsed, err := uuid.Parse(reporterID); err == nil {
		userID = parsed
	}

	message := "Ваш барьер был отклонен модератором"
	if comment != "" {
		message = "Ваш барьер был отклонен модератором: " + comment
	}

	notif := entity.NewNotification(
		userID,
		entity.NotificationBarrierRejected,
		"Барьер отклонен",
		message,
		map[string]string{"barrier_id": barrierID, "moderator_id": moderatorID},
	)

	_, err := s.repo.Create(notif)
	if err != nil {
		return err
	}

	return s.wsManager.Broadcast(map[string]interface{}{
		"type": "notification",
		"data": map[string]interface{}{
			"id":         notif.ID.String(),
			"user_id":    userID.String(),
			"type":       notif.Type,
			"title":      notif.Title,
			"message":    notif.Message,
			"payload":    notif.Payload,
			"created_at": notif.CreatedAt,
		},
	})
}

func (s *NotificationService) HandleRouteUpdated(event map[string]interface{}) error {
	payload, _ := event["payload"].(map[string]interface{})
	routeID, _ := payload["route_id"].(string)
	userIDStr, _ := payload["user_id"].(string)
	barrierID, _ := payload["barrier_id"].(string)

	userID := uuid.Nil
	if parsed, err := uuid.Parse(userIDStr); err == nil {
		userID = parsed
	}

	notif := entity.NewNotification(
		userID,
		entity.NotificationRouteUpdated,
		"Маршрут перестроен",
		"На вашем сохраненном маршруте появилось препятствие, маршрут изменен",
		map[string]string{"route_id": routeID, "barrier_id": barrierID},
	)

	_, err := s.repo.Create(notif)
	if err != nil {
		return err
	}

	return s.wsManager.Broadcast(map[string]interface{}{
		"type": "notification",
		"data": map[string]interface{}{
			"id":         notif.ID.String(),
			"user_id":    userID.String(),
			"type":       notif.Type,
			"title":      notif.Title,
			"message":    notif.Message,
			"payload":    notif.Payload,
			"created_at": notif.CreatedAt,
		},
	})
}