# perfmon — Gap Analysis for v1.0.0 Release

> **Audit Date:** 2026-05-14
> **Scope:** PRD requirements vs. implementation, cross-platform compatibility, error handling, release infrastructure, documentation accuracy

---

## 1. Release & Build Infrastructure

### 1.1 GoReleaser — Version Injection Mismatch
**Severity:** ~~Critical~~ → **Fixed**
**File:** `.goreleaser.yml`, `cmd/perfmon/main.go`
**Status:** Resolved — `version` is already declared as `var`, not `const`. Linker flags work correctly. Verified: `go build -ldflags="-X main.version=1.0.0"` produces correct version output.

### 1.2 GoReleaser — No Homebrew Tap Configuration
**Severity:** Medium
**File:** `.goreleaser.yml`
**Issue:** `README.md` says "Coming soon — brew tap", but GoReleaser has no `brews` section to auto-publish a Homebrew formula. Users on macOS can't `brew install perfmon`.

### 1.3 GoReleaser — Windows Binary Archive Extension
**Severity:** Low
**File:** `.goreleaser.yml:32`
**Issue:** Archive `name_template` doesn't differentiate `.exe` for Windows binaries. GoReleaser handles this internally, but if someone downloads `perfmon_windows_amd64` and tries to run it, it won't work without the `.exe` extension.
**Fix:** Use GoReleaser's `{{ .Ext }}` variable or ensure the release notes clarify the Windows binary name.

### 1.4 CI — Artifact Upload Only on `main` Branch
**Severity:** Low
**File:** `.github/workflows/ci.yml:45`
**Issue:** `if: github.ref == 'refs/heads/main'` — Artifacts only upload on pushes to main. PR builds don't produce downloadable binaries for testing.

---

## 2. Critical Crash / Runtime Bugs

### 2.1 `setupiOSProvider` Calls `os.Exit(1)` on Non-macOS
**Severity:** ~~High~~ → **Fixed**
**File:** `cmd/perfmon/main.go`
**Status:** Resolved — `setupiOSProvider()` no longer exists. The `tryiOSProvider()` function returns errors gracefully and the auto-detect flow handles "iOS not available" without hard-crashing.

### 2.2 `log.Fatalf` in Provider Setup Prevents Graceful Degradation
**Severity:** ~~Medium~~ → **Fixed**
**Files:** `cmd/perfmon/main.go`
**Status:** Resolved — only one `log.Fatalf` remains at line 263, which is for TUI startup failure after both mouse and non-mouse attempts. Provider setup uses error returns, not fatal exits.

### 2.3 Pre-flight Wizard Retry Ignores Errors
**Severity:** ~~Medium~~ → **Fixed**
**File:** `cmd/perfmon/main.go:467-486`
**Status:** Resolved — retry checks `aErr == nil` before assigning provider. If retry fails, falls through to iOS fallback. If all platforms fail, exits with `exitToolError` (code 3).

### 2.4 ADB Shell Pipe — No Read Timeout
**Severity:** ~~High~~ → **Fixed**
**File:** `internal/platform/android/provider.go:15-16,187-189`
**Status:** Resolved — `shellReadTimeout = 10 * time.Second` with `time.After()` in `execInShell()`. Pipe marks itself dead on timeout and caller falls back to one-shot `adbExec`.

### 2.5 ADB Shell Pipe — Busy-wait in `flushShell`
**Severity:** Low
**File:** `internal/platform/android/provider.go:133-151`
**Issue:** `flushShell()` busy-waits for up to 1 second (10ms sleep in a tight loop). While not catastrophic, this adds unnecessary CPU wake-ups during startup.

### 2.6 Android Sample Parsing — Delimiter Assumption
**Severity:** Medium
**File:** `internal/platform/android/telemetry.go:64`
**Issue:** `Sample()` splits output on `===MEM===\n`. If the `cat /proc/pid/status` output doesn't end with exactly `\n` (e.g., last line has no newline), or if the delimiter string appears naturally in `top` output, parsing fails or produces corrupted data.

### 2.7 Engine Poll Blocks TUI Render Loop
**Severity:** ~~High~~ → **Fixed**
**File:** `internal/tui/model.go:235-240`
**Status:** Resolved — `handleTick()` returns `func() tea.Msg { return m.Engine.Poll() }` as a Bubble Tea Cmd, which runs in a goroutine. The TUI remains responsive during polling.

### 2.8 Process PID Recycling — Silent Wrong Data
**Severity:** ~~Medium~~ → **Fixed**
**File:** `internal/platform/android/telemetry.go:176-180`
**Status:** Resolved — CPU delta calculation resets baseline if `deltaTicks > totalTicks` (indicating PID recycling) or `elapsed <= 0`. Returns 0% CPU and resets tracking state.

