# Docker Development Environment

This directory contains Docker infrastructure for the speak-to-ai project, providing isolated development and build environments.

## Quick Start

```bash
# Build Docker images
make docker-build

# Start development environment
make docker-dev

# Run linter
make docker-lint

# Run tests
make docker-test

# Build packages
make docker-appimage
make docker-flatpak
```

## Services

### `dev` - Development Environment
- **Image**: `docker/Dockerfile.dev`
- **Purpose**: Full development environment with GUI dependencies
- **Includes**: Go, golangci-lint, GUI libraries for systray
- **Usage**: `make docker-dev`

### `lint` - Linting Service  
- **Image**: `docker/Dockerfile.lint`
- **Purpose**: Lightweight linting and static analysis
- **Usage**: `make docker-lint`

### `test` - Testing Service
- **Image**: `docker/Dockerfile.dev` (reused)
- **Purpose**: Run tests with all dependencies
- **Usage**: `make docker-test`

### `build-appimage` - AppImage Builder
- **Image**: `docker/Dockerfile.appimage`
- **Purpose**: Build AppImage packages with all dependencies
- **Usage**: `make docker-appimage`

### `build-flatpak` - Flatpak Builder
- **Image**: `docker/Dockerfile.flatpak`
- **Purpose**: Build Flatpak packages (validation)
- **Usage**: `make docker-flatpak`

### `whisper-builder` - Whisper.cpp Builder
- **Image**: `docker/Dockerfile.dev` (reused)
- **Purpose**: Build whisper.cpp libraries shared between services
- **Usage**: `make docker-whisper`

## Docker Profiles

Services are organized into profiles for efficient resource usage:

- **`dev`**: Development environment
- **`lint`**: Linting only
- **`test`**: Testing only
- **`ci`**: CI pipeline (lint + test)
- **`build`**: Package building (AppImage + Flatpak)
- **`appimage`**: AppImage building only
- **`flatpak`**: Flatpak building only
- **`init`**: Whisper.cpp initialization

## Volumes

### Persistent Volumes
- **`go-cache`**: Go module cache (shared between services)
- **`whisper-libs`**: Built whisper.cpp libraries (shared)
- **`build-cache`**: Build artifacts cache
- **`appimage-dist`**: AppImage distribution files

### Bind Mounts
- **Source code**: Mounted read-only for most services
- **Development**: Full read-write access for dev service

## Common Workflows

### Development
```bash
# Start development environment
make docker-dev

# Inside container:
source bash-scripts/dev-env.sh
make build-systray
make test
golangci-lint run
```

### CI Pipeline
```bash
# Run full CI pipeline
make docker-ci
```

### Building Packages
```bash
# Build whisper.cpp first (shared dependency)
make docker-whisper

# Build packages
make docker-appimage
make docker-flatpak
```

### Troubleshooting
```bash
# Check container status
make docker-ps

# View logs
make docker-logs

# Clean up
make docker-clean

# Complete cleanup (including images)
make docker-clean-all
```

## Environment Variables

All containers have CGO environment variables pre-configured:
```bash
CGO_ENABLED=1
C_INCLUDE_PATH=/app/lib
LIBRARY_PATH=/app/lib
CGO_CFLAGS=-I/app/lib
CGO_LDFLAGS=-L/app/lib -lwhisper -lggml-cpu -lggml
LD_LIBRARY_PATH=/app/lib
PKG_CONFIG_PATH=/app/lib
```

## Benefits

1. **No System Dependencies**: No need to install GUI libraries on host
2. **Consistent Environment**: Same environment across all developers
3. **Parallel Development**: Multiple services can run simultaneously
4. **Easy CI/CD**: Ready-made pipeline for automated testing
5. **Package Building**: Complete isolation for building distributable packages

## Notes

- **Flatpak building**: Complex in containers, mainly for validation
- **AppImage building**: Fully functional in containers
- **GUI dependencies**: Required for systray functionality
- **Whisper.cpp**: Built once and shared between services via volumes