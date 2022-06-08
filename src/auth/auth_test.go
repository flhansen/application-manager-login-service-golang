package auth

import (
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	tokenString, err := GenerateToken(0, "test", jwt.SigningMethodHS256, []byte("supersecretsignkey"))
	if err != nil {
		t.Fatal(err)
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte("supersecretsignkey"), nil
	})

	claims := token.Claims.(jwt.MapClaims)

	assert.Nil(t, err)
	assert.NotNil(t, claims["userId"])
	assert.NotNil(t, claims["username"])
	assert.Equal(t, 0.0, claims["userId"])
	assert.Equal(t, "test", claims["username"])
}
