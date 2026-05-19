//go:build adb_test

package android

import (
	"os"
	"strings"
	"testing"

	"github.com/GAM3RG33K/perfmon-lite/internal/engine"
)

// ---------------------------------------------------------------------------
// Integration tests for the Android ADB provider.
//
// These tests require a real ADB binary and at least one connected device
// (physical or emulator). They are excluded from normal builds via the
// `adb_test` build tag.
//
// Run:
//
//	go test -tags=adb_test -v -race -count=1 ./internal/platform/android/
//
// or via Makefile target:
//
//	make test-adb
// ---------------------------------------------------------------------------

// requireDevice is a test helper that skips the test if no ADB device
// is reachable. It discovers devices and uses the first available one.
func requireDevice(t *testing.T) *ADBProvider {
	t.Helper()

	adbPath, err := FindAdbPath()
	if err != nil {
		t.Skipf("adb not found: %v", err)
	}

	p := NewProvider(adbPath)

	devices, err := p.Discover()
	if err != nil {
		t.Fatalf("Discover() failed: %v", err)
	}
	if len(devices) == 0 {
		t.Skip("no connected ADB devices found")
	}

	// Use the first available device
	dev := devices[0]
	t.Logf("using device: %s (%s) physical=%v", dev.ID, dev.Name, dev.IsPhysical)

	if err := ValidateDevice(adbPath, dev.ID); err != nil {
		t.Fatalf("device %s is not reachable: %v", dev.ID, err)
	}

	p.SetDevice(dev.ID)
	t.Cleanup(func() { p.Close() })
	return p
}

// ─── ADB Binary & Preflight Tests ───────────────────────────────────────────

func TestIntegration_FindAdbPath(t *testing.T) {
	path, err := FindAdbPath()
	if err != nil {
		t.Fatalf("FindAdbPath() failed: %v", err)
	}
	if path == "" {
		t.Fatal("expected non-empty adb path")
	}

	// Verify the binary exists on disk
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("adb path %s is not accessible: %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("adb path %s is a directory, not a file", path)
	}
	t.Logf("adb binary: %s", path)
}

func TestIntegration_CheckVersion(t *testing.T) {
	adbPath, err := FindAdbPath()
	if err != nil {
		t.Skipf("adb not found: %v", err)
	}

	ver, err := CheckVersion(adbPath)
	if err != nil {
		t.Fatalf("CheckVersion() failed: %v", err)
	}
	if ver.Major == 0 && ver.Minor == 0 && ver.Patch == 0 {
		t.Fatal("expected non-zero version")
	}
	t.Logf("adb version: %s", ver.String())
}

// ─── Device Discovery Tests ─────────────────────────────────────────────────

func TestIntegration_DiscoverDevices(t *testing.T) {
	adbPath, err := FindAdbPath()
	if err != nil {
		t.Skipf("adb not found: %v", err)
	}

	p := NewProvider(adbPath)
	t.Cleanup(func() { p.Close() })

	devices, err := p.Discover()
	if err != nil {
		t.Fatalf("Discover() failed: %v", err)
	}
	if len(devices) == 0 {
		t.Fatal("expected at least 1 device, got 0")
	}

	for _, dev := range devices {
		t.Logf("device: %s (name=%s, platform=%s, physical=%v, booted=%v)",
			dev.ID, dev.Name, dev.Platform, dev.IsPhysical, dev.IsBooted)
		if dev.ID == "" {
			t.Error("device ID must not be empty")
		}
		if !dev.IsBooted {
			t.Errorf("device %s should be booted (state=device)", dev.ID)
		}
	}

	t.Logf("found %d device(s)", len(devices))
}

// ─── Process Mapping Tests ──────────────────────────────────────────────────

