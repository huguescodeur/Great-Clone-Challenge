package follow

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/huguescodeur/insta-like/internal/pkg/ctxkeys"
)

type FollowHandler struct {
	followService *FollowService
}

func NewFollowHandler(s *FollowService) *FollowHandler {
	return &FollowHandler{followService: s}
}

// CreateHandler godoc
// @Summary      Suivre un utilisateur
// @Description  L'utilisateur connecté commence à suivre un autre utilisateur
// @Tags         Follow
// @Produce      json
// @Security     BearerAuth
// @Param        userID  path      string  true  "UUID de l'utilisateur à suivre"
// @Success      201     {object}  Follow
// @Failure      400     {string}  string  "ID incorrect ou tentative de se suivre soi-même"
// @Failure      409     {string}  string  "Déjà en train de suivre"
// @Failure      500     {string}  string  "Erreur interne"
// @Router       /users/{userID}/follow [post]
func (h *FollowHandler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := r.Context().Value(ctxkeys.UserContextKey).(ctxkeys.UserContext)

	followeeID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, "incorrect user ID", http.StatusBadRequest)
		return
	}

	f, err := h.followService.CreateFollow(ctx, userCtx.UserID, followeeID)
	if err != nil {
		if errors.Is(err, ErrAlreadyFollowing) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if errors.Is(err, ErrCannotFollowSelf) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, "erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(f)
}

// DeleteHandler godoc
// @Summary      Ne plus suivre un utilisateur
// @Description  L'utilisateur connecté arrête de suivre un autre utilisateur
// @Tags         Follow
// @Security     BearerAuth
// @Param        userID  path      string  true  "UUID de l'utilisateur à ne plus suivre"
// @Success      204     "Unfollow réussi"
// @Failure      400     {string}  string  "ID incorrect"
// @Failure      404     {string}  string  "Relation de follow non trouvée"
// @Failure      500     {string}  string  "Erreur interne"
// @Router       /users/{userID}/follow [delete]
func (h *FollowHandler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := r.Context().Value(ctxkeys.UserContextKey).(ctxkeys.UserContext)

	followeeID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, "incorrect user ID", http.StatusBadRequest)
		return
	}

	if err := h.followService.Unfollow(ctx, userCtx.UserID, followeeID); err != nil {
		if errors.Is(err, ErrFollowNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, "erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetStatusHandler godoc
// @Summary      Statut de follow
// @Description  Vérifie si l'utilisateur connecté suit un autre utilisateur (pending/accepted)
// @Tags         Follow
// @Produce      json
// @Security     BearerAuth
// @Param        userID  path      string  true  "UUID de l'utilisateur cible"
// @Success      200     {object}  Follow
// @Failure      400     {string}  string  "ID incorrect"
// @Failure      404     {string}  string  "Relation non trouvée"
// @Failure      500     {string}  string  "Erreur interne"
// @Router       /users/{userID}/follow/status [get]
func (h *FollowHandler) GetStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := r.Context().Value(ctxkeys.UserContextKey).(ctxkeys.UserContext)

	followeeID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, "incorrect user ID", http.StatusBadRequest)
		return
	}

	f, err := h.followService.GetStatus(ctx, userCtx.UserID, followeeID)
	if err != nil {
		if errors.Is(err, ErrFollowNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, "erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(f)
}

// GetFollowersHandler godoc
// @Summary      Followers d'un utilisateur
// @Description  Retourne la liste paginée des followers d'un utilisateur
// @Tags         Follow
// @Produce      json
// @Security     BearerAuth
// @Param        userID  path      string  true   "UUID de l'utilisateur"
// @Param        page    query     int     false  "Numéro de page (défaut: 1)"
// @Param        limit   query     int     false  "Nombre de résultats (défaut: 20)"
// @Success      200     {object}  map[string]any
// @Failure      400     {string}  string  "ID incorrect"
// @Failure      500     {string}  string  "Erreur interne"
// @Router       /users/{userID}/followers [get]
func (h *FollowHandler) GetFollowersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, "incorrect user ID", http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	followers, total, err := h.followService.GetFollowers(ctx, userID, limit, offset)
	if err != nil {
		http.Error(w, "erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"followers": followers,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

// GetFollowingHandler godoc
// @Summary      Abonnements d'un utilisateur
// @Description  Retourne la liste paginée des utilisateurs que suit un utilisateur
// @Tags         Follow
// @Produce      json
// @Security     BearerAuth
// @Param        userID  path      string  true   "UUID de l'utilisateur"
// @Param        page    query     int     false  "Numéro de page (défaut: 1)"
// @Param        limit   query     int     false  "Nombre de résultats (défaut: 20)"
// @Success      200     {object}  map[string]any
// @Failure      400     {string}  string  "ID incorrect"
// @Failure      500     {string}  string  "Erreur interne"
// @Router       /users/{userID}/following [get]
func (h *FollowHandler) GetFollowingHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, "incorrect user ID", http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	following, total, err := h.followService.GetFollowing(ctx, userID, limit, offset)
	if err != nil {
		http.Error(w, "erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"following": following,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}