---

## 3. Cross-Platform / Windows Compatibility

### 3.1 `adb` Discovery Ignores Windows SDK Paths
**Severity:** ~~High~~ → **Fixed**
**File:** `internal/platform/android/preflight.go`
**Status:** Resolved — Windows paths are included via `os.UserHomeDir()` → `AppData\Local\Android\Sdk`, and both `adb` and `adb.exe` binaries are searched.

### 3.2 Mouse Support Unavailable on Some Terminals
**Severity:** Low
**File:** `cmd/perfmon/main.go:240`
**Issue:** `tea.WithMouseCellMotion()` enables mouse support. Some terminals (SSH sessions, Windows Command Prompt without VT, tmux without mouse config) will either ignore this or produce escape sequence artifacts.

### 3.3 ANSI / Lipgloss Rendering on Windows
**Severity:** Medium
**File:** `internal/tui/styles/*.go`
**Issue:** Lipgloss uses ANSI escape sequences extensively. Windows terminals prior to Windows Terminal (or without VT escape sequence support enabled) will show raw escape codes instead of styled output.

### 3.4 No `SIGTERM` / `SIGINT` Handler for Pipe Cleanup
**Severity:** Low
**File:** `cmd/perfmon/main.go`
**Issue:** If the process receives `SIGTERM` (e.g., kill from CI/CD or system shutdown), the persistent ADB shell pipe may not be cleaned up. Bubble Tea handles `SIGINT` via `tea.Quit`, but `SIGTERM` has no handler.

---

## 4. iOS Provider Gaps

### 4.1 `Close()` is a No-op
**Severity:** Low
**File:** `internal/platform/ios/provider.go:83-85`
**Issue:** `Close()` returns `nil` unconditionally. No resources are managed. If future changes add persistent connections or temp files, they'll leak.

### 4.2 No Persistent Shell Pipe (Unlike Android)
**Severity:** Medium
**File:** `internal/platform/ios/provider.go`, `telemetry.go`
**Issue:** Each iOS telemetry sample spawns a new `xcrun simctl spawn` process. This is significantly higher latency than Android's persistent pipe approach. The PRD mitigation (risk #1) was implemented for Android but not iOS.

