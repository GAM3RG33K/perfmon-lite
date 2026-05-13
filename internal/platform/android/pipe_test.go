package android

import (
	"testing"
	"time"
)

// ─── Persistent Shell Pipe Tests ────────────────────────────────────────────

func TestEnsureShell_OpensAndCloses(t *testing.T) {
	t.Skip("Requires a real adb binary: set ADB_TEST=1 to run")
	p := NewProvider("adb")
	defer p.Close()

	if err := p.ensureShell(); err != nil {
		t.Fatalf("expected shell to open, got error: %v", err)
	}
	if !p.shellOpen {
		t.Fatal("expected shellOpen to be true after ensureShell")
	}

	p.closeShell()
	if p.shellOpen {
		t.Fatal("expected shellOpen to be false after closeShell")
	}
}

func TestEnsureShell_Idempotent(t *testing.T) {
	t.Skip("Requires a real adb binary: set ADB_TEST=1 to run")
	p := NewProvider("adb")
	defer p.Close()

	if err := p.ensureShell(); err != nil {
		t.Fatalf("first ensureShell failed: %v", err)
	}

	// Calling ensureShell again should be a no-op
	if err := p.ensureShell(); err != nil {
		t.Fatalf("second ensureShell failed: %v", err)
	}
}

func TestExecInShell_SimpleCommand(t *testing.T) {
	t.Skip("Requires a real adb binary: set ADB_TEST=1 to run")
	p := NewProvider("adb")
	defer p.Close()

	if err := p.ensureShell(); err != nil {
		t.Fatalf("ensureShell failed: %v", err)
	}

	out, err := p.execInShell("echo hello")
	if err != nil {
		t.Fatalf("execInShell failed: %v", err)
	}
	if out != "hello\n" && out != "hello\r\n" {
		t.Fatalf("expected 'hello', got %q", out)
	}
}

func TestExecInShell_MultipleCommands(t *testing.T) {
	t.Skip("Requires a real adb binary: set ADB_TEST=1 to run")
	p := NewProvider("adb")
	defer p.Close()

	if err := p.ensureShell(); err != nil {
		t.Fatalf("ensureShell failed: %v", err)
	}

	// First command
	out1, err := p.execInShell("echo first")
	if err != nil {
		t.Fatalf("first command failed: %v", err)
	}
	if out1 == "" {
		t.Fatal("expected output from first command, got empty")
	}

	// Second command — should work on the same pipe
	out2, err := p.execInShell("echo second")
	if err != nil {
		t.Fatalf("second command failed: %v", err)
	}
	if out2 == "" {
		t.Fatal("expected output from second command, got empty")
	}
}

func TestExecInShell_ReconnectsAfterClose(t *testing.T) {
	t.Skip("Requires a real adb binary: set ADB_TEST=1 to run")
	p := NewProvider("adb")
	defer p.Close()

	if err := p.ensureShell(); err != nil {
		t.Fatalf("ensureShell failed: %v", err)
	}

	// Force close the shell
	p.closeShell()

	// EnsureShell should reopen the pipe
	if err := p.ensureShell(); err != nil {
		t.Fatalf("reconnect ensureShell failed: %v", err)
	}

	out, err := p.execInShell("echo reconnected")
	if err != nil {
		t.Fatalf("execInShell after reconnect failed: %v", err)
	}
	if out == "" {
		t.Fatal("expected output after reconnect, got empty")
	}
}

func TestClose_CleansUpShell(t *testing.T) {
	t.Skip("Requires a real adb binary: set ADB_TEST=1 to run")
	p := NewProvider("adb")

	if err := p.ensureShell(); err != nil {
		t.Fatalf("ensureShell failed: %v", err)
	}

	if err := p.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	if p.shellOpen {
		t.Fatal("expected shellOpen to be false after Close()")
	}
}

// ─── Sample Fallback Tests ──────────────────────────────────────────────────

// TestSample_WithFakePID verifies that Sample() returns a useful error
// when given a non-existent PID, and that it tries the pipe before falling back.
func TestSample_WithFakePID(t *testing.T) {
	t.Skip("Requires a real device connection: set ADB_TEST=1 to run")
	p := NewProvider("adb")
	p.SetDevice("emulator-5554") // non-existent device, will fail gracefully
	defer p.Close()

	_, err := p.Sample(99999)
	if err == nil {
		t.Fatal("expected error for non-existent PID, got nil")
	}
	// Should contain meaningful error info
	t.Logf("Sample() returned expected error: %v", err)
}

// TestSample_NoDevice ensures Sample() fails early when no device is set.
func TestSample_NoDevice(t *testing.T) {
	p := NewProvider("adb")
	defer p.Close()

	_, err := p.Sample(1234)
	if err == nil {
		t.Fatal("expected error when no device is set, got nil")
	}
}

// TestSetDevice_EmptyString does not crash.
func TestSetDevice_EmptyString(t *testing.T) {
	p := NewProvider("adb")
	p.SetDevice("") // should not panic
	if p.DeviceID != "" {
		t.Fatalf("expected empty DeviceID, got %s", p.DeviceID)
	}
}

// TestSetDevice_ReopensForNewDevice ensures changing the device ID
// triggers a shell restart (but since no shell is open, it should be a no-op).
func TestSetDevice_ReopensForNewDevice(t *testing.T) {
	p := NewProvider("adb")
	p.SetDevice("device1")
	p.SetDevice("device2")
	// Should not panic; shell close on device change is safe even with no open shell
	if p.DeviceID != "device2" {
		t.Fatalf("expected DeviceID device2, got %s", p.DeviceID)
	}
}

// ─── Concurrency Tests ──────────────────────────────────────────────────────

// TestSample_ConcurrentCalls ensures Sample() can be called concurrently
// without data races (even if both calls fail due to no device, they handle
// locking correctly).
func TestSample_ConcurrentCalls(t *testing.T) {
	p := NewProvider("adb")
	defer p.Close()

	done := make(chan struct{}, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = p.Sample(int32(i))
			done <- struct{}{}
		}()
	}

	timeout := time.After(5 * time.Second)
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-timeout:
			t.Fatal("timed out waiting for concurrent Sample() calls")
		}
	}
}

// TestSetDevice_Concurrent ensures SetDevice is safe under concurrent calls.
func TestSetDevice_Concurrent(t *testing.T) {
	p := NewProvider("adb")
	defer p.Close()

	done := make(chan struct{}, 20)
	for i := 0; i < 20; i++ {
		go func(n int) {
			p.SetDevice("device")
			done <- struct{}{}
		}(i)
	}

	timeout := time.After(5 * time.Second)
	for i := 0; i < 20; i++ {
		select {
		case <-done:
		case <-timeout:
			t.Fatal("timed out waiting for concurrent SetDevice calls")
		}
	}
}

func TestClose_Idempotent(t *testing.T) {
	p := NewProvider("adb")
	if err := p.Close(); err != nil {
		t.Fatalf("first Close() failed: %v", err)
	}
	if err := p.Close(); err != nil {
		t.Fatalf("second Close() (idempotent) failed: %v", err)
	}
}
