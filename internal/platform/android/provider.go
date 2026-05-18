package android

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/w1n/perfmon/internal/engine"
)

// shellReadTimeout is how long execInShell waits for a line of output before giving up.
const shellReadTimeout = 10 * time.Second

// shellEOFMaker is a unique string used to delimit command output
// from the persistent ADB shell. Extremely unlikely to appear in real output.
const shellEOFMaker = "__PERFMON_EOF__"

// restartDelay is how long to wait before retrying a failed shell connection.
const restartDelay = 500 * time.Millisecond

// ADBProvider implements the PlatformProvider interface for Android devices.
// It communicates with devices via the ADB (Android Debug Bridge) command-line tool.
// Telemetry sampling uses a persistent `adb shell` pipe for efficiency.
type ADBProvider struct {
	AdbPath  string // path to the adb binary
	DeviceID string // target device serial for telemetry sampling

	// Persistent shell session for telemetry sampling
	shellCmd  *exec.Cmd
	shellIn   io.WriteCloser
	shellOut  *bufio.Reader
	shellMu   sync.Mutex // serialises access to the pipe
	shellOpen bool

	mu sync.Mutex // protects DeviceID

	// CPU delta tracking — /proc/<pid>/stat utime+stime from previous Sample()
	cpuMu        sync.Mutex
	prevPID      int32
	prevCPUTicks uint64
	prevCPUTime  time.Time
	firstSample  bool
}

// NewProvider creates a new Android provider using the given adb binary path.
func NewProvider(adbPath string) *ADBProvider {
	return &ADBProvider{
		AdbPath: adbPath,
	}
}

// SetDevice sets the target device for telemetry operations.
// If a shell session is already open for a different device, it is restarted.
func (p *ADBProvider) SetDevice(deviceID string) {
	p.mu.Lock()
	prev := p.DeviceID
	p.DeviceID = deviceID
	p.mu.Unlock()

	if prev != "" && prev != deviceID {
		p.closeShell() // device changed — restart shell on next Sample()
	}
}

// ─── Persistent ADB shell pipe ─────────────────────────────────────────────

