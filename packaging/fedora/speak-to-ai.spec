# Copyright (c) 2025 Asher Buk
# SPDX-License-Identifier: MIT
# https://github.com/AshBuk/speak-to-ai

# =============================================================================
# Version definitions (single source of truth)
# =============================================================================
%global app_version     1.6.0
%global go_version      1.21
%global whisper_version 1.8.3

# Exclude auto-requires for vendored whisper libraries
%global __requires_exclude libwhisper\\.so|libggml.*\\.so

Name:           speak-to-ai
Version:        %{app_version}
Release:        1%{?dist}
Summary:        Offline speech-to-text desktop application
License:        MIT
URL:            https://github.com/AshBuk/speak-to-ai

# Main source with vendored Go dependencies
Source0:        %{name}-%{version}-vendored.tar.gz
# whisper.cpp sources (pinned version for reproducible builds)
Source1:        https://github.com/ggml-org/whisper.cpp/archive/refs/tags/v%{whisper_version}.tar.gz#/whisper-cpp-%{whisper_version}.tar.gz

ExclusiveArch:  x86_64

# =============================================================================
# Build dependencies
# =============================================================================
BuildRequires:  golang >= %{go_version}
BuildRequires:  gcc
BuildRequires:  gcc-c++
BuildRequires:  cmake
BuildRequires:  make

# CGO/whisper.cpp build requirements
BuildRequires:  pkgconfig

# Systray support (GTK/AppIndicator)
BuildRequires:  pkgconfig(gtk+-3.0)
BuildRequires:  pkgconfig(ayatana-appindicator3-0.1)

# D-Bus for hotkey support
BuildRequires:  pkgconfig(dbus-1)

# Vulkan SDK for GPU acceleration
BuildRequires:  vulkan-devel
BuildRequires:  shaderc

# =============================================================================
# Runtime dependencies
# =============================================================================
# Audio recording (one of these required)
Requires:       alsa-utils
# Note: ffmpeg available in RPM Fusion, documented as optional
Suggests:       ffmpeg

# Clipboard operations
Requires:       xsel
Recommends:     wl-clipboard

# Text input automation (X11: xdotool, Wayland: ydotool or wtype)
Requires:       xdotool
Recommends:     ydotool
Recommends:     wtype

# Desktop notifications
Requires:       libnotify

# System tray (GNOME needs extension, KDE works out of box)
Requires:       libayatana-appindicator-gtk3
Requires:       gtk3

# D-Bus for hotkeys
Requires:       dbus

# Vulkan runtime for GPU acceleration (optional, falls back to CPU)
Recommends:     vulkan-loader

# =============================================================================
# Bundled provides (Fedora requires declaring all bundled deps)
# =============================================================================
# whisper.cpp C library
Provides:       bundled(whisper-cpp) = %{whisper_version}

# Vendored Go dependencies (generated from vendor/modules.txt)
Provides:       bundled(golang(fyne.io/systray)) = 1.11.0
Provides:       bundled(golang(github.com/ggerganov/whisper.cpp/bindings/go)) = 0.0.0.20251028185044.c62adfbd1ecd
Provides:       bundled(golang(github.com/go-audio/audio)) = 1.0.0
Provides:       bundled(golang(github.com/go-audio/riff)) = 1.0.0
Provides:       bundled(golang(github.com/go-audio/wav)) = 1.1.0
Provides:       bundled(golang(github.com/godbus/dbus/v5)) = 5.1.0
Provides:       bundled(golang(github.com/gorilla/websocket)) = 1.5.3
Provides:       bundled(golang(github.com/holoplot/go-evdev)) = 0.0.0.20250804134636.ab1d56a1fe83
Provides:       bundled(golang(github.com/kr/text)) = 0.2.0
Provides:       bundled(golang(go.uber.org/goleak)) = 1.3.0
Provides:       bundled(golang(golang.org/x/sys)) = 0.37.0
Provides:       bundled(golang(gopkg.in/yaml.v2)) = 2.4.0

%description
Speak-to-AI is a minimalist, privacy-focused desktop application for offline
speech-to-text. It converts voice input directly into any active window
(editors, browsers, IDEs, AI assistants) using the Whisper model locally.

Features:
- Offline speech-to-text with local processing (privacy-first)
- Cross-platform support for X11 and Wayland
- Native integration with GNOME, KDE, and other Linux DEs
- Voice typing or clipboard mode
- Flexible audio recording (ALSA or PulseAudio/PipeWire)
- Multi-language support, custom hotkey binding

The Whisper small-q5_1 model (~181 MB) is downloaded on first run.

Note: For evdev hotkey support (fallback method), add user to 'input' group:
  sudo usermod -a -G input $USER

%prep
%autosetup -n %{name}-%{version}

# Unpack whisper.cpp into build directory
mkdir -p build
tar -xzf %{SOURCE1} -C build
mv build/whisper.cpp-%{whisper_version} build/whisper.cpp

%build
# =============================================================================
# 1) Build whisper.cpp libraries
# =============================================================================
# Set RPATH for whisper libraries to find each other in private prefix
WHISPER_RPATH='%{_libdir}/%{name}'

pushd build/whisper.cpp
cmake -B build \
    -DCMAKE_BUILD_TYPE=Release \
    -DBUILD_SHARED_LIBS=ON \
    -DCMAKE_INSTALL_RPATH="$WHISPER_RPATH" \
    -DCMAKE_BUILD_WITH_INSTALL_RPATH=ON \
    -DGGML_NATIVE=OFF \
    -DGGML_AVX=ON \
    -DGGML_AVX2=ON \
    -DGGML_FMA=ON \
    -DGGML_F16C=ON \
    -DGGML_VULKAN=ON
