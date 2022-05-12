package security

import (
	"crypto/sha256"
	"encoding/base64"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

type RandomGenerator struct {
	Reader io.Reader
}

func (rng RandomGenerator) GenerateSalt(length int) ([]byte, error) {
	salt := make([]byte, length)

	if _, err := rng.Reader.Read(salt); err != nil {
		return nil, err
	}

	return salt, nil
}

func CreatePasswordHash(password string, salt []byte) []byte {
	passwordHash := pbkdf2.Key([]byte(password), salt, 4096, 32, sha256.New)
	return append(salt, passwordHash...)
}

func ValidatePassword(input string, passwordHashBase64 string) bool {
	decodedPasswordHash, _ := base64.StdEncoding.DecodeString(passwordHashBase64)
	salt := make([]byte, 16)
	copy(salt, decodedPasswordHash[:16])

	hashedInput := CreatePasswordHash(input, salt)
	areEqual := len(decodedPasswordHash) == len(hashedInput)

	for i := 0; i < len(hashedInput) && areEqual; i++ {
		areEqual = hashedInput[i] == decodedPasswordHash[i]
	}

	return areEqual
}
