package ios

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/GAM3RG33K/perfmon-lite/internal/engine"
)

// Sample collects a single telemetry snapshot for the given PID on the
// currently selected iOS device/simulator.
//
// iOS simulator processes are also visible as native macOS processes,
// so we can sample them directly from the host using /proc/pid equivalent
// commands (ps on macOS, proc_info on Linux).
//
// For physical devices, sampling is not supported due to sandbox restrictions.
func (p *iOSProvider) Sample(pid int32) (*engine.TelemetrySnapshot, error) {
	p.mu.Lock()
	deviceID := p.DeviceID
	p.mu.Unlock()

	if deviceID == "" {
		return nil, fmt.Errorf("no device ID set: call SetDevice() first")
	}

	device, err := p.getDevice(deviceID)
	if err != nil {
		return nil, fmt.Errorf("device lookup: %w", err)
	}

	if device.IsPhysical {
		return nil, fmt.Errorf("telemetry sampling not supported on physical iOS devices (sandbox restriction)")
	}

	return sampleHostProcess(pid)
}

// sampleHostProcess samples a running process using macOS host-level tools.
// iOS simulator processes run as native macOS processes and are visible
// from the host with the same PID.
func sampleHostProcess(pid int32) (*engine.TelemetrySnapshot, error) {
	// Use ps to get CPU% and RSS for the given PID
	// Format: pid cpu% rss (in KB)
	cmd := exec.Command("ps", "-p", strconv.FormatInt(int64(pid), 10), "-o", "pid=,%cpu=,rss=,comm=")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to sample PID %d: %w", pid, err)
	}

	snapshot := parsePSSample(string(out), pid)
	if snapshot == nil {
		return nil, fmt.Errorf("failed to parse telemetry for PID %d", pid)
	}

	if snapshot.CPUPercent > 50 {
		snapshot.Stack = readMacOSStack(pid)
	}

	return snapshot, nil
}

// parsePSSample extracts telemetry from `ps -p <pid> -o pid=,%cpu=,rss=,comm=` output.
//
// Expected format:
//
//	48356 29.7 267264 /path/to/tatva.app/tatva
//
// RSS is in KB on macOS.
func parsePSSample(output string, pid int32) *engine.TelemetrySnapshot {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil
	}

	fields := strings.Fields(output)
	if len(fields) < 4 {
		return nil
	}

	linePID, err := strconv.ParseInt(fields[0], 10, 32)
	if err != nil || int32(linePID) != pid {
		return nil
	}

	var cpuPercent float64
	var memKB int64
	var threads int32

	if len(fields) > 1 {
		cpuPercent, _ = strconv.ParseFloat(fields[1], 64)
	}
	if len(fields) > 2 {
		memKB, _ = strconv.ParseInt(fields[2], 10, 64)
	}

	if cpuPercent < 0 {
		cpuPercent = 0
	}
	if memKB < 0 {
		memKB = 0
	}

	// Thread count via ps -M is unreliable on macOS; leave as 0 for now.
	s := engine.NewTelemetrySnapshot(cpuPercent, memKB, threads)
	return &s
}

// readMacOSStack uses the macOS `sample` command to get a brief stack trace.
// sample is built into macOS and gives detailed user-space stacks.
func readMacOSStack(pid int32) string {
	cmd := exec.Command("sample", "-file", "/dev/stdout", strconv.FormatInt(int64(pid), 10), "1", "-mayDie")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	// Take first ~20 lines to keep it concise
	lines := strings.SplitN(string(out), "\n", 22)
	if len(lines) > 21 {
		return strings.Join(lines[:21], "\n") + "\n..."
	}
	return string(out)
}
