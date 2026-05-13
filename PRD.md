# Product Requirements Document (PRD)

**Project Name:** `perfmon` (Performance Monitor & Profiler)

**Document Version:** 1.0.0

**Stack:** Go (Golang), Bubble Tea / Lipgloss (TUI)

---

## 1. Executive Summary & Vision

### 1.1 The Problem

Mobile developers face immense friction when attempting to capture quick, lightweight performance metrics (CPU, Memory, Threads) across different platforms. Native tooling like Xcode Instruments and Android Studio Profiler are incredibly powerful but suffer from heavy resource footprints, long startup times, complex GUI workflows, and platform-locked ecosystems.

### 1.2 The Solution

`perfmon` is a blistering-fast, standalone, terminal-based mobile app profiler. Built entirely in Go, it acts as a lightweight orchestration layer that interacts natively with host platform tools (`adb`, `xcrun`) to aggregate real-time telemetry. It provides instant visual feedback via a highly polished Terminal User Interface (TUI) and exports deep offline metrics without requiring heavy IDEs.

### 1.3 Key Value Propositions

* **Zero Dependency Footprint:** Installs as a single statically linked binary.
* **Instant Start:** Boot-to-profiling takes sub-second execution time.
* **Unified Developer Experience:** One interface, regardless of targeting iOS or Android.

---

## 2. Target Audience & Personas

* **The Indie Mobile Developer:** Building cross-platform apps (Flutter/React Native) on macOS who wants to quickly verify memory leaks without context-switching to heavy IDEs.
* **The CI/CD Engineer:** Automating performance benchmarks in headless environments where GUI profilers cannot run.
* **The Low-Spec Hardware Contributor:** Running Linux or older hardware where running an Android Emulator alongside Android Studio causes system thrashing.

---

## 3. Product Requirements

### 3.1 Functional Requirements (Core Capabilities)

| ID | Feature | Description | Priority |
| --- | --- | --- | --- |
| **F-01** | Device Discovery | Automatically scan, identify, and list attached Android devices/emulators and booted iOS Simulators/Devices. | P0 |
| **F-02** | App / Process Mapping | Resolve running applications by Package Name (Android) or Bundle Identifier (iOS) and detect target Process IDs (PIDs). | P0 |
| **F-03** | Build Flag Detection | Parse application package manifests/entitlements to clearly badge targets as **Debug** or **Release** builds. | P1 |
| **F-04** | Telemetry Polling Engine | Safely sample CPU utilization (%), Memory footprint (PSS/RSS in KB), and active Thread Counts over time. | P0 |
| **F-05** | Offline Exporting | Snapshot current telemetry buffers to disk in JSON, Markdown, embedded HTML, and native PDF formats. | P1 |
| **F-06** | Pre-flight Setup Wizard | Detect missing native CLI bridges (`adb`) and offer interactive, inline downloads to tool-specific cache directories. | P2 |
| **F-07** | Mocking Subsystem | Provide a `--mock` execution flag to simulate target targets and dynamic telemetry graphs for local development. | P1 |

### 3.2 Non-Functional Requirements

* **Portability:** Must compile to standalone executables for `linux/amd64`, `linux/arm64`, `windows/amd64`, `darwin/amd64`, and `darwin/arm64`.
* **Footprint:** Final stripped release binary size must not exceed **20 MB**.
* **Offline Operation:** Zero reliance on active web connections during runtime; all CSS styles, graphing logic, and fonts for HTML/PDF exports must be statically bundled into the binary.
* **Host Performance:** The profiling loop must consume $< 2\%$ CPU overhead on the host machine to prevent skewing compilation or emulator performance.

---

## 4. Technical Architecture & System Design

### 4.1 System Topology

The application follows a decoupled unidirectional data flow model driven by the Bubble Tea architecture. The core polling engine communicates with the UI strictly through Go channels and standard messages.

```text
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

### 4.2 Module breakdown (Go Packages)

```text
cmd/
 └── perfmon/
      └── main.go                 # Entry point, CLI flag parsing, TUI boot
