package user

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UserHandler struct {
	store UserStore
}

func NewUserHandler(store UserStore) *UserHandler {
	return &UserHandler{store: store}
}

// GetByIDHandler returns a user's public profile including live follower, following, and post counts.
//
// @Summary      Get user profile
// @Description  Returns the public profile of a user by ID, including live followers, following, and posts counts.
// @Tags         Users
// @Security     BearerAuth
// @Param        userID  path      string  true  "User ID (UUID)"
// @Success      200     {object}  user.User
// @Failure      400     {string}  string  "Invalid user ID"
// @Failure      404     {string}  string  "User not found"
// @Failure      500     {string}  string  "Internal server error"
// @Router       /users/{userID} [get]
func (h *UserHandler) GetByIDHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	u, err := h.store.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

// GetByUsernameHandler returns a user's public profile by username.
//
// @Summary      Get user by username
// @Description  Returns the public profile of a user by their username, including live followers, following, and posts counts.
// @Tags         Users
// @Security     BearerAuth
// @Param        username  path      string  true  "Username"
// @Success      200       {object}  user.User
// @Failure      404       {string}  string  "User not found"
// @Failure      500       {string}  string  "Internal server error"
// @Router       /users/by-username/{username} [get]
func (h *UserHandler) GetByUsernameHandler(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	u, err := h.store.GetByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func (h *UserHandler) UserRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/{userID}", h.GetByIDHandler)
	return r
}