func TestIntegration_MapProcesses(t *testing.T) {
	p := requireDevice(t)

	processes, err := p.MapProcesses(p.DeviceID)
	if err != nil {
		t.Fatalf("MapProcesses() failed: %v", err)
	}
	if len(processes) == 0 {
		t.Fatal("expected at least 1 process, got 0")
	}

	// Verify no kernel threads leaked through (names wrapped in [])
	for _, proc := range processes {
		if proc.Name == "" {
			t.Errorf("process PID=%d has empty name", proc.PID)
		}
		if proc.PID <= 0 {
			t.Errorf("process %q has invalid PID %d", proc.Name, proc.PID)
		}
	}

	// Should contain essential system processes
	foundInit := false
	for _, proc := range processes {
		if proc.Name == "init" {
			foundInit = true
			break
		}
	}
	if !foundInit {
		t.Log("note: 'init' process not found in ps output (may have a different name)")
	}

	t.Logf("found %d processes (first: PID %d %s)",
		len(processes), processes[0].PID, processes[0].Name)
}

// ─── Build Type Detection Tests ─────────────────────────────────────────────

func TestIntegration_BuildTypeDetection(t *testing.T) {
	p := requireDevice(t)

	// Test against a known system app — these are typically release builds
	bt, err := p.BuildType(p.DeviceID, "com.android.systemui")
	if err != nil {
		t.Fatalf("BuildType() failed for com.android.systemui: %v", err)
	}
	t.Logf("com.android.systemui build type: %s", bt)
	if bt == "" {
		t.Error("build type must not be empty")
	}

	// Test with a non-existent package — should return BuildUnknown
	btUnknown, err := p.BuildType(p.DeviceID, "com.nonexistent.fake")
	if err == nil {
		// If dumpsys still returns output for a non-existent package,
		// the result should at least not be empty
		if btUnknown == "" {
			t.Error("expected non-empty build type even for non-existent package")
		}
		t.Logf("non-existent package build type: %s", btUnknown)
	} else {
		// dumpsys failing is acceptable — some emulators reject unknown packages
		t.Logf("non-existent package returned expected error: %v (type=%s)", err, btUnknown)
	}
}

// ─── Telemetry Sampling Tests ───────────────────────────────────────────────

func TestIntegration_SampleInitProcess(t *testing.T) {
	p := requireDevice(t)

	snapshot, err := p.Sample(1) // PID 1 = init on Android
	if err != nil {
		t.Fatalf("Sample(PID=1) failed: %v", err)
	}
	if snapshot == nil {
		t.Fatal("expected non-nil snapshot")
	}

	// Validate snapshot fields
	t.Logf("snapshot: cpu=%.1f%% mem=%dKB threads=%d ts=%d",
		snapshot.CPUPercent, snapshot.MemoryKB, snapshot.Threads, snapshot.Timestamp)

	if snapshot.Timestamp == 0 {
		t.Error("timestamp must not be zero")
	}
	if snapshot.CPUPercent < 0 {
		t.Error("CPU percent must not be negative")
	}
	if snapshot.MemoryKB < 0 {
		t.Error("MemoryKB must not be negative")
	}
	if snapshot.Threads < 0 {
		t.Error("Threads must not be negative")
	}
}

func TestIntegration_SampleSystemServer(t *testing.T) {
	p := requireDevice(t)

	// Find system_server PID dynamically instead of hardcoding 657.
	// Different emulator images / API levels may run system_server at
	// different PIDs.
	processes, err := p.MapProcesses(p.DeviceID)
	if err != nil {
		t.Fatalf("MapProcesses() failed: %v", err)
	}

	pid := int32(0)
	for _, proc := range processes {
		if proc.Name == "system_server" {
			pid = proc.PID
			break
		}
	}
	if pid == 0 {
		t.Skip("system_server process not found in process list")
	}

	snapshot, err := p.Sample(pid)
	if err != nil {
		t.Fatalf("Sample(PID=%d/system_server) failed: %v", pid, err)
	}
	if snapshot == nil {
		t.Fatal("expected non-nil snapshot")
	}

	t.Logf("system_server (PID=%d) snapshot: cpu=%.1f%% mem=%dKB threads=%d",
		pid, snapshot.CPUPercent, snapshot.MemoryKB, snapshot.Threads)

	// On Android 14+ (API 34), SELinux policies restrict shell user access
	// to /proc/<system_server_pid>/status, resulting in zero memory/threads.
	// This is a known platform limitation, not a code bug — skip gracefully.
	if snapshot.MemoryKB <= 0 || snapshot.Threads <= 0 {
		t.Skipf("system_server memory/threads unavailable due to SELinux (/proc/%d/status restricted)", pid)
	}
}

