package main

import (
	"flag"
	"flhansen/application-manager/login-service/src/database"
	"flhansen/application-manager/login-service/src/service"
	"fmt"
	"os"
)

func main() {
	os.Exit(runApplication())
}

func runApplication() int {
	port := flag.Int("port", 8080, "Listening port")
	host := flag.String("host", "localhost", "Host")
	configPath := flag.String("dbconfig", "", "Database Configuration File")
	dbHost := flag.String("dbhost", "localhost", "Database Host")
	dbPort := flag.Int("dbport", 5432, "Database Port")
	dbUsername := flag.String("dbusername", "username", "Database Username")
	dbPassword := flag.String("dbpassword", "password", "Database Password")
	dbDatabase := flag.String("dbdatabase", "database", "Database Name")
	flag.Parse()

	var err error
	var server service.LoginService

	if *configPath != "" {
		fmt.Printf("Using database configuration file: %s\n", *configPath)
		server, err = service.NewWithConfigFile(*host, *port, *configPath)
	} else {
		server = service.NewWithConfig(*host, *port, database.DatabaseConfig{
			Host:     *dbHost,
			Port:     *dbPort,
			Username: *dbUsername,
			Password: *dbPassword,
			Database: *dbDatabase,
		})
	}

	if err != nil {
		fmt.Printf("Could not create login service instance: %v\n", err)
		return 1
	}

	server.Start()
	return 0
}
