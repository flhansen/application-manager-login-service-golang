package database

import (
	"time"
)

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type Account struct {
	Id           int
	Username     string
	Password     string
	Email        string
	CreationDate time.Time
}
