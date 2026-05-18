# perfmon — Mobile Performance Monitor & Profiler

**Blistering-fast, terminal-based mobile app profiling** — CPU, memory, and thread telemetry for Android and iOS, right in your terminal.

> **v0.0.5 Beta** — Log capture, stack traces, app detection, auto-export on quit. Feedback welcome!

```text
┌─────────────────────────────────────────────────────────────┐
│ perfmon-tool v0.0.1              Device: Pixel 8  Uptime: 12:34 │
├─────────────────────────────────────────────────────────────┤
│ [Dashboard]  [Processes]  [System Logs]              (q) quit │
├─────────────────────────────────────────────────────────────┤
│ App: com.example.app  [DEBUG]  │ CPU: 8 cores  │ Temp: 52°C │
├─────────────────────────────────────────────────────────────┤
│ CPU Utilization (overall)  78.2%                           │
│ ┌───────────────────────────────────────────────────────┐ │
│ │ ████████████████████████████████████████████────── 78% │ │
│ └───────────────────────────────────────────────────────┘ │
│ Memory (Total: 8.0 GB)  312 MB                            │
│ ┌───────────────────────────────────────────────────────┐ │
│ │ Used:  ████████████████████████████████──────  215 MB │ │
│ │ Cache: ██████████████────────────────────────  97 MB  │ │
│ └───────────────────────────────────────────────────────┘ │
│ Threads: 42  │ Peak CPU: 78%  │ Peak RAM: 215 MB          │
├─────────────────────────────────────────────────────────────┤
│ [↑/↓] Navigate  [TAB] Switch  [e] Export  [?] Help  [q] Quit │
└─────────────────────────────────────────────────────────────┘
```

---

## Quick Start

```bash
# Try it with mock data (no device needed)
perfmon-tool --mock

# Profile a connected Android device
perfmon-tool

# Export telemetry to HTML report
perfmon-tool --mock --export html --output ./report
```

---

## Features

| Feature | Android | iOS |
|---------|---------|-----|
| Device discovery | ✅ `adb devices -l` | ✅ `xcrun simctl list` |
| Process mapping | ✅ `adb shell ps` | ✅ `launchctl list` |
| CPU sampling | ✅ `/proc/<pid>/stat` (tick delta) | ✅ macOS `ps` (host-level) |
| Memory sampling | ✅ `/proc/<pid>/status` (VmRSS) | ✅ macOS `ps` (RSS) |
| Thread counting | ✅ `/proc/<pid>/status` | ❌ |
| Build type detection | ✅ `dumpsys package` | ✅ entitlements |
| Persistent shell pipe | ✅ Single `adb shell` connection | N/A (macOS host tools) |
| **Export formats** | **All platforms** |
| JSON export | ✅ Structured data (PRD schema v1) |
| Markdown export | ✅ Report with ASCII tables + charts |
| HTML export | ✅ Standalone page with SVG vector charts |
| PDF export | ✅ Vector line graph report (go-pdf/fpdf) |

---

## Installation

### One-liner install

**macOS / Linux:**
```bash
curl -sfL https://get.perfmon.qzz.io | bash
```

**Windows (PowerShell):**
```powershell
iwr https://get.perfmon.qzz.io/windows -useb | iex
```

Installs as `perfmon-tool` on all platforms.

### One-liner update

```bash
# Built-in subcommand
perfmon-tool update

# Or via curl
curl -sfL https://get.perfmon.qzz.io/update | bash
```

### One-liner uninstall

```bash
perfmon-tool uninstall
# Or via curl
curl -sfL https://get.perfmon.qzz.io/uninstall | bash
```

### Manual download

Download from [GitHub Releases](https://github.com/GAM3RG33K/perfmon-lite/releases).

| Platform | File |
|----------|------|
| macOS (Intel) | `perfmon-tool-<version>-darwin-amd64` |
| macOS (Apple Silicon) | `perfmon-tool-<version>-darwin-arm64` |
| Linux (x86_64) | `perfmon-tool-<version>-linux-amd64` |
| Linux (ARM64) | `perfmon-tool-<version>-linux-arm64` |
| Windows (x86_64) | `perfmon-tool-<version>-windows-amd64.exe` |
| Windows (ARM64) | `perfmon-tool-<version>-windows-arm64.exe` |

**Manual install (macOS/Linux):**
```bash
chmod +x perfmon-tool-* && sudo mv perfmon-tool-* /usr/local/bin/perfmon-tool
```

**Manual install (Windows):**
```powershell
mkdir $env:LOCALAPPDATA\perfmon -Force
move .\perfmon-tool-*.exe $env:LOCALAPPDATA\perfmon\perfmon-tool.exe
```

### Prerequisites

- **Android**: [ADB](https://developer.android.com/studio/command-line/adb) (`brew install android-platform-tools`)
- **iOS (simulators)**: [Xcode](https://developer.apple.com/xcode/) (`xcode-select --install`)

---

## Usage

### Interactive TUI

```bash
# Auto-detect platform (Android → iOS on macOS)
perfmon-tool

# Target a specific device
perfmon-tool --device emulator-5554

# Target a specific app by package/bundle ID
perfmon-tool --id in.thetatva.tatva

# Mock mode for development
perfmon-tool --mock

# Custom polling interval and buffer size
perfmon-tool --interval 2 --buffer 600
```

### TUI Keybindings

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
| `?` | Toggle full-screen help overlay |
| `q` / `Ctrl+C` | Quit |

### Non-interactive Export

```bash
perfmon-tool --mock --export json
perfmon-tool --mock --export md --output ./reports/session-001
perfmon-tool --mock --export html --output ./reports/perf-report
perfmon-tool --device emulator-5554 --id in.thetatva.tatva --export json
```

See the full [Usage Guide](USAGE.md) for detailed documentation.

---

## Documentation

| Document | Description |
|----------|-------------|
| [Usage Guide](USAGE.md) | Complete install, usage, and configuration reference |
| [Architecture](docs/architecture.md) | System topology, module breakdown, data flow |
| [CLI Reference](docs/cli-reference.md) | Full flag reference, commands, exit codes |
| [Development Plan](docs/plan.md) | Phased implementation plan with task tracking |
| [Checklist](docs/checklist.md) | Detailed progress checklist across all phases |
| [Gap Analysis](docs/GAP_TO_FILL.md) | Known issues and pending improvements |
| [PRD](PRD.md) | Full product requirements document |

---

## Development

```bash
make build          # Build the binary
make mock           # Run with mock data
make test           # Full test suite with race detector
make test-adb       # Android integration tests
make cross-build    # Cross-compile for all platforms
make cut-release    # Create and push a release tag
make retag          # Re-trigger CI build (deletes existing tag)
make clean          # Clean build artifacts
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
web/                         # React landing page (GitHub Pages)
scripts/                     # Install/update/uninstall scripts
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
