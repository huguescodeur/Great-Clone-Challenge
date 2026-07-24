package comment

import "errors"

var (
	ErrCommentNotFound       = errors.New("commentaire introuvable")
	ErrParentCommentNotFound = errors.New("commentaire parent introuvable")
	ErrCommentPostMismatch   = errors.New("le commentaire parent n'appartient pas à ce post")
	ErrPostNotFound          = errors.New("post introuvable")
	ErrUpdateCountFailed     = errors.New("échec de la mise à jour du compteur de commentaires")
	ErrCreateCommentFailed   = errors.New("échec de la création du commentaire")
	ErrUnauthorized          = errors.New("vous n'avez pas l'autorisation d'effectuer cette action")
)
