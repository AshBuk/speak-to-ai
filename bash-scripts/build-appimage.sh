#!/bin/bash

# Speak-to-AI AppImage builder script

set -e  # Exit on error
set -x  # Show commands being executed

# Configuration
APP_NAME="speak-to-ai"
# Prefer explicit env var, then CI tag (GITHUB_REF_NAME), then latest git tag, else date
APP_VERSION="${APP_VERSION:-${GITHUB_REF_NAME:-$(git describe --tags --abbrev=0 2>/dev/null || date +%Y%m%d)}}"
ARCH="x86_64"
OUTPUT_DIR="dist"

echo "=== Starting AppImage build for ${APP_NAME} v${APP_VERSION} ==="

# Ensure whisper.cpp libraries are built and CGO is configured
echo "Preparing whisper.cpp libraries..."
# Try to source local dev env (non-fatal) to set CGO vars if available
if [ -f "bash-scripts/dev-env.sh" ]; then
    # shellcheck source=/dev/null
    source bash-scripts/dev-env.sh || true
fi
# Build libs only if headers not present (idempotent)
if [ ! -f "lib/whisper.h" ]; then
    make whisper-libs
fi

# Create necessary directories
echo "Creating AppDir structure..."
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/bin"
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/lib"
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/applications"
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/icons/hicolor/256x256/apps"
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/metainfo"
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/sources/language-models"
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/sources/core"

echo "Building ${APP_NAME} with systray support..."
# Always rebuild to ensure correct tags and versioned metadata
go build -tags systray -o "${APP_NAME}" cmd/daemon/main.go

echo "Copying main application..."
cp "${APP_NAME}" "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/bin/"
cp config.yaml "${OUTPUT_DIR}/${APP_NAME}.AppDir/"

# Bundle required shared libraries for whisper.cpp runtime
echo "Bundling whisper shared libraries..."
LIB_DST="${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/lib"
mkdir -p "$LIB_DST"
# Preserve symlinks (e.g., libwhisper.so -> libwhisper.so.1)
if compgen -G "lib/libwhisper.so*" > /dev/null; then
    cp -a lib/libwhisper.so* "$LIB_DST/" || true
fi
if compgen -G "lib/libggml*.so*" > /dev/null; then
    cp -a lib/libggml*.so* "$LIB_DST/" || true
fi
echo "Bundled libs:"
ls -la "$LIB_DST" || true

# Bundle system tray related libraries explicitly (for AppImage runtime)
echo "Bundling system tray libraries..."
LIB_ARGS=""
find_syslib() {
    local name="$1"
    for p in \
        /usr/lib/x86_64-linux-gnu \
        /lib/x86_64-linux-gnu \
        /usr/lib64 \
        /usr/lib; do
        if compgen -G "${p}/${name}" > /dev/null; then
            echo "${p}/${name}"
            return 0
        fi
    done
    return 1
}

add_lib() {
    local pattern="$1"
    local found
    found=$(find_syslib "$pattern") || { echo "Warning: not found: $pattern"; return 0; }
    echo "Including $(basename "$found") from $(dirname "$found")"
    cp -a "$found" "$LIB_DST/" || true
    LIB_ARGS+=" --library \"$found\""
}

# Primary tray libs (exact soversion names typical on Ubuntu)
add_lib "libayatana-appindicator3.so*"
add_lib "libayatana-indicator3.so*"
add_lib "libdbusmenu-gtk3.so*"
add_lib "libdbusmenu-glib.so*"

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
copy_if_exists "wtype"
copy_if_exists "ydotool"
copy_if_exists "notify-send"
copy_if_exists "arecord"
copy_if_exists "xdotool"

# Note: linuxdeploy will automatically handle library dependencies
# Manual library copying is only needed as fallback
echo "Libraries will be handled by linuxdeploy automatically..."

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

# Copy real application icon
echo "Copying application icon..."
if [ -f "icons/io.github.ashbuk.speak-to-ai.png" ]; then
    cp "icons/io.github.ashbuk.speak-to-ai.png" "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/icons/hicolor/256x256/apps/${APP_NAME}.png"
    echo "Real icon copied successfully"
else
    echo "Warning: Real icon not found, creating placeholder..."
    # Create a minimal PNG icon as fallback
    echo "iVBORw0KGgoAAAANSUhEUgAAAQAAAAEACAMAAABrrFhUAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAAAZQTFRF////AAAAVcLTfgAAAAF0Uk5TAEDm2GYAAAAqSURBVHja7cEBAQAAAIIg/69uSEABAAAAAAAAAAAAAAAAAAAAAAAAAHwZJsAAARqZF58AAAAASUVORK5CYII=" | base64 -d > "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/icons/hicolor/256x256/apps/${APP_NAME}.png"
