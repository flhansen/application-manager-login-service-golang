package main

import (
	"flag"
	"flhansen/application-manager/login-service/src/controller"
	"flhansen/application-manager/login-service/src/service"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func setTestEnv(envs map[string]string) func() {
	oldEnvs := map[string]string{}

	for name, value := range envs {
		if oldValue, ok := os.LookupEnv(name); ok {
			oldEnvs[name] = oldValue
		}

		os.Setenv(name, value)
	}

	return func() {
		for name := range envs {
			oldValue, ok := oldEnvs[name]

			if ok {
				os.Setenv(name, oldValue)
			} else {
				os.Unsetenv(name)
			}
		}
	}
}

func TestRunApplication(t *testing.T) {
	config := service.ServiceConfig{
		Jwt: service.JwtConfig{
			SignKey: "supersecretsigningkey",
		},
		Database: controller.DbConfig{
			Host:     "localhost",
			Port:     5432,
			Username: "test",
			Password: "test",
			Database: "test",
		},
	}

	configData, err := yaml.Marshal(config)
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
		os.Args = append([]string{"flags set"}, "-config="+configPath)
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
		os.Args = append([]string{"flags set"}, "-config=invalid.yml")
		done <- runApplication()
	}()

	select {
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Application is not terminating")
	case exitCode := <-done:
		assert.Equal(t, 1, exitCode)
	}
}

func TestRunApplicationStartError(t *testing.T) {
	config := service.ServiceConfig{
		Host: "test",
		Port: -1,
		Jwt: service.JwtConfig{
			SignKey: "supersecretsigningkey",
		},
		Database: controller.DbConfig{
			Host:     "localhost",
			Port:     -1,
			Username: "test",
			Password: "test",
			Database: "test",
		},
	}

	configData, err := yaml.Marshal(config)
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
		os.Args = append([]string{"flags set"}, "-config="+configPath)
		done <- runApplication()
	}()

	select {
	case <-time.After(500 * time.Millisecond):
		t.Fatal()
	case exitCode := <-done:
		assert.Equal(t, 1, exitCode)
	}
}

func TestInvalidConfigContent(t *testing.T) {
	configData := []byte(`{ "us`)

	configPath := filepath.Join(os.TempDir(), "test_config.yml")
	if err := ioutil.WriteFile(configPath, configData, 0777); err != nil {
		t.Fatal(err)
	}

	defer os.Remove(configPath)

	done := make(chan int, 1)

	go func() {
		flag.CommandLine = flag.NewFlagSet("flags set", flag.ExitOnError)
		os.Args = append([]string{"flags set"}, "-config="+configPath)
		done <- runApplication()
	}()

	select {
	case <-time.After(500 * time.Millisecond):
		t.Fatal()
	case exitCode := <-done:
		assert.Equal(t, 1, exitCode)
	}
}
func TestRunApplicationUsingEnv(t *testing.T) {
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()

	testEnvCloser := setTestEnv(map[string]string{
		"APPMAN_HOST":              "localhost",
		"APPMAN_PORT":              "8080",
		"APPMAN_JWT_SIGNKEY":       "secret",
		"APPMAN_DATABASE_HOST":     "localhost",
		"APPMAN_DATABASE_PORT":     "5432",
		"APPMAN_DATABASE_USERNAME": "test",
		"APPMAN_DATABASE_PASSWORD": "test",
		"APPMAN_DATABASE_NAME":     "test",
	})

	t.Cleanup(testEnvCloser)

	done := make(chan int)
	go func() {
		flag.CommandLine = flag.NewFlagSet("flags set", flag.ExitOnError)
		os.Args = []string{"flags set"}
		done <- runApplication()
	}()

	select {
	case <-time.After(200 * time.Millisecond):
		return
	case exitCode := <-done:
		t.Fatalf("The application terminated with code %d\n", exitCode)
	}
}
