package ios

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/w1n/perfmon/internal/engine"
)

// MapProcesses returns running processes for the given iOS device/simulator.
// For simulators, it uses `xcrun simctl spawn <udid> launchctl list`.
// For physical devices, it returns an error indicating limited support.
func (p *iOSProvider) MapProcesses(deviceID string) ([]engine.AppProcess, error) {
	device, err := p.getDevice(deviceID)
	if err != nil {
		return nil, err
	}

	if device.IsPhysical {
		// Physical iOS devices have very limited process visibility
		return nil, fmt.Errorf("process mapping not supported on physical iOS devices (sandbox restriction)")
	}

	return p.mapSimulatorProcesses(deviceID)
}

// mapSimulatorProcesses lists processes inside a booted simulator.
// Uses `launchctl list` to get all running processes, then `ps` for more detail.
func (p *iOSProvider) mapSimulatorProcesses(udid string) ([]engine.AppProcess, error) {
	// First try launchctl list for comprehensive process listing
	out, err := p.simctlSpawn(udid, "launchctl", "list")
	if err != nil {
		// Fall back to ps
		return p.mapSimulatorProcessesPS(udid)
	}

	processes := parseLaunchctlList(out)
	if len(processes) > 0 {
		return processes, nil
	}

	// Fall back to ps if launchctl returned nothing useful
	return p.mapSimulatorProcessesPS(udid)
}

// mapSimulatorProcessesPS uses `ps -A` inside the simulator as a fallback.
func (p *iOSProvider) mapSimulatorProcessesPS(udid string) ([]engine.AppProcess, error) {
	out, err := p.simctlSpawn(udid, "ps", "-A", "-o", "pid,comm")
	if err != nil {
		return nil, fmt.Errorf("failed to list processes in simulator %s: %w", udid, err)
	}

	return parsePSOutput(out), nil
}

// parseLaunchctlList parses the output of `launchctl list`.
//
// Expected output format:
//
//	PID	Status	Label
//	1234	0	com.apple.springboard
//	5678	0	com.example.app
//	-	78	com.apple.some.daemon
//	48356	0	UIKitApplication:in.thetatva.tatva[7140][rb-legacy]
//
// Lines with "-" as PID are not running (exit status only).
// UIKitApplication: prefix and [bracket] suffixes are stripped from labels.
func parseLaunchctlList(output string) []engine.AppProcess {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return nil
	}

	var processes []engine.AppProcess
	for _, line := range lines[1:] { // skip header
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		// Skip entries with "-" as PID (not running)
		if fields[0] == "-" {
			continue
		}

		pid, err := strconv.ParseInt(fields[0], 10, 32)
		if err != nil {
			continue
		}

		label := cleanLaunchLabel(fields[2])

		// Derive a friendly app name from the bundle identifier
		name := label
		if strings.Contains(label, ".") {
			parts := strings.Split(label, ".")
			if len(parts) > 0 {
				name = parts[len(parts)-1]
			}
		}

		processes = append(processes, engine.AppProcess{
			PID:         int32(pid),
			Name:        name,
			PackageName: label,
			BuildType:   engine.BuildUnknown,
		})
	}

	return processes
}

// cleanLaunchLabel strips UIKitApplication: prefix and [bracket] suffixes
// from a launchctl label.
//
//	"UIKitApplication:in.thetatva.tatva[7140][rb-legacy]" → "in.thetatva.tatva"
//	"com.apple.springboard" → "com.apple.springboard"
func cleanLaunchLabel(label string) string {
	// Strip UIKitApplication: prefix
	if idx := strings.Index(label, ":"); idx >= 0 {
		label = label[idx+1:]
	}
	// Strip [bracket] suffixes
	if idx := strings.Index(label, "["); idx >= 0 {
		label = label[:idx]
	}
	return label
}

// parsePSOutput parses the output of `ps -A -o pid,comm` from a simulator.
//
// Expected output format:
//
//	PID	COMM
//	1	/sbin/launchd
//	1234	/System/Library/CoreServices/SpringBoard.app/SpringBoard
//	5678	/var/containers/Bundle/Application/.../Example.app/Example
func parsePSOutput(output string) []engine.AppProcess {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return nil
	}

	var processes []engine.AppProcess
	for _, line := range lines[1:] { // skip header
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// PID is first, rest is the command path
		idx := strings.IndexAny(line, " \t")
		if idx < 0 {
			continue
		}

		pidStr := strings.TrimSpace(line[:idx])
		comm := strings.TrimSpace(line[idx:])

		pid, err := strconv.ParseInt(pidStr, 10, 32)
		if err != nil {
			continue
		}

		// Extract executable name from path
		name := comm
		if lastSlash := strings.LastIndex(comm, "/"); lastSlash >= 0 {
			name = comm[lastSlash+1:]
		}
		// Strip .app extension if present
		name = strings.TrimSuffix(name, ".app")

		// Use the full bundle path as package name reference
		pkgName := name
		if strings.Contains(comm, ".app/") {
			// Extract the .app name as the package identifier
			appIdx := strings.Index(comm, ".app")
			if appIdx > 0 {
				appPath := comm[:appIdx]
				if lastSlash := strings.LastIndex(appPath, "/"); lastSlash >= 0 {
					pkgName = appPath[lastSlash+1:]
				}
			}
		}

		// Skip kernel-like processes
		if strings.HasPrefix(name, "[") && strings.HasSuffix(name, "]") {
			continue
		}

		processes = append(processes, engine.AppProcess{
			PID:         int32(pid),
			Name:        name,
			PackageName: pkgName,
			BuildType:   engine.BuildUnknown,
		})
	}

	return processes
}
