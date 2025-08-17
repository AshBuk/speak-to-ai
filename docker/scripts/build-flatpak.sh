#!/bin/bash
set -euo pipefail

echo "=== Docker Flatpak Builder ==="

# Build whisper.cpp libraries first
echo "Building whisper.cpp libraries..."
make whisper-libs

# Source the development environment
echo "Setting up environment..."
source bash-scripts/dev-env.sh

# Build the Go application first
echo "Building Go application..."
make build-systray

# Check prerequisites for Flatpak
echo "Checking Flatpak prerequisites..."
bash bash-scripts/build-flatpak.sh --check

echo "=== Flatpak build completed ==="
echo "Note: Actual Flatpak building is complex in containers."
echo "This validates the environment and dependencies."
echo "For production Flatpak builds, use GitHub Actions."