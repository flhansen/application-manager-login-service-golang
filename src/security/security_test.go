package security

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSalt(t *testing.T) {
	randomGenerator := RandomGenerator{Reader: rand.Reader}
	salt, err := randomGenerator.GenerateSalt(32)

	if err != nil {
		t.Fatalf("Error while generating salt: %v\n", err)
	}

	assert.Equal(t, 32, len(salt))
}

func TestGenerateSaltInvalidReader(t *testing.T) {
	randomGenerator := RandomGenerator{Reader: iotest.ErrReader(errors.New("Invalid reader"))}
	_, err := randomGenerator.GenerateSalt(32)
	assert.NotNil(t, err)
}

func TestCreatePasswordHash(t *testing.T) {
	salt := []byte("thesaltthesaltth")
	hash := CreatePasswordHash("password", salt)
	hashString := base64.StdEncoding.EncodeToString(hash)

	assert.Equal(t, "dGhlc2FsdHRoZXNhbHR0aCH+CA0ZP72npZ/NA9AFhzcYzPW3V5jsDyc+23SG0Ugc", hashString)
}

func TestValidatePassword(t *testing.T) {
	saltHashBase64 := "dGhlc2FsdHRoZXNhbHR0aCH+CA0ZP72npZ/NA9AFhzcYzPW3V5jsDyc+23SG0Ugc"
	result := ValidatePassword("password", saltHashBase64)
	assert.True(t, result)
}
