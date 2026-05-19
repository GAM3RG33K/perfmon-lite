package android

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/GAM3RG33K/perfmon-lite/internal/engine"
)

// MapProcesses returns running processes for the given device by parsing
// the output of `adb shell ps -A`.
//
// Expected output format:
//
//	USER           PID  PPID        VSZ    RSS WCHAN            ADDR S NAME
//	u0_a123       4567  1234     123456    7890 0                   0 S com.example.app
//	root          1234   567      98765    4321 0                   0 S com.example.service
func (p *ADBProvider) MapProcesses(deviceID string) ([]engine.AppProcess, error) {
	out, err := p.adbExec("-s", deviceID, "shell", "ps", "-A")
	if err != nil {
		return nil, err
	}

	return parsePsOutput(out), nil
}

// parsePsOutput parses the output of `adb shell ps -A` into process entries.
// It uses whitespace-delimited field parsing since ps output has fixed-width columns.
func parsePsOutput(output string) []engine.AppProcess {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return nil
	}

	// Find the header line
	headerIdx := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "USER") || strings.HasPrefix(trimmed, "user") {
			headerIdx = i
			break
		}
	}
	if headerIdx < 0 {
		return nil
	}

	// Determine PID and NAME column indices from the header
	headerFields := strings.Fields(lines[headerIdx])
	pidFieldIdx := -1
	nameFieldIdx := -1
	for i, f := range headerFields {
		switch f {
		case "PID":
			pidFieldIdx = i
		case "NAME":
			nameFieldIdx = i
		}
	}
	if pidFieldIdx < 0 || nameFieldIdx < 0 || nameFieldIdx <= pidFieldIdx {
		return nil
	}

	var processes []engine.AppProcess
	for _, line := range lines[headerIdx+1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) <= nameFieldIdx {
			continue
		}

		pidStr := fields[pidFieldIdx]
		pid, err := strconv.ParseInt(pidStr, 10, 32)
		if err != nil {
			continue
		}

		name := fields[nameFieldIdx]

		// Skip kernel threads — they have names in brackets like [kthreadd]
		if strings.HasPrefix(name, "[") && strings.HasSuffix(name, "]") {
			continue
		}

		processes = append(processes, engine.AppProcess{
			PID:         int32(pid),
			Name:        name,
			PackageName: name,
			BuildType:   engine.BuildUnknown,
		})
	}

	return processes
}

// BuildType detects whether the given package is a debug or release build
// by parsing `adb shell dumpsys package <packageName>` for the DEBUGGABLE flag.
func (p *ADBProvider) BuildType(deviceID, packageName string) (engine.BuildType, error) {
	out, err := p.adbExec("-s", deviceID, "shell", "dumpsys", "package", packageName)
	if err != nil {
		// If the command fails (e.g., dumpsys not available), return unknown
		return engine.BuildUnknown, fmt.Errorf("build type detection failed for %s: %w", packageName, err)
	}

	return parseBuildType(out), nil
}

// parseBuildType checks dumpsys output for the DEBUGGABLE flag.
//
// The dumpsys output for a debuggable app contains either:
//   - A line like: flags=[0x1] DEBUGGABLE
//   - A line like: pkgFlags=DEBUGGABLE
//   - A line like: flags=0x1
func parseBuildType(output string) engine.BuildType {
	lines := strings.Split(output, "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return engine.BuildUnknown
	}

	for _, line := range lines {
		lower := strings.ToLower(line)
		// Check for explicit DEBUGGABLE flag
		if strings.Contains(lower, "debuggable") {
			return engine.BuildDebug
		}
		// Check for flags=[0x1] or flags=0x1 patterns
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "flags=") || strings.HasPrefix(trimmed, "pkgFlags=") {
			val := strings.TrimPrefix(trimmed, "flags=")
			val = strings.TrimPrefix(val, "pkgFlags=")
			val = strings.Trim(val, "[]") // remove brackets
			flags, err := strconv.ParseInt(val, 0, 64)
			if err == nil && (flags&0x1 != 0) {
				return engine.BuildDebug
			}
		}
	}
	return engine.BuildRelease
}
