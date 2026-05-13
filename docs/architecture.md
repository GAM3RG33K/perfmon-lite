# perfmon — System Architecture

> **Stack:** Go, Bubble Tea, Lipgloss
> **Pattern:** Unidirectional data flow (Elm architecture via Bubble Tea)

---

## 1. System Topology

```
┌────────────────────────────────────────────────────────┐
│                   TUI Presentation Layer               │
│     (Bubble Tea Model • Lipgloss Styling • Sparklines) │
└───────────────────▲────────────────────────▲───────────┘
     Msg Channel    │                        │ Cmd Exec
┌───────────────────┴────────────────────────┴───────────┐
│               Core Orchestration Engine                │
│  ┌────────────────────────┐ ┌───────────────────────┐  │
│  │ Target Manager         │ │ Exporter Subsystem    │  │
│  │ (Discovery & Mapping)  │ │ (JSON/MD/HTML/PDF)    │  │
│  └────────────────────────┘ └───────────────────────┘  │
└───────▲────────────────────────────────────────▲───────┘
        │ Implementation Interfaces              │
┌───────┴────────────────────┐      ┌────────────┴───────┐
│  Android Provider          │      │ iOS Provider       │
│  • adb devices -l          │      │ • xcrun simctl     │
│  • dumpsys / top           │      │ • xcrun devicectl  │
└────────────────────────────┘      └────────────────────┘
```

---

## 2. Module Breakdown (Go Packages)

```
cmd/
 └── perfmon/
      └── main.go                 # Entry point, CLI flag parsing, TUI boot

internal/
 ├── tui/
 │    ├── model.go                # Core Bubble Tea state machine
 │    ├── views/
 │    │    ├── dashboard.go       # Main telemetry dashboard
 │    │    ├── target_selector.go # Device & process list
 │    │    └── logs.go            # System log viewer
 │    └── styles/
 │         ├── colors.go          # ANSI color constants
 │         ├── badges.go          # Debug/Release badge styles
 │         └── borders.go         # Panel border definitions
 │
 ├── engine/
 │    ├── engine.go               # Telemetry loop scheduler & ring buffers
 │    ├── engine_test.go          # Ring buffer + Engine unit tests (20 tests)
 │    ├── types.go                # Domain types (Device, AppProcess, TelemetrySnapshot)
 │    ├── types_test.go           # MetricsSummary unit tests (7 tests)
 │    └── targets.go              # Shared interfaces
 │
 ├── platform/
 │    ├── mock/
 │    │    ├── mock.go            # Mock provider for --mock mode
 │    │    └── mock_test.go       # Mock provider unit tests (15 tests + 1 benchmark)
 │    ├── android/
 │    │    ├── provider.go        # ADBProvider struct, persistent shell pipe (ensureShell/execInShell/closeShell)
 │    │    ├── discovery.go       # `adb devices -l` parser, device listing
 │    │    ├── process.go         # `adb shell ps` parser + BuildType detection via dumpsys
 │    │    ├── telemetry.go       # CPU (top), memory (VmRSS), threads (/proc) sampling via pipe
 │    │    ├── preflight.go       # ADB path detection, version check, device validation
 │    │    ├── provider_test.go   # 50 tests covering parsers, preflight, errors
 │    │    └── pipe_test.go       # 11 tests for persistent pipe (ensureShell, exec, restart, concurrency)
 │    └── ios/                    # (not yet implemented)
 │
 └── export/                      # (not yet implemented)
      ├── generator.go
      └── templates/
           ├── export.json.tmpl
           ├── export.md.tmpl
           └── export.html.tmpl
```

---

## 3. Data Flow

```
User Input (keyboard)
     │
     ▼
tea.Msg ─────────────────┐
     │                   │
     ▼                   ▼
Update() ────► Engine (Cmd)
     │                   │
     │                   ▼
     │              Poll Platform
     │              (Mock/Android/iOS)
     │                   │
     │                   ▼
     │              TelemetrySnapshot
     │              (via channel)
     │                   │
     ▼                   ▼
Model State ────► Ring Buffer (300 pts)
     │
     ▼
View() ────► TUI Rendering (Lipgloss)
     │
     ▼
Terminal Output
```

---

## 4. Ring Buffer Design

Metrics are stored in a **Circular Queue** with fixed capacity (300 data points = 5 minutes at 1s intervals).

```
Buffer State = { (t₀, C₀, M₀, T₀), (t₁, C₁, M₁, T₁), ..., (tₙ, Cₙ, Mₙ, Tₙ) }

Where:
  t  = Unix timestamp
  C  = CPU percentage (float64)
  M  = Memory in KB (int64)
  T  = Active thread count (int32)
```

**Properties:**
- O(1) append operations
- Oldest data auto-evicted when buffer is full
- Thread-safe reads/writes via `sync.RWMutex`
- Protected against concurrent read/write races

---

## 5. Key Interfaces

```go
// internal/engine/targets.go

type DeviceDiscovery interface {
    Discover() ([]Device, error)
}

type ProcessMapper interface {
    MapProcesses(deviceID string) ([]AppProcess, error)
    BuildType(packageName string) (BuildType, error)
}

type TelemetryProvider interface {
    Sample(pid int32) (*TelemetrySnapshot, error)
    Close() error
}
```

---

## 6. Data Schema (JSON Export)

