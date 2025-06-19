#!/bin/bash

# Speak-to-AI AppImage builder script

set -e  # Exit on error

# Configuration
APP_NAME="speak-to-ai"
APP_VERSION="0.1.0"
ARCH="x86_64"
OUTPUT_DIR="dist"

echo "Creating AppDir structure..."
# Create necessary directories
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/bin"
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/lib"
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/applications"
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/icons/hicolor/256x256/apps"
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/metainfo"
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/sources/language-models"
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/sources/core"

echo "Building ${APP_NAME}..."
go build -o "${APP_NAME}" cmd/daemon/main.go

echo "Copying files to AppDir..."
cp "${APP_NAME}" "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/bin/"
cp config.yaml "${OUTPUT_DIR}/${APP_NAME}.AppDir/"

# If sources/core directory exists
if [ -d "sources/core" ]; then
    echo "Copying core sources..."
    cp -r sources/core/* "${OUTPUT_DIR}/${APP_NAME}.AppDir/sources/core/"
else
    echo "Warning: sources/core directory not found"
fi

# Include the pre-downloaded Whisper model
if [ -f "sources/language-models/base.bin" ]; then
    echo "Including pre-downloaded Whisper model..."
    cp sources/language-models/base.bin "${OUTPUT_DIR}/${APP_NAME}.AppDir/sources/language-models/"
else
    echo "Warning: Whisper model not found at sources/language-models/base.bin"
fi

# Copy required dependencies for offline use
echo "Copying required dependencies..."

# Function to copy binary with its libraries
copy_binary_with_libs() {
    local binary_name="$1"
    local binary_path=$(which "$binary_name" 2>/dev/null || echo "")
    
    if [ -n "$binary_path" ]; then
        echo "Including $binary_name dependency..."
        cp "$binary_path" "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/bin/"
        
        # Copy necessary shared libraries
        ldd "$binary_path" 2>/dev/null | grep "=>" | awk '{print $3}' | while read -r lib; do
            if [ -f "$lib" ] && [[ "$lib" != /lib* ]] && [[ "$lib" != /usr/lib* ]]; then
                echo "  Copying library: $lib"
                cp "$lib" "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/lib/" 2>/dev/null || true
            fi
        done
    else
        echo "Warning: $binary_name not found in PATH"
    fi
}

# Copy X11 clipboard tool
copy_binary_with_libs "xclip"

# Copy Wayland clipboard tools
copy_binary_with_libs "wl-copy"
copy_binary_with_libs "wl-paste"

# Copy notification tool
copy_binary_with_libs "notify-send"

# Copy audio recording tools if available
copy_binary_with_libs "arecord"
copy_binary_with_libs "ffmpeg"

# Create AppRun script with first-launch behavior
echo "Creating enhanced AppRun script..."
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

# Verify desktop file exists
if [ -f "$DESKTOP_FILE" ]; then
    echo "Desktop file created at: $DESKTOP_FILE"
else
    echo "Error: Desktop file was not created properly at: $DESKTOP_FILE"
    exit 1
fi

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

# Create a placeholder icon
# In a real application, you would use a proper icon file
echo "Creating placeholder icon (replace with a real icon)..."
cat > "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/icons/hicolor/256x256/apps/${APP_NAME}.png" << EOF
iVBORw0KGgoAAAANSUhEUgAAAQAAAAEACAMAAABrrFhUAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAAAZQTFRF////AAAAVcLTfgAAAAF0Uk5TAEDm2GYAAAAqSURBVHja7cEBAQAAAIIg/69uSEABAAAAAAAAAAAAAAAAAAAAAAAAAHwZJsAAARqZF58AAAAASUVORK5CYII=
EOF

# Link the icon to the root directory as required by AppImage
echo "Creating icon symlink..."
ln -sf "./usr/share/icons/hicolor/256x256/apps/${APP_NAME}.png" "${OUTPUT_DIR}/${APP_NAME}.AppDir/${APP_NAME}.png"

# Download AppImage tool if not present
if [ ! -f "${OUTPUT_DIR}/appimagetool-${ARCH}.AppImage" ]; then
    echo "Downloading appimagetool..."
    wget -q "https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-${ARCH}.AppImage" -O "${OUTPUT_DIR}/appimagetool-${ARCH}.AppImage"
    chmod +x "${OUTPUT_DIR}/appimagetool-${ARCH}.AppImage"
fi

# Verify AppDir setup
echo "Verifying AppDir structure..."
ls -la "${OUTPUT_DIR}/${APP_NAME}.AppDir/"
echo "Root files:"
ls -la "${OUTPUT_DIR}/${APP_NAME}.AppDir/"
echo "Desktop file:"
ls -la "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/applications/"
echo "Icon:"
ls -la "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/icons/hicolor/256x256/apps/"

# Build AppImage, bypassing the AppStream validation
echo "Building AppImage..."
cd "${OUTPUT_DIR}"

# Set environment variables to bypass validation
export ARCH="${ARCH}"  # Export explicitly
export NO_APPSTREAM_VALIDATE=1
export DISABLE_APPIMAGE_EXTRACT=1
export VERSION="${APP_VERSION}"

echo "Attempting to build AppImage with architecture: ${ARCH}"

# Try to build AppImage directly
if ./appimagetool-${ARCH}.AppImage --no-appstream "${APP_NAME}.AppDir"; then
    # Find the AppImage that was created
    APPIMAGE_FILE=$(find . -name "*.AppImage" ! -name "appimagetool*" -type f -print | head -n 1)
    
    if [ -n "$APPIMAGE_FILE" ]; then
        chmod +x "$APPIMAGE_FILE"
        echo "AppImage created successfully: $APPIMAGE_FILE"
        ls -lh "$APPIMAGE_FILE"
        echo "Done! AppImage is ready for use."
        exit 0
    else
        echo "Warning: AppImage was built but could not be found. Please check the 'dist' directory."
        exit 1
    fi
else
    echo "AppImage creation failed. Installation may be incomplete."
    exit 1
fi 