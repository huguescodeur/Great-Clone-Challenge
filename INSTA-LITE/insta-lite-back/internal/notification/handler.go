package notification

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/huguescodeur/insta-lite/internal/pkg/ctxkeys"
)

type NotificationHandler struct {
	service *NotificationService
}

func NewNotificationHandler(s *NotificationService) *NotificationHandler {
	return &NotificationHandler{service: s}
}

// @Summary      List notifications
// @Description  Returns the authenticated user's notifications in reverse chronological order with cursor-based pagination and an unread count
// @Tags         Notifications
// @Security     BearerAuth
// @Param        cursor  query     string  false  "Pagination cursor"
// @Param        limit   query     int     false  "Number of results (default: 20, max: 100)"
// @Success      200     {object}  notification.PaginatedNotificationResponse
// @Failure      500     {string}  string  "Internal server error"
// @Router       /notifications [get]
func (h *NotificationHandler) GetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := r.Context().Value(ctxkeys.UserContextKey).(ctxkeys.UserContext)

	cursorStr := r.URL.Query().Get("cursor")
	limit := 20
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 {
		limit = l
	}
	if limit > 100 {
		limit = 100
	}

	res, err := h.service.GetNotifications(ctx, userCtx.UserID, cursorStr, limit)
	if err != nil {
		http.Error(w, "erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

// @Summary      Mark notification as read
// @Description  Marks a single notification as read. Only the recipient can mark their own notification.
// @Tags         Notifications
// @Security     BearerAuth
// @Param        id   path      string  true  "Notification ID"
// @Success      204  "No content"
// @Failure      400  {string}  string  "Invalid notification ID"
// @Failure      404  {string}  string  "Notification not found"
// @Failure      500  {string}  string  "Internal server error"
// @Router       /notifications/{id}/read [patch]
func (h *NotificationHandler) MarkAsReadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := r.Context().Value(ctxkeys.UserContextKey).(ctxkeys.UserContext)

	notificationID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "incorrect notification ID", http.StatusBadRequest)
		return
	}

	if err := h.service.MarkAsRead(ctx, userCtx.UserID, notificationID); err != nil {
		if errors.Is(err, ErrNotificationNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *NotificationHandler) NotificationRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetHandler)
	r.Patch("/{id}/read", h.MarkAsReadHandler)
	return r
}
