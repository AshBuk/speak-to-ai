#!/bin/bash

# Speak-to-AI AppImage builder script

set -e  # Exit on error
set -x  # Show commands being executed

# Configuration
APP_NAME="speak-to-ai"
APP_VERSION="0.2.5"
ARCH="x86_64"
OUTPUT_DIR="dist"

echo "=== Starting AppImage build for ${APP_NAME} v${APP_VERSION} ==="

# Create necessary directories
echo "Creating AppDir structure..."
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/bin"
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/lib"
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/applications"
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/icons/hicolor/256x256/apps"
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/metainfo"
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/sources/language-models"
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/sources/core"

echo "Building ${APP_NAME}..."
if [ ! -f "${APP_NAME}" ]; then
    go build -tags systray -o "${APP_NAME}" cmd/daemon/main.go
fi

echo "Copying main application..."
cp "${APP_NAME}" "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/bin/"
cp config.yaml "${OUTPUT_DIR}/${APP_NAME}.AppDir/"

# Copy core sources if they exist
if [ -d "sources/core" ] && [ -f "sources/core/whisper" ]; then
    echo "Copying whisper binaries..."
    cp sources/core/whisper "${OUTPUT_DIR}/${APP_NAME}.AppDir/sources/core/"
    cp sources/core/quantize "${OUTPUT_DIR}/${APP_NAME}.AppDir/sources/core/" || true
else
    echo "Warning: Whisper binaries not found in sources/core"
fi

# Include the pre-downloaded Whisper model
if [ -f "sources/language-models/base.bin" ]; then
    echo "Including pre-downloaded Whisper model..."
    cp sources/language-models/base.bin "${OUTPUT_DIR}/${APP_NAME}.AppDir/sources/language-models/"
else
    echo "Warning: Whisper model not found at sources/language-models/base.bin"
fi

# Copy system dependencies
echo "Copying system dependencies..."
copy_if_exists() {
    local binary_name="$1"
    local binary_path=$(which "$binary_name" 2>/dev/null || echo "")
    
    if [ -n "$binary_path" ]; then
        echo "Including $binary_name dependency..."
        cp "$binary_path" "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/bin/"
    else
        echo "Warning: $binary_name not found in PATH"
    fi
}

copy_if_exists "xclip"
copy_if_exists "wl-copy"
copy_if_exists "wl-paste"
copy_if_exists "notify-send"
copy_if_exists "arecord"

