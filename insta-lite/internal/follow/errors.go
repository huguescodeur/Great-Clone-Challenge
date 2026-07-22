package follow

import "errors"

var (
	ErrFollowNotFound     = errors.New("relation de suivi introuvable")
	ErrAlreadyFollowing   = errors.New("vous suivez déjà cet utilisateur")
	ErrCannotFollowSelf   = errors.New("vous ne pouvez pas vous suivre vous-même")
	ErrInvalidStatus      = errors.New("statut invalide")
	ErrCreateFollowFailed = errors.New("échec de la création de la relation de suivi")
)
