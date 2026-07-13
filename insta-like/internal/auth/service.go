package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/huguescodeur/insta-like/internal/pkg/utils"
	"github.com/huguescodeur/insta-like/internal/user"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	authStore AuthStore
	secret    string
}

func NewServiceAuth(a AuthStore) *AuthService {
	return &AuthService{
		authStore: a,
		secret:    "mon_secret",
	}
}

func (s *AuthService) GenerateJWT(u *user.User) (string, error) {
	secret := []byte(s.secret)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": u.UserID.String(),
		"exp":    time.Now().Add(30 * 24 * time.Hour).Unix(),
	})

	return token.SignedString(secret)
}

func (s *AuthService) Register(ctx context.Context, u *user.User) (*user.User, string, error) {
	cleanUsername, ok := utils.SanitizeUsername(u.Username)
	if !ok {
		return nil, "", errors.New("format de username invalide (lettres, chiffres, underscore, doit commencer par une lettre)")
	}
	u.Username = cleanUsername

	u.UserID = uuid.New()

	hashed, err := bcrypt.GenerateFromPassword([]byte(u.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	u.PasswordHash = string(hashed)

	createdUser, err := s.authStore.Create(ctx, u)
	if err != nil {
		return nil, "", err
	}

	token, err := s.GenerateJWT(createdUser)
	if err != nil {
		return nil, "", err
	}

	return createdUser, token, nil
}

func (s *AuthService) Login(ctx context.Context, identifier, password string) (*user.User, string, error) {
	u, err := s.authStore.GetByEmailOrUsername(ctx, identifier)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "",
				errors.New("identifiants incorrect")
		}
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, "", fmt.Errorf("identifiants incorrect")
	}

	token, err := s.GenerateJWT(u)
	if err != nil {
		return nil, "", err
	}

	return u, token, nil
}