# Create AppRun script with first-launch behavior
echo "Creating AppRun script..."
cat > "${OUTPUT_DIR}/${APP_NAME}.AppDir/AppRun" << 'EOF'
#!/bin/bash
SELF=$(readlink -f "$0")
HERE=${SELF%/*}
export PATH="${HERE}/usr/bin:${PATH}"
export LD_LIBRARY_PATH="${HERE}/usr/lib:${LD_LIBRARY_PATH}"

# Create user config directory
CONFIG_DIR="${HOME}/.config/speak-to-ai"
mkdir -p "${CONFIG_DIR}"
mkdir -p "${CONFIG_DIR}/language-models"

# First launch checks
if [ ! -f "${CONFIG_DIR}/config.yaml" ]; then
    echo "First launch detected: Setting up configuration..."
    cp "${HERE}/config.yaml" "${CONFIG_DIR}/config.yaml"
    
    # Update the config to point to the correct model path
    sed -i "s|sources/language-models/base.bin|${CONFIG_DIR}/language-models/base.bin|g" "${CONFIG_DIR}/config.yaml"
fi

# Check if model exists in user directory, copy if not
if [ ! -f "${CONFIG_DIR}/language-models/base.bin" ]; then
    echo "Copying Whisper model to user directory..."
    if [ -f "${HERE}/sources/language-models/base.bin" ]; then
        cp "${HERE}/sources/language-models/base.bin" "${CONFIG_DIR}/language-models/base.bin"
    else
        echo "Warning: Model not found in AppImage. Please download it manually."
    fi
fi

# Check for input device permissions
if [ ! -r /dev/input/event0 ] 2>/dev/null; then
    echo "Warning: No access to input devices. Hotkeys may not work."
    echo "To enable hotkeys, please run:"
    echo "  sudo usermod -a -G input \$USER"
    echo "  sudo udevadm control --reload-rules"
    echo "  sudo udevadm trigger"
    echo "Then log out and log back in."
    echo ""
fi

# Run the application with user config
cd "${HERE}"
exec "${HERE}/usr/bin/speak-to-ai" --config "${CONFIG_DIR}/config.yaml" "$@"
EOF
chmod +x "${OUTPUT_DIR}/${APP_NAME}.AppDir/AppRun"

# Create desktop file
echo "Creating desktop file..."
DESKTOP_FILE="${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/applications/${APP_NAME}.desktop"
cat > "$DESKTOP_FILE" << EOF
[Desktop Entry]
Name=Speak-to-AI
Comment=Offline speech-to-text for AI assistants
Exec=speak-to-ai
Icon=speak-to-ai
Type=Application
Categories=Utility;
Terminal=false
StartupNotify=true
X-AppImage-Name=Speak-to-AI
X-AppImage-Version=${APP_VERSION}
X-AppImage-Arch=${ARCH}
EOF

# Create symlink for the desktop file
echo "Creating desktop file symlink..."
ln -sf "./usr/share/applications/${APP_NAME}.desktop" "${OUTPUT_DIR}/${APP_NAME}.AppDir/${APP_NAME}.desktop"

# Create AppStream metadata
echo "Creating AppStream metadata..."
cat > "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/metainfo/${APP_NAME}.appdata.xml" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<component type="desktop-application">
  <id>io.github.ashbuk.speak-to-ai</id>
  <metadata_license>MIT</metadata_license>
  <project_license>MIT</project_license>
  <name>Speak-to-AI</name>
  <summary>Offline speech-to-text for AI assistants</summary>
  <description>
    <p>A minimalist, offline desktop application that enables voice input for AI assistants without sending your voice to the cloud. Uses the Whisper model locally for speech recognition.</p>
  </description>
  <url type="homepage">https://github.com/AshBuk/speak-to-ai</url>
  <developer_name>Asher Buk</developer_name>
  <launchable type="desktop-id">speak-to-ai.desktop</launchable>
  <releases>
    <release version="${APP_VERSION}" date="$(date +%Y-%m-%d)"/>
  </releases>
  <provides>
    <binary>speak-to-ai</binary>
  </provides>
  <content_rating type="oars-1.1"/>
</component>
EOF

# Create a simple icon (placeholder)
echo "Creating placeholder icon..."
# Create a minimal PNG icon
echo "iVBORw0KGgoAAAANSUhEUgAAAQAAAAEACAMAAABrrFhUAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAAAZQTFRF////AAAAVcLTfgAAAAF0Uk5TAEDm2GYAAAAqSURBVHja7cEBAQAAAIIg/69uSEABAAAAAAAAAAAAAAAAAAAAAAAAAHwZJsAAARqZF58AAAAASUVORK5CYII=" | base64 -d > "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/icons/hicolor/256x256/apps/${APP_NAME}.png"

# Link the icon to the root directory as required by AppImage
echo "Creating icon symlink..."
ln -sf "./usr/share/icons/hicolor/256x256/apps/${APP_NAME}.png" "${OUTPUT_DIR}/${APP_NAME}.AppDir/${APP_NAME}.png"

# Download AppImage tool if not present (to a separate tools directory)
TOOLS_DIR="$(pwd)/tools"
mkdir -p "${TOOLS_DIR}"
if [ ! -f "${TOOLS_DIR}/appimagetool-${ARCH}.AppImage" ]; then
    echo "Downloading appimagetool..."
    wget -q "https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-${ARCH}.AppImage" -O "${TOOLS_DIR}/appimagetool-${ARCH}.AppImage"
    chmod +x "${TOOLS_DIR}/appimagetool-${ARCH}.AppImage"
fi

# Verify AppDir setup
echo "Verifying AppDir structure..."
echo "Root files:"
ls -la "${OUTPUT_DIR}/${APP_NAME}.AppDir/"
echo "Binary files:"
ls -la "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/bin/"

# Build AppImage
echo "Building AppImage..."
cd "${OUTPUT_DIR}"

# Set environment variables
export ARCH="${ARCH}"
export VERSION="${APP_VERSION}"

echo "Attempting to build AppImage with architecture: ${ARCH}"
echo "Tools directory: ${TOOLS_DIR}"
echo "Looking for appimagetool at: ${TOOLS_DIR}/appimagetool-${ARCH}.AppImage"
ls -la "${TOOLS_DIR}/" || echo "Tools directory not found"

# Try to build AppImage directly (with FUSE workaround for CI)
if "${TOOLS_DIR}/appimagetool-${ARCH}.AppImage" --appimage-extract-and-run --no-appstream "${APP_NAME}.AppDir"; then
    # Find the AppImage that was created
    APPIMAGE_FILE=$(find . -name "*.AppImage" ! -name "appimagetool*" -type f -print | head -n 1)
    
    if [ -n "$APPIMAGE_FILE" ]; then
        chmod +x "$APPIMAGE_FILE"
        echo "AppImage created successfully: $APPIMAGE_FILE"
        ls -lh "$APPIMAGE_FILE"
        echo "=== AppImage build completed successfully! ==="
        exit 0
    else
        echo "Error: AppImage was built but could not be found."
        ls -la
        exit 1
    fi
else
    echo "Primary method failed, trying alternative approach..."
    # Alternative: extract appimagetool and run directly
    if [ ! -d "appimagetool-extracted" ]; then
        echo "Extracting appimagetool..."
        "${TOOLS_DIR}/appimagetool-${ARCH}.AppImage" --appimage-extract
        mv squashfs-root appimagetool-extracted
    fi
    
    echo "Using extracted appimagetool..."
    if ./appimagetool-extracted/AppRun --no-appstream "${APP_NAME}.AppDir"; then
        # Find the AppImage that was created
        APPIMAGE_FILE=$(find . -name "*.AppImage" ! -name "appimagetool*" -type f -print | head -n 1)
        
        if [ -n "$APPIMAGE_FILE" ]; then
            chmod +x "$APPIMAGE_FILE"
            echo "AppImage created successfully: $APPIMAGE_FILE"
            ls -lh "$APPIMAGE_FILE"
            echo "=== AppImage build completed successfully! ==="
            exit 0
        else
            echo "Error: AppImage was built but could not be found."
            ls -la
            exit 1
        fi
    else
        echo "Error: Both AppImage creation methods failed."
        exit 1
    fi
fi 