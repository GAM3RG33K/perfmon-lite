# perfmon — Mobile Performance Monitor & Profiler

**Blistering-fast, terminal-based mobile app profiling** — CPU, memory, and thread telemetry for Android and iOS, right in your terminal.

```text
┌─ perfmon v1.0.0 ──────────────────────────────────────────────┐
│  Target: Pixel 8 (Physical)  │  App: com.example.app [DEBUG]  │
├───────────────────────────────────────────────────────────────┤
│  CPU Utilization (%)                                          │
│  78% ┤    ╭╮                                                  │
│  30% ┤ ╭──╯╰─╮╭──╮                                            │
│   0% └─╯     ╰╯  ╰──────────────────────────────────────────  │
│                                                               │
│  Memory Footprint (MB)                                        │
│  210 ┤      ╭───────────────────────────────────────────────  │
│  180 ┤   ╭──╯                                                 │
│    0 └───╯                                                    │
│                                                               │
│  Active Threads: 42  │  Peak RAM: 215 MB  │  Status: Active   │
├───────────────────────────────────────────────────────────────┤
│  ↑/↓ Navigate  TAB Switch  e Export  q Quit                   │
└───────────────────────────────────────────────────────────────┘
```

---

## Quick Start

```bash
# Try it with mock data (no device needed)
perfmon --mock

# Profile a connected Android device
perfmon

# Export 10 samples to HTML report
perfmon --mock --export html --output ./report
```

> **Windows users:** replace `perfmon` with `perfmon-tool` in all commands.

---

## Features

| Feature | Android | iOS |
|---------|---------|-----|
| Device discovery | ✅ `adb devices -l` | ✅ `xcrun simctl list` + `devicectl` |
| Process mapping | ✅ `adb shell ps` | ✅ `launchctl list` |
| CPU sampling | ✅ `/proc/<pid>/stat` (tick delta) | ✅ macOS `ps` (host-level) |
| Memory sampling | ✅ `/proc/<pid>/status` (VmRSS) | ✅ macOS `ps` (RSS) |
| Thread counting | ✅ `/proc/<pid>/status` (Threads) | ❌ (not available on iOS) |
| Build type detection | ✅ `dumpsys package` (DEBUGGABLE) | ✅ `_CodeSignature` + entitlements |
| Persistent shell pipe | ✅ Single `adb shell` connection | N/A (uses macOS host tools) |
| **Export formats** | **All platforms** |
| JSON export | ✅ Structured data (PRD schema v1) |
| Markdown export | ✅ Report with ASCII sparklines + tables |
| HTML export | ✅ Standalone page with SVG vector charts |
| PDF export | ✅ Vector line graph report (go-pdf/fpdf) |

---

## Installation

### One-liner install

**macOS / Linux:**
```bash
curl -sfL https://perfmon.qzz.io | bash
```

**Windows (PowerShell):**
```powershell
iwr https://perfmon.qzz.io/windows -useb | iex
```

> On macOS you may need to add `~/.local/bin` to your PATH.
> On Windows, the installer adds `%LOCALAPPDATA%\perfmon` to your user PATH — restart your terminal.

### One-liner uninstall

**macOS / Linux:**
```bash
curl -sfL https://perfmon.qzz.io/uninstall | bash
```

**Windows (PowerShell):**
```powershell
iwr https://perfmon.qzz.io/uninstall/windows -useb | iex
```

### Manual download

