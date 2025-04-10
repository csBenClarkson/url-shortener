package router

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const jwtExpireTime = 24 * time.Hour

func generateAccessToken(secret string) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "url-shortener",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtExpireTime)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func parseAccessToken(tokenStr string, secret string) (bool, error) {
	var claims jwt.RegisteredClaims
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return false, err
	}
	return token.Valid, nil
}
