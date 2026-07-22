package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func hmacKeyFunc(secret []byte) jwt.Keyfunc {
	return func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("méthode de signature inattendue: %v", t.Header["alg"])
		}
		return secret, nil
	}
}