Download the latest binary from the [Releases page](https://github.com/GAM3RG33K/perfmon-lite/releases).

| Platform | Binary |
|----------|--------|
| macOS (Intel) | `perfmon_darwin_amd64` |
| macOS (Apple Silicon) | `perfmon_darwin_arm64` |
| Linux (x86_64) | `perfmon_linux_amd64` |
| Linux (ARM64) | `perfmon_linux_arm64` |
| Windows (x86_64) | `perfmon-tool_windows_amd64.exe` |
| Windows (ARM64) | `perfmon-tool_windows_arm64.exe` |

### Prerequisites

- **Android**: [ADB](https://developer.android.com/studio/command-line/adb) (`brew install android-platform-tools`)
- **iOS (simulators)**: [Xcode](https://developer.apple.com/xcode/) (`xcode-select --install`)

---

## Usage

### Interactive TUI

```bash
# Auto-detect platform (Android → iOS on macOS)
perfmon

# Target a specific device
perfmon --device emulator-5554

# Target a specific app by package/bundle ID
perfmon --id in.thetatva.tatva

# Force mock mode for development
perfmon --mock

# Custom polling interval (1-60 seconds) and buffer size
perfmon --interval 2 --buffer 600
```

### TUI Keybindings

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate lists |
| `←` / `→` | Switch tabs |
| `Tab` | Cycle forward through tabs |
| `1`–`3` | Jump to tab by number |
| `Enter` | Select highlighted item |
| `e` | Export to JSON |
| `Shift+E` | Export to Markdown |
| `Ctrl+E` | Export to HTML |
| `r` | Refresh device list |
| `?` | Toggle full-screen help overlay |
| `q` / `Ctrl+C` | Quit |

### Non-interactive Export

```bash
# Export 10 samples to JSON
perfmon --mock --export json

# Export to Markdown with custom path
perfmon --mock --export md --output ./reports/session-001

# Export to HTML
perfmon --mock --export html --output ./reports/perf-report

# Target specific device and app for export
perfmon --device emulator-5554 --id in.thetatva.tatva --export json

# Default paths
perfmon --mock --export json          # → ./perfmon_export.json
perfmon --mock --export md            # → ./perfmon_export.md
perfmon --mock --export html          # → ./perfmon_export.html
perfmon --mock --export pdf           # → ./perfmon_export.pdf
```

### Export Formats

| Format | Description |
|--------|-------------|
| **JSON** | Structured data with metadata, metrics summary, and telemetry array — ideal for programmatic analysis |
| **Markdown** | Human-readable report with summary table, per-sample telemetry table, and ASCII sparkline charts |
| **HTML** | Standalone dark-themed webpage with SVG vector line charts for CPU, Memory, and Threads — no internet needed |
| **PDF** | Native PDF with multi-page vector line graphs — perfect for sharing |

---

## Documentation

| Document | Description |
|----------|-------------|
| [Architecture](docs/architecture.md) | System topology, module breakdown, data flow, interfaces |
| [CLI Reference](docs/cli-reference.md) | Full flag reference, commands, exit codes, environment variables |
| [Development Plan](docs/plan.md) | Phased implementation plan with task tracking |
| [Checklist](docs/checklist.md) | Detailed progress checklist across all phases |
| [Gap Analysis](docs/GAP_TO_FILL.md) | Known issues and pending improvements |
| [PRD](PRD.md) | Full product requirements document |
| [Domain Setup](docs/domain-setup.md) | Cloudflare + DNS configuration guide |

---

## Development

```bash
# Build
make build

# Run with mock data
make mock

# Run tests
make test               # Full suite with race detector
make test-short         # Quick run (no race detector)

# Run Android integration tests (requires emulator/device)
make test-adb

# Cross-compile for all platforms
make cross-build

# Create a release tag and push (triggers CI release build)
make cut-release

# Re-tag to trigger a new CI release build (deletes existing tag)
make retag

# Install/update from latest GitHub release
make install             # macOS/Linux only
make update              # macOS/Linux only

# Clean build artifacts
make clean
```

### Project Structure

```
cmd/perfmon/main.go          # Entry point, CLI flags, TUI boot
internal/
├── engine/                  # Telemetry engine, ring buffer, domain types
├── tui/                     # Bubble Tea TUI (dashboard, target selector, logs)
├── platform/
│   ├── mock/                # Simulated telemetry for --mock mode
│   ├── android/             # ADB-based provider (discovery, process, telemetry)
│   └── ios/                 # xcrun-based provider (simctl, devicectl)
└── export/                  # Export subsystem (JSON, MD, HTML, PDF)
```

### Test Stats

| Package | Tests | Status |
|---------|-------|--------|
| Engine + Types | 27 | ✅ |
| Mock Provider | 15 + 1 benchmark | ✅ |
| Android Provider | 59 | ✅ |
| iOS Provider | 34 | ✅ |
| Export Subsystem | 35 | ✅ |
| ADB Integration | 13 | ✅ (build tag: `adb_test`) |
| **Total** | **~183** | ✅ All pass with `-race` |

---

## License

MIT License — see [LICENSE](LICENSE).

---

*Built with [Go](https://go.dev/), [Bubble Tea](https://github.com/charmbracelet/bubbletea), and [Lipgloss](https://github.com/charmbracelet/lipgloss).*
