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

	// Warm up the shell by waiting briefly for the prompt
	// Not strictly necessary but helps flush startup noise
	_ = p.flushShell(time.Second)

	return nil
}

// flushShell drains any buffered data from the shell (e.g., the initial prompt).
func (p *ADBProvider) flushShell(timeout time.Duration) error {
	deadline := time.After(timeout)
	for {
		select {
		case <-deadline:
			return nil
		default:
			// Peek without blocking — if nothing available, we're done
			_, err := p.shellOut.Peek(1)
			if err != nil {
				return nil // no more data or EOF
			}
			// Drain one byte
			_, _ = p.shellOut.Discard(1)
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

	// Build the command with an EOF marker on a dedicated line.
	// Using printf for reliable marker echo.
	fullCmd := fmt.Sprintf("%s\nprintf '\\n%s\\n'\\n", command, shellEOFMaker)

	_, err := io.WriteString(p.shellIn, fullCmd)
	if err != nil {
		p.shellOpen = false // mark dead — next call will reopen
		return "", fmt.Errorf("shell write: %w", err)
	}

	// Read output line by line until we see the EOF marker
	var output strings.Builder
	for {
		line, err := p.shellOut.ReadString('\n')
		if err != nil {
			p.shellOpen = false // mark dead
			return "", fmt.Errorf("shell read: %w", err)
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == shellEOFMaker {
			break
		}
		output.WriteString(line)
	}

	return output.String(), nil
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

// Interface compliance checks.
var _ engine.PlatformProvider = (*ADBProvider)(nil)
var _ engine.DeviceDiscovery = (*ADBProvider)(nil)
var _ engine.ProcessMapper = (*ADBProvider)(nil)
var _ engine.TelemetryProvider = (*ADBProvider)(nil)
