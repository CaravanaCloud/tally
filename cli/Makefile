# Makefile for Go project

# Variables
APP_NAME := tally
SRC := $(wildcard *.go)
LINTER := golangci-lint
TEST_CMD := go test ./...
BUILD_DIR := ./bin

# Targets
all: build

build: $(SRC)
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(SRC)

lint:
	@echo "Running linter..."
	@$(LINTER) run

test:
	@echo "Running tests..."
	@$(TEST_CMD)

run: build
	@echo "Running $(APP_NAME)..."
	@$(BUILD_DIR)/$(APP_NAME)

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)

.PHONY: all build lint test run clean
