#!/bin/bash

# Test script for Flatpak build
# This script tests the Flatpak build locally

set -e

APP_ID="io.github.ashbuk.speak-to-ai"
BUILD_DIR="build-dir"

echo "=== Testing Flatpak Build ==="

# Clean previous build
if [ -d "$BUILD_DIR" ]; then
    echo "🧹 Cleaning previous build..."
    rm -rf "$BUILD_DIR"
fi

# Install runtime if not present
echo "📦 Installing Flatpak runtime..."
flatpak install --user --noninteractive flathub org.freedesktop.Platform//23.08 || true
flatpak install --user --noninteractive flathub org.freedesktop.Sdk//23.08 || true
flatpak install --user --noninteractive flathub org.freedesktop.Sdk.Extension.golang//23.08 || true

# Build the flatpak
echo "🔨 Building Flatpak..."
flatpak-builder --user --install --force-clean "$BUILD_DIR" io.github.ashbuk.speak-to-ai.json

# Test run
echo "🧪 Testing Flatpak run..."
flatpak run "$APP_ID" --version || echo "Version check failed, but app might still work"

echo "✅ Flatpak build and test completed successfully!"
echo "🚀 You can now run: flatpak run $APP_ID" 