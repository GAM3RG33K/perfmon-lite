# AI Conversation Summary — perfmon-lite

> **Project:** perfmon — Mobile Performance Monitor & Profiler
> **Date:** 2026-05-14 to 2026-05-18
> **Total Estimated Tokens:** ~785K sent / ~920K received

---

## Milestone 1: Initial Code Review & Gap Analysis

**Tokens sent:** ~45K / **Tokens received:** ~38K

- Reviewed PRD, architecture docs, development plan, and checklist
- Identified 40 gaps across 11 categories (critical, high, medium, low)
- Created `docs/GAP_TO_FILL.md` with prioritized issues
- Key findings: version injection bug (`const` vs `var`), ADB pipe deadlock, missing exit codes, env vars not wired

---

## Milestone 2: Release Infrastructure & CI Pipeline

**Tokens sent:** ~62K / **Tokens received:** ~55K

- Created `.goreleaser.yml` v2 config with post-build smoke test hook
- Fixed version injection: `const version` → `var version` for linker override
- Fixed post-build hook to skip --version check on cross-compiled binaries
- Added CI job chaining: `build-and-test → cross-build → release`
- Gated expensive jobs (cross-build, ADB integration) to run only on PRs/tags
- Added `FORCE_JAVASCRIPT_ACTIONS_TO_NODE24` for Node.js 24 compatibility
- Pinned goreleaser version to `'~> v2'` to suppress warnings

---

## Milestone 3: Pre-flight Wizard & CLI Restructure

**Tokens sent:** ~58K / **Tokens received:** ~52K

- Implemented interactive pre-flight setup wizard (ADB detection + brew install)
- Removed `perfmon export` and `perfmon version` subcommands
- Renamed `--target`/`-t` to `--device`/`-d`
- Removed `--ios` flag; replaced with auto-detect both platforms
- Removed `--export pdf` option from CLI
- Added `autoDetectProvider()` for cross-platform device selection
- Added `--id` flag for targeting apps by package/bundle name
- Added `uninstall` subcommand + `runUninstall()` function
- Added `update` subcommand + `runUpdate()` function

---

## Milestone 4: Cross-Platform & Windows Support

**Tokens sent:** ~52K / **Tokens received:** ~48K

- Added Windows ADB paths (`AppData\Local\Android\Sdk`) to `FindAdbPath()`
- Added `PERFMON_ADB_PATH` env var support
- Created `install.ps1` and `update.ps1` PowerShell scripts
- Created `uninstall.ps1` for Windows
- Renamed Windows binary to `perfmon-tool.exe` (avoided conflict with built-in Windows Performance Monitor)
- Added windows/arm64 builds to GoReleaser and CI matrix
- PowerShell scripts detect ARM64 and fallback to x64 binary via emulation
- Fixed Windows TUI crash: retry without mouse support on terminal error

---

## Milestone 5: CPU/RAM Telemetry Overhaul

**Tokens sent:** ~48K / **Tokens received:** ~55K

- Researched Android CPU sampling approaches (web research on `/proc/pid/stat` vs `top`)
- Replaced unreliable `top -n 1 -b` parsing with `/proc/<pid>/stat` tick delta
- Implemented `parseCPUStat()` with utime+stime tick delta calculation
- CPU% = `(delta_ticks / CLK_TCK / elapsed_seconds) * 100`
- Added `prevPID` tracking to reset baseline on PID changes
- Fixed `/proc/pid/stat` field index bug (off-by-2 in parser)
- Rewrote iOS simulator telemetry: removed non-existent `top`/`ps` inside sandbox
- iOS now uses macOS host `ps -p <pid> -o %cpu,rss` (simulator processes are host-visible)
- Added CPU delta tracking fields to `ADBProvider` struct
- Fixed export sample interval from 100ms to 500ms for real device CPU accumulation

---

## Milestone 6: Process Selection & App Targeting

**Tokens sent:** ~35K / **Tokens received:** ~32K

- Improved `selectBestProcess()`: filter system prefixes (`com.android.`, `com.google.`, `media.`, `zygote`)
- Added `hasAnyPrefix()` helper
- Prioritize non-`com.` package names (e.g. `in.thetatva.tatva` over `com.instagram.android`)
- Fixed export metadata: look up selected PID's name instead of using `processes[0]`
- `selectedProcess()` in TUI falls back to `AppID` when process not found
- `launchAppPrompt()`: interactive prompt asking user to launch non-running app
- After user confirms launch, re-discovers processes to find the new PID
- Android launch uses `monkey -p` with fallback to `am start`
- iOS launch uses `xcrun simctl launch`

