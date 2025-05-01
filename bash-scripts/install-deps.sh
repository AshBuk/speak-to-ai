#!/bin/bash

# Speak-to-AI dependencies installation script
# This script installs all necessary system dependencies for Speak-to-AI

set -e  # Exit on error

echo "Installing Speak-to-AI system dependencies..."

# Detect distribution
if [ -f /etc/os-release ]; then
    . /etc/os-release
    DISTRO=$ID
else
    echo "Cannot detect Linux distribution"
    exit 1
fi

install_debian_deps() {
    echo "Installing dependencies for Debian/Ubuntu..."
    sudo apt-get update
    sudo apt-get install -y \
        alsa-utils \
        ffmpeg \
        libx11-dev \
        libxtst-dev \
        libpng-dev \
        libjpeg-dev \
        xclip \
        xdotool \
        pulseaudio \
        libevdev-dev \
        libinput-dev \
        libudev-dev
    
    echo "Debian/Ubuntu dependencies installed successfully"
}

install_fedora_deps() {
    echo "Installing dependencies for Fedora..."
    sudo dnf install -y \
        alsa-utils \
        ffmpeg \
        libX11-devel \
        libXtst-devel \
        libpng-devel \
        libjpeg-turbo-devel \
        xclip \
        xdotool \
        pulseaudio \
        libevdev-devel \
        libinput-devel \
        systemd-devel
    
    echo "Fedora dependencies installed successfully"
}

install_arch_deps() {
    echo "Installing dependencies for Arch Linux..."
    sudo pacman -Sy --noconfirm \
        alsa-utils \
        ffmpeg \
        libx11 \
        libxtst \
        libpng \
        libjpeg-turbo \
        xclip \
        xdotool \
        pulseaudio \
        libevdev \
        libinput \
        systemd-libs
    
    echo "Arch Linux dependencies installed successfully"
}

install_wayland_tools() {
    echo "Installing Wayland-specific tools..."
    
    case $DISTRO in
        debian|ubuntu)
            sudo apt-get install -y wl-clipboard wtype
            ;;
        fedora)
            sudo dnf install -y wl-clipboard wtype
            ;;
        arch)
            sudo pacman -Sy --noconfirm wl-clipboard wtype
            ;;
        *)
            echo "Unknown distribution for Wayland tools installation"
            return 1
            ;;
    esac
    
    echo "Wayland tools installed successfully"
}

# Install dependencies based on distribution
case $DISTRO in
    debian|ubuntu)
        install_debian_deps
        ;;
    fedora)
        install_fedora_deps
        ;;
    arch)
        install_arch_deps
        ;;
    *)
        echo "Unsupported distribution: $DISTRO"
        echo "Please install the following packages manually:"
        echo "- alsa-utils (for sound recording)"
        echo "- ffmpeg"
        echo "- X11 development libraries"
        echo "- xclip (for clipboard operations)"
        echo "- xdotool (for typing simulation)"
        echo "- wl-clipboard (for Wayland clipboard support)"
        echo "- wtype (for Wayland typing simulation)"
        echo "- libevdev (for input device access)"
        exit 1
        ;;
esac

# Install Wayland-specific tools regardless of distribution
install_wayland_tools

echo "Checking Go installation..."
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Please install Go 1.20 or higher"
    exit 1
fi

echo "All dependencies installed successfully!"
echo "You can now build Speak-to-AI with 'go build -o speak-to-ai cmd/daemon/main.go'"
echo "Note: For Wayland support with evdev, you may need to run with elevated permissions" 