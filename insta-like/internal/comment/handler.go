package comment

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/huguescodeur/insta-like/internal/pkg/ctxkeys"
	"github.com/huguescodeur/insta-like/internal/pkg/responses"
)

type CommentHandler struct {
	commentService *CommentService
	validate       *validator.Validate
}

func NewCommentHandler(cs *CommentService) *CommentHandler {
	return &CommentHandler{commentService: cs, validate: validator.New()}
}

// GetAllByPostIDHandler godoc
// @Summary      Commentaires d'un post
// @Description  Retourne les commentaires racines d'un post, paginés par curseur
// @Tags         Comments
// @Produce      json
// @Security     BearerAuth
// @Param        postID  path      string  true   "UUID du post"
// @Param        cursor  query     string  false  "Curseur de pagination"
// @Param        limit   query     int     false  "Nombre de résultats (défaut: 10, max: 100)"
// @Success      200     {object}  PaginatedCommentResponse
// @Failure      400     {string}  string  "ID incorrect"
// @Failure      500     {string}  string  "Erreur interne"
// @Router       /posts/{postID}/comments [get]
func (h *CommentHandler) GetAllByPostIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	postIDStr := chi.URLParam(r, "postID")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		http.Error(w, "incorrect post ID", http.StatusBadRequest)
		return
	}

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

	res, err := h.commentService.GetAllByPostID(ctx, postID, cursorStr, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

// GetRepliesHandler godoc
// @Summary      Réponses à un commentaire
// @Description  Retourne toutes les réponses d'un commentaire parent
// @Tags         Comments
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      string  true  "UUID du commentaire parent"
// @Success      200 {array}   Comment
// @Failure      400 {string}  string  "ID incorrect"
// @Failure      500 {string}  string  "Erreur interne"
// @Router       /comments/{id}/replies [get]
func (h *CommentHandler) GetRepliesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	parentIDStr := chi.URLParam(r, "id")
	parentID, err := uuid.Parse(parentIDStr)
	if err != nil {
		http.Error(w, "incorrect comment ID", http.StatusBadRequest)
		return
	}

	replies, err := h.commentService.GetReplies(ctx, parentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(replies)
}

// CreateHandler godoc
// @Summary      Créer un commentaire
// @Description  Ajoute un commentaire (ou une réponse) sur un post. Passer parentCommentId pour une réponse.
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        postID  path      string                true  "UUID du post"
// @Param        body    body      CommentCreateRequest  true  "Contenu du commentaire"
// @Success      201     {object}  Comment
// @Failure      400     {string}  string  "Format invalide ou commentaire parent ne correspond pas au post"
// @Failure      404     {string}  string  "Post ou commentaire parent non trouvé"
// @Failure      500     {string}  string  "Erreur interne"
// @Router       /posts/{postID}/comments [post]
func (h *CommentHandler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := r.Context().Value(ctxkeys.UserContextKey).(ctxkeys.UserContext)

	postID, err := uuid.Parse(chi.URLParam(r, "postID"))
	if err != nil {
		http.Error(w, "incorrect post ID", http.StatusBadRequest)
		return
	}

	var req CommentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "format json invalide", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	commentData := Comment{
		PostID:          postID,
		UserID:          userCtx.UserID,
		ParentCommentID: req.ParentCommentID,
		Content:         req.Content,
	}

	res, err := h.commentService.CreateComment(ctx, commentData)
	if err != nil {
		if errors.Is(err, ErrParentCommentNotFound) || errors.Is(err, ErrPostNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, ErrCommentPostMismatch) {
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

// UpdateHandler godoc
// @Summary      Modifier un commentaire
// @Description  Met à jour le contenu d'un commentaire appartenant à l'utilisateur connecté
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                true  "UUID du commentaire"
// @Param        body  body      CommentUpdateRequest  true  "Nouveau contenu"
// @Success      200   {object}  Comment
// @Failure      400   {string}  string  "Format invalide"
// @Failure      404   {string}  string  "Commentaire non trouvé"
// @Failure      500   {string}  string  "Erreur interne"
// @Router       /comments/{id} [patch]
func (h *CommentHandler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := r.Context().Value(ctxkeys.UserContextKey).(ctxkeys.UserContext)

	commentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "incorrect comment ID", http.StatusBadRequest)
		return
	}

	var req CommentUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "format json invalide", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	c, err := h.commentService.UpdateComment(ctx, req.Content, userCtx.UserID, commentID)
	if err != nil {
		if errors.Is(err, ErrCommentNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, "erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(c)
}

// DeleteHandler godoc
// @Summary      Supprimer un commentaire
// @Description  Supprime un commentaire appartenant à l'utilisateur connecté
// @Tags         Comments
// @Security     BearerAuth
// @Param        id  path      string  true  "UUID du commentaire"
// @Success      204 "Commentaire supprimé"
// @Failure      400 {string}  string  "ID incorrect"
// @Failure      404 {string}  string  "Commentaire non trouvé"
// @Failure      500 {string}  string  "Erreur interne"
// @Router       /comments/{id} [delete]
func (h *CommentHandler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := r.Context().Value(ctxkeys.UserContextKey).(ctxkeys.UserContext)

	commentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "incorrect comment ID", http.StatusBadRequest)
		return
	}

	if err := h.commentService.DeleteComment(ctx, userCtx.UserID, commentID); err != nil {
		if errors.Is(err, ErrCommentNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, "erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CommentHandler) CommentRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.GetAllByPostIDHandler)
	r.Post("/", h.CreateHandler)

	return r
}
