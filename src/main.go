package main

import (
	"flag"
	"flhansen/application-manager/login-service/src/controller"
	"flhansen/application-manager/login-service/src/service"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

func main() {
	os.Exit(runApplication())
}

func runApplication() int {
	configPath := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	var serviceConfig service.ServiceConfig

	if *configPath != "" {
		fileContent, err := ioutil.ReadFile(*configPath)
		if err != nil {
			fmt.Printf("An error occured while reading the configuration file %s: %v\n", *configPath, err)
			return 1
		}

		if err := yaml.Unmarshal(fileContent, &serviceConfig); err != nil {
			fmt.Printf("An error occured while unmarshalling the configuration file content: %v\n", err)
			return 1
		}
	} else {
		serviceConfig.Host = os.Getenv("APPMAN_HOST")
		serviceConfig.Port, _ = strconv.Atoi(os.Getenv("APPMAN_PORT"))
		serviceConfig.Jwt = service.JwtConfig{}
		serviceConfig.Jwt.SignKey = []byte(os.Getenv("APPMAN_JWT_SIGNKEY"))
		serviceConfig.Database = controller.DbConfig{}
		serviceConfig.Database.Host = os.Getenv("APPMAN_DATABASE_HOST")
		serviceConfig.Database.Port, _ = strconv.Atoi(os.Getenv("APPMAN_DATABASE_PORT"))
		serviceConfig.Database.Username = os.Getenv("APPMAN_DATABASE_USERNAME")
		serviceConfig.Database.Password = os.Getenv("APPMAN_DATABASE_PASSWORD")
		serviceConfig.Database.Database = os.Getenv("APPMAN_DATABASE_NAME")
	}

	s := service.New(serviceConfig)

	if err := s.Start(); err != nil {
		fmt.Printf("An error occured while starting the service: %v\n", err)
		return 1
	}

	return 0
}
