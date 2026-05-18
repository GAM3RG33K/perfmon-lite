# perfmon — Usage Guide

## Installation

### Quick install (recommended)

**macOS / Linux:**
```bash
curl -sfL https://get.perfmon.qzz.io | bash
```

**Windows (PowerShell):**
```powershell
iwr https://get.perfmon.qzz.io/windows -useb | iex
```

### Manual download

Download the latest binary for your platform from the [releases page](https://github.com/GAM3RG33K/perfmon-lite/releases).

| Platform | Download |
|----------|----------|
| macOS (Intel) | `perfmon-tool-<version>-darwin-amd64` |
| macOS (Apple Silicon) | `perfmon-tool-<version>-darwin-arm64` |
| Linux (x86_64) | `perfmon-tool-<version>-linux-amd64` |
| Linux (ARM64) | `perfmon-tool-<version>-linux-arm64` |
| Windows (x86_64) | `perfmon-tool-<version>-windows-amd64.exe` |
| Windows (ARM64) | `perfmon-tool-<version>-windows-arm64.exe` |

**macOS / Linux manual install:**
```bash
chmod +x perfmon-tool-<version>-<platform>
sudo mv perfmon-tool-<version>-<platform> /usr/local/bin/perfmon-tool
```

**Windows manual install:**
```powershell
# Move to a directory in your PATH, e.g.:
mkdir %LOCALAPPDATA%\perfmon
move perfmon-tool-<version>-windows-amd64.exe %LOCALAPPDATA%\perfmon\perfmon-tool.exe
# Add to PATH if not already:
setx PATH "%LOCALAPPDATA%\perfmon;%PATH%"
```

### Update

```bash
perfmon-tool update
```

### Uninstall

```bash
# macOS / Linux
curl -sfL https://get.perfmon.qzz.io/uninstall | bash

# Windows
iwr https://get.perfmon.qzz.io/uninstall/windows -useb | iex

# Or use the built-in subcommand
perfmon-tool uninstall
```

---

## Usage

### Quick start

```bash
# Try with mock data (no device needed)
perfmon-tool --mock

# Auto-detect and profile connected devices
perfmon-tool

# List connected devices
perfmon-tool devices

# Target a specific device
perfmon-tool --device emulator-5554

# Target a specific app by package name / bundle ID
perfmon-tool --id in.thetatva.tatva
```

### Interactive TUI

When run without flags, perfmon opens an interactive terminal UI:

```
┌─ perfmon-tool v0.0.1 ──────────────────────────────────┐
│  Target: Pixel 8  │  App: com.example.app  [DEBUG]    │
├───────────────────────────────────────────────────────┤
│  CPU Utilization (%)                                   │
│  100 ┤      ╭╮                                        │
│   50 ┤  ╭──╯╰─╮╭──╮                                  │
│    0 └─╯     ╰╯  ╰──────────────────────────────     │
│  Memory Footprint (MB)                                 │
│  210 ┤      ╭───────────────────────────────          │
│    0 └──╯                                             │
│  Peak CPU: 78%  │  Peak RAM: 215 MB                   │
├───────────────────────────────────────────────────────┤
│  [↑/↓] Navigate  [TAB] Switch  [e] Export  [?] Help  │
└───────────────────────────────────────────────────────┘
```

### Keybindings

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate lists |
| `←` / `→` | Switch tabs |
| `Tab` / `Shift+Tab` | Cycle tabs |
| `1`–`3` | Jump to tab by number |
| `Enter` | Select highlighted item |
| `e` | Open export format picker |
| `Shift+E` | Export directly to Markdown |
| `Ctrl+E` | Export directly to HTML |
| `r` | Refresh device list |
| `?` | Toggle help overlay |
| `q` / `Ctrl+C` | Quit |

### Exporting data

```bash
# Non-interactive export
perfmon-tool --mock --export json
perfmon-tool --mock --export md
perfmon-tool --mock --export html

# Export from real device
perfmon-tool --device emulator-5554 --export json

# Custom output path
perfmon-tool --mock --export json --output ./reports/benchmark

# Custom buffer size (controls sample count)
perfmon-tool --mock --buffer 100 --export json
```

### Targeting devices and apps

```bash
# List available devices
perfmon-tool devices

# List with JSON output
perfmon-tool devices --json

# List with process details
perfmon-tool devices --build-info

# Target specific device
perfmon-tool --device emulator-5554

# Target specific app by package/bundle ID
perfmon-tool --id in.thetatva.tatva

# Combine device + app
perfmon-tool --device emulator-5554 --id in.thetatva.tatva
```

### Configuration

```bash
# Custom polling interval (1-60 seconds)
perfmon-tool --interval 2

# Custom ring buffer size
perfmon-tool --buffer 600

# Verbose logging
perfmon-tool --verbose

# Environment variables
#   PERFMON_ADB_PATH       - Path to adb binary
#   PERFMON_BUFFER_SIZE    - Buffer capacity (default: 300)
#   PERFMON_POLL_INTERVAL  - Polling interval in seconds (default: 1)
#   PERFMON_EXPORT_DIR     - Output path for exports
```

---

## Platform support

| Feature | Android | iOS (simulator) |
|---------|---------|----------------|
| Device discovery | ✅ `adb devices -l` | ✅ `xcrun simctl list` |
| Process mapping | ✅ `adb shell ps` | ✅ `launchctl list` |
| CPU sampling | ✅ `/proc/<pid>/stat` | ✅ macOS `ps` (host-level) |
| Memory sampling | ✅ `/proc/<pid>/status` | ✅ macOS `ps` (RSS) |
| Thread counting | ✅ `/proc/<pid>/status` | ❌ |
| Build type detection | ✅ `dumpsys package` | ✅ entitlements |

### Prerequisites

- **Android**: [ADB](https://developer.android.com/studio/command-line/adb) (`brew install android-platform-tools`)
- **iOS (simulators)**: [Xcode](https://developer.apple.com/xcode/) (`xcode-select --install`)
- **Windows**: Works in PowerShell, Windows Terminal, or Command Prompt with VT support enabled.

---

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Device error |
| `3` | Tool not configured |
| `4` | Export failed |
