package database

import (
	"context"
	"encoding/base64"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	db := NewContext("localhost", 5432, "test", "test", "test")
	db.CreateSchema()
	os.Exit(runAllTests(m))
}

func runAllTests(m *testing.M) int {
	return m.Run()
}

func TestDatabaseCreateSchemaBadConnection(t *testing.T) {
	db := NewContext("localhost", 5432, "test", "wrongpassword", "test")
	err := db.CreateSchema()

	assert.NotNil(t, err)
}

func TestDatabaseCreateSchema(t *testing.T) {
	db := NewContext("localhost", 5432, "test", "test", "test")
	err := db.CreateSchema()

	assert.Nil(t, err)
}

func TestDatabaseQueryBadConnection(t *testing.T) {
	db := NewContext("localhost", 5432, "test", "wrongpassword", "test")
	_, err := db.Query("SELECT * FROM account")

	assert.NotNil(t, err)
}

func TestDatabaseInsertAccountBadConnection(t *testing.T) {
	db := NewContext("localhost", 5432, "test", "wrongpassword", "test")
	_, err := db.InsertAccount("", "", "", time.Now())
	assert.NotNil(t, err)
}

func TestDatabaseDeleteAccountBadConnection(t *testing.T) {
	db := NewContext("localhost", 5432, "test", "wrongpassword", "test")
	err := db.DeleteAccount(0)
	assert.NotNil(t, err)
}

func TestDatabaseDeleteAccount(t *testing.T) {
	db := NewContext("localhost", 5432, "test", "test", "test")

	db.DeleteAccountByUsername("testuser")
	id, err := db.InsertAccount("testuser", "testpass", "testuser@test.com", time.Now())

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		conn, err := pgx.Connect(context.Background(), db.ConnectionString())

		if err != nil {
			t.Fatal(err)
		}

		conn.QueryRow(context.Background(), "DELETE FROM account WHERE username = $1", "testuser")
		conn.Close(context.Background())
	}()

	conn, err := pgx.Connect(context.Background(), db.ConnectionString())

	if err != nil {
		t.Fatal(err)
	}

	defer conn.Close(context.Background())

	var numberRowsBeforeDelete int
	conn.QueryRow(context.Background(), "SELECT count(*) FROM account WHERE username = $1", "testuser").Scan(&numberRowsBeforeDelete)

	if err = db.DeleteAccount(id); err != nil {
		t.Fatal(err)
	}

	var numberRows int
	conn.QueryRow(context.Background(), "SELECT count(*) FROM account WHERE username = $1", "testuser").Scan(&numberRows)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, numberRowsBeforeDelete)
	assert.Equal(t, 0, numberRows)
}

func TestDatabaseGetAccountByUsername(t *testing.T) {
	db := NewContext("localhost", 5432, "test", "test", "test")

	id, err := db.InsertAccount("testuser", "testpass", "testuser@test.com", time.Now())

	if err != nil {
		t.Fatal(err)
	}

	defer db.DeleteAccount(id)

	acc, err := db.GetAccountByUsername("testuser")

	if err != nil {
		t.Fatal(err)
	}

	hash, err := base64.StdEncoding.DecodeString(acc.Password)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, id, acc.Id)
	assert.Equal(t, "testuser", acc.Username)
	assert.NotEqual(t, "testpass", acc.Password)
	assert.Equal(t, 48, len(hash))
	assert.Equal(t, "testuser@test.com", acc.Email)
	assert.True(t, acc.CreationDate.Before(time.Now()))
}

func TestDatabaseGetAccountByUsernameBadConnection(t *testing.T) {
	db := NewContext("localhost", 5432, "test", "wrongpassword", "test")

	acc, err := db.GetAccountByUsername("'' OR 1=1;")

	assert.NotNil(t, err)
	assert.Equal(t, 0, acc.Id)
}

func TestDatabaseNewFromConfig(t *testing.T) {
	filePath := filepath.Join(os.TempDir(), "test_database_file.yml")
	err := ioutil.WriteFile(filePath, []byte(
		"host: localhost\n"+
			"port: 5432\n"+
			"username: test\n"+
			"password: test\n"+
			"database: test"), 0777)

	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(filePath)

	_, err = NewContextFromConfig(filePath)
	assert.Nil(t, err)
}

func TestDatabaseNewFromConfigInvalidFile(t *testing.T) {
	_, err := NewContextFromConfig("invalid/file/path.yml")
	assert.NotNil(t, err)
}

func TestDatabaseNewFromConfigParseError(t *testing.T) {
	filePath := filepath.Join(os.TempDir(), "test_database_file.yml")
	err := ioutil.WriteFile(filePath, []byte("invalid file content"), 0777)

	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(filePath)
	_, err = NewContextFromConfig(filePath)

	assert.NotNil(t, err)
}
