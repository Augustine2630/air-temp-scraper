.PHONY: build clean run test vet lint help

BINARY_NAME=temp-scraper
BIN_DIR=./bin
CMD_DIR=./cmd/scraper

help:
	@echo "Available targets:"
	@echo "  build       - Build the scraper binary"
	@echo "  run         - Build and run the scraper locally"
	@echo "  test        - Run tests (if any)"
	@echo "  vet         - Run go vet"
	@echo "  clean       - Remove build artifacts"

build:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=1 go build -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Built $(BIN_DIR)/$(BINARY_NAME)"

run: build
	TEMP_SCRAPER_PORT=9100 TEMP_SCRAPER_INTERVAL=30s $(BIN_DIR)/$(BINARY_NAME)

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -rf $(BIN_DIR)
	go clean

.DEFAULT_GOAL := help
