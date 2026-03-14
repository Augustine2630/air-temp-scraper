.PHONY: build build-arm64 build-amd64 clean run test vet help release install

GOOS=darwin
ARCH=$(shell uname -m)
BIN_DIR=./bin
CMD_DIR=./cmd/scraper

ifeq ($(ARCH),arm64)
	GOARCH=arm64
	BINARY_NAME=temp-scraper_darwin_arm64
else ifeq ($(ARCH),x86_64)
	GOARCH=amd64
	BINARY_NAME=temp-scraper_darwin_amd64
else
	BINARY_NAME=temp-scraper
endif

help:
	@echo "Available targets:"
	@echo "  build        - Build for current architecture (arm64 or amd64)"
	@echo "  build-arm64  - Build for Apple Silicon (arm64)"
	@echo "  build-amd64  - Build for Intel (amd64)"
	@echo "  run          - Build and run locally"
	@echo "  release      - Build release binaries for both architectures"
	@echo "  clean        - Remove build artifacts"
	@echo "  test         - Run go test"
	@echo "  vet          - Run go vet"

build: $(BIN_DIR)/$(BINARY_NAME)

$(BIN_DIR)/$(BINARY_NAME):
	@mkdir -p $(BIN_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=1 go build \
		-o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Built $(BIN_DIR)/$(BINARY_NAME)"

build-arm64:
	@mkdir -p $(BIN_DIR)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build \
		-o $(BIN_DIR)/temp-scraper_darwin_arm64 $(CMD_DIR)
	@echo "Built $(BIN_DIR)/temp-scraper_darwin_arm64"

build-amd64:
	@mkdir -p $(BIN_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build \
		-o $(BIN_DIR)/temp-scraper_darwin_amd64 $(CMD_DIR)
	@echo "Built $(BIN_DIR)/temp-scraper_darwin_amd64"

release: build-arm64 build-amd64
	@echo "Release binaries ready:"
	@ls -lh $(BIN_DIR)/temp-scraper_darwin_*

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
