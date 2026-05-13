package android

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/w1n/perfmon/internal/engine"
)

// Sample collects a single telemetry snapshot for the given PID on the
// currently selected device. It gathers:
//   - CPU: from `top -n 1 -b` output, %CPU column
//   - Memory: from `/proc/<pid>/status`, VmRSS field
//   - Threads: from `/proc/<pid>/status`, Threads field
//
// Sample first attempts to use the persistent ADB shell pipe for lower latency.
// If the pipe is unavailable or fails, it falls back to a one-shot adb exec.
func (p *ADBProvider) Sample(pid int32) (*engine.TelemetrySnapshot, error) {
	p.mu.Lock()
	deviceID := p.DeviceID
	p.mu.Unlock()

	if deviceID == "" {
		return nil, fmt.Errorf("no device ID set: call SetDevice() first")
	}

	// Build the sampling command — combined CPU + memory + threads
	sampleCmd := fmt.Sprintf(
		`top -n 1 -b -p %d 2>/dev/null; echo "===MEM==="; cat /proc/%d/status 2>/dev/null`,
		pid, pid,
	)

	var rawOutput string
	var err error

	// Try persistent shell pipe first
	err = p.ensureShell()
	if err == nil {
		rawOutput, err = p.execInShell(sampleCmd)
	}

	// Fall back to one-shot adb if pipe is unavailable
	if err != nil {
		rawOutput, err = p.adbExec("-s", deviceID, "shell", sampleCmd)
		if err != nil {
			return nil, fmt.Errorf("sample failed for PID %d: %w", pid, err)
		}
	}

	// Split output into sections
	parts := strings.SplitN(rawOutput, "===MEM===\n", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("unexpected sample output format for PID %d", pid)
	}

	topOutput := parts[0]
	statusOutput := parts[1]

	// Parse CPU from top output
	cpuPercent := parseCPU(topOutput, pid)

	// Parse memory and threads from /proc/status
	memKB := parseVmRSS(statusOutput)
	threads := parseThreads(statusOutput)

	// Validate parsed values
	if cpuPercent < 0 {
		cpuPercent = 0
	}
	if memKB < 0 {
		memKB = 0
	}
	if threads < 0 {
		threads = 0
	}

	snapshot := engine.NewTelemetrySnapshot(cpuPercent, memKB, threads)
	return &snapshot, nil
}

// parseCPU extracts the CPU percentage for the given PID from top output.
//
// Expected top output format (batch mode):
//
//	Tasks: 123 total, 1 running, 122 sleeping...
//	Mem: 123456k total, 78901k used...
//	Swap: 0k total, 0k used...
//	100%cpu  12%user   5%nice   30%sys  53%idle   0%iow   0%irq   0%sirq   0%host
//	  PID   USER   PR  NI  VIRT  RES  SHR  S  %CPU  %MEM   TIME+   ARGS
//	 4567   u0_a  20   0  1.2G  120M  80M  S  12.5   2.3    0:15.34  com.example.app
func parseCPU(topOutput string, pid int32) float64 {
	lines := strings.Split(topOutput, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		// First field should be the PID
		linePID, err := strconv.ParseInt(fields[0], 10, 32)
		if err != nil || int32(linePID) != pid {
			continue
		}

		// Find the %CPU column. In top batch output, the columns are:
		// PID USER PR NI VIRT RES SHR S %CPU %MEM TIME+ ARGS
		// %CPU is typically at index 8 (0-indexed)
		if len(fields) >= 10 {
			cpuStr := fields[8]
			cpu, err := strconv.ParseFloat(cpuStr, 64)
			if err == nil {
				return cpu
			}
		}
	}

	return -1
}

// parseVmRSS extracts the Resident Set Size in KB from /proc/<pid>/status.
//
// Expected format:
//
//	Name:   com.example.app
//	VmRSS:	    120480 kB
//	Threads:	15
func parseVmRSS(statusOutput string) int64 {
	for _, line := range strings.Split(statusOutput, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "VmRSS:") {
			continue
		}

		// Extract the numeric value before "kB"
		valueStr := strings.TrimPrefix(line, "VmRSS:")
		valueStr = strings.TrimSpace(valueStr)
		valueStr = strings.TrimSuffix(valueStr, "kB")
		valueStr = strings.TrimSpace(valueStr)

		// Remove any commas
		valueStr = strings.ReplaceAll(valueStr, ",", "")

		mem, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			continue
		}
		return mem
	}
	return -1
}

// parseThreads extracts the thread count from /proc/<pid>/status.
//
// Expected format:
//
//	Threads:	15
func parseThreads(statusOutput string) int32 {
	for _, line := range strings.Split(statusOutput, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "Threads:") {
			continue
		}

		valueStr := strings.TrimPrefix(line, "Threads:")
		valueStr = strings.TrimSpace(valueStr)

		threads, err := strconv.ParseInt(valueStr, 10, 32)
		if err != nil {
			continue
		}
		return int32(threads)
	}
	return -1
}
