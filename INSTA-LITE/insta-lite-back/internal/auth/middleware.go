package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/huguescodeur/insta-lite/internal/pkg/ctxkeys"
)

type TokenBlacklister interface {
	IsBlacklisted(ctx context.Context, token string) (bool, error)
}

type AuthMiddleware struct {
	blacklister TokenBlacklister
	secret      []byte
}

func NewAuthMiddleware(b TokenBlacklister, secret string) *AuthMiddleware {
	return &AuthMiddleware{
		blacklister: b,
		secret:      []byte(secret),
	}
}

func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Authentification requise (Format: Bearer <token>)", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		if m.blacklister != nil {
			isBanned, err := m.blacklister.IsBlacklisted(r.Context(), tokenString)
			if err != nil {
				log.Printf("[auth] vérification blacklist échouée, fail-open: %v", err)
			} else if isBanned {
				http.Error(w, "Session révoquée (déconnectée)", http.StatusUnauthorized)
				return
			}
		}

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Methode de signature inattendue: %v", t.Header["alg"])
			}
			return m.secret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Session invalide ou expirée", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Impossible de lire les données du token", http.StatusUnauthorized)
			return
		}

		uidRaw, okUID := claims["userID"]
		if !okUID || uidRaw == nil {
			http.Error(w, "Données utilisateur manquantes (userID)", http.StatusUnauthorized)
			return
		}

		userIDStr, okTypeUID := uidRaw.(string)
		if !okTypeUID {
			http.Error(w, "Format userUUID invalide (doit être une string)", http.StatusUnauthorized)
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			http.Error(w, "UUID invalide", http.StatusUnauthorized)
			return
		}

		userCtx := ctxkeys.UserContext{
			UserID: userID,
		}

		ctx := context.WithValue(
			r.Context(),
			ctxkeys.UserContextKey,
			userCtx,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
