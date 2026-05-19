# perfmon — Development Plan

> **Project:** Performance Monitor & Profiler (Mobile App TUI)
> **Stack:** Go, Bubble Tea, Lipgloss
> **Version:** 0.0.7
> **Last Updated:** 2026-05-19

---

## Phase 0: Project Scaffolding & Toolchain Setup

| #  | Task | Status |
|----|------|--------|
| 0.1 | Install Go (>=1.22) | ✅ Go 1.26.3 installed via Homebrew |
| 0.2 | Initialize Go module (`go mod init github.com/w1n/perfmon`) | ✅ `go.mod` created |
| 0.3 | Create directory structure (cmd/, internal/, etc.) | ✅ All dirs scaffolded |
| 0.4 | Install TUI dependencies (bubbletea, lipgloss, bubbles) | ✅ `go mod tidy` completed |
| 0.5 | Add Makefile / Taskfile (build, run, mock, test, lint, clean) | ✅ `Makefile` with 10 targets |
| 0.6 | Add `.goreleaser.yml` for multi-arch release builds | ✅ GoReleaser config v2, post-build smoke test hook |
| 0.7 | Add GitHub Actions CI (lint + test + build matrix) | ✅ Build, vet, test, cross-build, ADB integration, artifact upload |
| 0.8 | Add `.gitignore` for Go projects | ✅ Standard Go .gitignore |

---

## Phase 1: Core Scaffolding & Mock Engine (Weeks 1-2)

| #  | Task | Status |
|----|------|--------|
| 1.1 | `cmd/perfmon/main.go` — Entry point, CLI flag parsing (`--mock`, `--target`), Bubble Tea program boot | ✅ `devices`, `export`, `version` subcommands |
| 1.2 | `internal/engine/engine.go` — Telemetry loop scheduler + ring buffer (last 300 data points at 1s intervals) | ✅ Thread-safe ring buffer with mutex |
| 1.3 | `internal/engine/targets.go` — Shared interfaces: `DeviceDiscovery`, `ProcessMapper`, `TelemetryProvider` | ✅ Includes `PlatformProvider` composite interface |
| 1.4 | `internal/engine/types.go` — Domain types: `Device`, `AppProcess`, `TelemetrySnapshot` | ✅ Plus `MetricsSummary`, `ComputeMetricsSummary()` |
| 1.5 | `internal/tui/model.go` — Core Bubble Tea model: `Init()`, `Update()`, `View()` with tab support | ✅ 3 tabs, keybindings, resize handling |
| 1.6 | `internal/tui/views/dashboard.go` — Dashboard view: CPU/memory area charts, peak stats | ✅ Btop-style block charts via `internal/chart` |
| 1.6b | `internal/chart/chart.go` — Shared chart renderer for TUI, exports, and web demo | ✅ Catmull-Rom smoothing, 100-point window, block symbols |
| 1.7 | `internal/tui/views/target_selector.go` — Target selector view: device list + process list | ✅ Platform/build-type badges |
| 1.8 | `internal/tui/views/logs.go` — System log view | ✅ Scrollable log viewer |
| 1.9 | `internal/tui/styles/` — Lipgloss styling: colors (cyan, magenta), badges, borders | ✅ 3 files: colors.go, badges.go, borders.go |
| 1.10 | Mock provider engine — Sinusoidal CPU/memory/thread data when `--mock` flag is active | ✅ Deterministic seed, capped leak simulation |
| 1.11 | Window resize handling — `tea.WindowSizeMsg` for responsive layout | ✅ Dynamic width-based layout |
| 1.12 | Command footer / keybindings — Navigation hints, shortcuts | ✅ Footer with all keybinding hints |

---

## Phase 2: Android Subsystem Integration (Weeks 3-4)

| #  | Task | Status |
|----|------|--------|
| 2.1 | `internal/platform/android/discovery.go` — Parse `adb devices -l` output for device discovery | ✅ Parses serial, product, model, transport; filters offline/unauthorized; emulator detection |
| 2.2 | `internal/platform/android/process.go` — Parse `adb shell ps` / for app/PID mapping | ✅ Field-based ps parser, kernel thread filtering, BuildType via dumpsys package |
| 2.3 | `internal/platform/android/telemetry.go` — Poll CPU via `adb shell top -n 1`, memory via `/proc/<pid>/status` | ✅ CPU parser, VmRSS parser, threads parser, combined Sample() method |
| 2.4 | `internal/platform/android/buildinfo.go` — Detect debug/release via `adb shell dumpsys package` | ✅ Merged into process.go — `parseBuildType` detects DEBUGGABLE flag |
| 2.5 | `internal/platform/android/preflight.go` — Validate `adb` in PATH, device reachability, connection health | ✅ Default path list, version parser, ValidateDevice health check |
| 2.6 | Long-lived ADB shell connection — Persistent `adb shell` pipe instead of per-sample process spawns | ✅ Persistent pipe with ensureShell/execInShell/closeShell, automatic restart on failure, fallback to one-shot adbExec |

