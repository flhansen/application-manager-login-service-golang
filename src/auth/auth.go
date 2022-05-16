package auth

import (
	"flhansen/application-manager/login-service/src/database"
	"time"

	"github.com/golang-jwt/jwt"
)

type JwtClaims struct {
	UserId   int    `json:"userId"`
	Username string `json:"username"`
	jwt.StandardClaims
}

type JwtConfig struct {
	SignKey string `yaml:"signkey"`
}

func GenerateToken(acc database.Account, signingMethod jwt.SigningMethod, key interface{}) (string, error) {
	claims := JwtClaims{
		UserId:   acc.Id,
		Username: acc.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(5 * time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	signedToken, err := token.SignedString(key)
	return signedToken, err
}
