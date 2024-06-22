# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME=universum
GOFMT=gofmt

configure:
		@echo "CONFIGURING THE PACKAGE..."
		$(GOCMD) mod tidy
		$(GOCMD) mod verify
		$(GOFMT) -s -w .
		staticcheck ./...

test:
		@echo "RUNNING UNIT TESTS...\n"
		$(GOCMD) run tools/gotest_exec.go

# Build for current OS
build:
		@echo "BUILDING THE PACKAGE...\n"
		$(GOBUILD) -o ./bin/$(BINARY_NAME)

clean:
		@echo "CLEANING THE PREVIOUS BUILDS...\n"
		$(GOCLEAN)
		rm -f ./bin/$(BINARY_NAME)
		rm -f ./bin/$(BINARY_NAME)-linux
		rm -f ./bin/$(BINARY_NAME)-darwin
		rm -f ./bin/$(BINARY_NAME)-windows.exe

run:
		@echo "LAUNCHING THE BINARY...\n"
		./bin/$(BINARY_NAME)


all: configure test

# Cross compilation examples
build-linux:
		GOOS=linux GOARCH=amd64 $(GOBUILD) -o ./bin/$(BINARY_NAME)-linux -v

build-mac:
		GOOS=darwin GOARCH=arm64 $(GOBUILD) -o ./bin/$(BINARY_NAME)-darwin -v
