package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestValidateToken(t *testing.T) {
	secret := []byte("test_secret")
	manager := NewJwtManager(secret).(*jwtToken)

	t.Run("valid token", func(t *testing.T) {
		token, err := manager.GenerateToken()
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := manager.ValidateToken(token)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, "token", claims.Subject)
	})

	t.Run("invalid signing method", func(t *testing.T) {
		// создаём токен с неправильным алгоритмом (например, none)
		token := jwt.NewWithClaims(jwt.SigningMethodNone, &Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			},
		})
		tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		assert.NoError(t, err)

		_, err = manager.ValidateToken(tokenStr)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected signing method")
	})

	t.Run("expired token", func(t *testing.T) {
		claims := &Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // уже истёк
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			},
		}
		expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, err := expiredToken.SignedString(secret)
		assert.NoError(t, err)

		_, err = manager.ValidateToken(tokenStr)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token is expired")
	})

	t.Run("malformed token", func(t *testing.T) {
		_, err := manager.ValidateToken("not.a.valid.token")
		assert.Error(t, err)
	})
}
