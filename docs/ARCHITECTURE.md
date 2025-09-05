# Speak-to-AI Architecture Documentation

The application follows a **modular daemon architecture** with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────────┐
│                        Application Layer                        │
├─────────────────────────────────────────────────────────────────┤
│  cmd/daemon/main.go  │  Entry point, CLI parsing, environment   │
└─────────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────────┐
│                        Business Logic                           │
├─────────────────────────────────────────────────────────────────┤
│  internal/app/       │  Core application orchestration          │
└─────────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────────┐
│                       Service Modules                           │
├─────────────────────────────────────────────────────────────────┤
│  audio/     │  hotkeys/   │  whisper/   │  output/    │  config/ │
│  websocket/ │                                                   │
└─────────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────────┐
│                      Internal Utilities                         │
├─────────────────────────────────────────────────────────────────┤
│  internal/logger/ │ internal/notify/ │ internal/platform/       │
│  internal/tray/   │ internal/utils/  │                          │
└─────────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────────┐
│                       System Integration                        │
├─────────────────────────────────────────────────────────────────┤
│  X11/Wayland │ DBus Portals │ System Tray │ Audio System │ FS   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Module Breakdown

### 📁 **Core Application**

#### `cmd/daemon/main.go`
- **Purpose**: Application entry point and CLI argument parsing
- **Responsibilities**:
  - Command-line flag processing (`--config`, `--model`, `--debug`)
  - Environment detection (AppImage, Flatpak)
  - Path adjustment for portable packages
  - Application initialization and lifecycle management

#### `internal/app/` - Application Core
- **`app.go`**: Main application struct and lifecycle management
- **`initialization.go`**: Component initialization and dependency injection
- **`init_components.go`**: Individual component setup (tray, hotkeys, etc.)
- **`init_model.go`**: Whisper model initialization and validation
- **`runtime.go`**: Application runtime loop and shutdown handling
- **`handlers_*.go`**: Event handlers for different subsystems
  - `handlers_config.go`: Configuration management
  - `handlers_hotkeys.go`: Global hotkey processing
  - `handlers_recording.go`: Audio recording lifecycle
  - `handlers_streaming.go`: Real-time transcription
  - `handlers_vad.go`: Voice activity detection

### 🎤 **Audio Processing Module** (`audio/`)

#### Core Audio Components
- **`interfaces/recorder.go`**: AudioRecorder interface definition
- **`recorders/`**: Recorder implementations
  - `base_recorder.go`: Common functionality for all audio recorders
  - `arecord_recorder.go`: ALSA-based recording implementation
  - `ffmpeg_recorder.go`: FFmpeg-based recording implementation
- **`factory/factory.go`**: Factory pattern for creating appropriate recorders

#### Streaming & Processing (`processing/`)
- **`chunk_processor.go`**: Real-time audio chunk processing
- **`vad.go`**: Voice Activity Detection implementation
- **`tempfile_manager.go`**: Temporary audio file lifecycle management

#### Testing & Mocking
- **`mocks/mock_recorder.go`**: Mock implementation for testing
- **`interfaces/interface_validation_test.go`**: Interface compliance tests
- **`*_test.go`**: Comprehensive unit tests for all components

### ⌨️ **Hotkey Management** (`hotkeys/`)

#### Core Hotkey System
- **`interfaces/provider.go`**: KeyboardEventProvider interface and types
- **`manager/manager.go`**: Cross-platform hotkey manager
- **`manager/manager_linux.go`**: Linux-specific hotkey implementation
- **`manager/manager_stub.go`**: Stub implementation for unsupported platforms
- **`adapters/adapter.go`**: Abstraction layer for different providers

#### Provider Implementations
- **`providers/dbus_provider.go`**: DBus GlobalShortcuts portal (preferred for GNOME/KDE)
- **`providers/evdev_provider.go`**: Direct evdev input handling (fallback for other DEs)
- **`manager/provider_fallback.go`**: Fallback logic and hotkey re-registration
- **`providers/dummy_provider.go`**: Dummy provider for testing

