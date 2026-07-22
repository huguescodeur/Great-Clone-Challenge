package like

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/huguescodeur/insta-lite/internal/pkg/ctxkeys"
	"github.com/huguescodeur/insta-lite/internal/pkg/responses"
)

type LikeHandler struct {
	likeService *LikeService
	validate    *validator.Validate
}

func NewLikeHandler(l *LikeService) *LikeHandler {
	return &LikeHandler{likeService: l, validate: validator.New()}
}

// GetAllHandler godoc
// @Summary      Likes d'un post
// @Description  Retourne tous les likes d'un post
// @Tags         Post Likes
// @Produce      json
// @Security     BearerAuth
// @Param        postID  path      string  true  "UUID du post"
// @Success      200     {array}   Like
// @Failure      400     {string}  string  "ID invalide"
// @Failure      500     {string}  string  "Erreur interne"
// @Router       /posts/{postID}/likes [get]
func (h *LikeHandler) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	postID, err := uuid.Parse(chi.URLParam(r, "postID"))
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	likes, err := h.likeService.GetAllLikes(ctx, postID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(likes)
}

// GetByUserAndPostHandler godoc
// @Summary      Mon like sur un post
// @Description  Retourne le like de l'utilisateur connecté sur un post
// @Tags         Post Likes
// @Produce      json
// @Security     BearerAuth
// @Param        postID  path      string  true  "UUID du post"
// @Success      200     {object}  Like
// @Failure      400     {string}  string  "ID invalide"
// @Failure      404     {string}  string  "Like non trouvé"
// @Failure      500     {string}  string  "Erreur interne"
// @Router       /posts/{postID}/likes/me [get]
func (h *LikeHandler) GetByUserAndPostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := r.Context().Value(ctxkeys.UserContextKey).(ctxkeys.UserContext)
	postID, err := uuid.Parse(chi.URLParam(r, "postID"))
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	likes, err := h.likeService.GetLikeByUserAndPost(ctx, userCtx.UserID, postID)
	if err != nil {
		if errors.Is(err, ErrReactionNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(likes)

}

// UpdateReactionHandler godoc
// @Summary      Modifier une réaction sur un post
// @Description  Change le type de réaction (like/love/laugh) de l'utilisateur connecté sur un post
// @Tags         Post Likes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        postID  path      string       true  "UUID du post"
// @Param        body    body      LikeRequest  true  "Nouvelle réaction"
// @Success      200     {object}  Like
// @Failure      400     {string}  string  "ID ou type de réaction invalide"
// @Failure      404     {string}  string  "Like non trouvé"
// @Failure      500     {string}  string  "Erreur interne"
// @Router       /posts/{postID}/likes [patch]
func (h *LikeHandler) UpdateReactionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := r.Context().Value(ctxkeys.UserContextKey).(ctxkeys.UserContext)
	postID, err := uuid.Parse(chi.URLParam(r, "postID"))
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	var req LikeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "format json invalide", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	likeUpdated, err := h.likeService.UpdateReaction(ctx, req.ReactionType, userCtx.UserID, postID)
	if err != nil {
		if errors.Is(err, ErrReactionNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if errors.Is(err, ErrInvalidReactionType) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, "erreur interne du server", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(likeUpdated)

}

// DeleteHandler godoc
// @Summary      Supprimer un like sur un post
// @Description  Retire la réaction de l'utilisateur connecté sur un post
// @Tags         Post Likes
// @Security     BearerAuth
// @Param        postID  path      string  true  "UUID du post"
// @Success      204     "Like supprimé"
// @Failure      400     {string}  string  "ID invalide"
// @Failure      404     {string}  string  "Like non trouvé"
// @Failure      500     {string}  string  "Erreur interne"
// @Router       /posts/{postID}/likes [delete]
func (h *LikeHandler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := r.Context().Value(ctxkeys.UserContextKey).(ctxkeys.UserContext)
	postID, err := uuid.Parse(chi.URLParam(r, "postID"))
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	if err := h.likeService.Delete(ctx, postID, userCtx.UserID); err != nil {

		if errors.Is(err, ErrUpdateCountFailed) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if errors.Is(err, ErrReactionNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, "erreur interne du server", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

// CreateHandler godoc
// @Summary      Liker un post
// @Description  Ajoute une réaction (like/love/laugh) à un post
// @Tags         Post Likes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        postID  path      string       true  "UUID du post"
// @Param        body    body      LikeRequest  true  "Type de réaction"
// @Success      201     {object}  Like
// @Failure      400     {string}  string  "Type de réaction invalide"
// @Failure      404     {string}  string  "Post non trouvé"
// @Failure      409     {string}  string  "Déjà liké"
// @Failure      500     {string}  string  "Erreur interne"
// @Router       /posts/{postID}/likes [post]
func (h *LikeHandler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := r.Context().Value(ctxkeys.UserContextKey).(ctxkeys.UserContext)
	postID, err := uuid.Parse(chi.URLParam(r, "postID"))
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	var req LikeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "format json invalide", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	l := Like{
		PostID:       postID,
		UserID:       userCtx.UserID,
		ReactionType: req.ReactionType,
	}

	likeCreated, err := h.likeService.CreateLike(ctx, l)
	if err != nil {
		if errors.Is(err, ErrCreateReactionFailed) || errors.Is(err, ErrUpdateCountFailed) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if errors.Is(err, ErrInvalidReactionType) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if errors.Is(err, ErrAlreadyLiked) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		http.Error(w, "erreur interne du server", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(likeCreated)

}

func (h *LikeHandler) LikeRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.GetAllHandler)
	r.Get("/me", h.GetByUserAndPostHandler)
	r.Patch("/", h.UpdateReactionHandler)
	r.Delete("/", h.DeleteHandler)
	r.Post("/", h.CreateHandler)

	return r
}
