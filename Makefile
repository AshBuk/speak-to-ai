.PHONY: all build build-systray cli test clean deps whisper-libs appimage flatpak help fmt lint test-integration test-integration-full docker-% docker-build docker-dev docker-lint docker-clean

# Variables
GO_VERSION := 1.24.1
BINARY_NAME := speak-to-ai
CLI_BINARY_NAME := speak-to-ai-cli
BUILD_DIR := build
LIB_DIR := lib
DIST_DIR := dist
# Optional: set to a tag or commit hash to pin whisper.cpp version for reproducible builds
# Example (CI recommended): make WHISPER_CPP_REF=v1.7.6
WHISPER_CPP_REF ?= v1.7.6

# CGO environment
# These variables are necessary for CGO to find the whisper.cpp libraries.
# They tell the Go compiler where to find the C header files (.h) and the compiled C libraries (.so, .a).
# Because we are building whisper.cpp locally into the `lib` directory, we need to
# explicitly tell CGO where to look. Without these, `go build` and `go test` will fail
# because they won't be able to find the required C dependencies.
export C_INCLUDE_PATH := $(PWD)/$(LIB_DIR)
export LIBRARY_PATH := $(PWD)/$(LIB_DIR)
export CGO_CFLAGS := -I$(PWD)/$(LIB_DIR)
export CGO_LDFLAGS := -L$(PWD)/$(LIB_DIR) -lwhisper -lggml-cpu -lggml
export LD_LIBRARY_PATH := $(PWD)/$(LIB_DIR):$(LD_LIBRARY_PATH)
export PKG_CONFIG_PATH := $(PWD)/$(LIB_DIR):$(PKG_CONFIG_PATH)

# ============================================================================
# Formatting & Lint
# ============================================================================

# Format source code (full: go fmt + goimports)
fmt:
	@echo "=== Running go fmt and goimports in Docker ==="
	docker compose --profile dev run --rm dev sh -c "go fmt ./... && go install golang.org/x/tools/cmd/goimports@latest && goimports -w ."

# Lint source code
lint:
	@echo "=== Running linter in Docker ==="
	docker compose --profile lint run --rm lint

# -----------------------------------------------------------------------------
# Test targets
# -----------------------------------------------------------------------------

# Run tests via Docker (reuses dev image with CGO + whisper.cpp)
test:
	@echo "=== Running tests via Docker ==="
	docker compose --profile dev run --rm dev go test -v -cover ./...

test-integration:
	@echo "=== Running integration tests (fast mode) via Docker ==="
	docker compose --profile dev run --rm -e CGO_ENABLED=0 dev go test -tags=integration ./tests/integration/... -short -v

test-integration-full:
	@echo "=== Running full integration tests via Docker ==="
	docker compose --profile dev run --rm dev go test -tags=integration ./tests/integration/... -v

# ============================================================================
# Build & Dependencies
# ============================================================================

# Default target (with systray support for desktop usage)
all: deps whisper-libs build-systray

# Download Go dependencies
deps:
	@echo "=== Downloading Go dependencies ==="
	go mod download
	go mod tidy
	go mod verify

# Build whisper.cpp libraries
whisper-libs: $(LIB_DIR)/whisper.h

