# Makefile for Universum

# This Makefile defines several commands to manage and build the project universum.
# The commands include building, testing, cleaning, and running the project,
# as well as configuring the project and generating test coverage reports.

# Colors
RED         := \033[0;31m
GREEN       := \033[0;32m
YELLOW      := \033[1;33m
BLUE        := \033[0;34m
NC          := \033[0m

BINARYNAME := universum

# Go Commands
GOCMD       := go # Command to run Go
GOBUILD     := $(GOCMD) build  # Command to build Go binaries
GOCLEAN     := $(GOCMD) clean  # Command to clean Go binaries and caches
GOGET       := $(GOCMD) get    # Command to download Go modules
GOFMT       := gofmt           # Command to format Go source code
GOINSTALL   := $(GOCMD) install # Command to install Go packages
STATICCHECK := staticcheck     # Command to run static analysis on Go code
LINTER      := golangci-lint   # Command for comprehensive linting

# Build Variables for app version and build timestamp
GITHASH     := $(shell git rev-parse HEAD)
VERSION     := $(shell git tag --points-at HEAD)
BUILDTIME   := $(shell date +%s)

# Linker flags for various environments, ie dev, prod, and unknown
LDFLAGSDEF  := -X 'main.GitHash=$(GITHASH)' -X 'main.AppVersion=$(VERSION)' -X 'main.BuildTime=$(BUILDTIME)' -X 'main.AppEnv=unknown'
LDFLAGSDEV  := -X 'main.GitHash=$(GITHASH)' -X 'main.AppVersion=$(VERSION)' -X 'main.BuildTime=$(BUILDTIME)' -X 'main.AppEnv=development'
LDFLAGSPROD := -X 'main.GitHash=$(GITHASH)' -X 'main.AppVersion=$(VERSION)' -X 'main.BuildTime=$(BUILDTIME)' -X 'main.AppEnv=production'

# Target: help
# Description: List all make targets with descriptions
.PHONY: help
help:
	@printf "\n${YELLOW}Makefile Targets:${NC}\n\n"
	@printf "  ${GREEN}configure${NC}      - Configure the project\n"
	@printf "  ${GREEN}test${NC}           - Run unit tests\n"
	@printf "  ${GREEN}testcoverage${NC}   - Run unit tests with coverage report\n"
	@printf "  ${GREEN}build${NC}          - Build the application (ENV=unknown)\n"
	@printf "  ${GREEN}build-dev${NC}      - Build the application (ENV=development)\n"
	@printf "  ${GREEN}build-prod${NC}     - Build the application (ENV=production)\n"
	@printf "  ${GREEN}run${NC}            - Run the application\n"
	@printf "  ${GREEN}clean${NC}          - Clean the previous builds\n"
	@printf "  ${GREEN}lint${NC}           - Run static code analysis\n"
	@printf "  ${GREEN}docker-build${NC}   - Build the Docker image\n"
	@printf "  ${GREEN}docker-run${NC}     - Run the Docker container\n"
	@printf "  ${GREEN}all-dev${NC}        - Configure, test, build, and run the application for DEV\n\n"
	@printf "  ${GREEN}all-prod${NC}       - Configure, test, build, and run the application for PROD\n\n"

# Target: configure
# Description: Configure the project by tidying and verifying the modules, formatting the code, and running static analysis.
.PHONY: configure
configure:
	@printf "\n${YELLOW}CONFIGURING THE PACKAGE...${NC}\n\n"
	$(GOINSTALL) honnef.co/go/tools/cmd/staticcheck@latest
	$(GOCMD) mod tidy
	$(GOCMD) mod verify
	$(GOFMT) -s -w .
	$(STATICCHECK) ./...
	@printf "\n"

# Target: test
# Description: Run unit tests for the project.
.PHONY: test
test:
	@printf "\n${YELLOW}RUNNING UNIT TESTS...${NC}\n\n"
	rm -rf ./coverage.txt
	$(GOCMD) run tools/gotest_exec.go
	@printf "\n"

# Target: testcoverage
# Description: Run unit tests and generate a coverage report for the project.
.PHONY: testcoverage
testcoverage:
	@printf "\n${YELLOW}RUNNING UNIT TESTS WITH COVERAGE REPORT...${NC}\n\n"
	rm -rf ./coverage.txt
	$(GOCMD) run tools/gotest_exec.go
	$(GOCMD) tool cover -func ./coverage.txt
	@printf "\n\n"

# Target: build
# Description: Build the project for the current OS with an unknown environment.
.PHONY: build
build:
	@printf "\n${YELLOW}Building the application with: ENV=unknown, VERSION=$(VERSION), BUILDTIME=$(BUILDTIME) ...${NC}\n\n"
	$(GOBUILD) -ldflags "$(LDFLAGSDEF)" -buildvcs=false -o ./bin/$(BINARYNAME) && chmod +x ./bin/$(BINARYNAME)

# Target: build-dev
# Description: Build the project for the current OS with a development environment.
.PHONY: build-dev
build-dev:
	@printf "\n${YELLOW}Building the application with: ENV=development, VERSION=$(VERSION), BUILDTIME=$(BUILDTIME)...${NC}\n\n"
	$(GOBUILD) -ldflags "$(LDFLAGSDEV)" -buildvcs=false -o ./bin/$(BINARYNAME) && chmod +x ./bin/$(BINARYNAME)

# Target: build-prod
# Description: Build the project for the current OS with a production environment.
.PHONY: build-prod
build-prod:
	@printf "\n${YELLOW}Building the application with: ENV=production, VERSION=$(VERSION), BUILDTIME=$(BUILDTIME)...${NC}\n\n"
	$(GOBUILD) -ldflags "$(LDFLAGSPROD)" -buildvcs=false -o ./bin/$(BINARYNAME) && chmod +x ./bin/$(BINARYNAME)

# Target: run
# Description: Run the project binary.
.PHONY: run
run:
	@printf "\n${YELLOW}LAUNCHING THE BINARY...${NC}\n\n"
	./bin/$(BINARYNAME) -config=/etc/universum/config.toml

# Target: clean
# Description: Clean the previous builds and remove the binary.
.PHONY: clean
clean:
	@printf "\n${YELLOW}CLEANING THE PREVIOUS BUILDS...${NC}\n\n"
	$(GOCLEAN)
	rm -f ./bin/$(BINARYNAME)

# Target: lint
# Description: Run static code analysis using golangci-lint
.PHONY: lint
lint:
	@printf "\n${YELLOW}LINTING THE CODEBASE...${NC}\n\n"
	$(LINTER) run ./...

# Target: docker-build
# Description: Build the Docker image for the project.
.PHONY: docker-build
docker-build:
	@printf "\n${YELLOW}BUILDING DOCKER IMAGE...${NC}\n\n"
	docker build -t $(BINARYNAME):latest .

# Target: docker-run
# Description: Run the Docker container for the project.
.PHONY: docker-run
docker-run:
	@printf "\n${YELLOW}RUNNING DOCKER CONTAINER...${NC}\n\n"
	docker run -p 11191:11191 $(BINARYNAME):latest

# Target: configure-build
# Description: Clean, configure, test, and build the project.
.PHONY: configure-build
configure-build: clean configure test build-prod

# Target: all-dev
# Description: Clean, configure, test, build-dev and run the project.
.PHONY: all-dev
all-dev: clean configure test build-dev run

# Target: all-prod
# Description: Clean, configure, test, build-prod and run the project.
.PHONY: all-prod
all-prod: clean configure test build-prod run
