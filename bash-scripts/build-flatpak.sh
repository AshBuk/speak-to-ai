#!/bin/bash

# Speak-to-AI Flatpak development helper script
# For production builds, use GitHub Actions CI/CD pipeline

set -e  # Exit on error

# Configuration
APP_ID="io.github.ashbuk.speak-to-ai"

# Functions
print_header() {
    echo "=== Speak-to-AI Flatpak Development Helper ==="
    echo "⚠️  This script is for local development only"
    echo "⚠️  Production builds should use GitHub Actions CI/CD"
    echo ""
}

check_flatpak_builder() {
    if ! command -v flatpak-builder &> /dev/null; then
        echo "❌ flatpak-builder is not installed"
        echo "💡 For production builds, use GitHub Actions instead"
        echo "💡 To install locally: sudo dnf install flatpak-builder"
        exit 1
    fi
}

check_runtime() {
    # If running in container, skip runtime presence check
    if [ -f "/.dockerenv" ]; then
        echo "🟡 Detected Docker environment: skipping runtime presence check"
        echo "✅ Runtime check skipped (Docker)"
        return
    fi
    
    # Align local runtime check with manifest (GNOME 47)
    if ! flatpak list --runtime | grep -q "org.gnome.Platform.*47"; then
        echo "❌ Required Flatpak runtime (org.gnome.Platform//47) not found"
        echo "💡 For local setup: flatpak install flathub org.gnome.Platform//47 org.gnome.Sdk//47"
        exit 1
    fi
}

check_prerequisites() {
    echo "Checking prerequisites..."
    check_flatpak_builder
    check_runtime
    echo "✅ Prerequisites check passed"
}

check_whisper_cpp() {
    echo "Checking whisper.cpp dependencies..."
    
    # Skip in Docker, CI/CD builds from sources
    if [ -f "/.dockerenv" ]; then
        echo "🟡 Detected Docker environment: skipping whisper.cpp binaries check"
        echo "✅ whisper.cpp check skipped (Docker)"
        return
    fi
    
    if [ ! -f "sources/core/whisper" ] || [ ! -f "sources/core/quantize" ]; then
        echo "❌ whisper.cpp not found"
        echo "💡 Run whisper.cpp build first or let CI/CD handle it"
        exit 1
    fi
    
    echo "✅ whisper.cpp available"
}

check_whisper_model() {
    echo "Checking Whisper model..."
    
    # Skip in Docker, CI/CD handles model download
    if [ -f "/.dockerenv" ]; then
        echo "🟡 Detected Docker environment: skipping model presence check"
        echo "✅ Model check skipped (Docker)"
        return
    fi
    
    if [ ! -f "sources/language-models/base.bin" ]; then
        echo "❌ Whisper model not found"
        echo "💡 Download the model first or let CI/CD handle it"
        exit 1
    fi
    
    echo "✅ Whisper model available"
}

setup_flatpak_environment() {
    echo "🚧 Setting up Flatpak environment..."
    
    # Ensure flathub is added
    flatpak remote-add --if-not-exists flathub https://flathub.org/repo/flathub.flatpakrepo || true

    # Install required runtimes (system-level inside container)
    echo "Installing Flatpak runtimes (org.gnome 47, extensions 24.08)..."
    flatpak install -y flathub org.gnome.Platform//47 org.gnome.Sdk//47 || true
    flatpak install -y flathub org.freedesktop.Sdk.Extension.golang//24.08 || true
    flatpak install -y flathub org.freedesktop.Sdk.Extension.vala//24.08 || true
}

prepare_build_directories() {
    echo "Preparing build directories..."
    mkdir -p dist/flatpak/build-dir dist/flatpak/repo
}

build_flatpak_package() {
    echo "Building Flatpak package..."
    
    BUILDER_FLAGS="--force-clean --install-deps-from=flathub --ccache --repo=dist/flatpak/repo"
    if [ -f "/.dockerenv" ]; then
        BUILDER_FLAGS="${BUILDER_FLAGS} --disable-sandbox"
    fi
    
    # shellcheck disable=SC2086
    flatpak-builder \
      $BUILDER_FLAGS \
      dist/flatpak/build-dir \
      io.github.ashbuk.speak-to-ai.json
}

create_flatpak_bundle() {
    echo "Creating Flatpak bundle..."
    
    OUT_NAME="speak-to-ai-dev.flatpak"
    flatpak build-bundle \
      --runtime-repo=https://flathub.org/repo/flathub.flatpakrepo \
      dist/flatpak/repo \
      "dist/${OUT_NAME}" \
      ${APP_ID}
    
    echo "✅ Flatpak built: dist/${OUT_NAME}"
    ls -lh dist/*.flatpak || true
}

do_build() {
    echo "🚧 Starting Flatpak build..."
    setup_flatpak_environment
    prepare_build_directories
    build_flatpak_package
    create_flatpak_bundle
}

show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --help     Show this help message"
    echo "  --check    Only check prerequisites"
    echo "  --build    Build Flatpak bundle inside container"
    echo ""
    echo "For production builds, use GitHub Actions CI/CD:"
    echo "  - Push to main/develop branch"
    echo "  - Create a pull request"
    echo "  - Create a release"
    echo ""
    echo "The CI/CD pipeline will:"
    echo "  - Build whisper.cpp automatically"
    echo "  - Download the Whisper model"
    echo "  - Fix SHA256 hashes automatically"
    echo "  - Test the flatpak package"
    echo "  - Upload artifacts and releases"
}

parse_arguments() {
    MODE="default"
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --help)
                show_help
                exit 0
                ;;
            --check)
                MODE="check"
                shift
                ;;
            --build)
                MODE="build"
                shift
                ;;
            *)
                echo "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

run_development_check() {
    echo "🚀 Starting development build check..."
    check_prerequisites
    check_whisper_cpp
    check_whisper_model
    echo ""
    echo "✅ All prerequisites are met!"
    echo "💡 To build flatpak:"
    echo "   1. Push your changes to GitHub"
    echo "   2. GitHub Actions will build automatically"
    echo "   3. Download artifacts from Actions tab"
    echo ""
    echo "📝 To monitor the build:"
    echo "   Visit: https://github.com/ashbuk/speak-to-ai/actions"
    echo ""
    echo "🔧 For local testing, all dependencies are ready"
}

# Main function
main() {
    print_header
    parse_arguments "$@"
    
    if [ "$MODE" = "check" ] || [ "$MODE" = "default" ]; then
        run_development_check
    fi

    if [ "$MODE" = "build" ]; then
        check_prerequisites
        do_build
    fi
}

# Run main function
main "$@" 