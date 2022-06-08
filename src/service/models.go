package service

import (
	"encoding/json"
	"flhansen/application-manager/login-service/src/controller"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type ServiceConfig struct {
	Host     string              `yaml:"host"`
	Port     int                 `yaml:"port"`
	Jwt      JwtConfig           `yaml:"jwt"`
	Database controller.DbConfig `yaml:"database"`
}

func NewApiResponse(status int, message string) string {
	// Create the default response message
	response := map[string]interface{}{
		"status":  status,
		"message": message,
	}

	// Encode JSON object to string
	jsonObj, _ := json.Marshal(response)
	return string(jsonObj)
}

func NewApiResponseObject(status int, message string, moreProps map[string]interface{}) string {
	// Create the default response message
	response := map[string]interface{}{
		"status":  status,
		"message": message,
	}

	// Copy all other properties to the response
	for k, v := range moreProps {
		if _, ok := moreProps[k]; ok {
			response[k] = v
		}
	}

	// Encode JSON object to string
	jsonObj, _ := json.Marshal(response)
	return string(jsonObj)
}
