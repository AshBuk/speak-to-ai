#!/bin/bash

# Speak-to-AI AppImage builder script

set -e  # Exit on error
set -x  # Show commands being executed

# Configuration
APP_NAME="speak-to-ai"
APP_VERSION_RAW="${APP_VERSION:-${GITHUB_REF_NAME:-$(git describe --tags --abbrev=0 2>/dev/null || date +%Y%m%d)}}"
APP_VERSION=$(echo "${APP_VERSION_RAW}" | sed 's/^v//')
ARCH="x86_64"
OUTPUT_DIR="dist"
APPDIR="${OUTPUT_DIR}/${APP_NAME}.AppDir"
TOOLS_DIR="$(pwd)/tools"
# Determine script directory (works both locally and in Docker)
if [ -f "packaging/appimage/AppRun" ]; then
    SCRIPT_DIR="$(pwd)/packaging/appimage"
else
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
fi

echo "=== Starting AppImage build for ${APP_NAME} v${APP_VERSION} ==="

# Prepare build environment
prepare_environment() {
    echo "Preparing environment..."
    if [ -f "bash-scripts/dev-env.sh" ]; then
        source bash-scripts/dev-env.sh || true
    fi
    if [ ! -f "lib/whisper.h" ]; then
        make internal-whisper-libs
    fi
}

# Download Whisper model if missing
download_model() {
    if [ -f "sources/language-models/small-q5_1.bin" ]; then
        return
    fi
    echo "Downloading Whisper small-q5_1 model..."
    mkdir -p sources/language-models
    curl -fsSL "https://raw.githubusercontent.com/ggml-org/whisper.cpp/master/models/download-ggml-model.sh" | bash -s small-q5_1 || {
        echo "Failed to download model"; exit 1
    }
    mv ggml-small-q5_1.bin "sources/language-models/small-q5_1.bin"
}

# Create AppDir structure and build application
build_appdir() {
    echo "Creating AppDir structure..."
    mkdir -p "${APPDIR}"/{usr/bin,usr/lib,usr/share/applications,usr/share/icons/hicolor/256x256/apps,usr/share/metainfo,sources/language-models}

    # Build application
    echo "Building ${APP_NAME}..."
    go build -tags systray -ldflags "-s -w -X main.version=${APP_VERSION}" -o "${APP_NAME}" ./cmd/speak-to-ai
    cp "${APP_NAME}" "${APPDIR}/usr/bin/"
    cp config.yaml "${APPDIR}/"

    # Copy Whisper libraries
    cp -a lib/libwhisper.so* lib/libggml*.so* "${APPDIR}/usr/lib/" 2>/dev/null || true

    # Copy Whisper model
    [ -f "sources/language-models/small-q5_1.bin" ] && \
        cp sources/language-models/small-q5_1.bin "${APPDIR}/sources/language-models/"

    # Copy external binaries
    for bin in xsel wl-copy wtype ydotool notify-send arecord xdotool ffmpeg; do
        local path=$(which "$bin" 2>/dev/null || true)
        [ -n "$path" ] && cp "$path" "${APPDIR}/usr/bin/" && echo "Bundled: $bin"
    done

    # Copy AppRun from template
    cp "${SCRIPT_DIR}/AppRun" "${APPDIR}/AppRun"
    chmod +x "${APPDIR}/AppRun"

    # Copy desktop file and icon (use existing files from repo)
    cp "io.github.ashbuk.speak-to-ai.desktop" "${APPDIR}/usr/share/applications/${APP_NAME}.desktop"
    cp "io.github.ashbuk.speak-to-ai.appdata.xml" "${APPDIR}/usr/share/metainfo/${APP_NAME}.appdata.xml"
    # Copy icon with name matching Icon= field in .desktop file
    mkdir -p "${APPDIR}/usr/share/icons/hicolor/128x128/apps"
    cp "icons/io.github.ashbuk.speak-to-ai.png" "${APPDIR}/usr/share/icons/hicolor/128x128/apps/io.github.ashbuk.speak-to-ai.png"
    cp "icons/io.github.ashbuk.speak-to-ai.png" "${APPDIR}/usr/share/icons/hicolor/128x128/apps/${APP_NAME}.png"

    # Create symlinks for AppImage standard
    ln -sf "usr/share/applications/${APP_NAME}.desktop" "${APPDIR}/${APP_NAME}.desktop"
    ln -sf "usr/share/icons/hicolor/128x128/apps/${APP_NAME}.png" "${APPDIR}/${APP_NAME}.png"
    ln -sf "usr/share/icons/hicolor/128x128/apps/io.github.ashbuk.speak-to-ai.png" "${APPDIR}/io.github.ashbuk.speak-to-ai.png"
}