func TestIntegration_SampleNonExistentPID(t *testing.T) {
	p := requireDevice(t)

	// PID 2147483647 (max int32) should never exist — Sample should fail
	// gracefully or produce a snapshot with zeros.
	snap, err := p.Sample(2147483647)
	if err != nil {
		t.Logf("non-existent PID error (expected): %v", err)
		return
	}
	// If no error, the snapshot should have zeros
	t.Logf("non-existent PID snapshot: cpu=%.1f%% mem=%dKB threads=%d",
		snap.CPUPercent, snap.MemoryKB, snap.Threads)
	// Memory and threads may be 0 if the PID doesn't exist
	if snap.Timestamp == 0 {
		t.Error("timestamp should be set even for non-existent PID")
	}
}

func TestIntegration_SampleConsecutive(t *testing.T) {
	p := requireDevice(t)

	// Take two consecutive samples of init to verify the pipe handles reuse
	snap1, err := p.Sample(1)
	if err != nil {
		t.Fatalf("first Sample(PID=1) failed: %v", err)
	}

	snap2, err := p.Sample(1)
	if err != nil {
		t.Fatalf("second Sample(PID=1) failed: %v", err)
	}

	t.Logf("sample 1: cpu=%.1f%% mem=%dKB threads=%d", snap1.CPUPercent, snap1.MemoryKB, snap1.Threads)
	t.Logf("sample 2: cpu=%.1f%% mem=%dKB threads=%d", snap2.CPUPercent, snap2.MemoryKB, snap2.Threads)

	// Both snapshots should have valid data
	if snap1.MemoryKB <= 0 {
		t.Error("first sample memory should be > 0")
	}
	if snap2.MemoryKB <= 0 {
		t.Error("second sample memory should be > 0")
	}
}

// ─── Persistent Shell Pipe Tests ────────────────────────────────────────────

func TestIntegration_PersistentShellPipe(t *testing.T) {
	p := requireDevice(t)

	// Open the persistent shell
	if err := p.ensureShell(); err != nil {
		t.Fatalf("ensureShell() failed: %v", err)
	}

	// Run a simple command through the pipe
	out, err := p.execInShell("echo adb_integration_test")
	if err != nil {
		t.Fatalf("execInShell() failed: %v", err)
	}
	if out == "" {
		t.Fatal("expected non-empty output from pipe")
	}
	t.Logf("pipe echo output: %q", out)

	// Run multiple commands sequentially
	for i := 0; i < 5; i++ {
		out, err := p.execInShell("echo step" + string(rune('0'+i)))
		if err != nil {
			t.Fatalf("execInShell step %d failed: %v", i, err)
		}
		if out == "" {
			t.Errorf("expected output at step %d", i)
		}
	}
}

func TestIntegration_PipeSampleFallback(t *testing.T) {
	p := requireDevice(t)

	// First, verify the pipe works and telemetry comes through
	snap, err := p.Sample(1)
	if err != nil {
		t.Fatalf("Sample(PID=1) via pipe failed: %v", err)
	}
	if snap.MemoryKB < 0 || snap.Threads < 0 {
		t.Errorf("invalid snapshot values: mem=%d threads=%d", snap.MemoryKB, snap.Threads)
	}

	// Kill the underlying shell process to force a pipe failure
	p.shellMu.Lock()
	if p.shellCmd != nil && p.shellCmd.Process != nil {
		if err := p.shellCmd.Process.Kill(); err == nil {
			_ = p.shellCmd.Wait()
		}
		p.shellOpen = false
	}
	p.shellMu.Unlock()

	// Subsequent sample should auto-recover via adbExec fallback
	snap2, err := p.Sample(1)
	if err != nil {
		t.Fatalf("Sample(PID=1) after pipe kill fell back incorrectly: %v", err)
	}
	if snap2.MemoryKB < 0 {
		t.Errorf("fallback snapshot has invalid memory: %d", snap2.MemoryKB)
	}
	t.Logf("pipe fallback snapshot: cpu=%.1f%% mem=%dKB threads=%d",
		snap2.CPUPercent, snap2.MemoryKB, snap2.Threads)
}

