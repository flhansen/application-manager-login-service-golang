package database

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"flhansen/application-manager/login-service/src/security"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/jackc/pgx/v4"
	"gopkg.in/yaml.v3"
)

type PostgresContext struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

func NewContext(host string, port int, username string, password string, database string) *PostgresContext {
	return &PostgresContext{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Database: database,
	}
}

func NewContextFromConfig(configPath string) (*PostgresContext, error) {
	fileBytes, err := ioutil.ReadFile(configPath)

	if err != nil {
		return nil, err
	}

	var config DatabaseConfig
	if err = yaml.Unmarshal(fileBytes, &config); err != nil {
		return nil, err
	}

	db := NewContext(config.Host, config.Port, config.Username, config.Password, config.Database)
	return db, nil
}

func (ctx PostgresContext) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", ctx.Username, ctx.Password, ctx.Host, ctx.Port, ctx.Database)
}

func (ctx PostgresContext) Query(query string, args ...interface{}) (pgx.Row, error) {
	conn, err := pgx.Connect(context.Background(), ctx.ConnectionString())

	if err != nil {
		return nil, err
	}

	defer conn.Close(context.Background())

	row := conn.QueryRow(context.Background(), query, args...)
	return row, nil
}

func (ctx PostgresContext) CreateSchema() error {
	_, err := ctx.Query("DROP TABLE IF EXISTS account")

	if err != nil {
		return err
	}

	_, err = ctx.Query(
		`CREATE TABLE account (
			id SERIAL PRIMARY KEY,
			username VARCHAR(80) UNIQUE NOT NULL,
			password VARCHAR(80) NOT NULL,
			email VARCHAR(80) UNIQUE NOT NULL,
			creation_date TIMESTAMP WITH TIME ZONE DEFAULT now()
		)`)

	return err
}

func (ctx PostgresContext) InsertAccount(username string, password string, email string, creationDate time.Time) (int, error) {
	rng := security.RandomGenerator{Reader: rand.Reader}
	salt, err := rng.GenerateSalt(16)

	if err != nil {
		return -1, err
	}

	passwordHash := security.CreatePasswordHash(password, salt)
	passwordHashString := base64.StdEncoding.EncodeToString(passwordHash)

	row, err := ctx.Query("INSERT INTO account (username, password, email, creation_date) VALUES ($1, $2, $3, $4) RETURNING id",
		username, passwordHashString, email, creationDate)

	if err != nil {
		return -1, err
	}

	id := -1
	err = row.Scan(&id)
	return id, err
}

func (ctx PostgresContext) DeleteAccount(accountId int) error {
	_, err := ctx.Query("DELETE FROM account WHERE id = $1", accountId)
	return err
}

func (ctx PostgresContext) DeleteAccountByUsername(username string) error {
	_, err := ctx.Query("DELETE FROM account WHERE username = $1", username)
	return err
}

func (ctx PostgresContext) GetAccountByUsername(username string) (Account, error) {
	row, err := ctx.Query("SELECT id, username, password, email, creation_date FROM account WHERE username = $1", username)

	if err != nil {
		return Account{}, err
	}

	var account Account
	err = row.Scan(&account.Id, &account.Username, &account.Password, &account.Email, &account.CreationDate)

	return account, err
}