---

## Phase 3: iOS Subsystem Integration (Weeks 5-6)

| #  | Task | Status |
|----|------|--------|
| 3.1 | `internal/platform/ios/discovery.go` — Parse `xcrun simctl list` for simulator + `xcrun devicectl list` for physical devices | ✅ `simctl` + `devicectl` discovery |
| 3.2 | `internal/platform/ios/process.go` — Resolve bundle IDs and PIDs on booted simulators via `launchctl list` / `ps -A` | ✅ Simulator process mapping |
| 3.3 | `internal/platform/ios/telemetry.go` — Poll via `simctl spawn top -l 1 -n 1 -pid <PID>` with PS fallback | ✅ Simulator telemetry via top/ps; physical devices limited by sandbox |
| 3.4 | `internal/platform/ios/buildinfo.go` — Detect debug/release from `.app` entitlements or Info.plist | ✅ Entitlements + _CodeSignature checks |
| 3.5 | Wire into `cmd/perfmon/main.go` — Auto-detect Android → fallback to iOS on macOS; `--ios` flag to force iOS mode | ✅ Auto-fallback + `--ios` flag |

---

## Phase 4: Export & Reporting Subsystem (Weeks 7-8)

| #  | Task | Status |
|----|------|--------|
| 4.1 | `internal/export/export.go` — Orchestrator: format dispatcher, `ResolveOutputPath`, `EnsureOutputDir`, `Export` | ✅ Format dispatch + path resolution |
| 4.2 | `internal/export/json.go` — JSON schema exporter (matching PRD schema) | ✅ `ExportJSON` with metadata + metrics + telemetry |
| 4.3 | `internal/export/markdown.go` — Markdown template with stats table + block area charts | ✅ Uses `chart.RenderCPUChart` / `RenderMemoryChart` |
| 4.4 | `internal/export/html.go` — Embedded HTML with inline CSS + chart.js renderer | ✅ `templates/chart.js` embedded; matches TUI block charts |
| 4.5 | `internal/export/pdf.go` — PDF export using `go-pdf/fpdf` with vector line charts | ✅ `ExportPDF` with multi-page vector line graphs |
| 4.6 | Static asset embedding — `//go:embed` for CSS, styles | ✅ `templates/style.css` embedded in binary |
| 4.7 | TUI keybindings — `e` for JSON, `Shift+E` for Markdown, `Ctrl+E` for HTML | ✅ Wired into TUI model |
| 4.8 | CLI `--export` flag — Non-interactive export mode | ✅ `--export json|md|html|pdf` with 10 samples |
| 4.9 | Unit tests — 35 tests covering all format generators + utilities | ✅ Pass with `-race` |

---

## Phase 5: Polish & Release Engineering

| #  | Task | Status |
|----|------|--------|
| 5.1 | Keyboard shortcut system — TAB-switch, `e` export, `q` quit, `↑/↓` navigate, `/` search | ✅ Implemented in TUI model (TAB, arrows, q, Ctrl+C, resize) |
| 5.2 | Performance optimization — Profiling loop <2% host CPU overhead | ✅ Measured 0–0.1% CPU overhead, well under target |
| 5.3 | Binary stripping & size check — `go build -ldflags="-s -w"`, verify <20MB target | ✅ `make build` configured with `-ldflags="-s -w"`, 5.5MB darwin/arm64 (max 6.1MB windows) |
| 5.4 | Pre-flight setup wizard — Detect missing `adb`, offer guided install | ✅ Interactive wizard with Homebrew install, retry, mock/iOS fallback |
| 5.5 | Comprehensive test suite — Unit tests for engine, mock provider, platform parsers | ✅ **181+ tests** across all packages: engine (27), mock (15), android (59), ios (34), export (35), adb integration (13) |
| 5.6 | Documentation — README, CLI `--help` output, architecture docs | ✅ README.md + 4 docs in `docs/`: plan, architecture, checklist, CLI reference |
| 5.7 | GitHub Release workflow — Automated releases with GoReleaser | ✅ `.goreleaser.yml` + CI + post-build hook |

---

## Environment Prerequisites

| Requirement | Status | Notes |
|-------------|--------|-------|
| Go (>=1.22) | ✅ Go 1.26.3 | Installed via Homebrew |
| adb | ✅ Available | `~/Library/Android/sdk/platform-tools/adb` v36.0.0 |
| xcrun | ✅ Available | `/usr/bin/xcrun` v72 |
| macOS (development) | ✅ | Darwin arm64 |
