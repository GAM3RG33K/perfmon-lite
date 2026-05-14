# perfmon — CLI Reference

> **Command:** `perfmon`
> **Version:** 1.0.0
> **Stack:** Go, Bubble Tea, Lipgloss

---

## 1. Synopsis

```
perfmon [flags] [command]
```

A blistering-fast, standalone, terminal-based mobile app profiler that provides real-time CPU, memory, and thread telemetry for Android and iOS apps without requiring heavy IDEs.

---

## 2. Commands

| Command | Description |
|---------|-------------|
| `perfmon [flags]` | Launch interactive TUI (auto-detect platform) |
| `perfmon devices [flags]` | List connected devices and simulators |
| `perfmon uninstall` | Remove perfmon binary from system |

## 3. Global Flags

| Flag | Shorthand | Type | Default | Description |
|------|-----------|------|---------|-------------|
| `--mock` | | | `false` | Run with simulated telemetry data (no device required) |
| `--device` | `-d` | `string` | `""` | Target device by serial/UUID (e.g. `emulator-5554`) |
| `--id` | | `string` | `""` | Target app by package name or bundle ID (e.g. `com.example.app`) |
| `--interval` | | `int` | `1` (or `$PERFMON_POLL_INTERVAL`) | Polling interval in seconds. Range: 1-60 |
| `--output` | | `string` | `"./perfmon_export"` (or `$PERFMON_EXPORT_DIR`) | Output path for export file (without extension) |
| `--buffer` | | `int` | `300` (or `$PERFMON_BUFFER_SIZE`) | Ring buffer capacity (number of data points kept in memory) |
| `--export` | | `string` | `""` | Export format: `json`, `md`, `html` |
| `--verbose` | `-V` | | `false` | Enable verbose logging to stderr |
| `--help` | `-h` | | `false` | Show help message and exit |
| `--version` | `-v` | | `false` | Show version information and exit |

---

## 3. Commands

### 3.1 `perfmon devices`

List connected Android devices and booted iOS simulators, then exit.

```
perfmon devices [flags]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--json` | `false` | Output device list as JSON |
| `--platform` | `"all"` | Filter by platform: `android`, `ios`, `all` |
| `--build-info` | `false` | Show build type (Debug/Release) for detected apps on each device |

**Example output:**

```
$ perfmon devices
DEVICE ID               NAME                            PLATFORM      TYPE
────────────────────────────────────────────────────────────────────────────
  emulator-5554         sdk_gphone16k_arm64             android       emulator
  F39098E7-...          iPhone 17 Pro                   ios           simulator

$ perfmon devices --json --platform android
[
  {
    "device": {
      "id": "emulator-5554",
      "name": "sdk_gphone16k_arm64",
      "platform": "android",
      "is_physical": false,
      "is_booted": true
    }
  }
]
```

---

### 3.2 `perfmon export` / `--export`

Export telemetry data to a file. Supports two modes:

**Non-interactive CLI mode:**
```
perfmon [--mock] --export <format> [--output <path>]
```
This samples 10 data points (with simulated polling) and writes the report.

**Interactive TUI mode:**
Press `e` (JSON), `Shift+E` (Markdown), or `Ctrl+E` (HTML) during a live session.

**Arguments:**

| Argument | Description |
|----------|-------------|
| `format` | Export format: `json`, `md`|`markdown`, `html`, `pdf` |

**Flags:**

| Flag | Shorthand | Default | Description |
|------|-----------|---------|-------------|
| `--output` | `-o` | `"./perfmon_export"` | Output path (extension auto-appended) |

**Examples:**

```bash
# Non-interactive export (CLI mode)
perfmon --mock --export json
perfmon --mock --export md --output ./reports/session-001
perfmon --mock --export html --output ./reports/perf-report
perfmon --device emulator-5554 --id in.thetatva.tatva --export json

# Interactive mode (press e/E/Ctrl+E inside TUI)
perfmon
```

**Generated files:**

