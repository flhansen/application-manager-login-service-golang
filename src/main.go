package main

import (
	"flag"
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
	flag.Parse()

	fmt.Printf("Using database configuration file: %s\n", *configPath)
	server, err := service.NewWithConfigFile(*host, *port, *configPath)

	if err != nil {
		fmt.Printf("Could not create login service instance: %v\n", err)
		return 1
	}

	server.Start()
	return 0
}
