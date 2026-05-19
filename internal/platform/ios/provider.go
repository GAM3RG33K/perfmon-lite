package ios

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/GAM3RG33K/perfmon-lite/internal/engine"
)

// iOSProvider implements the PlatformProvider interface for iOS simulators
// and physical devices. It communicates with devices via xcrun (simctl/devicectl)
// and uses xcrun simctl spawn to run commands inside booted simulators.
type iOSProvider struct {
	XcrunPath string // path to the xcrun binary
	DeviceID  string // target device UDID for telemetry sampling

	// Device cache — populated lazily by Discover()
	devices []engine.Device
	devMu   sync.RWMutex

	mu sync.Mutex // protects DeviceID
}

// NewProvider creates a new iOS provider using the given xcrun binary path.
func NewProvider(xcrunPath string) *iOSProvider {
	return &iOSProvider{
		XcrunPath: xcrunPath,
	}
}

// SetDevice sets the target device for telemetry operations.
func (p *iOSProvider) SetDevice(deviceID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.DeviceID = deviceID
}

// getDevice returns the device info for the given device ID from the cache.
func (p *iOSProvider) getDevice(deviceID string) (engine.Device, error) {
	p.devMu.RLock()
	defer p.devMu.RUnlock()

	for _, d := range p.devices {
		if d.ID == deviceID {
			return d, nil
		}
	}

	return engine.Device{}, fmt.Errorf("device %s not found in cache", deviceID)
}

// CacheDevices stores the discovered device list for later lookups.
// Called after Discover() to cache device metadata.
func (p *iOSProvider) CacheDevices(devices []engine.Device) {
	p.devMu.Lock()
	defer p.devMu.Unlock()
	p.devices = devices
}

// ─── xcrun command helpers ──────────────────────────────────────────────────

// xcrunExec runs xcrun with the given arguments and returns the combined output.
func (p *iOSProvider) xcrunExec(args ...string) (string, error) {
	cmd := exec.Command(p.XcrunPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("xcrun %s: %s (%w)", strings.Join(args, " "), strings.TrimSpace(string(out)), err)
	}
	return string(out), nil
}

// simctlSpawn runs a command inside a booted simulator via `xcrun simctl spawn`.
func (p *iOSProvider) simctlSpawn(udid string, args ...string) (string, error) {
	spawnArgs := append([]string{"simctl", "spawn", udid}, args...)
	return p.xcrunExec(spawnArgs...)
}

// ─── Interface compliance ────────────────────────────────────────────────────

// Close releases provider resources.
func (p *iOSProvider) Close() error {
	return nil
}

// ─── Log capture (simulator log stream) ────────────────────────────────────

// CaptureLogs fetches recent log entries from the iOS simulator for the given bundle.
func (p *iOSProvider) CaptureLogs(pid int32) ([]string, error) {
	p.mu.Lock()
	deviceID := p.DeviceID
	p.mu.Unlock()

	if deviceID == "" {
		return nil, fmt.Errorf("no device ID set")
	}

	// Get the device info to check if it's a simulator
	device, err := p.getDevice(deviceID)
	if err != nil {
		return nil, fmt.Errorf("device lookup: %w", err)
	}
	if device.IsPhysical {
		return nil, fmt.Errorf("log capture not supported on physical iOS devices")
	}

	// Use log stream --last to get recent logs, then grep for the PID
	out, err := p.simctlSpawn(deviceID, "log", "stream", "--last", "2m", "--style", "compact")
	if err != nil {
		return nil, fmt.Errorf("log stream: %w", err)
	}

	// Filter for lines containing our PID
	lines := strings.Split(out, "\n")
	var filtered []string
	pidStr := strconv.FormatInt(int64(pid), 10)
	for _, line := range lines {
		if strings.Contains(line, pidStr) {
			filtered = append(filtered, line)
		}
	}
	if len(filtered) > 20 {
		filtered = filtered[len(filtered)-20:]
	}
	return filtered, nil
}

// ─── Interface compliance checks ────────────────────────────────────────────
var _ engine.PlatformProvider = (*iOSProvider)(nil)
var _ engine.LogCapturer = (*iOSProvider)(nil)
var _ engine.DeviceDiscovery = (*iOSProvider)(nil)
var _ engine.ProcessMapper = (*iOSProvider)(nil)
var _ engine.TelemetryProvider = (*iOSProvider)(nil)
