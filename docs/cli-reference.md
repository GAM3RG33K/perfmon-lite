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

## 2. Global Flags

| Flag | Shorthand | Type | Default | Description |
|------|-----------|------|---------|-------------|
| `--mock` | | | `false` | Run with simulated telemetry data (no device required). Useful for UI development and demos |
| `--ios` | | | `false` | Force iOS mode (use xcrun instead of ADB auto-detection) |
| `--target` | `-t` | `string` | `""` | Specify target device/app (e.g., `emulator-5554`, `Pixel8`, `com.example.app`) |
| `--interval` | | `int` | `1` (or `$PERFMON_POLL_INTERVAL`) | Polling interval in seconds. Range: 1-60 |
| `--output` | | `string` | `"./perfmon_export"` (or `$PERFMON_EXPORT_DIR`) | Output path for export file (without extension) |
| `--buffer` | | `int` | `300` (or `$PERFMON_BUFFER_SIZE`) | Ring buffer capacity (number of data points kept in memory) |
| `--export` | | `string` | `""` | Export format for non-interactive mode: `json`, `md`, `html`, `pdf` |
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
Available Devices:
──────────────────────────────
  • emulator-5554  sdk_gphone16k_arm64  (android, emulator)
  • F39098E7-...   iPhone 17 Pro        (ios, simulator)

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
perfmon --export pdf --target emulator-5554

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
| `↑` / `↓` | Navigate lists (devices, processes, tabs) |
| `←` / `→` | Switch between tabs (Dashboard, Threads/Procs, System Logs) |
| `Tab` | Cycle forward through tabs |
| `Shift+Tab` | Cycle backward through tabs |
| `Enter` | Select highlighted item |
| `/` | Enter search/filter mode |
| `Esc` | Exit search mode / close panel |
| `r` | Refresh device list |
| `e` | Export current telemetry data (prompts for format: `json`, `md`, `html`, `pdf`) |
| `q` / `Ctrl+C` | Quit perfmon |
| `?` | Toggle help overlay |

---

## 5. Usage Examples

### 5.1 Launch interactive TUI with mock data

```bash
perfmon --mock
```

Starts the TUI dashboard with simulated sinusoidal CPU/memory/thread telemetry. Ideal for UI development and testing.

### 5.2 Launch interactive TUI targeting a specific device

```bash
perfmon --target emulator-5554
```

Skips the device selection screen and immediately connects to `emulator-5554`.

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
| `PERFMON_ADB_PATH` | Path to `adb` executable (checked before PATH and ANDROID_HOME) | Auto-detected | ✅ Implemented |
| `PERFMON_BUFFER_SIZE` | Ring buffer data point capacity | `300` | ✅ Implemented |
| `PERFMON_POLL_INTERVAL` | Default polling interval in seconds | `1` | ✅ Implemented |
| `PERFMON_EXPORT_DIR` | Default export directory / output path | `./perfmon_export` | ✅ Implemented |

---

## 8. Complete Command Tree

```
perfmon [--mock] [--target <device>] [--interval <n>] [--buffer <n>]
       [--output <path>] [--verbose]
       [--help | -h] [--version | -v]

perfmon devices [--json] [--platform <p>] [--build-info]

perfmon export <format> [--output <path>] [--pretty]
       <format> := json | md | html | pdf

perfmon version [--json]
```

---

## 9. Common Workflows

### Quick check — mock mode

```bash
perfmon --mock
# → See live animated telemetry immediately
```

### Profile a real Android app

```bash
# Step 1: List connected devices
perfmon devices

# Step 2: Connect and profile
perfmon --target emulator-5554

# Step 3: Press 'e' to export during session
# or after exiting, run:
perfmon export json --output ./results/my-profile
```

### Headless export for CI/CD

```bash
perfmon devices --json | jq '.[] | select(.platform == "android") | .id'
perfmon --target <device_id> export json
```
