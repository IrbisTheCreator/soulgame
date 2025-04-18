package main

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	_ "golang.org/x/crypto/bcrypt"
)

func homeHandler(w http.ResponseWriter, r *http.Request) (string, int, int, error) {
	// Получаем токен из куки
	c, err := r.Cookie("token")
	if err != nil {
		return "", 0, 0, err
	}

	tokenStr := c.Value
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return signingKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return "", 0, 0, err
		}
		return "", 0, 0, err
	}

	if !token.Valid {
		return "", 0, 0, err
	}

	return claims.Phone, claims.User_id, claims.Dostup, nil
}
