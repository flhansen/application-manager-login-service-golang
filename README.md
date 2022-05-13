# Application Manager Login Service
The microservice of the Application Manager for authentication/authorization tasks.

![go test](https://github.com/flhansen/application-manager-login-service-golang/actions/workflows/test.yml/badge.svg)

## Requirements

- Git
- Golang (version 1.18.x)
- PostgreSQL

## Get the project
First, be sure Golang is installed on your system. To get the project run the
following command.

    git clone https://github.com/flhansen/application-manager-login-service-golang

Then get all the packages and dependencies using

    go install

## Run the tests
It is important, that the test **do not** run in parallel, because they could
interfere each other while executing queries in the database.

    go test -p 1 ./src/...

To obtain test coverage information you may want to execute the following command.

    go test -p 1 --cover ./src/... -coverpkg=./src/... -coverprofile /tmp/cover.out

After that you can create a nicer view for the test coverage using

    go tool cover -html /tmp/cover.out -o /tmp/cover.html

## Endpoints

- `POST` `/api/auth/register` Register a new user
- `POST` `/api/auth/login` Create auth token for user