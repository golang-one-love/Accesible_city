package http

type NotificationResponse struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Title     string            `json:"title"`
	Message   string            `json:"message"`
	Payload   map[string]string `json:"payload"`
	IsRead    bool              `json:"is_read"`
	CreatedAt string            `json:"created_at"`
	ReadAt    *string           `json:"read_at,omitempty"`
}

type NotificationsListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	Total         int                    `json:"total"`
	Limit         int                    `json:"limit"`
	Offset        int                    `json:"offset"`
}

type UnreadCountResponse struct {
	Count int `json:"count"`
}

type MarkAsReadRequest struct {
	NotificationID string `json:"notification_id" validate:"required"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type WSMessage struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}