#!/bin/bash

# Speak-to-AI Flatpak builder script

set -e  # Exit on error

# Configuration
APP_ID="io.github.ashbuk.speak-to-ai"
OUTPUT_DIR="dist"
FLATPAK_DIR="${OUTPUT_DIR}/flatpak"

echo "=== Speak-to-AI Flatpak Builder ==="

# Check if flatpak-builder is available
if ! command -v flatpak-builder &> /dev/null; then
    echo "Error: flatpak-builder is not installed"
    echo "Install it with: sudo dnf install flatpak-builder"
    exit 1
fi

# Check required runtime
if ! flatpak list --runtime | grep -q "org.freedesktop.Platform.*23.08"; then
    echo "Installing required Flatpak runtime..."
    flatpak install -y flathub org.freedesktop.Platform//23.08 org.freedesktop.Sdk//23.08
fi

# Check SDK extension
if ! flatpak list | grep -q "org.freedesktop.Sdk.Extension.golang"; then
    echo "Installing Golang SDK extension..."
    flatpak install -y flathub org.freedesktop.Sdk.Extension.golang//23.08
fi

# Create output directory
mkdir -p "${FLATPAK_DIR}"

echo "Building Flatpak package..."

# Build the Flatpak
flatpak-builder \
    --force-clean \
    --sandbox \
    --user \
    --install-deps-from=flathub \
    --ccache \
    --mirror-screenshots-url=https://dl.flathub.org/media/ \
    --repo="${FLATPAK_DIR}/repo" \
    "${FLATPAK_DIR}/build-dir" \
    "${APP_ID}.json"

echo "Creating Flatpak bundle..."

# Create bundle (.flatpak file)
flatpak build-bundle \
    "${FLATPAK_DIR}/repo" \
    "${OUTPUT_DIR}/${APP_ID}.flatpak" \
    "${APP_ID}"

echo "✅ Flatpak build completed successfully!"
echo "📦 Bundle created: ${OUTPUT_DIR}/${APP_ID}.flatpak"
echo ""
echo "To install locally: flatpak install ${OUTPUT_DIR}/${APP_ID}.flatpak"
echo "To run: flatpak run ${APP_ID}" 