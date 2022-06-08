# Application Manager Login Service
The microservice of the Application Manager for authentication/authorization tasks.

![go test](https://github.com/flhansen/application-manager-login-service-golang/actions/workflows/test.yml/badge.svg)
[![codecov](https://codecov.io/gh/flhansen/application-manager-login-service-golang/branch/master/graph/badge.svg?token=CU2QV6EDA7)](https://codecov.io/gh/flhansen/application-manager-login-service-golang)

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

## Prepare the database

    DROP TABLE IF EXISTS account;
    CREATE TABLE account (
        id SERIAL PRIMARY KEY,
        username VARCHAR(80) UNIQUE NOT NULL,
        password VARCHAR(80) NOT NULL,
        email VARCHAR(80) UNIQUE NOT NULL,
        creation_date TIMESTAMP WITH TIME ZONE DEFAULT now()
    );

## Run the tests
Make sure you have a local instance of the PostgreSQL database running. The
tests expect the database running on `localhost` and port `5432`. Also, for
running the tests, make sure the test database `test` and the user `test:test`
is configured. You don't need to create entities, because the tests theirselves
will create those automatically. Here is an example of how you can configure the
database using [Docker](https://docker.com/).

    docker run --name postgres -dp 5432:5432 -e POSTGRES_PASSWORD=test -e POSTGRES_USER=test -e POSTGRES_DB=test -d postgres

It is important to **not** run the tests in parallel, because they could
interfere each other while executing queries on the database.

    go test -p 1 ./src/...

To obtain test coverage information you may want to execute the following command.

    go test -p 1 --cover ./src/... -coverpkg=./src/... -coverprofile /tmp/cover.out

After that you can create a nicer view for the test coverage using

    go tool cover -html /tmp/cover.out -o /tmp/cover.html

## Run using Docker
You can build the image yourself by executing

    docker build -t login-service .

Then create and start the container using

    docker run -itd --name=login-service -p 7043:7043 -t login-service

To setup the container, you can use environment variables described in the following table.

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `APPMAN_LOGIN_HOST` | localhost | |
| `APPMAN_LOGIN_PORT` | 7043      | |
| `APPMAN_DATABASE_HOST` | localhost | |
| `APPMAN_DATABASE_PORT` | 5432 | |
| `APPMAN_DATABASE_USERNAME` | postgres | |
| `APPMAN_DATABASE_PASSWORD` | password | |
| `APPMAN_DATABASE_NAME` | database | |

## Endpoints

- `POST` `/api/auth/register` Register a new account
- `POST` `/api/auth/login` Create auth token for account
- `DELETE` `/api/auth/delete` Delete account