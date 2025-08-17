#!/bin/bash
set -euo pipefail

echo "=== Docker AppImage Builder ==="

# Build whisper.cpp libraries first
echo "Building whisper.cpp libraries..."
make whisper-libs

# Source the development environment
echo "Setting up environment..."
source bash-scripts/dev-env.sh

# Run the AppImage build script
echo "Building AppImage..."
bash bash-scripts/build-appimage.sh

echo "=== AppImage build completed ==="
echo "Output files:"
find dist/ -name "*.AppImage" -exec ls -lh {} \;