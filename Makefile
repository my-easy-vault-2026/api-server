APP_NAME := easy_vault_api
BIN_DIR := bin
BIN_PATH := $(BIN_DIR)/$(APP_NAME)

.PHONY: build docker clean

build:
	@mkdir -p $(BIN_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BIN_PATH) main.go

docker: build
	@docker build -f docker/Dockerfile -t easy_vault_api .

clean:
	@rm -f $(BIN_PATH)