---

## Milestone 7: TUI UX Overhaul

**Tokens sent:** ~68K / **Tokens received:** ~82K

- Added status bar notifications that auto-clear after 3 seconds
- Replaced fullscreen help overlay with categorized keybinding sections
- Fixed title badge: always showed `PlatformMock`, now shows actual platform
- Small centered export format modal (replaces fullscreen overlay)
- `renderFormatPicker()` overlays on top of normal TUI view
- `renderMainView()` extracted for reuse
- Removed Threads/Processes tab (simplified to dashboard + logs layout)
- Bottom log console panel (always visible, shows last 5 lines)
- `RenderRecent(5)` for compact log display
- Replaced vertical bar sparklines with half-block line charts (`█▓▄` Unicode)
- `renderLineChart()` with filled area underneath
- Auto-scaling Y-axis based on data max + 20% headroom
- Charts capped at `maxChartPoints = 100` for readability
- Website terminal mockup updated to match new TUI layout

---

## Milestone 8: Stack Traces & Log Capture

**Tokens sent:** ~55K / **Tokens received:** ~62K

- Added `Stack` field to `TelemetrySnapshot` struct
- When CPU > 50%, fetches kernel stack trace
- Android: reads `/proc/<pid>/stack` via persistent ADB shell pipe
- iOS/macOS: uses built-in `sample` command for user-space stacks
- Added `LogCapturer` interface to `engine/targets.go`
- Android `CaptureLogs()`: `adb shell logcat -d --pid=<pid>` via persistent pipe
- iOS `CaptureLogs()`: `xcrun simctl spawn log stream --last 2m` filtered by PID
- TUI runs `logCaptureCmd` every 2 seconds
- Captured logs added to system logs with `[APP]` tag
- Per-tick CPU/RAM logged to system logs with `[TICK]` tag
- High CPU (>70%) triggers `[ALERT]` with stack trace
- High RAM (>500MB) triggers `[ALERT]`

---

## Milestone 9: Export Reports Enhancement

**Tokens sent:** ~42K / **Tokens received:** ~48K

- Added `Logs []string` field to `ExportData` struct
- Added `Logs` field to `Options` struct
- `BuildExportData()` now copies logs into export data
- Markdown: stack column in telemetry table, expandable stack sections, captured app logs section
- HTML: collapsible details/summary for stack traces, captured app logs section, high CPU row highlighting (`td.alert`)
- CSS: styles for `.stack-section`, `.stack-details`, `.stack-summary`, `.stack-pre`, `.stack-badge`
- Fixed HTML template: added `inc` function for 1-based sample numbering
- Auto-export logs on quit (`q`/`Ctrl+C`) to `perfmon_logs_<timestamp>.log`
- Removed manual `l` keybinding for log export

---

## Milestone 10: Dynamic Versioning & Release Management

**Tokens sent:** ~38K / **Tokens received:** ~35K

- Created `VERSION` file as single source of truth
- Makefile injects version via `-X main.version=$(cat VERSION)`
- GoReleaser uses `{{ .Version }}` from git tag
- Web workflow reads `VERSION` file, passes as `VITE_APP_VERSION`
- React app reads `import.meta.env.VITE_APP_VERSION`
- Added commit hash: `-X main.commit=$(git rev-parse --short HEAD)`
- `--version` now shows: `perfmon-tool v0.0.5 (abc1234)`
- `scripts/release.sh` reads version from `VERSION` file
- `make retag` and `make cut-release` targets

---

## Milestone 11: Website & Cloudflare Deployment

**Tokens sent:** ~72K / **Tokens received:** ~85K