cmake --build build --parallel %{_smp_build_ncpus}
popd

# Prepare lib directory for Go build
mkdir -p lib
cp build/whisper.cpp/build/src/libwhisper.so* lib/
cp build/whisper.cpp/include/whisper.h lib/
cp build/whisper.cpp/ggml/include/*.h lib/ 2>/dev/null || :
cp build/whisper.cpp/build/ggml/src/libggml*.so* lib/ 2>/dev/null || :

# =============================================================================
# 2) Build Go binary with systray support (using vendored deps)
# =============================================================================
export CGO_ENABLED=1
export C_INCLUDE_PATH=$(pwd)/lib
export LIBRARY_PATH=$(pwd)/lib
export CGO_CFLAGS="-I$(pwd)/lib"
export CGO_LDFLAGS="-L$(pwd)/lib -lwhisper -lggml -lggml-cpu"
export LD_LIBRARY_PATH=$(pwd)/lib

# Set RPATH at build time to find bundled libraries at runtime
VENDOR_RPATH='$ORIGIN/../lib64/%{name}'
go build -v \
    -mod=vendor \
    -tags systray \
    -ldflags "-s -w -X main.version=%{version} -linkmode=external -extldflags '-Wl,-rpath,${VENDOR_RPATH}'" \
    -o %{name} \
    ./cmd/speak-to-ai

%install
# Binary
install -D -m 0755 %{name} %{buildroot}%{_bindir}/%{name}

# Bundled whisper libraries (private prefix to avoid conflicts)
install -d %{buildroot}%{_libdir}/%{name}
cp -a lib/libwhisper.so* %{buildroot}%{_libdir}/%{name}/
cp -a lib/libggml*.so* %{buildroot}%{_libdir}/%{name}/ 2>/dev/null || :

# Desktop entry
install -D -m 0644 io.github.ashbuk.speak-to-ai.desktop \
    %{buildroot}%{_datadir}/applications/io.github.ashbuk.speak-to-ai.desktop

# AppStream metainfo
install -D -m 0644 io.github.ashbuk.speak-to-ai.appdata.xml \
    %{buildroot}%{_metainfodir}/io.github.ashbuk.speak-to-ai.appdata.xml

# Icons (multiple sizes for HiDPI support)
install -D -m 0644 icons/io.github.ashbuk.speak-to-ai.png \
    %{buildroot}%{_datadir}/icons/hicolor/128x128/apps/io.github.ashbuk.speak-to-ai.png
install -D -m 0644 icons/io.github.ashbuk.speak-to-ai.svg \
    %{buildroot}%{_datadir}/icons/hicolor/scalable/apps/io.github.ashbuk.speak-to-ai.svg

# Documentation
install -D -m 0644 README.md %{buildroot}%{_docdir}/%{name}/README.md
install -D -m 0644 CHANGELOG.md %{buildroot}%{_docdir}/%{name}/CHANGELOG.md
install -D -m 0644 docs/Desktop_Environment_Support.md \
    %{buildroot}%{_docdir}/%{name}/Desktop_Environment_Support.md

%check
# Sanity check - verify binary runs and shows help
export LD_LIBRARY_PATH=%{buildroot}%{_libdir}/%{name}
%{buildroot}%{_bindir}/%{name} -help 2>&1 | grep -q "speak-to-ai"

%post
# Update icon cache
/bin/touch --no-create %{_datadir}/icons/hicolor &>/dev/null || :

%postun
if [ $1 -eq 0 ] ; then
    /bin/touch --no-create %{_datadir}/icons/hicolor &>/dev/null
    /usr/bin/gtk-update-icon-cache %{_datadir}/icons/hicolor &>/dev/null || :
fi

%posttrans
/usr/bin/gtk-update-icon-cache %{_datadir}/icons/hicolor &>/dev/null || :

%files
%license LICENSE
%doc %{_docdir}/%{name}/
# Binary
%{_bindir}/%{name}
# Bundled libraries
%dir %{_libdir}/%{name}
%{_libdir}/%{name}/libwhisper.so*
%{_libdir}/%{name}/libggml*.so*
# Desktop integration
%{_datadir}/applications/io.github.ashbuk.speak-to-ai.desktop
%{_metainfodir}/io.github.ashbuk.speak-to-ai.appdata.xml
%{_datadir}/icons/hicolor/128x128/apps/io.github.ashbuk.speak-to-ai.png
%{_datadir}/icons/hicolor/scalable/apps/io.github.ashbuk.speak-to-ai.svg

%changelog
* Tue Jan 14 2026 Asher Buk <AshBuk@users.noreply.github.com> - 1.6.0-1
- GPU acceleration: Vulkan backend support (auto-fallback to CPU)

* Tue Jan 07 2026 Asher Buk <AshBuk@users.noreply.github.com> - 1.5.2-1
- Hotkey provider: evdev is now the default (more reliable)
- Config path: use XDG path for all installations

* Mon Jan 06 2026 Asher Buk <AshBuk@users.noreply.github.com> - 1.5.1-1
- Auto-download whisper model to ~/.local/share/speak-to-ai/models/ on first run

* Mon Jan 06 2026 Asher Buk <AshBuk@users.noreply.github.com> - 1.5.0-1
- Add --version flag for version information display

* Sun Jan 04 2026 Asher Buk <AshBuk@users.noreply.github.com> - 1.4.2-1
- Initial RPM package for Fedora
- Bundled whisper.cpp 1.8.2 for offline speech recognition
- Vendored Go dependencies for offline builds
- System tray integration with GNOME/KDE support
- X11 and Wayland compatibility