// ensureShell opens a persistent `adb shell` session if one is not already running.
// The session is used for low-latency telemetry sampling.
func (p *ADBProvider) ensureShell() error {
	p.shellMu.Lock()
	defer p.shellMu.Unlock()

	if p.shellOpen {
		return nil
	}

	p.mu.Lock()
	adbPath := p.AdbPath
	deviceID := p.DeviceID
	p.mu.Unlock()

	// Build command: adb [-s device] shell
	var args []string
	if deviceID != "" {
		args = append(args, "-s", deviceID)
	}
	args = append(args, "shell")

	cmd := exec.Command(adbPath, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("shell stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("shell stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("shell stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("shell start: %w", err)
	}

	// Merge stdout and stderr so error output is captured inline
	merged := io.MultiReader(stdout, stderr)

	p.shellCmd = cmd
	p.shellIn = stdin
	p.shellOut = bufio.NewReader(merged)
	p.shellOpen = true

	// Give the shell a moment to start, then drain any startup output.
	// The adb shell doesn't print a prompt by default, but this handles
	// unexpected banner output or initial echo.
	p.flushShell(time.Second)

	return nil
}

// flushShell drains any data the shell may have emitted at startup
// (e.g., ANSI reset sequences or a shell prompt).
// Uses only non-blocking operations to avoid data races.
func (p *ADBProvider) flushShell(timeout time.Duration) {
	// Wait a brief moment for the shell to start producing output.
	// Then drain only what's already in the in-memory buffer.
	// Any data that arrives later will be consumed by execInShell's
	// ReadString loop as noise before the actual command output.
	deadline := time.After(timeout)
	for {
		if n := p.shellOut.Buffered(); n > 0 {
			p.shellOut.Discard(n)
			continue
		}
		select {
		case <-deadline:
			return
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// execInShell sends a command to the persistent ADB shell and reads the output
// until the shellEOFMaker delimiter is seen. Returns the output before the marker.
// If the pipe fails, it returns an error (the caller may fall back to adbExec).
func (p *ADBProvider) execInShell(command string) (string, error) {
	p.shellMu.Lock()
	defer p.shellMu.Unlock()

	if !p.shellOpen {
		return "", fmt.Errorf("shell not open")
	}

	// Send the command followed by an EOF marker on its own line.
	// printf interprets \n as newline (even inside single quotes in bash/sh),
	// so the marker is echoed as a standalone line for ReadString to detect.
	fullCmd := fmt.Sprintf("%s\nprintf '\\n%s\\n'\n", command, shellEOFMaker)

	_, err := io.WriteString(p.shellIn, fullCmd)
	if err != nil {
		p.shellOpen = false // mark dead — next call will reopen
		return "", fmt.Errorf("shell write: %w", err)
	}

	// Read output line by line until we see the EOF marker
	var output strings.Builder
	readCh := make(chan readResult, 1)
	go p.readShellLine(readCh)

	for {
		var line string
		var err error

		select {
		case res := <-readCh:
			line, err = res.line, res.err
		case <-time.After(shellReadTimeout):
			p.shellOpen = false
			return "", fmt.Errorf("shell read timed out after %v", shellReadTimeout)
		}

		if err != nil {
			p.shellOpen = false // mark dead
			return "", fmt.Errorf("shell read: %w", err)
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == shellEOFMaker {
			break
		}
		output.WriteString(line)

		// Start reading the next line
		go p.readShellLine(readCh)
	}

	return output.String(), nil
}

// readResult holds a single line read from the shell and any error.
type readResult struct {
	line string
	err  error
}

// readShellLine reads one line from the shell's buffered reader and sends it on ch.
// Must be called in a goroutine since ReadString blocks.
func (p *ADBProvider) readShellLine(ch chan<- readResult) {
	line, err := p.shellOut.ReadString('\n')
	ch <- readResult{line, err}
}

// closeShell kills the persistent ADB shell session, if any.
func (p *ADBProvider) closeShell() {
	p.shellMu.Lock()
	defer p.shellMu.Unlock()

	if !p.shellOpen {
		return
	}

	p.shellOpen = false

	if p.shellIn != nil {
		p.shellIn.Close()
		p.shellIn = nil
	}

	if p.shellCmd != nil && p.shellCmd.Process != nil {
		// Try graceful shutdown first, then force kill
		p.shellCmd.Process.Kill()
		_ = p.shellCmd.Wait()
		p.shellCmd = nil
	}

	p.shellOut = nil
}

// ─── ADB command helpers ────────────────────────────────────────────────────

// adbCommand returns an exec.Cmd for running adb with the given arguments.
// If a device ID is set, it inserts "-s <deviceID>" after the adb binary.
func (p *ADBProvider) adbCommand(args ...string) *exec.Cmd {
	p.mu.Lock()
	deviceID := p.DeviceID
	p.mu.Unlock()

	var cmdArgs []string
	if deviceID != "" {
		cmdArgs = append(cmdArgs, "-s", deviceID)
	}
	cmdArgs = append(cmdArgs, args...)
	return exec.Command(p.AdbPath, cmdArgs...)
}

// adbExec runs adb with the given arguments and returns the combined output.
func (p *ADBProvider) adbExec(args ...string) (string, error) {
	cmd := p.adbCommand(args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", &AdbError{
			Args:   args,
			Output: strings.TrimSpace(string(out)),
			Err:    err,
		}
	}
	return string(out), nil
}

// ─── Interface compliance and helpers ───────────────────────────────────────

// Close releases provider resources (persistent shell session, if any).
func (p *ADBProvider) Close() error {
	p.closeShell()
	return nil
}

// AdbError represents an error from an ADB command execution.
type AdbError struct {
	Args   []string
	Output string
	Err    error
}

func (e *AdbError) Error() string {
	return "adb " + strings.Join(e.Args, " ") + ": " + e.Output + " (" + e.Err.Error() + ")"
}

func (e *AdbError) Unwrap() error {
	return e.Err
}

// ─── Log capture (logcat) ─────────────────────────────────────────────────

// CaptureLogs fetches recent logcat entries for the current device.
// Returns new log lines since the last call (tracked by logcatCursor).
func (p *ADBProvider) CaptureLogs(pid int32) ([]string, error) {
	p.mu.Lock()
	deviceID := p.DeviceID
	p.mu.Unlock()

	if deviceID == "" {
		return nil, fmt.Errorf("no device ID set")
	}

	// Capture last 10 lines of logcat for this PID, clear buffer after
	cmd := fmt.Sprintf("logcat -d --pid=%d -t 10 2>/dev/null", pid)

	var rawOutput string
	var err error

	err = p.ensureShell()
	if err == nil {
		rawOutput, err = p.execInShell(cmd)
	}
	if err != nil {
		rawOutput, err = p.adbExec("-s", deviceID, "shell", cmd)
		if err != nil {
			return nil, err
		}
	}

	lines := strings.Split(strings.TrimSpace(rawOutput), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return nil, nil
	}

	return lines, nil
}

// ─── Interface compliance checks ──────────────────────────────────────────
var _ engine.PlatformProvider = (*ADBProvider)(nil)
var _ engine.LogCapturer = (*ADBProvider)(nil)
var _ engine.DeviceDiscovery = (*ADBProvider)(nil)
var _ engine.ProcessMapper = (*ADBProvider)(nil)
var _ engine.TelemetryProvider = (*ADBProvider)(nil)