fi

# Link the icon to the root directory as required by AppImage
echo "Creating icon symlink..."
ln -sf "./usr/share/icons/hicolor/256x256/apps/${APP_NAME}.png" "${OUTPUT_DIR}/${APP_NAME}.AppDir/${APP_NAME}.png"

# Download AppImage tools if not present (to a separate tools directory)
TOOLS_DIR="$(pwd)/tools"
mkdir -p "${TOOLS_DIR}"

if [ ! -f "${TOOLS_DIR}/linuxdeploy-${ARCH}.AppImage" ]; then
    echo "Downloading linuxdeploy..."
    wget -q "https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-${ARCH}.AppImage" -O "${TOOLS_DIR}/linuxdeploy-${ARCH}.AppImage"
    chmod +x "${TOOLS_DIR}/linuxdeploy-${ARCH}.AppImage"
fi

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

# Use linuxdeploy to automatically include dependencies
echo "Using linuxdeploy to gather dependencies..."
cd "${OUTPUT_DIR}"

# Set environment variables
export ARCH="${ARCH}"
export VERSION="${APP_VERSION}"

echo "Running linuxdeploy to prepare AppDir with dependencies..."

# Use linuxdeploy to automatically copy libraries and dependencies
if "${TOOLS_DIR}/linuxdeploy-${ARCH}.AppImage" --appimage-extract-and-run \
    --appdir "${APP_NAME}.AppDir" \
    --executable "${APP_NAME}.AppDir/usr/bin/${APP_NAME}" \
    --executable "${APP_NAME}.AppDir/usr/bin/xdotool" \
    --executable "${APP_NAME}.AppDir/usr/bin/wtype" \
    --executable "${APP_NAME}.AppDir/usr/bin/ydotool" \
    --executable "${APP_NAME}.AppDir/usr/bin/wl-copy" \
    --executable "${APP_NAME}.AppDir/usr/bin/xclip" \
    --executable "${APP_NAME}.AppDir/usr/bin/arecord" \
    --executable "${APP_NAME}.AppDir/usr/bin/notify-send" \
    ${LIB_ARGS} \
    --desktop-file "${APP_NAME}.AppDir/${APP_NAME}.desktop" \
    --icon-file "${APP_NAME}.AppDir/${APP_NAME}.png" \
    --output appimage; then
    
    # Find the AppImage that was created
    APPIMAGE_FILE=$(find . -name "*.AppImage" ! -name "*tool*" -type f -print | head -n 1)
    
    if [ -n "$APPIMAGE_FILE" ]; then
        chmod +x "$APPIMAGE_FILE"
        # Rename to include version for clarity and distribution (unified naming)
        TARGET_NAME="speak-to-ai-${APP_VERSION}.AppImage"
        mv -f "$APPIMAGE_FILE" "$TARGET_NAME"
        echo "AppImage created successfully with linuxdeploy: $TARGET_NAME"
        ls -lh "$TARGET_NAME"
        echo "=== AppImage build completed successfully! ==="
        exit 0
    else
        echo "Warning: linuxdeploy completed but AppImage not found, trying manual approach..."
    fi
fi

echo "Linuxdeploy failed or didn't produce AppImage, falling back to manual appimagetool..."

# Try to build AppImage directly (with FUSE workaround for CI)
if "${TOOLS_DIR}/appimagetool-${ARCH}.AppImage" --appimage-extract-and-run --no-appstream "${APP_NAME}.AppDir"; then
    # Find the AppImage that was created
    APPIMAGE_FILE=$(find . -name "*.AppImage" ! -name "appimagetool*" -type f -print | head -n 1)
    
    if [ -n "$APPIMAGE_FILE" ]; then
        chmod +x "$APPIMAGE_FILE"
        TARGET_NAME="speak-to-ai-${APP_VERSION}.AppImage"
        mv -f "$APPIMAGE_FILE" "$TARGET_NAME"
        echo "AppImage created successfully: $TARGET_NAME"
        ls -lh "$TARGET_NAME"
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
            TARGET_NAME="speak-to-ai-${APP_VERSION}.AppImage"
            mv -f "$APPIMAGE_FILE" "$TARGET_NAME"
            echo "AppImage created successfully: $TARGET_NAME"
            ls -lh "$TARGET_NAME"
            echo "=== AppImage build completed successfully! ==="
            exit 0
        else
            echo "Error: AppImage was built but could not be found."
            ls -la
            exit 1
        fi
    else
        echo "Error: All AppImage creation methods failed."
        exit 1
    fi
fi 