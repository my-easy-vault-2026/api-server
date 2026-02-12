APP_NAME := easy_vault_api
VERSION ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)
BIN_DIR := bin
BIN_PATH := $(BIN_DIR)/$(APP_NAME)

.PHONY: build docker clean prune

build:
	@mkdir -p $(BIN_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BIN_PATH) main.go
	
docker: 
	@docker build -f docker/Dockerfile -t $(APP_NAME):$(VERSION) .

clean:
	@rm -f $(BIN_PATH)

prune:
	@docker image prune -f