package jwt

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type jwtToken struct {
	secret []byte
	ttl    time.Duration
}

type Claims struct {
	jwt.RegisteredClaims
}

type TokenManager interface {
	GenerateToken() (string, error)
	ValidateToken(accessToken string) (*Claims, error)
}

func NewJwtManager(secret []byte) TokenManager {
	return &jwtToken{secret: secret, ttl: 15 * time.Minute}
}

func (t *jwtToken) GenerateToken() (string, error) {
	claims := &Claims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "token",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(t.secret)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func (t *jwtToken) ValidateToken(accessToken string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(accessToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return t.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)

	if !ok {
		return nil, errors.New("invalid claims")
	}

	return claims, nil
}
