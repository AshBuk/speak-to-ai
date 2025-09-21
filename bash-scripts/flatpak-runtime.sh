#!/bin/bash

# Flatpak wrapper script for Speak-to-AI
# This script runs before the main application to set up environment and check configuration

echo "=== Speak-to-AI Flatpak Wrapper ==="

# Set up Flatpak-specific environment variables
export FLATPAK_ID="io.github.ashbuk.speak-to-ai"

# Configure library and binary paths for libayatana-appindicator
export LD_LIBRARY_PATH="/app/lib:${LD_LIBRARY_PATH}"
export PATH="/app/bin:${PATH}"

# Set up XDG directories
export XDG_DATA_HOME="${XDG_DATA_HOME:-$HOME/.local/share}"
export XDG_CONFIG_HOME="${XDG_CONFIG_HOME:-$HOME/.config}"

# Flatpak config directory
FLATPAK_CONFIG_DIR="${XDG_CONFIG_HOME}/speak-to-ai"

# Function to check permissions
check_permissions() {
    local has_warnings=false
    
    # Check for input device permissions
    if [ ! -r /dev/input/event0 ] 2>/dev/null; then
        echo "⚠️  Warning: Limited access to input devices detected."
        echo "   Hotkeys may not work properly in Flatpak environment."
        echo ""
        echo "💡 To enable full hotkey support:"
        echo "   flatpak override --user --device=input io.github.ashbuk.speak-to-ai"
        echo ""
        has_warnings=true
    fi
    
    # Check if system tray is available
    if [ -z "$XDG_CURRENT_DESKTOP" ]; then
        echo "⚠️  Warning: Desktop environment not detected."
        echo "   System tray may not be available."
        echo ""
        has_warnings=true
    fi
    
    # Check for indicator services
    if ! pgrep -f "indicator" > /dev/null 2>&1; then
        echo "⚠️  Warning: No indicator services detected."
        echo "   System tray integration may not work properly."
        echo ""
        has_warnings=true
    fi
    
    if [ "$has_warnings" = true ]; then
        echo "🔄 Continuing with potentially limited functionality..."
        echo ""
    fi
}

# Function to set up configuration on first launch
setup_configuration() {
    echo "🔧 Setting up configuration..."
    
    # Create configuration directory structure
    mkdir -p "${FLATPAK_CONFIG_DIR}/models"
    mkdir -p "${FLATPAK_CONFIG_DIR}/logs"
    
    # Copy default config if not exists
    if [ ! -f "${FLATPAK_CONFIG_DIR}/config.yaml" ]; then
        echo "📋 Creating default configuration..."
        cp "/app/share/speak-to-ai/config.yaml" "${FLATPAK_CONFIG_DIR}/config.yaml"
        
        # Update config paths for Flatpak environment
        sed -i "s|sources/language-models/small-q5_1.bin|${FLATPAK_CONFIG_DIR}/models/small-q5_1.bin|g" "${FLATPAK_CONFIG_DIR}/config.yaml"
        # Note: whisper library is used via Go bindings, not CLI binary
        
        echo "✅ Configuration created successfully"
    else
        echo "✅ Configuration already exists"
    fi
}

# Function to set up Whisper model on first launch
setup_whisper_model() {
    echo "📥 Setting up Whisper model..."
    
    # Copy model to user directory if not exists
    if [ ! -f "${FLATPAK_CONFIG_DIR}/models/small-q5_1.bin" ]; then
        echo "📦 Copying Whisper model to user directory..."
        if [ -f "/app/share/speak-to-ai/models/small-q5_1.bin" ]; then
            cp "/app/share/speak-to-ai/models/small-q5_1.bin" "${FLATPAK_CONFIG_DIR}/models/small-q5_1.bin"
            echo "✅ Model copied successfully"
        else
            echo "❌ Error: Model not found in Flatpak installation"
            echo "   Please check the Flatpak build includes the model"
            return 1
        fi
    else
        echo "✅ Model already exists"
    fi
}

# Function to verify installation
verify_installation() {
    echo "🔍 Verifying installation..."
    
    local has_errors=false
    
    # Check main binary
    if [ ! -f "/app/bin/speak-to-ai" ]; then
        echo "❌ Main binary not found"
        has_errors=true
    fi
    
    # Check whisper library (we use library bindings, not CLI binary)
    if [ ! -f "/app/lib/libwhisper.so" ] && ! ls /app/lib/libwhisper.so* >/dev/null 2>&1; then
        echo "❌ Whisper library not found"
        has_errors=true
    fi
    
    # Check model
    if [ ! -f "${FLATPAK_CONFIG_DIR}/models/small-q5_1.bin" ]; then
        echo "❌ Whisper model not found"
        has_errors=true
    fi
    
    # Check config
    if [ ! -f "${FLATPAK_CONFIG_DIR}/config.yaml" ]; then
        echo "❌ Configuration file not found"
        has_errors=true
    fi
    
    if [ "$has_errors" = true ]; then
        echo "❌ Installation verification failed"
        return 1
    else
        echo "✅ Installation verified successfully"
    fi
}

# Main initialization
echo "🚀 Initializing Speak-to-AI..."

# Check permissions and show warnings
check_permissions

# Set up configuration and model
setup_configuration
setup_whisper_model

# Verify installation
if ! verify_installation; then
    echo "❌ Installation verification failed. Exiting."
    exit 1
fi

echo "✅ Initialization completed successfully"
echo "🎤 Starting Speak-to-AI application..."
echo ""

# Run the main application with user config
exec /app/bin/speak-to-ai --config "${FLATPAK_CONFIG_DIR}/config.yaml" "$@" 