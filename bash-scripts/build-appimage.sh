#!/bin/bash

# Speak-to-AI AppImage builder script

set -e  # Exit on error
set -x  # Show commands being executed

# Configuration
APP_NAME="speak-to-ai"
APP_VERSION="${APP_VERSION:-${GITHUB_REF_NAME:-$(git describe --tags --abbrev=0 2>/dev/null || date +%Y%m%d)}}"
ARCH="x86_64"
OUTPUT_DIR="dist"

echo "=== Starting AppImage build for ${APP_NAME} v${APP_VERSION} ==="

# Functions
prepare_environment() {
    echo "Preparing environment..."
    if [ -f "bash-scripts/dev-env.sh" ]; then
        source bash-scripts/dev-env.sh || true
    fi
    if [ ! -f "lib/whisper.h" ]; then
        make whisper-libs
    fi
}

download_model() {
    if [ ! -f "sources/language-models/small-q5_1.bin" ]; then
        echo "Downloading Whisper small-q5_1 model..."
        mkdir -p sources/language-models
        curl -fsSL "https://raw.githubusercontent.com/ggml-org/whisper.cpp/master/models/download-ggml-model.sh" | bash -s small-q5_1 || {
            echo "❌ Failed to download small-q5_1 model"; exit 1;
        }
        mv ggml-small-q5_1.bin "sources/language-models/small-q5_1.bin"
        echo "Model downloaded: $(ls -lh sources/language-models/small-q5_1.bin)"
    fi
}

create_appdir_structure() {
    echo "Creating AppDir structure..."
    mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/bin"
    mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/lib"
    mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/applications"
    mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/icons/hicolor/256x256/apps"
    mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/metainfo"
    mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/sources/language-models"
    mkdir -p "${OUTPUT_DIR}/${APP_NAME}.AppDir/sources/core"
}

build_application() {
    echo "Building ${APP_NAME} with systray support..."
    go build -tags systray -o "${APP_NAME}" cmd/daemon/main.go
    cp "${APP_NAME}" "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/bin/"

    echo "Building ${APP_NAME}-cli..."
    go build -o "${APP_NAME}-cli" cmd/cli/main.go
    cp "${APP_NAME}-cli" "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/bin/"

    cp config.yaml "${OUTPUT_DIR}/${APP_NAME}.AppDir/"
}

copy_libraries() {
    echo "Bundling libraries..."
    LIB_DST="${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/lib"
    mkdir -p "$LIB_DST"
    
    # Copy whisper libraries
    if compgen -G "lib/libwhisper.so*" > /dev/null; then
        cp -a lib/libwhisper.so* "$LIB_DST/" || true
    fi
    if compgen -G "lib/libggml*.so*" > /dev/null; then
        cp -a lib/libggml*.so* "$LIB_DST/" || true
    fi
    
    # Copy system tray libraries
    copy_system_lib "libayatana-appindicator3.so*"
    copy_system_lib "libayatana-indicator3.so*"
    copy_system_lib "libdbusmenu-gtk3.so*"
    copy_system_lib "libdbusmenu-glib.so*"
}

copy_system_lib() {
    local pattern="$1"
    for d in /usr/lib/x86_64-linux-gnu /usr/lib64 /usr/lib; do
        if compgen -G "$d/${pattern}" > /dev/null; then
            for f in $d/${pattern}; do
                echo "Including $f"
                cp -a "$f" "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/lib/" || true
            done
            break
        fi
    done
}

copy_binaries() {
    echo "Copying system dependencies..."
    copy_binary_if_exists "xsel"
    copy_binary_if_exists "wl-copy"
    copy_binary_if_exists "wtype"
    copy_binary_if_exists "ydotool"
    copy_binary_if_exists "notify-send"
    copy_binary_if_exists "arecord"
    copy_binary_if_exists "xdotool"
    copy_binary_if_exists "ffmpeg"
}

