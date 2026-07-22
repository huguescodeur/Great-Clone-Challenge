package post

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/huguescodeur/insta-lite/internal/pkg/ctxkeys"
	"github.com/huguescodeur/insta-lite/internal/pkg/responses"
)

type PostHandler struct {
	postService *PostService
	validate    *validator.Validate
}

func NewPostHandler(ps *PostService) *PostHandler {
	return &PostHandler{postService: ps, validate: validator.New()}
}

// GetAllHandler godoc
// @Summary      Liste tous les posts (feed)
// @Description  Retourne les posts paginés par curseur (cursor-based pagination)
// @Tags         Posts
// @Produce      json
// @Security     BearerAuth
// @Param        cursor  query     string  false  "Curseur de pagination (opaque string)"
// @Param        limit   query     int     false  "Nombre de résultats (défaut: 10, max: 100)"
// @Success      200     {object}  PaginatedPostResponse
// @Failure      500     {string}  string  "Erreur interne"
// @Router       /posts [get]
func (h *PostHandler) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	res, err := h.postService.GetAllWithMedia(ctx, cursorStr, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

// GetAllByIDHandler godoc
// @Summary      Posts d'un utilisateur
// @Description  Retourne les posts d'un utilisateur précis, paginés par curseur
// @Tags         Posts
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string  true   "UUID de l'utilisateur"
// @Param        cursor  query     string  false  "Curseur de pagination"
// @Param        limit   query     int     false  "Nombre de résultats (défaut: 10, max: 100)"
// @Success      200     {object}  PaginatedPostResponse
// @Failure      400     {string}  string  "ID incorrect"
// @Failure      500     {string}  string  "Erreur interne"
// @Router       /posts/{id} [get]
func (h *PostHandler) GetAllByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "incorrect ID", http.StatusBadRequest)
		return
	}

	res, err := h.postService.GetAllWithMediaByID(ctx, userID, cursorStr, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)

}

// UpdateHandler godoc
// @Summary      Modifier un post
// @Description  Met à jour le contenu d'un post appartenant à l'utilisateur connecté
// @Tags         Posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      PostUpdateRequest  true  "Données de mise à jour"
// @Success      200   {object}  Post
// @Failure      400   {string}  string  "Format JSON invalide ou validation échouée"
// @Failure      404   {string}  string  "Post introuvable ou non autorisé"
// @Failure      500   {string}  string  "Erreur interne"
// @Router       /posts [patch]
func (h *PostHandler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := r.Context().Value(ctxkeys.UserContextKey).(ctxkeys.UserContext)

	var req PostUpdateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "format json invalide", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	p, err := h.postService.UpdatePost(ctx, req.Content, userCtx.UserID, req.PostID)
	if err != nil {
		if err.Error() == "post introuvable ou vous n'avez pas l'autorisation de le modifier" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, "erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(p)

}

// DeleteHandler godoc
// @Summary      Supprimer un post
// @Description  Supprime un post appartenant à l'utilisateur connecté
// @Tags         Posts
// @Accept       json
// @Security     BearerAuth
// @Param        body  body      PostDeleteRequest  true  "ID du post à supprimer"
// @Success      204   "Post supprimé"
// @Failure      400   {string}  string  "Format JSON invalide"
// @Failure      404   {string}  string  "Post introuvable"
// @Failure      500   {string}  string  "Erreur interne"
// @Router       /posts [delete]
func (h *PostHandler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := r.Context().Value(ctxkeys.UserContextKey).(ctxkeys.UserContext)

	var req PostDeleteRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "format json invalide", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	if err := h.postService.DeletePost(ctx, userCtx.UserID, req.PostID); err != nil {
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		if err.Error() == "post n'existe pas ou déjà supprimé" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		// http.Error(w, err.Error(), http.StatusInternalServerError)
		http.Error(w, "erreur interne du serveur", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateHandler godoc
// @Summary      Créer un post
// @Description  Crée un nouveau post avec au moins un média
// @Tags         Posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      PostCreateRequest  true  "Contenu et médias du post"
// @Success      201   {object}  PostResponse
// @Failure      400   {string}  string  "Format invalide ou aucun média"
// @Failure      500   {string}  string  "Erreur interne"
// @Router       /posts [post]
func (h *PostHandler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := r.Context().Value(ctxkeys.UserContextKey).(ctxkeys.UserContext)

	var req PostCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "format json invalide", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	postData := Post{
		UserID:  userCtx.UserID,
		Content: req.Content,
	}

	res, err := h.postService.CreatePost(ctx, postData, req.Medias)
	if err != nil {
		if err.Error() == "un post doit contenir au moins un média" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, "erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *PostHandler) PostRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.GetAllHandler)
	r.Get("/{id}", h.GetAllByIDHandler)
	r.Patch("/", h.UpdateHandler)
	r.Delete("/", h.DeleteHandler)
	r.Post("/", h.CreateHandler)

	return r
}