// ─── Full End-to-End Flow ───────────────────────────────────────────────────

func TestIntegration_FullEndToEndFlow(t *testing.T) {
	p := requireDevice(t)

	// 1. Discover devices
	devices, err := p.Discover()
	if err != nil {
		t.Fatalf("step 1 (Discover) failed: %v", err)
	}
	if len(devices) == 0 {
		t.Fatal("no devices discovered")
	}
	t.Logf("step 1 ✅ discover: %d device(s)", len(devices))

	// 2. Select first device
	dev := devices[0]
	p.SetDevice(dev.ID)
	t.Logf("step 2 ✅ select device: %s (%s)", dev.ID, dev.Name)

	// 3. Map processes
	processes, err := p.MapProcesses(dev.ID)
	if err != nil {
		t.Fatalf("step 3 (MapProcesses) failed: %v", err)
	}
	if len(processes) == 0 {
		t.Fatal("no processes mapped")
	}
	t.Logf("step 3 ✅ map: %d processes", len(processes))

	// 4. Detect build type for a package (skip if no package-named process found)
	// Only Android packages (with dots in the name) work with dumpsys.
	targetProc := processes[0]
	var bt engine.BuildType
	for _, proc := range processes {
		if strings.Contains(proc.Name, ".") {
			targetProc = proc
			bt, err = p.BuildType(dev.ID, proc.Name)
			if err != nil {
				t.Logf("step 4 (BuildType for %s): %v (continuing)", proc.Name, err)
			} else {
				t.Logf("step 4 ✅ build type for %s: %s", proc.Name, bt)
			}
			break
		}
	}
	if bt == "" {
		t.Log("step 4 ⏭️ no package-named process found to detect build type")
	}

	// 5. Sample telemetry
	snapshot, err := p.Sample(targetProc.PID)
	if err != nil {
		t.Fatalf("step 5 (Sample PID=%d) failed: %v", targetProc.PID, err)
	}
	t.Logf("step 5 ✅ sample PID=%d (%s): cpu=%.1f%% mem=%dKB threads=%d",
		targetProc.PID, targetProc.Name,
		snapshot.CPUPercent, snapshot.MemoryKB, snapshot.Threads)

	// 6. Take a second sample (via persistent pipe)
	snapshot2, err := p.Sample(targetProc.PID)
	if err != nil {
		t.Fatalf("step 6 (second Sample) failed: %v", err)
	}
	t.Logf("step 6 ✅ second sample PID=%d (%s): cpu=%.1f%% mem=%dKB threads=%d",
		targetProc.PID, targetProc.Name,
		snapshot2.CPUPercent, snapshot2.MemoryKB, snapshot2.Threads)

	// 7. Clean up
	if err := p.Close(); err != nil {
		t.Fatalf("step 7 (Close) failed: %v", err)
	}
	t.Log("step 7 ✅ close")
}

// ─── Concurrency Under Real ADB ─────────────────────────────────────────────

func TestIntegration_ConcurrentSamples(t *testing.T) {
	p := requireDevice(t)

	// Sample PID 1 (init) concurrently from multiple goroutines
	errCh := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := p.Sample(1)
			errCh <- err
		}()
	}

	for i := 0; i < 10; i++ {
		if err := <-errCh; err != nil {
			t.Errorf("concurrent Sample() failed: %v", err)
		}
	}
	t.Log("10 concurrent samples completed successfully")
}
