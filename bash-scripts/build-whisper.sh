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
# Clean previous build
rm -rf build
# Build using CMake (modern whisper.cpp uses CMake)
cmake -B build
cmake --build build --config Release

# Create symlink for easy access
echo "Setting up library paths..."
cd ../..
mkdir -p lib
# Copy shared library (modern whisper.cpp builds .so instead of .a)
cp build/whisper.cpp/build/src/libwhisper.so* lib/
cp build/whisper.cpp/include/whisper.h lib/

echo "Build complete!"
echo "Library files:"
ls -la lib/
echo ""
echo "To use with Go:"
echo "export C_INCLUDE_PATH=$(pwd)/lib"
echo "export LIBRARY_PATH=$(pwd)/lib"
echo "export LD_LIBRARY_PATH=$(pwd)/lib"
echo "export CGO_LDFLAGS=\"-L$(pwd)/lib -lwhisper\""