copy_binary_if_exists() {
    local binary_name="$1"
    local binary_path=$(which "$binary_name" 2>/dev/null || echo "")
    
    if [ -n "$binary_path" ]; then
        echo "Including $binary_name dependency..."
        cp "$binary_path" "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/bin/"
    else
        echo "Warning: $binary_name not found in PATH"
    fi
}

copy_sources() {
    if [ -d "sources/core" ] && [ -f "sources/core/whisper" ]; then
        echo "Copying whisper binaries..."
        cp sources/core/whisper "${OUTPUT_DIR}/${APP_NAME}.AppDir/sources/core/"
    else
        echo "Warning: Whisper binaries not found in sources/core"
    fi

    if [ -f "sources/language-models/small-q5_1.bin" ]; then
        echo "Including pre-downloaded Whisper model..."
        cp sources/language-models/small-q5_1.bin "${OUTPUT_DIR}/${APP_NAME}.AppDir/sources/language-models/"
    else
        echo "Warning: Whisper model not found at sources/language-models/small-q5_1.bin"
    fi
}

create_apprun() {
    echo "Creating AppRun script..."
    cat > "${OUTPUT_DIR}/${APP_NAME}.AppDir/AppRun" << 'EOF'
#!/bin/bash
SELF=$(readlink -f "$0")
HERE=${SELF%/*}
export PATH="${HERE}/usr/bin:${PATH}"
export LD_LIBRARY_PATH="${HERE}/usr/lib:${LD_LIBRARY_PATH}"

# Export PulseAudio/PipeWire environment for audio recording
export PULSE_SERVER="${PULSE_SERVER:-unix:${XDG_RUNTIME_DIR}/pulse/native}"
export PULSE_RUNTIME_PATH="${PULSE_RUNTIME_PATH:-${XDG_RUNTIME_DIR}/pulse}"

# Create user config directory
CONFIG_DIR="${HOME}/.config/speak-to-ai"
mkdir -p "${CONFIG_DIR}"

# First launch config setup
if [ ! -f "${CONFIG_DIR}/config.yaml" ]; then
    echo "First launch detected: Setting up configuration..."
    cp "${HERE}/config.yaml" "${CONFIG_DIR}/config.yaml"
fi

# Check hotkey support (prioritize D-Bus over input devices)
if command -v busctl >/dev/null 2>&1 && busctl --user status >/dev/null 2>&1; then
    # D-Bus session available - hotkeys should work on modern DEs
    echo "D-Bus session available - hotkeys supported"
elif [ -r /dev/input/event0 ] 2>/dev/null; then
    # evdev available - hotkeys should work
    echo "Input devices accessible - hotkeys supported"  
else
    # No hotkey support detected
    echo "Warning: Hotkeys may not work without additional setup."
    echo "For hotkeys on GNOME/KDE: Ensure D-Bus session is running"
    echo "For hotkeys on other DEs: sudo usermod -a -G input $USER"
    echo "Then log out and log back in."
    echo ""
fi

# Auto-integrate with desktop menu function
integrate_with_desktop() {
    local desktop_file="${HOME}/.local/share/applications/speak-to-ai.desktop"
    local icon_file="${HOME}/.local/share/icons/hicolor/256x256/apps/speak-to-ai.png"
    
    # Create .desktop file if not exists
    if [ ! -f "$desktop_file" ]; then
        echo "Creating desktop menu integration..."
        mkdir -p "$(dirname "$desktop_file")"
        cat > "$desktop_file" << DESKTOP_EOF
[Desktop Entry]
Name=Speak-to-AI
Comment=Offline speech-to-text for AI assistants
Exec="${SELF}" %U
Icon=speak-to-ai
Type=Application
Categories=Utility;Audio;Accessibility;
Terminal=false
StartupNotify=true
DESKTOP_EOF
        chmod +x "$desktop_file"
        echo "✅ Desktop menu integration created"
    fi
    
    # Copy icon if not exists
    if [ ! -f "$icon_file" ]; then
        mkdir -p "$(dirname "$icon_file")"
        cp "${HERE}/speak-to-ai.png" "$icon_file" 2>/dev/null || true
    fi
    
    # Update desktop database
    if command -v update-desktop-database >/dev/null 2>&1; then
        update-desktop-database "${HOME}/.local/share/applications"
    fi
}

# Check for AppImageLauncher integration
if command -v appimaged >/dev/null 2>&1; then
    echo "AppImageLauncher detected - will handle desktop integration"
elif command -v appimageupdatetool >/dev/null 2>&1; then
    echo "AppImageUpdateTool detected - will handle desktop integration"
else
    echo "No AppImageLauncher found - creating manual menu integration..."
    integrate_with_desktop
fi

# Determine which binary to run based on arguments
cd "${HERE}"

# Check if first argument is a CLI command
case "$1" in
    start|stop|status)
        # Route to CLI binary for CLI commands
        exec "${HERE}/usr/bin/speak-to-ai-cli" "$@"
        ;;
    --help|-h)
        # Show help that includes both daemon and CLI options
        # Use ARGV0 if available (original AppImage path), otherwise use generic name
        APPIMAGE_NAME="${ARGV0:-speak-to-ai.AppImage}"
        # Get just the basename if it's a full path
        APPIMAGE_NAME="$(basename "$APPIMAGE_NAME")"

        echo "Speak-to-AI - Offline speech-to-text for Linux"
        echo ""
        echo "Usage:"
        echo "  ./$APPIMAGE_NAME                    Start the daemon (GUI mode)"
        echo "  ./$APPIMAGE_NAME start              Begin recording (CLI mode)"
        echo "  ./$APPIMAGE_NAME stop               Stop recording and transcribe (CLI mode)"
        echo "  ./$APPIMAGE_NAME status             Check recording status (CLI mode)"
        echo ""
        echo "CLI Options:"
        echo "  --json                Output in JSON format"
        echo "  --socket PATH         Use custom IPC socket path"
        echo ""
        echo "Daemon Options:"
        echo "  --config PATH         Use custom config file"
        exit 0
        ;;
    *)
        # Run daemon for all other cases (no args, or daemon-specific args)
        exec "${HERE}/usr/bin/speak-to-ai" --config "${CONFIG_DIR}/config.yaml" "$@"
        ;;
esac
EOF
    chmod +x "${OUTPUT_DIR}/${APP_NAME}.AppDir/AppRun"
}

create_desktop_file() {
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
    ln -sf "./usr/share/applications/${APP_NAME}.desktop" "${OUTPUT_DIR}/${APP_NAME}.AppDir/${APP_NAME}.desktop"
}

create_appstream_metadata() {
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
  <categories>
    <category>Utility</category>
    <category>Audio</category>
    <category>Accessibility</category>
  </categories>
</component>
EOF
}


copy_icon() {
    echo "Copying application icon..."
    if [ -f "icons/io.github.ashbuk.speak-to-ai.png" ]; then
        cp "icons/io.github.ashbuk.speak-to-ai.png" "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/icons/hicolor/256x256/apps/${APP_NAME}.png"
        echo "Real icon copied successfully"
    else
        echo "Warning: Real icon not found, creating placeholder..."
        echo "iVBORw0KGgoAAAANSUhEUgAAAQAAAAEACAMAAABrrFhUAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAAAZQTFRF////AAAAVcLTfgAAAAF0Uk5TAEDm2GYAAAAqSURBVHja7cEBAQAAAIIg/69uSEABAAAAAAAAAAAAAAAAAAAAAAAAAHwZJsAAARqZF58AAAAASUVORK5CYII=" | base64 -d > "${OUTPUT_DIR}/${APP_NAME}.AppDir/usr/share/icons/hicolor/256x256/apps/${APP_NAME}.png"
    fi
    ln -sf "./usr/share/icons/hicolor/256x256/apps/${APP_NAME}.png" "${OUTPUT_DIR}/${APP_NAME}.AppDir/${APP_NAME}.png"
}

download_appimage_tools() {
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

    if [ ! -f "${TOOLS_DIR}/linuxdeploy-plugin-gtk-${ARCH}.AppImage" ]; then
        echo "Downloading linuxdeploy-plugin-gtk..."
        wget -q "https://github.com/linuxdeploy/linuxdeploy-plugin-gtk/releases/download/continuous/linuxdeploy-plugin-gtk-${ARCH}.AppImage" \
            -O "${TOOLS_DIR}/linuxdeploy-plugin-gtk-${ARCH}.AppImage" || true
        chmod +x "${TOOLS_DIR}/linuxdeploy-plugin-gtk-${ARCH}.AppImage" 2>/dev/null || true
    fi
}

build_appimage() {
    echo "Using linuxdeploy to gather dependencies..."
    cd "${OUTPUT_DIR}"

    export ARCH="${ARCH}"
    export VERSION="${APP_VERSION}"

    # Build list of executables dynamically
    EXEC_ARGS=""
    for exe in \
      "usr/bin/${APP_NAME}" \
      "usr/bin/${APP_NAME}-cli" \
      "usr/bin/xdotool" \
      "usr/bin/wtype" \
      "usr/bin/ydotool" \
      "usr/bin/wl-copy" \
      "usr/bin/xsel" \
      "usr/bin/arecord" \
      "usr/bin/notify-send"; do
      if [ -f "${APP_NAME}.AppDir/${exe}" ]; then
        EXEC_ARGS="$EXEC_ARGS --executable \"${APP_NAME}.AppDir/${exe}\""
      fi
    done

    # Force-bundle libs that are often dlopened at runtime
    LIB_ARGS=""
    for lib in \
      libayatana-appindicator3.so \
      libayatana-indicator3.so \
      libdbusmenu-gtk3.so \
      libdbusmenu-glib.so; do
      for d in /usr/lib/x86_64-linux-gnu /usr/lib64 /usr/lib; do
        cand=$(ls -1 $d/${lib}* 2>/dev/null | sort -V | tail -n1 || true)
        if [ -n "$cand" ]; then
          echo "Will bundle library: $cand"
          LIB_ARGS="$LIB_ARGS --library $cand"
          break
        fi
      done
    done

    # Try alternative appindicator libraries (Fedora/non-Ayatana systems)
    for lib in \
      libappindicator3.so \
      libindicator3.so; do
      for d in /usr/lib/x86_64-linux-gnu /usr/lib64 /usr/lib; do
        cand=$(ls -1 $d/${lib}* 2>/dev/null | sort -V | tail -n1 || true)
        if [ -n "$cand" ]; then
          echo "Will bundle library: $cand"
          LIB_ARGS="$LIB_ARGS --library $cand"
          break
        fi
      done
    done

    # Use linuxdeploy to automatically copy libraries and dependencies
    if eval "\"${TOOLS_DIR}/linuxdeploy-${ARCH}.AppImage\" --appimage-extract-and-run \
        --appdir \"${APP_NAME}.AppDir\" \
        ${EXEC_ARGS} \
        ${LIB_ARGS} \
        --desktop-file \"${APP_NAME}.AppDir/${APP_NAME}.desktop\" \
        --icon-file \"${APP_NAME}.AppDir/${APP_NAME}.png\" \
        --plugin gtk \
        --output appimage"; then
        
        APPIMAGE_FILE=$(find . -name "*.AppImage" ! -name "*tool*" -type f -print | head -n 1)
        
        if [ -n "$APPIMAGE_FILE" ]; then
            chmod +x "$APPIMAGE_FILE"
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
    
    if "${TOOLS_DIR}/appimagetool-${ARCH}.AppImage" --appimage-extract-and-run --no-appstream "${APP_NAME}.AppDir"; then
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
}

# Main execution
main() {
    prepare_environment
    download_model
    create_appdir_structure
    build_application
    copy_libraries
    copy_binaries
    copy_sources
    create_apprun
    create_desktop_file
    create_appstream_metadata
    copy_icon
    download_appimage_tools
    build_appimage
}

# Run main function
main "$@" 