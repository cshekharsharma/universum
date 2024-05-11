# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME=universum

all: test build

test:
    $(GOCMD) run tools/gotest_exec.go

# Build for current OS
build:
    $(GOBUILD) -o $(BINARY_NAME) -v

clean:
    $(GOCLEAN)
    rm -f $(BINARY_NAME)
    rm -f $(BINARY_NAME)-linux
    rm -f $(BINARY_NAME)-darwin
    rm -f $(BINARY_NAME)-windows.exe

run:
    ./$(BINARY_NAME)

# Cross compilation examples
build-linux:
    GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-linux -v

build-mac:
    GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BINARY_NAME)-darwin -v