internal/
 ├── tui/
 │    ├── model.go                # Core Bubble Tea state machine
 │    ├── views/                  # Dashboard, Target Selector, Logs
 │    └── styles/                 # Lipgloss definitions (Colors, Borders)
 ├── engine/
 │    ├── engine.go               # Telemetry loop scheduler & ring buffers
 │    └── targets.go              # Shared interfaces for Device/App mapping
 ├── platform/
 │    ├── android/                # ADB bridging, logcat stream parsing
 │    └── ios/                    # simctl, devicectl, instruments hooks
 └── export/
      ├── templates/              # Embedded HTML/MD/PDF layouts (//go:embed)
      └── generator.go            # File output orchestration

```

---

## 5. Interface & User Experience (TUI) Design

### 5.1 Design System (Lipgloss)

* **Primary Accent:** ANSI Cyan (`#00FFFF`) / Terminal Default Bold for selection highlights.
* **Secondary Accent:** Magenta (`#FF00FF`) for charts and telemetry peaks.
* **Badges:** Green for `[DEBUG]`, Amber/Red for `[RELEASE]` (warning users about stripped telemetry limits).
* **Layout:** Fixed Header, dynamic tab-swappable body panes, persistent command footer.

### 5.2 Wireframe Layout

```text
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

## 6. Core Subsystem Specifications

### 6.1 Telemetry Engine & Buffers

To prevent memory exhaustion during long profiling sessions, metrics must be stored in a **Ring Buffer** (Circular Queue) with a fixed maximum capacity (e.g., last 300 data points, representing 5 minutes of polling at 1-second intervals).

$$\text{Buffer State} = \{ (t_0, C_0, M_0, T_0), (t_1, C_1, M_1, T_1), \dots, (t_n, C_n, M_n, T_n) \}$$

Where $t$ is Timestamp, $C$ is CPU percentage, $M$ is Memory allocated, and $T$ is active thread count.

### 6.2 Data Schema for JSON Export

```json
{
  "$schema": "https://perfmon.dev/schemas/export-v1.json",
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

## 7. Phased Implementation Plan

### Phase 1: Core Scaffolding & Mock Engine (Weeks 1-2)

* Initialize project repo, configure pre-commit hooks, set up GitHub Actions matrix.
* Build out the underlying Bubble Tea models, custom Lipgloss layout borders, and sparkline view integrations.
* Implement the `--mock` flag to pipe static sinusoidal curves into the TUI to perfect the dashboard views.

### Phase 2: Android Subsystem Integration (Weeks 3-4)

* Implement `adb` path discovery and pre-flight validation logic.
* Implement parsers for `adb shell dumpsys meminfo` and `adb shell top` commands.
* Wire up dynamic process selection directly from the UI target selector.

### Phase 3: iOS Subsystem Integration (Weeks 5-6)

* Add `xcrun simctl` parsers to discover active simulator targets.
* Implement application process identification routines for booted virtual instances.
* Integrate hooks for native macOS command-line tracking (`xcrun instruments` / `devicectl` telemetry streams).

### Phase 4: Reporting Subsystem & Final Polish (Weeks 7-8)

* Build standalone format generators (JSON schema serialization, Markdown template generation).
* Embed CSS and charts inside base64 HTML string generators using Go's `//go:embed` directive.
* Leverage `go-pdf/gopdf` to cleanly render vector line graphs directly onto output layout documents.
* Execute final binary stripping tests and test multi-architecture GitHub releases distribution scripts.

---

## 8. Open Risks & Mitigations

| Risk | Impact | Proposed Mitigation Strategy |
| --- | --- | --- |
| **ADB Command Overhead Skew** | High | Polling via explicit process execution loops can inflate target CPU values. **Mitigation:** Execute single long-lived persistent shell connections over stdin/stdout pipes rather than spawning isolated execution sub-shells per tick. |
| **iOS Hardware Sandbox Restrictions** | High | Unsigned third-party binaries face deep OS blocks when fetching internal process metrics on physical hardware. **Mitigation:** Explicitly document hardware monitoring limits within the CLI setup guide; direct physical iOS workflows toward native host fallback triggers (`xctrace`). |
| **Window Resizing Visual Artifacts** | Medium | Sudden terminal geometry changes can snap raw graph strings or break view boundaries. **Mitigation:** Intercept standard Bubble Tea `tea.WindowSizeMsg` updates to instantly recalculate buffer metrics and clear UI output states. |
