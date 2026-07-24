package feed

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/huguescodeur/insta-lite/internal/pkg/ctxkeys"
)

type FeedHandler struct {
	feedService *FeedService
}

func NewFeedHandler(s *FeedService) *FeedHandler {
	return &FeedHandler{feedService: s}
}

// GetHandler godoc
// @Summary      Fil d'actualité personnalisé
// @Description  Retourne les posts des comptes suivis par l'utilisateur connecté, paginés par curseur
// @Tags         Feed
// @Produce      json
// @Security     BearerAuth
// @Param        cursor  query     string  false  "Curseur de pagination"
// @Param        limit   query     int     false  "Nombre de résultats (défaut: 10, max: 100)"
// @Success      200     {object}  post.PaginatedPostResponse
// @Failure      500     {string}  string  "Erreur interne"
// @Router       /feed [get]
func (h *FeedHandler) GetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := r.Context().Value(ctxkeys.UserContextKey).(ctxkeys.UserContext)

	cursorStr := r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if limit > 100 {
		limit = 100
	}

	res, err := h.feedService.GetFeed(ctx, userCtx.UserID, cursorStr, limit)
	if err != nil {
		http.Error(w, "erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *FeedHandler) FeedRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetHandler)
	return r
}
