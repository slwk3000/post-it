APP_NAME := Post-it
BINARY := post-it
BUILD_DIR := build
APP_BUNDLE := $(BUILD_DIR)/$(APP_NAME).app

.PHONY: all build test app run clean

all: test build

test:
	go test -v ./...

build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY) ./cmd/post-it

app: build
	@./scripts/bundle_app.sh

run: app
	@killall post-it 2>/dev/null || true
	open $(APP_BUNDLE)

clean:
	rm -rf $(BUILD_DIR)
