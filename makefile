GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get

BINARY_NAME = ./bin/chirpy

build:
	$(GOBUILD) -o $(BINARY_NAME) -v cmd/chirpy/main.go 

run: build
	$(BINARY_NAME) --debug
