package android

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/w1n/perfmon/internal/engine"
)

// clkTick is the standard Linux clock tick rate used by Android.
// Almost all Android devices use 100 ticks/sec.
const clkTick = 100

// Sample collects a single telemetry snapshot for the given PID on the
// currently selected device. It gathers:
//   - CPU: from `/proc/<pid>/stat`, utime + stime fields, computed as delta
//   - Memory: from `/proc/<pid>/status`, VmRSS field
//   - Threads: from `/proc/<pid>/status`, Threads field
//
// CPU is calculated as a delta between consecutive samples. The first
// sample returns 0% CPU (no previous data point to diff against).
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

	// Build the sampling command — stat + status
	// /proc/<pid>/stat gives CPU ticks; /proc/<pid>/status gives memory + threads
	sampleCmd := fmt.Sprintf(
		`cat /proc/%d/stat 2>/dev/null; echo "===MEM==="; cat /proc/%d/status 2>/dev/null`,
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

	statOutput := parts[0]
	statusOutput := parts[1]

	// Parse CPU from /proc/<pid>/stat ticks
	now := time.Now()
	cpuPercent := p.parseCPUStat(statOutput, pid, now)

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

// parseCPUStat extracts CPU percentage from /proc/<pid>/stat using
// utime (field 14) and stime (field 15) with delta calculation.
//
// Expected format (see man 5 proc):
//
//	pid (comm) S ppid pgrp session tty_nr tpgid flags minflt cminflt majflt cmajflt
//	utime stime cutime cstime priority nice num_threads itrealvalue starttime ...
//
// utime and stime are measured in clock ticks (typically 100 Hz on Android).
func (p *ADBProvider) parseCPUStat(output string, pid int32, now time.Time) float64 {
	p.cpuMu.Lock()
	defer p.cpuMu.Unlock()

	// Find the closing paren of comm — it's the last ')' before the state char
	// The format is: pid (comm) state ...
	closeParen := strings.LastIndex(output, ") ")
	if closeParen < 0 {
		return -1
	}

	// Everything after ") " is the remaining fields
	rest := strings.TrimSpace(output[closeParen+2:])
	fields := strings.Fields(rest)
	if len(fields) < 15 {
		return -1
	}

	// utime is field 14 (1-indexed) = index 11 after removing "pid (comm) "
	// stime is field 15 (1-indexed) = index 12 after removing "pid (comm) "
	if len(fields) < 15 {
		return -1
	}
	utime, err1 := strconv.ParseUint(fields[11], 10, 64)
	stime, err2 := strconv.ParseUint(fields[12], 10, 64)
	if err1 != nil || err2 != nil {
		return -1
	}

	totalTicks := utime + stime

	if p.firstSample || p.prevPID != pid {
		// First sample or PID changed — store baseline, return 0
		p.prevPID = pid
		p.prevCPUTicks = totalTicks
		p.prevCPUTime = now
		p.firstSample = false
		return 0
	}

	// Compute delta since last sample
	deltaTicks := totalTicks - p.prevCPUTicks
	elapsed := now.Sub(p.prevCPUTime).Seconds()
	if elapsed <= 0 || deltaTicks > totalTicks {
		// Reset on invalid state (e.g. PID recycled)
		p.prevCPUTicks = totalTicks
		p.prevCPUTime = now
		return 0
	}

	cpuPercent := float64(deltaTicks) / clkTick / elapsed * 100

	// Update baseline for next sample
	p.prevCPUTicks = totalTicks
	p.prevCPUTime = now

	if cpuPercent < 0 {
		return 0
	}
	return cpuPercent
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

		valueStr := strings.TrimPrefix(line, "VmRSS:")
		valueStr = strings.TrimSpace(valueStr)
		valueStr = strings.TrimSuffix(valueStr, "kB")
		valueStr = strings.TrimSpace(valueStr)
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
