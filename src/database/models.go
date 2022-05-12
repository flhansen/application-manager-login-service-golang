package database

import (
	"time"
)

type DatabaseConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

type Account struct {
	Id           int
	Username     string
	Password     string
	Email        string
	CreationDate time.Time
}
