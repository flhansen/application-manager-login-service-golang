package main

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunApplication(t *testing.T) {
	done := make(chan int, 1)

	go func() {
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
