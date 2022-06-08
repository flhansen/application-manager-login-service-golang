package service

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"flhansen/application-manager/login-service/src/auth"
	"flhansen/application-manager/login-service/src/controller"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
)

var loginService *LoginService

func TestMain(m *testing.M) {
	os.Exit(runAllTests(m))
}

func runAllTests(m *testing.M) int {
	loginService = New(
		ServiceConfig{
			Host: "localhost",
			Port: 8080,
			Jwt: JwtConfig{
				SignKey: []byte("supersecretsigningkey"),
			},
			Database: controller.DbConfig{
				Host:     "localhost",
				Port:     5432,
				Username: "test",
				Password: "test",
				Database: "test",
			}})

	// Make sure the user 'testuser' does not exist and then create it
	loginService.Database.CreateSchema()
	_, err := loginService.Database.InsertAccount("testuser", "testpass", "testuser@test.com", time.Now())

	if err != nil {
		fmt.Printf("Could not insert test user: %v\n", err)
		return 1
	}

	defer func() {
		if err := loginService.Database.DeleteAccountByUsername("testuser"); err != nil {
			fmt.Printf("Error while deleting account: %v", err)
		}
	}()

	// Start the login service
	go func() {
		if err := loginService.Start(); err != nil {
			fmt.Printf("Could not start service: %v\n", err)
			os.Exit(1)
		}
	}()

	// Wait a second for the login service to be online
	time.Sleep(1 * time.Second)

	// Start all of the tests
	return m.Run()
}

func TestLoginSuccess(t *testing.T) {
	loginRequestBody := LoginRequest{
		Username: "testuser",
		Password: "testpass",
	}

	body, err := json.Marshal(loginRequestBody)

	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Post("http://localhost:8080/api/auth/login", "application/json", bytes.NewBuffer(body))

	if err != nil {
		t.Fatal(err)
	}

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	tokenString := fmt.Sprintf("%v", res["token"])
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte("supersecretsigningkey"), nil
	})

	if err != nil {
		t.Fatal(err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("Invalid token")
	}

	assert.Equal(t, 200, resp.StatusCode)
	assert.NotNil(t, res["status"])
	assert.NotNil(t, res["message"])
	assert.NotNil(t, res["token"])
	assert.Equal(t, loginRequestBody.Username, claims["username"])
	assert.NotNil(t, claims["userId"])
}

func TestLoginWrongPassword(t *testing.T) {
	loginRequestBody := LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	body, err := json.Marshal(loginRequestBody)

	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Post("http://localhost:8080/api/auth/login", "application/json", bytes.NewBuffer(body))

	if err != nil {
		t.Fatal(err)
	}

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	assert.Equal(t, 401, resp.StatusCode)
	assert.NotNil(t, res["status"])
	assert.NotNil(t, res["message"])
}

func TestLoginWrongUsername(t *testing.T) {
	loginService.Database.DeleteAccountByUsername("testuser")
	defer loginService.Database.InsertAccount("testuser", "testpass", "testuser@test.com", time.Now())

	loginRequestBody := LoginRequest{
		Username: "testuser",
		Password: "testpass",
	}

	body, err := json.Marshal(loginRequestBody)

	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Post("http://localhost:8080/api/auth/login", "application/json", bytes.NewBuffer(body))

	if err != nil {
		t.Fatal(err)
	}

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.NotNil(t, res["status"])
	assert.NotNil(t, res["message"])
}

func TestLoginInvalidJsonRequest(t *testing.T) {
	invalidJsonString := `{ "username": "testuser", "password": "testpass`
	resp, err := http.Post("http://localhost:8080/api/auth/login", "application/json", bytes.NewBuffer([]byte(invalidJsonString)))

	if err != nil {
		t.Fatal(err)
	}

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.NotNil(t, res["status"])
	assert.NotNil(t, res["message"])
}

func TestLoginTokenSignError(t *testing.T) {
	loginRequestBody := LoginRequest{
		Username: "testuser",
		Password: "testpass",
	}

	body, err := json.Marshal(loginRequestBody)

	if err != nil {
		t.Fatal(err)
	}

	oldSignKey := loginService.JwtSignKey
	loginService.JwtSignKey = ""
	defer func() {
		loginService.JwtSignKey = oldSignKey
	}()

	resp, err := http.Post("http://localhost:8080/api/auth/login", "application/json", bytes.NewBuffer(body))

	if err != nil {
		t.Fatal(err)
	}

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.NotNil(t, res["status"])
	assert.NotNil(t, res["message"])
}

func TestRegisterUsernameAlreadyExists(t *testing.T) {
	registerRequestBody := RegisterRequest{
		Username: "testuser",
		Password: "testpass",
		Email:    "testuserisnotatestuser@test.com",
	}

	body, err := json.Marshal(registerRequestBody)

	if err != nil {
		t.Fatal(err)
	}

	// At this point the user 'testuser' already exists (see TestMain)
	resp, err := http.Post("http://localhost:8080/api/auth/register", "application/json", bytes.NewBuffer(body))

	if err != nil {
		t.Fatal(err)
	}

	var res map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&res)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.NotNil(t, res["status"])
	assert.NotNil(t, res["message"])
}

