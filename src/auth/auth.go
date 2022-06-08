package auth

import (
	"time"

	"github.com/golang-jwt/jwt"
)

type JwtClaims struct {
	UserId   int    `json:"userId"`
	Username string `json:"username"`
	jwt.StandardClaims
}

func GenerateToken(id int, username string, signingMethod jwt.SigningMethod, key interface{}) (string, error) {
	claims := JwtClaims{
		UserId:   id,
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(5 * time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	signedToken, err := token.SignedString(key)
	return signedToken, err
}
