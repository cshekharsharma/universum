#!/bin/bash

# Halt on any error
set -e

# Define your application's binary name
APP_NAME="universum"

# Define the build version and other ldflags
VERSION="1.0.0"
BUILD_TIME=$(date -u '+%Y-%m-%d_%I:%M:%S%p')
GIT_HASH=$(git rev-parse HEAD)
LDFLAGS="-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitHash=${GIT_HASH}"

printf "\e[42m\n\nBUILDING GO APPLICATION: $APP_NAME...\n\n\e[0m"

# Format the Go code
printf "\e[32m\nFORMATTING SOURCE CODE...\n\n\e[0m"
printf ">> gofmt -w\n"
gofmt -w .

# Run go static analysis on source code
printf "\e[32m\nRUNNING GO STATIC-CHECK...\n\n\e[0m"
printf ">> staticcheck ./...\n"
staticcheck ./...

# Run tests
printf "\e[32m\nRUNNING UNIT TESTS...\n\n\e[0m"
printf ">> rm -rf ./coverage.txt && go run tools/gotest_exec.go && go tool cover -func ./coverage.txt \n"
rm -rf /tmp/coverage.txt
go run tools/gotest_exec.go
go tool cover -func ./coverage.txt

go clean -modcache

# Tidy up dependencies
printf "\e[32m\nVERIFYING MODULE DEPEDENCIES...\n\n\e[0m"
printf ">> go mod verify\n"
go mod verify

# Build the binary
printf "\e[32m\nBUILDING THE BINARY...\n\n\e[0m"

set -x
go build -ldflags "$LDFLAGS" -o bin/$APP_NAME && chmod +x bin/$APP_NAME
set +x

printf "\e[32m\n\nBUILD SUCCESSFUL. BINARY GENERATED AT: ./bin/$APP_NAME \e[0m\n\n"
