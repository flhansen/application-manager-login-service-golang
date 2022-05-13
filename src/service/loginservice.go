package service

import (
	"encoding/json"
	"flhansen/application-manager/login-service/src/database"
	"flhansen/application-manager/login-service/src/security"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
)

type LoginService struct {
	Port     int
	Host     string
	Router   *httprouter.Router
	Database *database.PostgresContext
}

type JwtClaims struct {
	UserId   int    `json:"userId"`
	Username string `json:"username"`
	jwt.StandardClaims
}

func (service LoginService) LoginHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, NewApiResponse(http.StatusInternalServerError, "An error occured while parsing the request body"))
		return
	}

	acc, err := service.Database.GetAccountByUsername(req.Username)

	if err != nil || !security.ValidatePassword(req.Password, acc.Password) || acc.Id == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, NewApiResponse(http.StatusUnauthorized, "Wrong credentials"))
		return
	}

	claims := JwtClaims{
		UserId:   acc.Id,
		Username: acc.Username,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().UnixMilli(),
			ExpiresAt: time.Now().UnixMilli() + 10*60*60*1000,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, _ := token.SignedString([]byte("supersecretsigningkey"))

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, NewApiResponseObject(http.StatusOK, "User has been logged in", map[string]interface{}{"token": signedToken}))
}

func (service LoginService) RegisterHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, NewApiResponse(http.StatusInternalServerError, "An error occured while parsing the request body"))
		return
	}

	_, err := service.Database.InsertAccount(req.Username, req.Password, req.Email, time.Now())

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, NewApiResponse(http.StatusBadRequest, "User already exists"))
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, NewApiResponse(http.StatusOK, "User registered"))
}

func NewWithContext(host string, port int, context *database.PostgresContext) LoginService {
	service := LoginService{
		Port:     port,
		Host:     host,
		Router:   httprouter.New(),
		Database: context,
	}

	service.Router.POST("/api/auth/login", service.LoginHandler)
	service.Router.POST("/api/auth/register", service.RegisterHandler)

	return service
}

func NewWithConfig(host string, port int, config database.DatabaseConfig) LoginService {
	context := database.NewContext(config.Host, config.Port, config.Username, config.Password, config.Database)
	service := NewWithContext(host, port, context)
	return service
}

func NewWithConfigFile(host string, port int, configPath string) (LoginService, error) {
	context, err := database.NewContextFromConfig(configPath)

	if err != nil {
		return LoginService{}, err
	}

	service := NewWithContext(host, port, context)
	return service, nil
}

func (service LoginService) Start() error {
	return http.ListenAndServe(fmt.Sprintf("%s:%d", service.Host, service.Port), service.Router)
}
