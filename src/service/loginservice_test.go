package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

var loginService LoginService

func TestMain(m *testing.M) {
	os.Exit(runAllTests(m))
}

func runAllTests(m *testing.M) int {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Printf("Could not get the home directory of the current user: %v", err)
		return 1
	}

	configPath := filepath.Join(homeDir, ".secret/application_manager_login_service_db.yml")
	loginService, err = New("localhost", 8080, configPath)

	if err != nil {
		fmt.Printf("Could not create login service instance: %v", err)
		return 1
	}

	// Make sure the user 'testuser' does not exist and then create it
	loginService.Database.DeleteAccountByUsername("testuser")
	_, err = loginService.Database.InsertAccount("testuser", "testpass", "testuser@test.com", time.Now())

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
	token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte("supersecretsigningkey"), nil
	})

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

func TestLoginServiceNewInvalidConfig(t *testing.T) {
	filePath := filepath.Join(os.TempDir(), "test_database_file.yml")
	err := ioutil.WriteFile(filePath, []byte("invalid file content"), 0777)

	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(filePath)
	_, err = New("localhost", 8080, filePath)

	assert.NotNil(t, err)
}
