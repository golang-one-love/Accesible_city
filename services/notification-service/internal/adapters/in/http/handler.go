package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/accessible-path/notification-service/internal/domain/entity"
	"github.com/accessible-path/notification-service/internal/ports/in"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type NotificationHandler struct {
	notificationUseCase in.NotificationUseCase
	wsManager           interface {
		Register(userID uuid.UUID, conn interface{}) error
		Unregister(userID uuid.UUID) error
	}
}

func NewNotificationHandler(notificationUseCase in.NotificationUseCase, wsManager interface {
	Register(userID uuid.UUID, conn interface{}) error
	Unregister(userID uuid.UUID) error
}) *NotificationHandler {
	return &NotificationHandler{
		notificationUseCase: notificationUseCase,
		wsManager:           wsManager,
	}
}

func (h *NotificationHandler) RegisterRoutes(g *echo.Group) {
	g.GET("", h.GetNotifications)
	g.GET("/unread-count", h.GetUnreadCount)
	g.POST("/read", h.MarkAsRead)
	g.POST("/read-all", h.MarkAllAsRead)
	g.GET("/ws", h.WebSocket)
}

func (h *NotificationHandler) GetNotifications(c echo.Context) error {
	userID, err := userIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid user id"})
	}
	
	limit := 50
	offset := 0
	if l := c.QueryParam("limit"); l != "" {
		// parse limit
	}
	if o := c.QueryParam("offset"); o != "" {
		// parse offset
	}

	notifications, total, err := h.notificationUseCase.GetNotifications(userID, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, NotificationsListResponse{
		Notifications: toNotificationResponses(notifications),
		Total:         total,
		Limit:         limit,
		Offset:        offset,
	})
}

func (h *NotificationHandler) GetUnreadCount(c echo.Context) error {
	userID, err := userIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid user id"})
	}

	count, err := h.notificationUseCase.GetUnreadCount(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, UnreadCountResponse{Count: count})
}

func (h *NotificationHandler) MarkAsRead(c echo.Context) error {
	userID, err := userIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid user id"})
	}

	var req MarkAsReadRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	notificationID, err := uuid.Parse(req.NotificationID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid notification ID"})
	}

	if err := h.notificationUseCase.MarkAsRead(userID, notificationID); err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
	}

	return c.NoContent(http.StatusOK)
}

func (h *NotificationHandler) MarkAllAsRead(c echo.Context) error {
	userID, err := userIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid user id"})
	}

	if err := h.notificationUseCase.MarkAllAsRead(userID); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.NoContent(http.StatusOK)
}

func (h *NotificationHandler) WebSocket(c echo.Context) error {
	userID, err := userIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid user id"})
	}

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	if err := h.wsManager.Register(userID, ws); err != nil {
		ws.Close()
		return err
	}

	defer h.wsManager.Unregister(userID)

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
	
	return nil
}

func userIDFromContext(c echo.Context) (uuid.UUID, error) {
	value := c.Get("user_id")
	s, ok := value.(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("user id not found")
	}
	return uuid.Parse(s)
}

func toNotificationResponses(notifications []*entity.Notification) []NotificationResponse {
	resp := make([]NotificationResponse, len(notifications))
	for i, n := range notifications {
		var readAt *string
		if n.ReadAt != nil {
			s := n.ReadAt.Format(time.RFC3339)
			readAt = &s
		}
		resp[i] = NotificationResponse{
			ID:        n.ID.String(),
			Type:      string(n.Type),
			Title:     n.Title,
			Message:   n.Message,
			Payload:   n.Payload,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt.Format(time.RFC3339),
			ReadAt:    readAt,
		}
	}
	return resp
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}