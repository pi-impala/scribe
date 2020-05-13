# Go build system commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOTIDY=$(GOCMD) mod tidy

# Binary specific parameters
BIN_NAME=scribe

all: test build

test:
	$(GOTEST) -v ./...

build:
	$(GOTIDY)
	$(GOBUILD) -o $(BIN_NAME) -v 

clean:
	$(GOCLEAN)

build-prod-linux: clean test
	$(GOTIDY)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="-s -w" -o $(BIN_NAME) -v
	upx --brute $(BIN_NAME)

build-prod-darwin: clean test
	$(GOTIDY)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags="-s -w" -o $(BIN_NAME) -v
	upx --brute $(BIN_NAME)