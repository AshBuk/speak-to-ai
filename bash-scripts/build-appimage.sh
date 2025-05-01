#!/bin/bash

# Speak-to-AI AppImage builder script

set -e  # Exit on error

# Configuration
APP_NAME="speak-to-ai"
APP_VERSION="0.1.0"
ARCH="x86_64"
OUTPUT_DIR="dist"

# Create necessary directories
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/bin"
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/applications"
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/icons/hicolor/256x256/apps"
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/metainfo"
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/sources/language-models"
mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/sources/core"

echo "Building ${APP_NAME}..."
go build -o "${APP_NAME}" cmd/daemon/*.go

echo "Copying files to AppDir..."
cp "${APP_NAME}" "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/bin/"
cp config.yaml "${OUTPUT_DIR}/${APP_NAME}.AppDir/"
cp -r sources/core/* "${OUTPUT_DIR}/${APP_NAME}.AppDir/sources/core/"

# Create AppRun script
cat > "${OUTPUT_DIR}/${APP_NAME}.AppDir/AppRun" << 'EOF'
#!/bin/bash
SELF=$(readlink -f "$0")
HERE=${SELF%/*}
export PATH="${HERE}/usr/bin:${PATH}"
cd "${HERE}"
exec "${HERE}/usr/bin/speak-to-ai" "$@"
EOF
chmod +x "${OUTPUT_DIR}/${APP_NAME}.AppDir/AppRun"

# Create desktop file
cat > "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/applications/${APP_NAME}.desktop" << EOF
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

# Create AppStream metadata
cat > "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/metainfo/${APP_NAME}.appdata.xml" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<component type="desktop-application">
  <id>io.github.ashbuk.speak-to-ai</id>
  <name>Speak-to-AI</name>
  <summary>Offline speech-to-text for AI assistants</summary>
  <description>
    <p>A minimalist, offline desktop application that enables voice input for AI assistants without sending your voice to the cloud. Uses the Whisper model locally for speech recognition.</p>
  </description>
  <url type="homepage">https://github.com/AshBuk/speak-to-ai</url>
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
ln -sf "./usr/share/icons/hicolor/256x256/apps/${APP_NAME}.png" "${OUTPUT_DIR}/${APP_NAME}.AppDir/${APP_NAME}.png"

# Download AppImage tool if not present
if [ ! -f "${OUTPUT_DIR}/appimagetool-${ARCH}.AppImage" ]; then
    echo "Downloading appimagetool..."
    wget -q "https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-${ARCH}.AppImage" -O "${OUTPUT_DIR}/appimagetool-${ARCH}.AppImage"
    chmod +x "${OUTPUT_DIR}/appimagetool-${ARCH}.AppImage"
fi

# Build AppImage
echo "Building AppImage..."
cd "${OUTPUT_DIR}"
./appimagetool-${ARCH}.AppImage "${APP_NAME}.AppDir" "${APP_NAME}-${APP_VERSION}-${ARCH}.AppImage"

echo "AppImage created: ${OUTPUT_DIR}/${APP_NAME}-${APP_VERSION}-${ARCH}.AppImage" 