#### Fallback Strategy
- **GNOME/KDE**: No fallback - D-Bus portal
- **i3/XFCE/MATE**: Auto-fallback from D-Bus to evdev on provider failure
- **Failover**: Seamless hotkey re-registration on provider switching

### 🗣️ **Speech Recognition** (`whisper/`)

#### Engine Components
- **`engine.go`**: Main Whisper engine using CGO bindings
- **`engine_stub.go`**: Stub implementation when CGO is disabled
- **`streaming_engine.go`**: Real-time streaming transcription
- **`streaming_engine_stub.go`**: Stub for streaming when CGO disabled
- **`model_manager.go`**: Whisper model lifecycle management

### 📤 **Output Management** (`output/`)

#### Output Implementations
- **`interfaces/outputter.go`**: Outputter interface definition
- **`factory/factory.go`**: Factory for creating appropriate output handlers
- **`outputters/`**: Output implementations
  - `clipboard_outputter.go`: System clipboard integration
  - `type_outputter.go`: Active window typing simulation
  - `combined_outputter.go`: Combined clipboard + typing output
  - `mock_outputter.go`: Mock implementation for testing

### 🌐 **WebSocket API** (`websocket/`)

- **`server.go`**: WebSocket server for external integrations
- **`message_handler.go`**: WebSocket message processing
- **`authentication.go`**: API authentication and authorization
- **`retry_manager.go`**: Connection retry logic

### ⚙️ **Configuration System** (`config/`)

#### Configuration Management
- **`models/config.go`**: Configuration data structures and constants
- **`loaders/yaml_loader.go`**: YAML configuration loading, parsing and defaults
- **`validators/standard_validator.go`**: Configuration validation and sanitization
- **`security/utils.go`**: Configuration security and integrity checks

### 🔧 **Internal Utilities** (`internal/`)

#### System Integration
- **`logger/logger.go`**: Structured logging with levels
- **`notify/notification.go`**: Desktop notification system
- **`platform/environment.go`**: Platform detection (X11/Wayland)
- **`tray/`**: System tray integration
  - `interface.go`: TrayManager interface
  - `default_manager.go`: Standard system tray implementation
  - `default_systray.go`: Systray library integration
  - `mock_tray.go`: Mock implementation for testing
- **`utils/`**: General utilities
  - `files.go`: File system utilities
  - `disk_linux.go`: Disk space checking (Linux)
  - `disk_stub.go`: Stub implementation

### 🏗️ **Build & Packaging**

#### Bash Scripts (`bash-scripts/`)
- **`build-appimage.sh`**: AppImage creation with dependencies
- **`build-flatpak.sh`**: Flatpak development helper
- **`check-license-headers.sh`**: License compliance validation
- **`dev-env.sh`**: Development environment setup
- **`flatpak-runtime.sh`**: Flatpak runtime wrapper

#### Docker Configuration (`docker/`)
- **`Dockerfile.dev`**: Development environment
- **`Dockerfile.appimage`**: AppImage build environment
- **`Dockerfile.flatpak`**: Flatpak build environment
- **`docker-compose.yml`**: Multi-service development setup

#### Package Metadata
- **`io.github.ashbuk.speak-to-ai.json`**: Flatpak manifest
- **`io.github.ashbuk.speak-to-ai.desktop`**: Desktop entry
- **`io.github.ashbuk.speak-to-ai.appdata.xml`**: AppStream metadata

### 🧪 **Testing Infrastructure** (`tests/`)

#### Integration Tests (`tests/integration/`)
- **`integration_test.go`**: Main integration test suite
- **`end_to_end_test.go`**: Full workflow testing
- **`audio_integration_test.go`**: Audio system integration
- **`platform_integration_test.go`**: Platform-specific testing

#### Test Utilities
- **`tests/mocks/logger.go`**: Mock logger implementation
- **Build tags**: Tests use `-tags=integration` for selective execution

---

*This architecture documentation is maintained alongside the codebase. Last updated: 2025-09-05*