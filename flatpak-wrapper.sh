#!/bin/bash

# Flatpak wrapper script for Speak-to-AI
# This script runs before the main application to check permissions and provide guidance

echo "=== Speak-to-AI Flatpak Wrapper ==="

# Check for input device permissions
if [ ! -r /dev/input/event0 ] 2>/dev/null; then
    echo "⚠️  Warning: Limited access to input devices detected."
    echo "   Hotkeys may not work properly in Flatpak environment."
    echo ""
    echo "💡 To enable full hotkey support:"
    echo "   1. Install the AppImage version for better hardware access"
    echo "   2. Or grant additional permissions to this Flatpak:"
    echo "      flatpak override --user --device=input io.github.ashbuk.speak-to-ai"
    echo ""
    echo "🔄 Continuing with limited functionality..."
    echo ""
fi

# Check if system tray is available
if [ -z "$XDG_CURRENT_DESKTOP" ]; then
    echo "⚠️  Warning: Desktop environment not detected."
    echo "   System tray may not be available."
    echo ""
fi

# Set up Flatpak-specific environment
export FLATPAK_ID="io.github.ashbuk.speak-to-ai"

# Create user config directory in Flatpak data location
FLATPAK_CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.var/app/io.github.ashbuk.speak-to-ai/config}/speak-to-ai"
mkdir -p "${FLATPAK_CONFIG_DIR}/language-models"

# Copy default config if not exists
if [ ! -f "${FLATPAK_CONFIG_DIR}/config.yaml" ]; then
    echo "🔧 Setting up first-time configuration..."
    cp "/app/share/speak-to-ai/config.yaml" "${FLATPAK_CONFIG_DIR}/config.yaml"
    
    # Update config to point to user model location
    sed -i "s|sources/language-models/base.bin|${FLATPAK_CONFIG_DIR}/language-models/base.bin|g" "${FLATPAK_CONFIG_DIR}/config.yaml"
fi

# Copy model to user directory if not exists
if [ ! -f "${FLATPAK_CONFIG_DIR}/language-models/base.bin" ]; then
    echo "📥 Copying Whisper model to user directory..."
    if [ -f "/app/share/speak-to-ai/models/base.bin" ]; then
        cp "/app/share/speak-to-ai/models/base.bin" "${FLATPAK_CONFIG_DIR}/language-models/base.bin"
        echo "✅ Model copied successfully"
    else
        echo "❌ Error: Model not found in Flatpak installation"
    fi
fi

echo "🚀 Starting Speak-to-AI..."
echo ""

# Run the main application with user config
exec /app/bin/speak-to-ai --config "${FLATPAK_CONFIG_DIR}/config.yaml" "$@" 