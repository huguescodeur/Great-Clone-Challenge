package commentlike

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

type CommentLikeHandler struct {
	commentLikeService *CommentLikeService
	validate           *validator.Validate
}

func NewCommentLikeHandler(s *CommentLikeService) *CommentLikeHandler {
	return &CommentLikeHandler{commentLikeService: s, validate: validator.New()}
}

// GetAllHandler godoc
// @Summary      Likes d'un commentaire
// @Description  Retourne tous les likes d'un commentaire
// @Tags         Comment Likes
// @Produce      json
// @Security     BearerAuth
// @Param        commentID  path      string  true  "UUID du commentaire"
// @Success      200        {array}   CommentLike
// @Failure      400        {string}  string  "ID incorrect"
// @Failure      500        {string}  string  "Erreur interne"
// @Router       /comments/{commentID}/likes [get]
func (h *CommentLikeHandler) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	commentID, err := uuid.Parse(chi.URLParam(r, "commentID"))
	if err != nil {
		http.Error(w, "incorrect comment ID", http.StatusBadRequest)
		return
	}

	likes, err := h.commentLikeService.GetAllLikes(ctx, commentID)
	if err != nil {
		http.Error(w, "erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(likes)
}

// GetByUserAndCommentHandler godoc
// @Summary      Mon like sur un commentaire
// @Description  Retourne le like de l'utilisateur connecté sur un commentaire
// @Tags         Comment Likes
// @Produce      json
// @Security     BearerAuth
// @Param        commentID  path      string  true  "UUID du commentaire"
// @Success      200        {object}  CommentLike
// @Failure      400        {string}  string  "ID incorrect"
// @Failure      404        {string}  string  "Like non trouvé"
// @Failure      500        {string}  string  "Erreur interne"
// @Router       /comments/{commentID}/likes/me [get]
func (h *CommentLikeHandler) GetByUserAndCommentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := r.Context().Value(ctxkeys.UserContextKey).(ctxkeys.UserContext)

	commentID, err := uuid.Parse(chi.URLParam(r, "commentID"))
	if err != nil {
		http.Error(w, "incorrect comment ID", http.StatusBadRequest)
		return
	}

	l, err := h.commentLikeService.GetLikeByUserAndComment(ctx, userCtx.UserID, commentID)
	if err != nil {
		if errors.Is(err, ErrReactionNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(l)
}

// CreateHandler godoc
// @Summary      Liker un commentaire
// @Description  Ajoute une réaction (like/love/laugh) à un commentaire
// @Tags         Comment Likes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        commentID  path      string           true  "UUID du commentaire"
// @Param        body       body      reactionRequest  true  "Type de réaction"
// @Success      201        {object}  CommentLike
// @Failure      400        {string}  string  "Type de réaction invalide"
// @Failure      409        {string}  string  "Déjà liké"
// @Failure      500        {string}  string  "Erreur interne"
// @Router       /comments/{commentID}/likes [post]
func (h *CommentLikeHandler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := r.Context().Value(ctxkeys.UserContextKey).(ctxkeys.UserContext)

	commentID, err := uuid.Parse(chi.URLParam(r, "commentID"))
	if err != nil {
		http.Error(w, "incorrect comment ID", http.StatusBadRequest)
		return
	}

	var req reactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "format json invalide", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	commentLikeData := CommentLike{
		CommentID:    commentID,
		UserID:       userCtx.UserID,
		ReactionType: req.ReactionType,
	}

	created, err := h.commentLikeService.CreateLike(ctx, commentLikeData)
	if err != nil {
		if errors.Is(err, ErrAlreadyLiked) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if errors.Is(err, ErrInvalidReactionType) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// UpdateReactionHandler godoc
// @Summary      Modifier une réaction sur un commentaire
// @Description  Change le type de réaction de l'utilisateur connecté sur un commentaire
// @Tags         Comment Likes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        commentID  path      string           true  "UUID du commentaire"
// @Param        body       body      reactionRequest  true  "Nouvelle réaction"
// @Success      200        {object}  CommentLike
// @Failure      400        {string}  string  "Type invalide"
// @Failure      404        {string}  string  "Like non trouvé"
// @Failure      500        {string}  string  "Erreur interne"
// @Router       /comments/{commentID}/likes [patch]
func (h *CommentLikeHandler) UpdateReactionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := r.Context().Value(ctxkeys.UserContextKey).(ctxkeys.UserContext)

	commentID, err := uuid.Parse(chi.URLParam(r, "commentID"))
	if err != nil {
		http.Error(w, "incorrect comment ID", http.StatusBadRequest)
		return
	}

	var req reactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "format json invalide", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	updated, err := h.commentLikeService.UpdateReaction(ctx, req.ReactionType, userCtx.UserID, commentID)
	if err != nil {
		if errors.Is(err, ErrReactionNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, ErrInvalidReactionType) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

// DeleteHandler godoc
// @Summary      Supprimer un like sur un commentaire
// @Description  Retire la réaction de l'utilisateur connecté sur un commentaire
// @Tags         Comment Likes
// @Security     BearerAuth
// @Param        commentID  path      string  true  "UUID du commentaire"
// @Success      204        "Like supprimé"
// @Failure      400        {string}  string  "ID incorrect"
// @Failure      404        {string}  string  "Like non trouvé"
// @Failure      500        {string}  string  "Erreur interne"
// @Router       /comments/{commentID}/likes [delete]
func (h *CommentLikeHandler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := r.Context().Value(ctxkeys.UserContextKey).(ctxkeys.UserContext)

	commentID, err := uuid.Parse(chi.URLParam(r, "commentID"))
	if err != nil {
		http.Error(w, "incorrect comment ID", http.StatusBadRequest)
		return
	}

	if err := h.commentLikeService.Delete(ctx, commentID, userCtx.UserID); err != nil {
		if errors.Is(err, ErrReactionNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CommentLikeHandler) CommentLikeRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.GetAllHandler)
	r.Get("/me", h.GetByUserAndCommentHandler)
	r.Patch("/", h.UpdateReactionHandler)
	r.Delete("/", h.DeleteHandler)
	r.Post("/", h.CreateHandler)

	return r
}
