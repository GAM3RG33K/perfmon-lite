# perfmon — Development Checklist

> Use this checklist to track progress through each phase.
> Mark items as `[x]` when complete.

---

## Prerequisites

- [x] Go >=1.22 installed (`go version`) — Go 1.26.3 via Homebrew
- [x] Project directory initialized with `go mod init github.com/GAM3RG33K/perfmon-lite`

---

## Phase 0: Scaffolding

- [x] `cmd/perfmon/main.go` created
- [x] `internal/` package directories created
- [x] `go.mod` with bubbletea, lipgloss dependencies
- [x] `Makefile` with targets: `build`, `run`, `mock`, `test`, `lint`, `clean`
- [x] `.gitignore` (Go standard)
- [ ] `.goreleaser.yml` for multi-arch releases
- [ ] GitHub Actions workflow (`.github/workflows/ci.yml`)

---

## Phase 1: Core Engine + Mock TUI

### Engine
- [x] `internal/engine/types.go` — Domain types
- [x] `internal/engine/engine.go` — Scheduler + ring buffer
- [x] `internal/engine/targets.go` — Shared interfaces
- [x] `internal/platform/mock/mock.go` — Mock telemetry provider

### Chart (shared renderer)
- [x] `internal/chart/chart.go` — Btop-style block area charts, Catmull-Rom smoothing, gauges
- [x] `internal/chart/chart_test.go` — Chart renderer unit tests

### TUI
- [x] `internal/tui/model.go` — Core Bubble Tea model
- [x] `internal/tui/views/dashboard.go` — CPU and memory area charts
- [x] `internal/tui/views/area_chart.go` — Lipgloss-colored chart wrapper
- [x] `internal/tui/views/layout.go` — Responsive line truncation helpers
- [x] `internal/tui/views/target_selector.go` — Device/process list
- [x] `internal/tui/views/logs.go` — System logs
- [x] `internal/tui/styles/colors.go` — Color constants
- [x] `internal/tui/styles/badges.go` — Debug/Release badges
- [x] `internal/tui/styles/borders.go` — Panel borders
- [x] Window resize handling implemented
- [x] Command footer with keybindings
- [x] `--mock` flag produces live sinusoidal telemetry in TUI

---

## Phase 2: Android Subsystem

- [x] `internal/platform/android/discovery.go` — `adb devices -l` parser
- [x] `internal/platform/android/process.go` — `adb shell ps` parser + BuildType detection
- [x] `internal/platform/android/telemetry.go` — `top` CPU + `/proc/pid/status` memory/threads parsers
- [x] `internal/platform/android/buildinfo.go` — Debug/Release detection via dumpsys package (merged into process.go)
- [x] `internal/platform/android/preflight.go` — ADB health check, version parsing, device validation
- [x] Long-lived ADB pipe connection (peristent `adb shell`, ensureShell/execInShell/closeShell, auto-restart on failure)
- [x] End-to-end: auto-discover Android device, select best process, pipe-based telemetry in TUI
- [x] `internal/platform/android/provider.go` — ADBProvider struct, adb helper, SetDevice, interface compliance
- [x] `internal/platform/android/provider_test.go` — 50 tests covering discovery, process, telemetry, preflight, adb errors
- [x] `internal/platform/android/pipe_test.go` — 11 tests covering ensureShell, execInShell, pipe restart, fallback, concurrency

---

## Phase 3: iOS Subsystem

- [x] `internal/platform/ios/discovery.go` — `xcrun simctl list` + `devicectl list` parser
- [x] `internal/platform/ios/process.go` — Bundle ID/PID resolution via launchctl list / ps
- [x] `internal/platform/ios/telemetry.go` — Metric polling via top / ps fallback
- [x] `internal/platform/ios/buildinfo.go` — Debug/Release detection via entitlements
- [x] `internal/platform/ios/preflight.go` — xcrun path detection, version check, xcode-select
- [x] `internal/platform/ios/provider.go` — iOSProvider struct with interface compliance
- [x] Wire into main.go: auto-detect Android → iOS fallback on macOS; `--ios` flag
- [x] End-to-end: select iOS simulator → see live telemetry in TUI

---

## Phase 4: Export Subsystem

- [x] `internal/export/types.go` — ExportData, Options, BuildExportData
- [x] `internal/export/export.go` — Format dispatcher + path resolution
- [x] `internal/export/json.go` — JSON export (PRD schema v1)
- [x] `internal/export/markdown.go` — Markdown report with block area charts + tables
- [x] `internal/export/html.go` — HTML export with embedded chart.js + CSS
- [x] `internal/export/templates/chart.js` — Client-side chart renderer for HTML export
- [x] `web/src/tuiChart.js` — Landing-page chart demo (matches TUI renderer)
- [x] `internal/export/pdf.go` — PDF export with vector line graphs (go-pdf/fpdf)
- [x] `internal/export/templates/style.css` — Dark-theme CSS (`//go:embed`)
- [x] `internal/export/export_test.go` — 35 unit tests covering all formats
- [x] TUI keybindings: `e` → JSON, `Shift+E` → Markdown, `Ctrl+E` → HTML
- [x] CLI `--export` flag for non-interactive export mode

---

## Phase 5: Polish & Release

- [x] Keyboard shortcuts: TAB, arrows, `q`/`Ctrl+C`, resize handling
- [x] Host CPU overhead <2% verified (measured 0–0.1%)
- [x] Binary stripped, <20MB confirmed (5.5MB darwin/arm64, max 6.1MB windows/amd64)
- [x] Pre-flight setup wizard (interactive ADB detection + guided install)
- [x] Unit tests cover engine (27 tests)
- [x] Unit tests cover mock provider (15 tests + 1 benchmark)
- [x] Unit tests cover platform parsers (Android 59, iOS 34, Export 35)
- [x] README with installation, usage, examples
- [x] CLI `--help` is comprehensive
- [x] GitHub Releases workflow (`.goreleaser.yml` + CI + post-build hook)
- [x] Multi-arch dry-run build passes (5 platforms, all <6.5MB)

---

## Progress Tracking

| Phase | Total Tasks | Done | % |
|-------|-------------|------|---|
| 0: Scaffolding | 7 | 7 | 100% |
| 1: Engine + TUI | 12 | 12 | 100% |
| 2: Android | 9 | 9 | 100% |
| 3: iOS | 8 | 8 | 100% |
| 4: Export | 10 | 10 | 100% |
| 5: Polish | 11 | 11 | 100% |
| **Total** | **57** | **57** | **100%** |
