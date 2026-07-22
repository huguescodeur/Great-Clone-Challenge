package notification

type PaginatedNotificationResponse struct {
	Data        []*Notification `json:"data"`
	NextCursor  string          `json:"nextCursor"`
	HasMore     bool            `json:"hasMore"`
	UnreadCount int64           `json:"unreadCount"`
}
