package commentlike

import "errors"

var (
	ErrCommentLikeNotFound  = errors.New("réaction introuvable")
	ErrReactionNotFound     = errors.New("réaction introuvable")
	ErrInvalidReactionType  = errors.New("type de réaction invalide")
	ErrAlreadyLiked         = errors.New("vous avez déjà réagi à ce commentaire")
	ErrCreateReactionFailed = errors.New("échec de la création de la réaction")
	ErrUpdateCountFailed    = errors.New("échec de la mise à jour du compteur de likes")
)