```json
{
  "$schema": "https://perfmon.qzz.io/schemas/export-v1.json",
  "metadata": {
    "generated_at": "2026-05-13T16:19:38Z",
    "perfmon_version": "1.0.0",
    "target_platform": "android",
    "device_name": "Google Pixel 8",
    "app_package": "com.example.app",
    "build_type": "debug"
  },
  "metrics_summary": {
    "duration_seconds": 120,
    "peak_memory_kb": 220160,
    "average_cpu_percentage": 14.2
  },
  "telemetry": [
    { "timestamp": 1778689178, "cpu": 12.5, "memory_kb": 210450, "threads": 41 },
    { "timestamp": 1778689179, "cpu": 45.0, "memory_kb": 215000, "threads": 45 }
  ]
}
```

---

## 7. Design System (Lipgloss)

| Element | Color | Usage |
|---------|-------|-------|
| Primary Accent | ANSI Cyan (`#00FFFF`) | Selection highlights, headers |
| Secondary Accent | Magenta (`#FF00FF`) | Charts, telemetry peaks |
| Debug Badge | Green | `[DEBUG]` build indicator |
| Release Badge | Amber/Red | `[RELEASE]` warning |
| Background | Terminal default | Default background |
| Borders | Dim white | Panel separators |

---

## 8. TUI Layout

```
┌─ perfmon v1.0.0 ──────────────────────────────────────────────┐
│  Target: Pixel 8 (Physical)  │  App: com.example.app [DEBUG]  │
├───────────────────────────────────────────────────────────────┤
│  [ Dashboard ]   Threads/Procs   System Logs                  │
│                                                               │
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
│  [↑/↓] Navigate  [TAB] Switch Tabs  [e] Export  [q] Quit      │
└───────────────────────────────────────────────────────────────┘
```

---

## 9. Non-Functional Constraints

| Constraint | Target | Current Status |
|------------|--------|----------------|
| Binary size | < 20 MB (stripped release build) | ~4.7 MB (arm64, unstripped) |
| Host CPU overhead | < 2% during profiling | Not yet measured |
| Polling interval | 1 second (configurable via `--interval`) | ✅ Configurable |
| Ring buffer capacity | 300 data points (configurable via `--buffer`) | ✅ Configurable |
| Export formats | JSON, Markdown, HTML, PDF | ❌ Not yet implemented |
| Target platforms | linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64 | ❌ Only darwin/arm64 tested |
| Offline operation | Zero web dependencies at runtime | ✅ No runtime web dependencies |
| Race condition free | All tests pass with `-race` | ✅ Verified |
| Test coverage | ≥80% for engine + platform packages | ✅ Engine + mock + android provider covered |

---

## 10. Testing & Coverage

### Test Files

| File | Tests | What it covers |
|------|-------|----------------|
| `internal/engine/engine_test.go` | 20 tests | Ring buffer: new/empty, push/chronological order, eviction, latest, count, single-element, non-mutation, concurrent push, concurrent read/write. Engine: new, start/stop, set-target, poll, no-target error, provider error, close, concurrent poll safety |
| `internal/engine/types_test.go` | 7 tests | MetricsSummary: empty, single, multiple, uniform, floating-point, zero values. NewTelemetrySnapshot |
| `internal/platform/mock/mock_test.go` | 15 tests + 1 benchmark | Provider: deterministic output, different seeds, valid ranges, sinusoidal variation, step increment, PID ignored, memory leak after 100, leak cap, thread variation, copy semantics, close. Static helpers: MockDevice, MockProcess. BenchmarkSample |
| `internal/platform/android/provider_test.go` | 50 tests | Discovery: empty/headers/emulators/physical/offline/unauthorized. Process: empty/header-only/app/kernel-threads/multiple. BuildType: debuggable/release/alt-format/empty. Telemetry: CPU found/not-found/empty/zero/varied, VmRSS found/not-found/empty/zero, Threads found/not-found/empty/many. Preflight: version parse/no-match/empty/string. Provider: new/set-device/adb-command/close/interfaces/AdbError |
| `internal/platform/android/pipe_test.go` | 11 tests | Pipe: ensureShell opens/closes, idempotent, simple command, multiple commands, reconnects after close, Close cleanup. Sample: fake PID error, no device error, SetDevice edge cases (empty, reopen). Concurrency: 10 concurrent Sample calls, 20 concurrent SetDevice calls, Close idempotent |
| `internal/platform/android/adb_integration_test.go` | 13 tests | **Integration tests** (build tag: `adb_test`): ADB binary discovery, version check, device discovery, process mapping, build type detection (systemui + non-existent), sample init/system_server/non-existent PID, consecutive samples, persistent shell pipe echo + multi-command, pipe fallback after kill, full end-to-end flow (discover→select→map→buildtype→sample→close), 10 concurrent samples |

### Test Properties

- **All tests pass** with `-race` flag (race detector enabled)
- **Zero data races** detected
- **`go vet`** passes clean across the entire project
- **Benchmark:** `BenchmarkSample` available for mock provider performance profiling

### Running Tests

```bash
make test              # Full suite with race detector + coverage
make test-short        # Quick run without race detector
make test-adb          # ADB integration tests (requires connected device/emulator)
go test -v ./internal/engine/          # Engine tests only
go test -v ./internal/platform/mock/   # Mock provider tests only
go test -tags=adb_test -v ./internal/platform/android/  # ADB integration tests
```

### Integration Test Notes

- **Build tag:** `adb_test` — excluded from normal `go test ./...` runs
- **Prerequisite:** A connected ADB device (physical or emulator) with `adb` in PATH
- **Skip behavior:** Tests skip gracefully with `t.Skip()` when no ADB device is connected
- **Tested device:** `emulator-5554` (Android emulator, API 34+)

### Verification Commands

```bash
make build             # Compile the binary
go vet ./...           # Static analysis
```