| Format | File | Contents |
|--------|------|----------|
| `json` | `<output>.json` | Structured data matching PRD schema v1: metadata, metrics_summary, telemetry array |
| `md` | `<output>.md` | Human-readable Markdown report with summary table, telemetry table, ASCII sparklines |
| `html` | `<output>.html` | Standalone HTML with dark theme CSS, SVG vector charts for CPU/Memory/Threads |
| `pdf` | `<output>.pdf` | Native PDF with vector line charts on multiple pages (via `go-pdf/fpdf`) |

**JSON Schema:** The JSON export conforms to `https://perfmon.qzz.io/schemas/export-v1.json`. See `docs/architecture.md` §6 for the full schema specification, including `metadata`, `metrics_summary`, and `telemetry` arrays.

---

### 3.3 `perfmon version`

Print version information.

```
perfmon version [flags]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--json` | `false` | Output version info as JSON |

---

## 4. Interactive TUI Keybindings

When running `perfmon` without a command (or with `--mock`), the interactive TUI starts. The following keybindings are available:

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate lists |
| `←` / `→` | Switch tabs |
| `Tab` / `Shift+Tab` | Cycle tabs |
| `1`–`3` | Jump to tab by number |
| `Enter` | Select highlighted item |
| `r` | Refresh device list |
| `e` | Export to JSON |
| `Shift+E` | Export to Markdown |
| `Ctrl+E` | Export to HTML |
| `?` | Toggle full-screen help overlay |
| `q` / `Ctrl+C` | Quit |

---

## 5. Usage Examples

### 5.1 Launch interactive TUI with mock data

```bash
perfmon --mock
```

Starts the TUI dashboard with simulated sinusoidal CPU/memory/thread telemetry. Ideal for UI development and testing.

### 5.2 Launch interactive TUI targeting a specific device

```bash
perfmon --device emulator-5554
```

Skips the device selection screen and immediately connects to the specified device.

### 5.3 Launch with custom polling interval

```bash
perfmon --mock --interval 2
```

Samples telemetry every 2 seconds (default: 1s).

### 5.4 Export telemetry to JSON

```bash
perfmon export json --output ./data/metrics
# -> ./data/metrics.json
```

### 5.5 List all available devices

```bash
perfmon devices
```

### 5.6 List only Android devices as JSON

```bash
perfmon devices --platform android --json
```

---

## 6. Exit Codes

| Code | Meaning | Used When |
|------|---------|-----------|
| `0` | Success | Normal exit, `--version`, `--help` |
| `1` | General error | Invalid flags, runtime failure, no telemetry data |
| `2` | Device error | *(reserved for future use)* |
| `3` | Tool not configured | ADB not found, xcrun not found, no devices, no processes |
| `4` | Export failed | Write error, directory creation failure |

---

## 7. Environment Variables

| Variable | Description | Default | Status |
|----------|-------------|---------|--------|
| `PERFMON_ADB_PATH` | Path to `adb` executable | Auto-detected | ✅ |
| `PERFMON_BUFFER_SIZE` | Ring buffer capacity | `300` | ✅ |
| `PERFMON_POLL_INTERVAL` | Polling interval in seconds | `1` | ✅ |
| `PERFMON_EXPORT_DIR` | Output path for exports | `./perfmon_export` | ✅ |

---

## 8. Complete Command Tree

```
perfmon [--mock] [--device <id>] [--id <app>] [--interval <n>]
       [--buffer <n>] [--output <path>] [--verbose]
       [--help | -h] [--version | -v]

perfmon devices [--json] [--platform <p>] [--build-info]

perfmon uninstall
```

---

## 9. Common Workflows

### Quick check — mock mode

```bash
perfmon --mock
# → See live animated telemetry immediately
```

### Profile a specific app

```bash
# Step 1: List connected devices
perfmon devices

# Step 2: Connect and profile by device
perfmon --device emulator-5554

# Or target a specific app by package name
perfmon --id in.thetatva.tatva

# Or combine device + app
perfmon --device emulator-5554 --id in.thetatva.tatva

# Step 3: Press 'e' to export during session
# or use --export for non-interactive export:
perfmon --id in.thetatva.tatva --export json
```

### Headless export for CI/CD

```bash
perfmon devices --json | jq '.[] | select(.platform == "android") | .id'
perfmon --device <device_id> --export json
```
