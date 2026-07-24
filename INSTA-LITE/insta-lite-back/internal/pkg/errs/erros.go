package errs

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	ErrNotFound       = errors.New("ressource non trouvée")
	ErrAlreadyExists  = errors.New("la ressource existe déjà")
	ErrInvalidInput   = errors.New("données invalides")
	ErrInternalServer = errors.New("erreur interne du serveur")
	ErrUnauthorized   = errors.New("Vous n'êtes pas autorisé")
)

// NewBadRequest crée une erreur 400 avec un message lisible côté client.
func NewBadRequest(msg string) error {
	return &businessError{msg: msg}
}

type businessError struct{ msg string }

func (e *businessError) Error() string { return e.msg }
func (e *businessError) Unwrap() error { return ErrInvalidInput }

// SafeMessage retourne le message d'erreur pour le client.
// Pour les erreurs 5xx, on ne divulgue pas les détails internes.
func SafeMessage(err error, status int) string {
	if status >= 500 {
		return "Une erreur interne s'est produite"
	}
	return err.Error()
}

func MapHTTPError(err error) int {
	if errors.Is(err, pgx.ErrNoRows) {
		return http.StatusNotFound
	}
	if errors.Is(err, ErrNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, ErrAlreadyExists) {
		return http.StatusConflict
	}
	if errors.Is(err, ErrInvalidInput) {
		return http.StatusBadRequest
	}
	if errors.Is(err, ErrUnauthorized) {
		return http.StatusUnauthorized
	}
	return http.StatusInternalServerError
}
