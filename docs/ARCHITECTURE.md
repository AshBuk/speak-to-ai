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
│  internal/tray/   │ internal/utils/  │ internal/constants/      │
│  internal/services/                                              │
└─────────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────────┐
│                       System Integration                        │
├─────────────────────────────────────────────────────────────────┤
│  X11/Wayland │ DBus Portals │ System Tray │ Audio System │ FS   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Service Pipelines

```
// Audio pipeline
  AudioService -> AudioRecorder -> (arecord/ffmpeg)
  AudioService -> TempFileManager -> cleanup logic
  AudioService -> WhisperEngine -> transcription

// Config pipeline  
  ConfigService -> YAMLLoader -> validation
  ConfigService -> StandardValidator -> sanitization

// UI pipeline
  UIService -> TrayManager -> systray integration
  UIService -> NotificationManager -> system notifications

// Hotkey pipeline
  HotkeyService -> HotkeyManager -> (evdev/dbus providers)
  HotkeyService -> ConfigAdapter -> hotkey mapping

// Output pipeline
  IOService -> OutputManager -> (clipboard/typing/websocket)
  
// Project Pattern
  Service -> Manager -> Provider/Adapter
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
- **`app.go`**: Application orchestrator with ServiceContainer and RuntimeContext
- **`handlers.go`**: Hotkey handlers that delegate to services

### **Service Layer** (`internal/services/`)
- **`interfaces.go`**: Service contracts and ServiceContainer definition
- **`component_factory.go`**: Components initialization (recorder, whisper, output, hotkeys, tray, notify, websocket)
- **`service_assembler.go`**: Services assembly and cross-dependencies
- **`callback_wirer.go`**: Tray menu callbacks wiring
- **`factory.go`**: Facade delegating to ComponentFactory, ServiceAssembler, CallbackWirer
- **`audio_service.go`**: Audio recording, Whisper transcription
- **`ui_service.go`**: System tray, notifications, UI state
- **`io_service.go`**: Text output, WebSocket server
- **`config_service.go`**: Configuration file operations
- **`hotkey_service.go`**: Hotkey registration and callbacks

Related constants:
- **`internal/constants/ui.go`**: UI icons/messages/titles centralization

### **Audio Processing Module** (`audio/`)

#### Core Audio Components
- **`interfaces/recorder.go`**: AudioRecorder interface definition
- **`recorders/`**: Recorder implementations
  - `base_recorder.go`: Common functionality for all audio recorders
  - `arecord_recorder.go`: ALSA-based recording implementation
  - `ffmpeg_recorder.go`: FFmpeg-based recording implementation
- **`factory/factory.go`**: Factory pattern for creating appropriate recorders

#### Processing (`processing/`)
- **`tempfile_manager.go`**: Temporary audio file lifecycle management

#### Testing & Mocking
- **`mocks/mock_recorder.go`**: Mock implementation for testing
- **`interfaces/interface_validation_test.go`**: Interface compliance tests
- **`*_test.go`**: Comprehensive unit tests for all components

### **Hotkey Management** (`hotkeys/`)

#### Core Hotkey System
- **`interfaces/provider.go`**: KeyboardEventProvider interface and types
- **`manager/manager.go`**: Cross-platform hotkey manager
- **`manager/manager_linux.go`**: Linux-specific hotkey implementation
- **`manager/manager_stub.go`**: Stub implementation for unsupported platforms
- **`adapters/adapter.go`**: Abstraction layer for different providers

#### Provider Implementations
- **`providers/dbus_provider.go`**: DBus GlobalShortcuts portal (preferred for GNOME/KDE)
- **`providers/evdev_provider.go`**: Direct evdev input handling (primary for AppImage)
- **`manager/provider_fallback.go`**: Fallback logic and hotkey re-registration
- **`providers/dummy_provider.go`**: Dummy provider for testing

#### Provider Selection & Fallback
- **Override** (config): `hotkeys.provider: auto | dbus | evdev` (default: `auto`)
- **System (GNOME/KDE)**: D-Bus primary → evdev fallback
- **AppImage**: evdev primary → D-Bus fallback
- **Flatpak**: D-Bus only (evdev blocked by sandbox)
- **Failover**: Seamless re-registration on provider switching

### **Speech Recognition** (`whisper/`)

#### Engine Components
- **`engine.go`**: Main Whisper engine using CGO bindings
- **`engine_stub.go`**: Stub implementation when CGO is disabled
- **`manager/model_manager.go`**: Whisper model lifecycle management
- **`providers/model_path_resolver.go`**: Bundled model path resolution for different environments
- **`whisper.go`**: Public facade and type re-exports for external use

### **Output Management** (`output/`)

#### Output Implementations
- **`interfaces/outputter.go`**: Outputter interface definition
- **`factory/factory.go`**: Factory for creating appropriate output handlers
- **`outputters/`**: Output implementations
  - `clipboard_outputter.go`: System clipboard integration
  - `type_outputter.go`: Active window typing simulation
  - `mock_outputter.go`: Mock implementation for testing

### **WebSocket API** (`websocket/`)

- **`server.go`**: WebSocket server for external integrations
- **`message_handler.go`**: WebSocket message processing
- **`authentication.go`**: API authentication and authorization
- **`retry_manager.go`**: Connection retry logic

### **Configuration System** (`config/`)

#### Configuration Management
- **`models/config.go`**: Configuration data structures and constants
- **`loaders/yaml_loader.go`**: YAML configuration loading, parsing and defaults
- **`validators/standard_validator.go`**: Configuration validation and sanitization
- **`security/utils.go`**: Configuration security and integrity checks

### **Internal Utilities** (`internal/`)

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
  - `sanitize.go`: Transcript sanitization (token cleanup, whitespace)
  - `async.go`: Goroutine tracking and graceful shutdown coordination

### **Build & Packaging**

#### Bash Scripts (`bash-scripts/`)
- **`build-appimage.sh`**: AppImage creation with dependencies
- **`build-flatpak.sh`**: Flatpak development helper
- **`check-license-headers.sh`**: License compliance validation
- **`dev-env.sh`**: Development environment setup
- **`flatpak-runtime.sh`**: Flatpak runtime wrapper

#### Docker Configuration (`docker/`)
- **`Dockerfile.dev`**: Development environment
- **`Dockerfile.appimage`**: AppImage build environment
- **`Dockerfile.flatpak`**: Flatpak build environment (WIP - not in release pipeline yet)
- **`docker-compose.yml`**: Multi-service development setup

#### Package Metadata
- **`io.github.ashbuk.speak-to-ai.json`**: Flatpak manifest
- **`io.github.ashbuk.speak-to-ai.desktop`**: Desktop entry
- **`io.github.ashbuk.speak-to-ai.appdata.xml`**: AppStream metadata

###  **Testing Infrastructure** (`tests/`)

#### Integration Tests (`tests/integration/`)
- **`integration_test.go`**: Main integration test suite
- **`end_to_end_test.go`**: Full workflow testing
- **`audio_integration_test.go`**: Audio system integration
- **`platform_integration_test.go`**: Platform-specific testing
- **`goroutine_lifecycle_test.go`**: Goroutine leak detection and graceful shutdown
- **`notification_integration_test.go`**: Desktop notification system integration

#### Test Utilities
- **`tests/mocks/logger.go`**: Mock logger implementation
- **Build tags**: Tests use `-tags=integration` for selective execution

---

*This architecture documentation is maintained alongside the codebase. Last updated: 2025-10-11*
