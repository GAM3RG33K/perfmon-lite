# perfmon — Development Checklist

> Use this checklist to track progress through each phase.
> Mark items as `[x]` when complete.

---

## Prerequisites

- [x] Go >=1.22 installed (`go version`) — Go 1.26.3 via Homebrew
- [x] Project directory initialized with `go mod init github.com/w1n/perfmon`

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

### TUI
- [x] `internal/tui/model.go` — Core Bubble Tea model
- [x] `internal/tui/views/dashboard.go` — CPU, memory, thread charts
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

- [ ] `internal/platform/android/discovery.go` — `adb devices -l` parser
- [ ] `internal/platform/android/process.go` — `adb shell ps` parser
- [ ] `internal/platform/android/telemetry.go` — `dumpsys meminfo` + `top` parsers
- [ ] `internal/platform/android/buildinfo.go` — Debug/Release detection
- [ ] `internal/platform/android/preflight.go` — ADB health check
- [ ] Long-lived ADB pipe connection (instead of per-sample exec)
- [ ] End-to-end: select Android device → see live telemetry in TUI

---

## Phase 3: iOS Subsystem

- [ ] `internal/platform/ios/discovery.go` — `xcrun simctl list` parser
- [ ] `internal/platform/ios/process.go` — Bundle ID/PID resolution
- [ ] `internal/platform/ios/telemetry.go` — Metric polling
- [ ] `internal/platform/ios/buildinfo.go` — Debug/Release detection
- [ ] End-to-end: select iOS simulator → see live telemetry in TUI

---

## Phase 4: Export Subsystem

- [ ] `internal/export/generator.go` — Format selection + file writing
- [ ] `internal/export/templates/export.json` — JSON export
- [ ] `internal/export/templates/export.md` — Markdown export
- [ ] `internal/export/templates/export.html` — HTML export with inline CSS
- [ ] PDF export with `go-pdf/gopdf`
- [ ] Static assets embedded via `//go:embed`

---

## Phase 5: Polish & Release

- [x] Keyboard shortcuts: TAB, arrows, `q`/`Ctrl+C`, resize handling
- [ ] Host CPU overhead <2% verified
- [ ] Binary stripped, <20MB confirmed (~4.7MB current build)
- [ ] Pre-flight setup wizard (`adb` detection + guided install)
- [x] Unit tests cover engine (12 ring buffer + 8 Engine + 7 MetricsSummary = 27 tests)
- [x] Unit tests cover mock provider (15 tests + 1 benchmark)
- [ ] Unit tests cover platform parsers
- [ ] README with installation, usage, examples
- [x] CLI `--help` is comprehensive
- [ ] GitHub Releases workflow working
- [ ] Multi-arch dry-run build passes

---

## Progress Tracking

| Phase | Total Tasks | Done | % |
|-------|-------------|------|---|
| 0: Scaffolding | 7 | 5 | 71% |
| 1: Engine + TUI | 12 | 12 | 100% |
| 2: Android | 6 | 0 | 0% |
| 3: iOS | 4 | 0 | 0% |
| 4: Export | 6 | 0 | 0% |
| 5: Polish | 10 | 4 | 40% |
| **Total** | **45** | **21** | **47%** |