- Created `web/` React + Vite landing page with dark terminal theme
- Animated typewriter terminal, live CPU/RAM sparkline charts, scroll reveals
- parallax effects: floating code symbols, mouse-tracking 3D terminal
- Install tab switcher (macOS/Linux + Windows)
- Copy-to-clipboard button for install commands
- Manual download table with platform-specific asset names
- Set up Cloudflare Worker on `perfmon.qzz.io` with redirect routing
- Moved install scripts to `get.perfmon.qzz.io` subdomain
- Connected domain to Cloudflare nameservers, DNS, and Workers
- Real device name detection from `navigator.userAgent`
- Real total RAM from `navigator.deviceMemory`
- Smoother live metrics (slow drift, not random jumps)
- Responsive mobile layout (font shrink, scrollable terminal, single-column features)
- GitHub Actions workflow (`web.yml`) builds and deploys to GitHub Pages
- Web deploy runs on both tag pushes and main branch pushes

---

## Milestone 12: Asset Naming & Binary Distribution

**Tokens sent:** ~28K / **Tokens received:** ~25K

- Renamed release assets to `perfmon-tool` naming convention
- Final naming: `perfmon-tool-{version}-linux-{arch}`, `perfmon-tool-{version}-darwin-{arch}`, `perfmon-tool-{version}-windows-{arch}.exe`
- Removed unsupported `universal_binary` (GoReleaser v2 limitation)
- Updated `install.sh`, `install.ps1`, `update.sh`, `update.ps1` for new asset names
- Updated `perfmon update` command for new asset URLs
- Updated all scripts to use `perfmon-tool` as final binary name on all platforms
- Created `USAGE.md` for release distribution

---

## Milestone 13: Documentation & Final Polish

**Tokens sent:** ~42K / **Tokens received:** ~38K

- Updated README with `perfmon-tool` binary name, new TUI mockup, download table
- Updated CLI reference with all flag changes, exit codes, environment variables
- Created `docs/domain-setup.md` for Cloudflare configuration
- Created `docs/AI_CONVERSATION_SUMMARY.md` (this file)
- Updated `docs/architecture.md` with current binary size and CPU overhead
- Version tracking: v0.0.1 (beta) → v0.0.2 → v0.0.3 → v0.0.4 → v0.0.5
- All doc references synced across README, CLI reference, USAGE.md, and landing page

---

## Final Technical Summary

**Last Commit:** `3594f59` — Fix HTML template: add 'inc' function, replace 'add $i 1' in stack traces
**Latest Tag:** `v0.0.5`
**Total Commits (session):** ~65
**Files Changed (session):** ~85+
**Tests:** All 183+ tests pass with `-race`

### Architecture Overview

```
perfmon-tool [--device <id>] [--id <app>] [--mock]
  ├── Auto-detect: Android → iOS → mock wizard
  ├── Telemetry Engine (ring buffer, 1s poll, 100 chart pts)
  │   ├── Android: /proc/<pid>/stat (CPU) + /proc/<pid>/status (RAM)
  │   ├── iOS Sim: macOS ps (host-level)
  │   └── Stack traces on high CPU: /proc/pid/stack (Android) / sample (macOS)
  ├── Log Capture (2s interval)
  │   ├── Android: adb logcat --pid=<pid>
  │   └── iOS: xcrun simctl spawn log stream
  ├── TUI (Bubble Tea)
  │   ├── Dashboard (line charts, status, stats)
  │   ├── Bottom log panel (5 most recent lines)
  │   └── Help overlay, export modal, auto-export on quit
  └── Export (JSON, MD, HTML, PDF)
      ├── Telemetry data + stack traces
      ├── Captured app logs
      └── HTML: collapsible sections, alert highlights
```

### Distribution

| Platform | Asset | Install |
|----------|-------|---------|
| macOS (Intel/ARM) | `perfmon-tool-{ver}-darwin-{arch}` | `curl -sfL https://get.perfmon.qzz.io \| bash` |
| Linux (amd64/arm64) | `perfmon-tool-{ver}-linux-{arch}` | `curl -sfL https://get.perfmon.qzz.io \| bash` |
| Windows (amd64/arm64) | `perfmon-tool-{ver}-windows-{arch}.exe` | `iwr https://get.perfmon.qzz.io/windows -useb \| iex` |

### Hosting

| Domain | Purpose | Provider |
|--------|---------|----------|
| `perfmon.qzz.io` | Landing page (React) | GitHub Pages |
| `get.perfmon.qzz.io` | Script redirects | Cloudflare Worker |

### Token Estimate Notes

Token counts are estimates based on conversation duration, file sizes changed per milestone, and message complexity. Actual token usage may vary. Calculated at approximately 4 chars per token for code, 5 chars per token for prose.
