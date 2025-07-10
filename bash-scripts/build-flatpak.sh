#!/bin/bash

# Speak-to-AI Flatpak development helper script
# For production builds, use GitHub Actions CI/CD pipeline

set -e  # Exit on error

# Configuration
APP_ID="io.github.ashbuk.speak-to-ai"

echo "=== Speak-to-AI Flatpak Development Helper ==="
echo "⚠️  This script is for local development only"
echo "⚠️  Production builds should use GitHub Actions CI/CD"
echo ""

# Function to check prerequisites
check_prerequisites() {
    echo "Checking prerequisites..."
    
    if ! command -v flatpak-builder &> /dev/null; then
        echo "❌ flatpak-builder is not installed"
        echo "💡 For production builds, use GitHub Actions instead"
        echo "💡 To install locally: sudo dnf install flatpak-builder"
        exit 1
    fi
    
    if ! flatpak list --runtime | grep -q "org.freedesktop.Platform.*23.08"; then
        echo "❌ Required Flatpak runtime not found"
        echo "💡 For production builds, use GitHub Actions instead"
        echo "💡 To install locally: flatpak install flathub org.freedesktop.Platform//23.08"
        exit 1
    fi
    
    echo "✅ Prerequisites check passed"
}

# Function to check whisper.cpp
check_whisper_cpp() {
    echo "Checking whisper.cpp dependencies..."
    
    if [ ! -f "sources/core/whisper" ] || [ ! -f "sources/core/quantize" ]; then
        echo "❌ whisper.cpp not found"
        echo "💡 Run whisper.cpp build first or let CI/CD handle it"
        exit 1
    fi
    
    echo "✅ whisper.cpp available"
}

# Function to check model
check_whisper_model() {
    echo "Checking Whisper model..."
    
    if [ ! -f "sources/language-models/base.bin" ]; then
        echo "❌ Whisper model not found"
        echo "💡 Download the model first or let CI/CD handle it"
        exit 1
    fi
    
    echo "✅ Whisper model available"
}

# Function to show help
show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --help     Show this help message"
    echo "  --check    Only check prerequisites"
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

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --help)
            show_help
            exit 0
            ;;
        --check)
            check_prerequisites
            check_whisper_cpp
            check_whisper_model
            echo "✅ All checks passed. Ready for CI/CD build."
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
    shift
done

# Main function
main() {
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

# Run main function
main "$@" 