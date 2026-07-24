package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/huguescodeur/insta-lite/internal/pkg/utils"
	"github.com/huguescodeur/insta-lite/internal/user"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	authStore   AuthStore
	blacklister *RedisBlacklist
	secret      string
}

func NewServiceAuth(a AuthStore, b *RedisBlacklist) *AuthService {
	return &AuthService{
		authStore:   a,
		blacklister: b,
		secret:      "mon_secret",
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

func (s *AuthService) Logout(ctx context.Context, tokenString string) error {
	token, err := jwt.Parse(tokenString, hmacKeyFunc([]byte(s.secret)))
	if err != nil {
		return errors.New("token invalide")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("impossible de lire les claims")
	}

	expRaw, okExp := claims["exp"]
	if !okExp {
		return errors.New("le token n'a pas de date d'expiration")
	}

	expFloat, okType := expRaw.(float64)
	if !okType {
		return errors.New("format exp invalide")
	}

	expirationTime := time.Unix(int64(expFloat), 0)
	timeLeft := time.Until(expirationTime)

	if timeLeft <= 0 {
		return nil
	}

	return s.blacklister.Blacklist(ctx, tokenString, timeLeft)
}
