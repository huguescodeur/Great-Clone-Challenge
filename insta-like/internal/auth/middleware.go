package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/huguescodeur/insta-like/internal/pkg/ctxkeys"
)

func AuthMidlleware(next http.Handler) http.Handler {
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
		secret := []byte("mon_secret")

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Methode de signature inattendue: %v", t.Header["alg"])
			}

			return secret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Session invalide ou expirée", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Impossible de lire les donne2es du token", http.StatusUnauthorized)
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