$(LIB_DIR)/whisper.h:
	@echo "=== Building whisper.cpp libraries ==="
	mkdir -p $(BUILD_DIR)
	cd $(BUILD_DIR) && \
	if [ ! -d "whisper.cpp" ]; then \
		echo "Cloning whisper.cpp..."; \
		git clone https://github.com/ggml-org/whisper.cpp.git; \
	fi
	cd $(BUILD_DIR)/whisper.cpp && \
		if [ -n "$(WHISPER_CPP_REF)" ]; then \
			echo "Checking out whisper.cpp ref $(WHISPER_CPP_REF)"; \
			git fetch --tags; \
			git checkout $(WHISPER_CPP_REF); \
		fi; \
		rm -rf build && \
		cmake -B build && \
		cmake --build build --config Release
	mkdir -p $(LIB_DIR)
	cp $(BUILD_DIR)/whisper.cpp/build/src/libwhisper.so* $(LIB_DIR)/ || true
	cp $(BUILD_DIR)/whisper.cpp/build/src/libwhisper.a $(LIB_DIR)/ || true
	cp $(BUILD_DIR)/whisper.cpp/include/whisper.h $(LIB_DIR)/
	cp $(BUILD_DIR)/whisper.cpp/ggml/include/*.h $(LIB_DIR)/ || true
	cp $(BUILD_DIR)/whisper.cpp/build/ggml/src/libggml*.* $(LIB_DIR)/ || true
	@echo "Library files:"
	@ls -la $(LIB_DIR)/

# Build the main binary
build: deps whisper-libs
	@echo "=== Building $(BINARY_NAME) ==="
	go build -v -o $(BINARY_NAME) cmd/daemon/main.go
	@echo "Build completed: $(BINARY_NAME)"
	@ls -lh $(BINARY_NAME)

# Build with systray support
build-systray: deps whisper-libs
	@echo "=== Building $(BINARY_NAME) with systray support ==="
	go build -tags systray -v -o $(BINARY_NAME) cmd/daemon/main.go
	@echo "Build completed: $(BINARY_NAME)"
	@ls -lh $(BINARY_NAME)

# Build standalone CLI helper (AppImage build already bundles this)
cli: deps
	@echo "=== Building CLI helper $(CLI_BINARY_NAME) ==="
	go build -v -o $(CLI_BINARY_NAME) cmd/cli/main.go
	@echo "Build completed: $(CLI_BINARY_NAME)"
	@ls -lh $(CLI_BINARY_NAME)


# ============================================================================
# Packaging
# ============================================================================

# Build AppImage
appimage: build
	@echo "=== Building AppImage ==="
	bash bash-scripts/build-appimage.sh

# Build Flatpak
flatpak: build
	@echo "=== Building Flatpak ==="
	bash bash-scripts/build-flatpak.sh

# Clean build artifacts
clean:
	@echo "=== Cleaning build artifacts ==="
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)
	rm -rf $(LIB_DIR)
	rm -rf $(DIST_DIR)
	go clean -cache
	@echo "Clean completed"

# ============================================================================
# Utilities
# ============================================================================

# Check if required tools are available
check-tools:
	@echo "=== Checking required tools ==="
	@command -v go >/dev/null 2>&1 || { echo "Go is required but not installed"; exit 1; }
	@command -v cmake >/dev/null 2>&1 || { echo "CMake is required but not installed"; exit 1; }
	@command -v git >/dev/null 2>&1 || { echo "Git is required but not installed"; exit 1; }
	@echo "All required tools are available"

# ============================================================================
# Docker Development Commands
# ============================================================================

# Docker build commands
docker-build:
	@echo "=== Building all Docker images ==="
	docker compose build

docker-build-dev:
	@echo "=== Building development Docker image ==="
	docker compose build dev

docker-build-lint:
	@echo "=== Skipping build: using official golangci-lint image ==="
	@true

# Docker development environment
docker-up:
	@echo "=== Starting Docker development services ==="
	docker compose --profile dev up -d

docker-dev:
	@echo "=== Starting Docker development environment ==="
	docker compose --profile dev up -d dev
	@echo "=== Entering development container ==="
	docker compose exec dev bash

docker-dev-stop:
	@echo "=== Stopping Docker development environment ==="
	docker compose --profile dev down

docker-down:
	@echo "=== Stopping all Docker services ==="
	docker compose down -v --remove-orphans

# Docker whisper.cpp setup
docker-whisper:
	@echo "=== Building whisper.cpp libraries in Docker ==="
	docker compose --profile init up whisper-builder

# Docker building packages
docker-appimage:
	@echo "=== Building AppImage via docker build (multi-stage) ==="
	docker build -f docker/Dockerfile.appimage --target artifacts --output type=local,dest=$(DIST_DIR)/appimage .

docker-flatpak:
	@echo "=== Building Flatpak via docker build (multi-stage) ==="
	docker build -f docker/Dockerfile.flatpak --target artifacts --output type=local,dest=$(DIST_DIR)/flatpak .

# Docker CI pipeline
docker-ci:
	@echo "=== Running full CI pipeline in Docker ==="
	docker compose --profile init up whisper-builder
	docker compose --profile ci run --rm lint
	$(MAKE) test
	$(MAKE) docker-appimage
	$(MAKE) docker-flatpak
	@echo "=== CI pipeline completed successfully ==="

# Docker cleanup
docker-clean:
	@echo "=== Cleaning Docker resources ==="
	docker compose down --volumes --remove-orphans
	docker system prune -f

docker-clean-all:
	@echo "=== Cleaning all Docker resources including images ==="
	docker compose down --volumes --remove-orphans --rmi all
	docker system prune -af

# Docker utility commands
docker-logs:
	@echo "=== Showing Docker logs ==="
	docker compose logs -f

docker-ps:
	@echo "=== Showing Docker containers ==="
	docker compose ps

docker-shell:
	@echo "=== Opening shell in development container ==="
	docker compose --profile dev run --rm dev bash

# ============================================================================
# Help
# ============================================================================

help:
	@echo "Available targets:"
	@echo ""
	@echo "General:"
	@echo "  all                   - Build everything (deps + whisper + binary)"
	@echo "  build                 - Build Go binary only"
	@echo "  build-systray         - Build with systray support"
	@echo "  deps                  - Download Go dependencies"
	@echo "  whisper-libs          - Build whisper.cpp libraries"
	@echo "  fmt                   - Format Go code (go fmt + goimports)"
	@echo "  lint                  - Run linter and code quality checks"
	@echo "  clean                 - Clean build artifacts"
	@echo ""
	@echo "Tests:"
	@echo "  test                  - Run unit tests"
	@echo "  test-integration      - Run integration tests (fast mode)"
	@echo "  test-integration-full - Run full integration tests (with CGO)"
	@echo ""
	@echo "Packaging:"
	@echo "  appimage              - Build AppImage"
	@echo "  flatpak               - Build Flatpak"
	@echo ""
	@echo "Docker:"
	@echo "  docker-up             - Start development services (docker compose up -d)"
	@echo "  docker-down           - Stop all services (docker compose down)"
	@echo "  docker-dev            - Start and enter development environment"
	@echo "  docker-dev-stop       - Stop development environment"
	@echo "  docker-shell          - Open shell in dev container"
	@echo "  docker-build          - Build all Docker images"
	@echo "  docker-build-dev      - Build development Docker image"
	@echo "  docker-build-lint     - Skipping build: using official golangci-lint image"
	@echo "  docker-whisper        - Build whisper.cpp libraries in Docker"
	@echo "  docker-appimage       - Build AppImage in Docker"
	@echo "  docker-flatpak        - Build Flatpak in Docker"
	@echo "  docker-ci             - Run full CI pipeline"
	@echo "  docker-logs           - Show Docker logs"
	@echo "  docker-ps             - Show Docker containers"
	@echo "  docker-clean          - Clean Docker resources"
	@echo "  docker-clean-all      - Clean everything including images"
