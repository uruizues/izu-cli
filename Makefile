APP_NAME := izu
VERSION := 0.1.0
BUILD_DIR := dist

.PHONY: build install clean

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/izu/

install: build
	sudo cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/

clean:
	rm -rf $(BUILD_DIR)
