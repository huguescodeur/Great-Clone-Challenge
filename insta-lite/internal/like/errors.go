package like

import "errors"

var (
	ErrReactionNotFound     = errors.New("impossible de modifier la reaction")
	ErrInvalidReactionType  = errors.New("type de réaction invalide")
	ErrUpdateCountFailed    = errors.New("update like count echoué")
	ErrCreateReactionFailed = errors.New("impossible de creer la reaction")
	ErrAlreadyLiked         = errors.New("vous avez déjà liké ce post")
	ErrLikeNotFound         = errors.New("like not found")
)
