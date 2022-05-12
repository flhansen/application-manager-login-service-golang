# Application Manager Login Service

## Requirements

- Git
- Golang (version go1.18.1)
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