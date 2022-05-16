package main

import (
	"encoding/json"
	"flag"
	"flhansen/application-manager/login-service/src/auth"
	"flhansen/application-manager/login-service/src/database"
	"flhansen/application-manager/login-service/src/service"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunApplication(t *testing.T) {
	config := service.ServiceConfig{
		JwtConfig: auth.JwtConfig{
			SignKey: "supersecretsigningkey",
		},
		DatabaseConfig: database.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Username: "test",
			Password: "test",
			Database: "test",
		},
	}

	configData, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(os.TempDir(), "test_config.yml")
	if err = ioutil.WriteFile(configPath, configData, 0777); err != nil {
		t.Fatal(err)
	}

	defer os.Remove(configPath)

	done := make(chan int, 1)

	go func() {
		flag.CommandLine = flag.NewFlagSet("flags set", flag.ExitOnError)
		os.Args = append([]string{"flags set"}, "-dbconfig="+configPath)
		done <- runApplication()
	}()

	select {
	case <-time.After(500 * time.Millisecond):
		return
	case exitCode := <-done:
		assert.Equal(t, 0, exitCode)
	}
}

func TestRunApplicationInvalidConfig(t *testing.T) {
	done := make(chan int, 1)

	go func() {
		flag.CommandLine = flag.NewFlagSet("flags set", flag.ExitOnError)
		os.Args = append([]string{"flags set"}, "-dbconfig=invalid.yml")
		done <- runApplication()
	}()

	select {
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Application is not terminating")
	case exitCode := <-done:
		assert.Equal(t, 1, exitCode)
	}
}
