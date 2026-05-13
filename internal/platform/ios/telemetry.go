package ios

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/w1n/perfmon/internal/engine"
)

// Sample collects a single telemetry snapshot for the given PID on the
// currently selected iOS device/simulator. It gathers:
//   - CPU: from `top` output (%CPU column)
//   - Memory: from `top` output (RES/MEM column)
//   - Threads: from `top` output (#TH column)
//
// For simulators, sampling uses `xcrun simctl spawn <udid> top -l 1 -n 1 -pid <PID>`.
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

	// Sample from simulator via top
	return p.sampleSimulator(deviceID, pid)
}

// sampleSimulator collects telemetry from a booted simulator using top.
func (p *iOSProvider) sampleSimulator(udid string, pid int32) (*engine.TelemetrySnapshot, error) {
	// Use top -l 1 -n 1 -pid <PID> for a single-shot sample
	out, err := p.simctlSpawn(udid, "top", "-l", "1", "-n", "1", "-pid", strconv.FormatInt(int64(pid), 10))
	if err != nil {
		// Fall back to ps-based sampling
		return p.sampleSimulatorPS(udid, pid)
	}

	snapshot := parseTopOutput(out, pid)
	if snapshot != nil {
		return snapshot, nil
	}

	// Fall back to ps-based sampling
	return p.sampleSimulatorPS(udid, pid)
}

// sampleSimulatorPS uses ps as a fallback to collect basic telemetry.
func (p *iOSProvider) sampleSimulatorPS(udid string, pid int32) (*engine.TelemetrySnapshot, error) {
	// Get CPU and memory from ps
	out, err := p.simctlSpawn(udid, "ps", "-o", "pid,%cpu,rss,comm", "-p", strconv.FormatInt(int64(pid), 10))
	if err != nil {
		return nil, fmt.Errorf("failed to sample PID %d in simulator %s: %w", pid, udid, err)
	}

	snapshot := parsePSSample(out, pid)
	if snapshot != nil {
		return snapshot, nil
	}

	return nil, fmt.Errorf("unable to parse telemetry for PID %d", pid)
}

// ─── top output parsing ─────────────────────────────────────────────────────

// parseTopOutput extracts CPU, memory, and thread info from iOS simulator top output.
//
// Expected output format (single iteration):
//
//	Processes: 123 total, 2 running, 121 sleeping...
//	Load: 2.34  ...
//	...
//	PID    COMMAND     %CPU  TIME     #TH   #WQ   #PORTS MEM    PURG   CMPRS  PGRP  PPID  STATE    BOOSTED
//	1234   ExampleApp  12.5  00:15.34 15    2     123    45M    0B     0B     1234  1     running  *
//
// Note: iOS top may have slightly different column layouts depending on the version.
func parseTopOutput(output string, pid int32) *engine.TelemetrySnapshot {
	lines := strings.Split(output, "\n")

	// Find the process line by PID
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

		// Parse columns. iOS top format typically has these columns:
		// PID COMMAND %CPU TIME #TH #WQ #PORTS MEM PURG CMPRS PGRP PPID STATE BOOSTED
		// Index:    0       1     2    3   4   5      6   7    8     9   10    11

		var cpuPercent float64
		var memBytes int64
		var threads int32

		// %CPU is typically at index 2
		if len(fields) > 2 {
			cpuPercent, _ = strconv.ParseFloat(fields[2], 64)
		}

		// #TH (threads) is typically at index 4
		if len(fields) > 4 {
			if t, err := strconv.ParseInt(fields[4], 10, 32); err == nil {
				threads = int32(t)
			}
		}

		// MEM is at index 7 (index 6 is #PORTS), format like "45M", "120M", "1.2G"
		if len(fields) > 7 {
			memBytes = parseMemoryValue(fields[7])
		}

		if cpuPercent < 0 {
			cpuPercent = 0
		}
		if memBytes < 0 {
			memBytes = 0
		}
		if threads < 0 {
			threads = 0
		}

		snapshot := engine.NewTelemetrySnapshot(cpuPercent, memBytes/1024, threads)
		return &snapshot
	}

	return nil
}

// parsePSSample extracts telemetry from `ps -o pid,%cpu,rss,comm` output.
//
// Expected format:
//
//	PID %CPU  RSS      COMM
//	1234 12.5 46080    /path/to/ExampleApp
//
// RSS is in KB on iOS/macOS.
func parsePSSample(output string, pid int32) *engine.TelemetrySnapshot {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return nil
	}

	for _, line := range lines[1:] { // skip header
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		linePID, err := strconv.ParseInt(fields[0], 10, 32)
		if err != nil || int32(linePID) != pid {
			continue
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

		// ps doesn't give thread count directly; set to 0 (unknown)
		if cpuPercent < 0 {
			cpuPercent = 0
		}
		if memKB < 0 {
			memKB = 0
		}

		snapshot := engine.NewTelemetrySnapshot(cpuPercent, memKB, threads)
		return &snapshot
	}

	return nil
}

// parseMemoryValue parses memory strings like "45M", "120M", "1.2G", "46080K".
func parseMemoryValue(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" {
		return 0
	}

	var multiplier int64 = 1
	switch {
	case strings.HasSuffix(s, "G") || strings.HasSuffix(s, "g"):
		multiplier = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "G")
		s = strings.TrimSuffix(s, "g")
	case strings.HasSuffix(s, "M") || strings.HasSuffix(s, "m"):
		multiplier = 1024 * 1024
		s = strings.TrimSuffix(s, "M")
		s = strings.TrimSuffix(s, "m")
	case strings.HasSuffix(s, "K") || strings.HasSuffix(s, "k"):
		multiplier = 1024
		s = strings.TrimSuffix(s, "K")
		s = strings.TrimSuffix(s, "k")
	case strings.HasSuffix(s, "B") || strings.HasSuffix(s, "b"):
		s = strings.TrimSuffix(s, "B")
		s = strings.TrimSuffix(s, "b")
	}

	// Parse float to handle values like "1.2G"
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}

	return int64(val * float64(multiplier))
}
