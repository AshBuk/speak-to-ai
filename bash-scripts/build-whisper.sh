#!/bin/bash

# Script to build whisper.cpp library for Go bindings
set -e

echo "Building whisper.cpp library..."

# Create build directory
mkdir -p build
cd build

# Clone whisper.cpp if not exists
if [ ! -d "whisper.cpp" ]; then
    echo "Cloning whisper.cpp..."
    git clone https://github.com/ggerganov/whisper.cpp.git
fi

cd whisper.cpp

# Build the library
echo "Building libwhisper.a..."
make clean
make libwhisper.a

# Create symlink for easy access
echo "Setting up library paths..."
cd ../..
mkdir -p lib
cp build/whisper.cpp/libwhisper.a lib/
cp build/whisper.cpp/whisper.h lib/

echo "Build complete!"
echo "Library files:"
echo "  - lib/libwhisper.a"
echo "  - lib/whisper.h"
echo ""
echo "To use with Go:"
echo "export C_INCLUDE_PATH=$(pwd)/lib"
echo "export LIBRARY_PATH=$(pwd)/lib"
echo "export CGO_LDFLAGS=\"-L$(pwd)/lib -lwhisper\""