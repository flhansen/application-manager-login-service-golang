package main

import (
	"flag"
	"flhansen/application-manager/login-service/src/service"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	os.Exit(runApplication())
}

func runApplication() int {
	homeDir, _ := os.UserHomeDir()
	port := flag.Int("port", 8080, "Listening port")
	host := flag.String("host", "localhost", "Host")
	configPath := flag.String("dbconfig", filepath.Join(homeDir, ".secret/application_manager_login_service_db.yml"), "dbconfig")
	flag.Parse()

	fmt.Printf("Using database configuration file: %s\n", *configPath)
	service, err := service.New(*host, *port, *configPath)

	if err != nil {
		fmt.Printf("Could not create login service instance: %v\n", err)
		return 1
	}

	service.Start()
	return 0
}