### 4.3 Physical iOS Device Limitations Not User-facing
**Severity:** Medium
**File:** `internal/platform/ios/telemetry.go`
**Issue:** The PRD explicitly warns about "iOS Hardware Sandbox Restrictions" (risk #2), but there's no user-facing warning when connecting to a physical iOS device. Users will silently get no data or errors without understanding why.

### 4.4 No Integration Tests
**Severity:** Medium
**File:** (missing) `internal/platform/ios/*_integration_test.go`
**Issue:** Unlike Android (which has `adb_integration_test.go` with 13 tests), iOS has zero integration tests. No tests verify actual simulator interaction.

---

## 5. TUI / UX Gaps

### 5.1 No Loading / Error States for Empty Data
**Severity:** ~~Medium~~ → **Fixed**
**File:** `internal/tui/views/dashboard.go:48-50`
**Status:** Resolved — `renderHeader()` shows "No data — waiting for telemetry..." when latest is nil. Chart renderers show "CPU: waiting for data..." when insufficient data.

### 5.2 `Model.Err` Never Assigned
**Severity:** Low
**File:** `internal/tui/model.go:55,283-285`
**Issue:** `m.Err` is checked in `View()` to display errors, but it's never set anywhere in the code. The error display path is dead code.

### 5.3 Tab Arrow Key Conflicts
**Severity:** Low
**File:** `internal/tui/model.go:163-165`
**Issue:** `←` / `→` always switch tabs, even when the user is focused on a selectable list (e.g., Threads tab where up/down navigates items). There's no way to prevent tab switching while browsing a list.

### 5.4 Export Keybinding — Docs vs. Code Inconsistency
**Severity:** ~~Low~~ → **Fixed**
**File:** `internal/tui/model.go:184-191`, `docs/cli-reference.md:169-171`
**Status:** Resolved — `e` opens format picker modal (JSON/Markdown/HTML), `Shift+E` exports Markdown, `Ctrl+E` exports HTML. CLI reference updated to match.

### 5.5 No Help Overlay
**Severity:** ~~Low~~ → **Fixed**
**File:** `internal/tui/model.go:435-474`
**Status:** Resolved — `?` toggles `m.ShowHelp` which renders a full-screen help overlay via `renderHelp()`. Closes on `?`, `Esc`, `Enter`, or `q`.

### 6.1 `ResolveOutputPath` Generates Colliding Filenames
**Severity:** Low
**File:** `internal/export/export.go:31-36`
**Issue:** With an empty `OutputPath`, the function generates `perfmon_export_<N>` where N is the snapshot count. Two exports with the same buffer size overwrite each other. No timestamp is included.

### 6.2 `.json` Schema URL is Inaccessible
**Severity:** Low
**File:** `internal/export/types.go:63`
**Issue:** The JSON export includes `"$schema": "https://perfmon.qzz.io/schemas/export-v1.json"` but this domain/URL doesn't resolve. Consumers relying on schema validation will get 404s.

### 6.3 PDF Colors May Print Poorly
**Severity:** Low
**File:** `internal/export/pdf.go:20-21,85-86,107-108`
**Issue:** PDF uses bright display-optimized colors (cyan `0,255,255`, magenta `255,0,255`, green `0,255,0`). These don't print well on white paper. No print-friendly mode.

### 6.4 Export Timestamp Uses Local Time
**Severity:** ~~Low~~ → **Fixed**
**File:** `internal/export/types.go:65`
**Status:** Resolved — `GeneratedAt` uses `time.Now().UTC().Format(time.RFC3339)` which includes timezone offset (UTC).

---

## 7. Engine / Architecture Issues

### 7.1 GitHub Module Path Mismatch
**Severity:** ~~Medium~~ → **Fixed**
**File:** `go.mod` (path: `github.com/GAM3RG33K/perfmon-lite`)
**Status:** Resolved — module path updated to match repository. `go install github.com/GAM3RG33K/perfmon-lite/cmd/perfmon@latest` now works correctly.

### 7.2 `selectBestProcess` Android-specific Logic Used for iOS
**Severity:** Low
**File:** `cmd/perfmon/main.go:467-493`
**Issue:** `selectBestProcess()` filters out `android.` and `com.apple.` prefixed processes. For Android this makes sense. For iOS, filtering `com.apple.` is correct but the comment mentions "not a system daemon" in Android terms.

### 7.3 No Resource Limits on Ring Buffer
**Severity:** Low
**File:** `internal/engine/engine.go:104-111`
**Issue:** No upper limit enforcement on `capacity` or `interval` beyond CLI validation (min 10). A user could set `--buffer 10000000` and allocate ~1.6GB for the ring buffer.

---

## 8. Testing Gaps

### 8.1 TUI Has Zero Tests
**Severity:** High
**File:** `internal/tui/` (no test files)
**Issue:** No tests for the TUI model, views, or styles. All TUI packages show 0% coverage. Key logic (keyboard handling, tab switching, export flow) is untested.

### 8.2 No Export Format Validation Tests
**Severity:** Medium
**File:** `internal/export/export_test.go`
**Issue:** Tests verify that export functions return success but don't validate output file integrity (valid JSON syntax, valid PDF structure, valid HTML parsing).

### 8.3 No iOS Integration Tests
**Severity:** Medium
**File:** (missing) `internal/platform/ios/*_integration_test.go`
**Issue:** (duplicate of 4.4 — noted here for completeness)

### 8.4 No Persistent Pipe Stress Tests
**Severity:** Low
**File:** `internal/platform/android/pipe_test.go`
**Issue:** Pipe tests cover concurrency (10 concurrent calls) but don't test long-running scenarios (1000+ samples, device disconnection/reconnection, network latency spikes).

---

## 9. Documentation Inaccuracies

### 9.1 `devices` Subcommand Flags Not Implemented
**Severity:** ~~High~~ → **Fixed**
**File:** `cmd/perfmon/main.go:1014-1110`
**Status:** Resolved — `runDevices()` implements all three flags: `--json`, `--platform`, `--build-info`. Platform filtering works for `android`, `ios`, and `all`. JSON output with indent is supported. Build info fetches processes per device.

### 9.2 `--target -t` Shorthand Not Implemented
**Severity:** Medium
**File:** `docs/cli-reference.md:25`, `cmd/perfmon/main.go`
**Issue:** CLI reference lists `-t` as shorthand for `--target`, but main.go only defines `--target` with no shorthand flag.

### 9.3 Environment Variables Not Wired
**Severity:** ~~High~~ → **Fixed**
**File:** `cmd/perfmon/main.go`, `internal/platform/android/preflight.go`
**Status:** Resolved — all four env vars are wired:
- `PERFMON_ADB_PATH` — checked first in `FindAdbPath()` (preflight.go:28)
- `PERFMON_BUFFER_SIZE` — default for `--buffer` flag (main.go:47)
- `PERFMON_POLL_INTERVAL` — default for `--interval` flag (main.go:46)
- `PERFMON_EXPORT_DIR` — default for `--output` flag (main.go:49)

### 9.4 Exit Codes Not Implemented
**Severity:** ~~Medium~~ → **Fixed**
**File:** `cmd/perfmon/main.go:28-35`
**Status:** Resolved — all five exit codes are defined and used: `exitSuccess` (0), `exitGeneralError` (1), `exitDeviceError` (2), `exitToolError` (3), `exitExportError` (4).

### 9.5 Architecture Doc References Stale Binary Size
**Severity:** Low
**File:** `docs/architecture.md:242`
**Issue:** Says "~4.7 MB (arm64, unstripped)" — actual stripped size is 5.5MB. Minor but misleading.

### 9.6 Plan Doc Mentions Dumpsys Meminfo Not Used
**Severity:** Low
**File:** `docs/plan.md:50`, `PRD.md:198`
**Issue:** PRD phase 2 mentions `adb shell dumpsys meminfo`, but the implementation uses `/proc/<pid>/status` (VmRSS). The docs should reflect the actual approach.

---

## 10. Feature-level Gaps (vs. PRD)

### 10.1 F-01 Device Discovery — No Auto-refresh
**Severity:** Low
**File:** `internal/tui/model.go:168-169`
**Issue:** PRD says "Automatically scan, identify, and list attached devices." The `r` key refreshes, but there's no automatic polling for device hot-plug. If a device is connected after startup, the user must manually press `r`.

### 10.2 F-06 Pre-flight Wizard — No Inline Download
**Severity:** Medium
**File:** `cmd/perfmon/main.go:495-563`
**Issue:** PRD specifies "interactive, inline downloads to tool-specific cache directories." The current wizard only offers `brew install` or shows URL. It doesn't download platform-tools to a `.perfmon/cache` directory as specified.

### 10.3 F-02 Process Mapping — No Process Re-selection
**Severity:** Low
**File:** `internal/tui/model.go:163-166`
**Issue:** PRD expects dynamic process selection from UI. The `Enter` key on the Threads tab triggers `ShowProcesses = true`, but there's no way to re-select a different process once profiling has started.

### 10.4 F-04 Telemetry — No PSS (Proportional Set Size)
**Severity:** Low
**File:** `internal/platform/android/telemetry.go`
**Issue:** PRD mentions "Memory footprint (PSS/RSS in KB)". The implementation uses RSS (VmRSS) but never measures PSS. PSS is more accurate for shared memory but requires root or `adb shell dumpsys meminfo`.

---

## 11. Miscellaneous

### 11.1 `exportData()` Contains Stub Placeholder Text
**Severity:** ~~Low~~ → **Fixed**
**File:** `cmd/perfmon/main.go`
**Status:** Resolved — no `exportData()` function exists. Export subsystem is fully implemented with JSON, Markdown, HTML, and PDF formats.

### 11.2 TUI Title Always Shows Mock Badge
**Severity:** ~~Low~~ → **Fixed**
**File:** `internal/tui/model.go:341`
**Status:** Resolved — title uses `styles.PlatformBadge(m.Platform)` which reflects the actual platform (Android/iOS/Mock), not always mock.

### 11.3 No Graceful Shutdown on ADB Pipe Failure During Profiling
**Severity:** Medium
**File:** `internal/platform/android/telemetry.go:37-49`
**Issue:** If the ADB pipe fails mid-session (device unplugged, ADB crash), `Sample()` falls back to one-shot `adbExec`. If that also fails, it returns an error. But the error is logged and the engine continues polling forever, logging errors each second. No automatic re-discovery or graceful degradation.

---

## Severity Summary

| Severity | Count | Key Items |
|----------|-------|-----------|
| **Critical** | 0 | All critical issues resolved |
| **High** | 1 | No TUI tests (8.1) |
| **Medium** | 7 | Delimiter parsing (2.6); iOS limitations not user-facing (4.3); No iOS integration tests (4.4/8.3); Pre-flight no-download (10.2); ANSI on Windows (3.3); Export format validation (8.2); ADB pipe graceful shutdown (11.3) |
| **Low** | 14 | CI artifacts only on main (1.4); Busy-wait (2.5); Dead error path (5.2); PDF print colors (6.3); Schema URL dead (6.2); Colliding filenames (6.1); etc. |

> **Note:** Many issues documented here were already resolved in prior iterations. The remaining items are mostly polish, edge cases, or nice-to-haves that don't block a v1.0.0-beta launch.
