# Makefile for Go project

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=bin/otto
UI_BINARY_NAME=bin/otto-ui
BINARY_UNIX=$(BINARY_NAME)_unix

default: build

# All target
all: test build

# Build the project
build:
	$(GOBUILD) -o $(BINARY_NAME) -v

build-ui:
	cd ./ui && $(GOBUILD) -o ../$(UI_BINARY_NAME) -v .

run-ui:
	cd ui && air

gen-ui:
	cd ui && templ generate

# Run tests
test:
	$(GOTEST) -v ./...

# Clean the project
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

# Run the project
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME) server

.PHONY: all build build-ui clean test run build-linux