func TestRegisterEmailAlreadyExists(t *testing.T) {
	registerRequestBody := RegisterRequest{
		Username: "testuserisnotatestuser",
		Password: "testpass",
		Email:    "testuser@test.com",
	}

	body, err := json.Marshal(registerRequestBody)

	if err != nil {
		t.Fatal(err)
	}

	// At this point the user 'testuser' already exists (see TestMain)
	resp, err := http.Post("http://localhost:8080/api/auth/register", "application/json", bytes.NewBuffer(body))

	if err != nil {
		t.Fatal(err)
	}

	var res map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&res)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.NotNil(t, res["status"])
	assert.NotNil(t, res["message"])
}

func TestRegisterSuccess(t *testing.T) {
	loginService.Database.DeleteAccountByUsername("testuser")
	defer loginService.Database.InsertAccount("testuser", "testpass", "testuser@test.com", time.Now())

	registerRequestBody := RegisterRequest{
		Username: "testuser",
		Password: "testpass",
		Email:    "testuser@test.com",
	}

	body, err := json.Marshal(registerRequestBody)

	if err != nil {
		t.Fatal(err)
	}

	// At this point the user 'testuser' already exists (see TestMain)
	resp, err := http.Post("http://localhost:8080/api/auth/register", "application/json", bytes.NewBuffer(body))

	if err != nil {
		t.Fatal(err)
	}

	var res map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&res)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotNil(t, res["status"])
	assert.NotNil(t, res["message"])
}

func TestRegisterInvalidJsonRequest(t *testing.T) {
	registerRequestBody := `{ "username": "testuser", "password": "testpass", "email": "`

	resp, err := http.Post("http://localhost:8080/api/auth/register", "application/json", bytes.NewBuffer([]byte(registerRequestBody)))

	if err != nil {
		t.Fatal(err)
	}

	var res map[string]interface{}
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.NotNil(t, res["status"])
	assert.NotNil(t, res["message"])
}

func TestDelete(t *testing.T) {
	loginService.Database.CreateSchema()
	loginService.Database.InsertAccount("test", "test", "testuser@test.com", time.Now())
	acc, _ := loginService.Database.GetAccountByUsername("test")
	token, _ := auth.GenerateToken(acc.Id, acc.Username, jwt.SigningMethodHS256, []byte("supersecretsigningkey"))

	client := &http.Client{}

	req, err := http.NewRequest(http.MethodDelete, "http://localhost:8080/api/auth/delete", nil)

	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Authorization", token)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	var res map[string]interface{}
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotNil(t, res["status"])
	assert.NotNil(t, res["message"])
}

func TestDeleteWrongSigningMethod(t *testing.T) {
	loginService.Database.CreateSchema()
	loginService.Database.InsertAccount("test", "test", "testuser@test.com", time.Now())
	acc, _ := loginService.Database.GetAccountByUsername("test")

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	token, _ := auth.GenerateToken(acc.Id, acc.Username, jwt.SigningMethodRS256, privateKey)

	client := &http.Client{}

	req, err := http.NewRequest(http.MethodDelete, "http://localhost:8080/api/auth/delete", nil)

	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Authorization", token)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	var res map[string]interface{}
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.NotNil(t, res["status"])
	assert.NotNil(t, res["message"])
}

func TestDeleteInvalidClaims(t *testing.T) {
	loginService.Database.CreateSchema()
	loginService.Database.InsertAccount("test", "test", "testuser@test.com", time.Now())

	header := map[string]interface{}{
		"alg": jwt.SigningMethodHS256.Alg(),
		"typ": "JWT",
	}
	headerBytes, _ := json.Marshal(header)

	claimsBytes := []byte(`{ "userId":`)
	preTokenString := base64.RawURLEncoding.EncodeToString(headerBytes) + "." + base64.RawURLEncoding.EncodeToString(claimsBytes)
	signature, _ := jwt.SigningMethodHS256.Sign(preTokenString, []byte("supersecretsigningkey"))
	tokenString := strings.Join([]string{preTokenString, signature}, ".")

	client := &http.Client{}

	req, err := http.NewRequest(http.MethodDelete, "http://localhost:8080/api/auth/delete", nil)

	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Authorization", tokenString)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	var res map[string]interface{}
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.NotNil(t, res["status"])
	assert.NotNil(t, res["message"])
}

func TestDeleteInvalidQuery(t *testing.T) {
	loginService.Database.CreateSchema()
	loginService.Database.InsertAccount("test", "test", "testuser@test.com", time.Now())
	acc, _ := loginService.Database.GetAccountByUsername("test")
	oldPort := loginService.Database.Port
	loginService.Database.Port = 0

	defer func() {
		loginService.Database.Port = oldPort
	}()

	tokenString, err := auth.GenerateToken(acc.Id, acc.Username, jwt.SigningMethodHS256, []byte("supersecretsigningkey"))
	if err != nil {
		t.Fatal(err)
	}

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, "http://localhost:8080/api/auth/delete", nil)

	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Authorization", tokenString)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	var res map[string]interface{}
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.NotNil(t, res["status"])
	assert.NotNil(t, res["message"])
}