# Download AppImage tools
download_tools() {
    mkdir -p "${TOOLS_DIR}"
    local base_url="https://github.com"

    [ ! -f "${TOOLS_DIR}/linuxdeploy-${ARCH}.AppImage" ] && \
        wget -q "${base_url}/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-${ARCH}.AppImage" \
            -O "${TOOLS_DIR}/linuxdeploy-${ARCH}.AppImage" && \
        chmod +x "${TOOLS_DIR}/linuxdeploy-${ARCH}.AppImage"

    [ ! -f "${TOOLS_DIR}/appimagetool-${ARCH}.AppImage" ] && \
        wget -q "${base_url}/AppImage/AppImageKit/releases/download/continuous/appimagetool-${ARCH}.AppImage" \
            -O "${TOOLS_DIR}/appimagetool-${ARCH}.AppImage" && \
        chmod +x "${TOOLS_DIR}/appimagetool-${ARCH}.AppImage"

    [ ! -f "${TOOLS_DIR}/linuxdeploy-plugin-gtk-${ARCH}.AppImage" ] && \
        wget -q "${base_url}/linuxdeploy/linuxdeploy-plugin-gtk/releases/download/continuous/linuxdeploy-plugin-gtk-${ARCH}.AppImage" \
            -O "${TOOLS_DIR}/linuxdeploy-plugin-gtk-${ARCH}.AppImage" || true
    chmod +x "${TOOLS_DIR}/linuxdeploy-plugin-gtk-${ARCH}.AppImage" 2>/dev/null || true
}

# Find and bundle system library
find_lib() {
    local pattern="$1"
    for d in /usr/lib/x86_64-linux-gnu /usr/lib64 /usr/lib; do
        local lib=$(ls -1 $d/${pattern}* 2>/dev/null | sort -V | tail -n1 || true)
        [ -n "$lib" ] && echo "$lib" && return
    done
}

# Build final AppImage
build_appimage() {
    cd "${OUTPUT_DIR}"
    export ARCH VERSION="${APP_VERSION}"

    # Build executable arguments for linuxdeploy
    local exec_args=""
    for exe in ${APP_NAME} xdotool wtype ydotool wl-copy xsel arecord notify-send; do
        [ -f "${APP_NAME}.AppDir/usr/bin/${exe}" ] && \
            exec_args="$exec_args --executable ${APP_NAME}.AppDir/usr/bin/${exe}"
    done

    # Build library arguments for system tray support
    local lib_args=""
    for lib in libayatana-appindicator3.so libayatana-indicator3.so libdbusmenu-gtk3.so libdbusmenu-glib.so libappindicator3.so libindicator3.so; do
        local found=$(find_lib "$lib")
        [ -n "$found" ] && lib_args="$lib_args --library $found" && echo "Will bundle: $found"
    done

    # Step 1: Use linuxdeploy to gather dependencies
    "${TOOLS_DIR}/linuxdeploy-${ARCH}.AppImage" --appimage-extract-and-run \
        --appdir "${APP_NAME}.AppDir" \
        $exec_args \
        $lib_args \
        --desktop-file "${APP_NAME}.AppDir/${APP_NAME}.desktop" \
        --icon-file "${APP_NAME}.AppDir/io.github.ashbuk.speak-to-ai.png" \
        --plugin gtk || echo "Warning: linuxdeploy had issues, continuing..."

    # Remove unnecessary docs (licenses available in source repos)
    rm -rf "${APP_NAME}.AppDir/usr/share/doc"

    # Create final AppImage
    echo "Creating AppImage..."
    "${TOOLS_DIR}/appimagetool-${ARCH}.AppImage" --appimage-extract-and-run "${APP_NAME}.AppDir"

    # Rename output
    local output=$(find . -maxdepth 1 -name "*.AppImage" ! -name "appimagetool*" -type f | head -n1)
    if [ -n "$output" ]; then
        mv "$output" "${APP_NAME}-${APP_VERSION}-${ARCH}.AppImage"
        chmod +x "${APP_NAME}-${APP_VERSION}-${ARCH}.AppImage"
        ls -lh "${APP_NAME}-${APP_VERSION}-${ARCH}.AppImage"
        echo "=== Build completed successfully! ==="
    else
        echo "Error: AppImage not found"
        exit 1
    fi
}

# Main
main() {
    prepare_environment
    download_model
    build_appdir
    download_tools
    build_appimage
}

main "$@"
