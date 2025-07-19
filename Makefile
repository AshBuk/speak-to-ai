.PHONY: all build build-systray test clean deps whisper-libs appimage flatpak help

# Variables
GO_VERSION := 1.22
BINARY_NAME := speak-to-ai
BUILD_DIR := build
LIB_DIR := lib
DIST_DIR := dist

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

# Default target
all: deps whisper-libs build

# Help target
help:
	@echo "Available targets:"
	@echo "  all          - Build everything (deps + whisper + binary)"
	@echo "  build        - Build Go binary only"
	@echo "  build-systray- Build with systray support"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download Go dependencies"
	@echo "  whisper-libs - Build whisper.cpp libraries"
	@echo "  appimage     - Build AppImage"
	@echo "  flatpak      - Build Flatpak"

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
		git clone https://github.com/ggerganov/whisper.cpp.git; \
	fi
	cd $(BUILD_DIR)/whisper.cpp && \
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

# Run tests
# It is important to use `make test` instead of `go test ./...` directly.
# This target ensures that the CGO environment variables are set correctly before running the tests.
# It also ensures that the whisper.cpp libraries are built and available.
test: deps whisper-libs
	@echo "=== Running tests ==="
	go test -v -cover ./...

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

# Check if required tools are available
check-tools:
	@echo "=== Checking required tools ==="
	@command -v go >/dev/null 2>&1 || { echo "Go is required but not installed"; exit 1; }
	@command -v cmake >/dev/null 2>&1 || { echo "CMake is required but not installed"; exit 1; }
	@command -v git >/dev/null 2>&1 || { echo "Git is required but not installed"; exit 1; }
	@echo "All required tools are available"