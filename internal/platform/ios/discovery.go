package ios

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/w1n/perfmon/internal/engine"
)

// ─── Simulator Discovery via `xcrun simctl list devices available` ───────────

// Discover returns a list of booted iOS simulators and connected physical devices.
func (p *iOSProvider) Discover() ([]engine.Device, error) {
	// Discover simulators
	simulators, err := p.listSimulators()
	if err != nil {
		return nil, fmt.Errorf("simulator discovery: %w", err)
	}

	// Discover physical devices
	physical, err := p.listPhysicalDevices()
	if err != nil {
		// Physical device discovery is best-effort; return simulators only
		if len(simulators) == 0 {
			return nil, fmt.Errorf("device discovery failed (simulators and physical): simctl: %v, devicectl: %v", simulators, err)
		}
	}

	return append(simulators, physical...), nil
}

// listSimulators discovers booted iOS simulators via `xcrun simctl list devices available`.
func (p *iOSProvider) listSimulators() ([]engine.Device, error) {
	out, err := p.xcrunExec("simctl", "list", "devices", "available")
	if err != nil {
		return nil, err
	}

	return parseSimctlDevices(out), nil
}

// parseSimctlDevices parses the output of `xcrun simctl list devices available`.
//
// Expected output format:
//
//	-- iOS 26.0 --
//	    iPhone 17 Pro (XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX) (Shutdown)
//	    iPad Air 11-inch (M3) (XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX) (Shutdown)
//	-- tvOS 18.0 --
//	    ...
//	-- watchOS 11.0 --
//	    ...
func parseSimctlDevices(output string) []engine.Device {
	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		return nil
	}

	// Regex for runtime version headers: "-- iOS 26.0 --"
	runtimeRe := regexp.MustCompile(`^--\s+(.+?)\s+(\d+\.\d+)\s+--$`)
	// Simple pattern to find any UUID in the line (8-4-4-4-12 hex format)
	uuidRe := regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)

	var devices []engine.Device

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")

		// Skip runtime header lines
		if runtimeRe.MatchString(line) {
			continue
		}

		// Find the UUID in the line
		loc := uuidRe.FindStringIndex(line)
		if loc == nil {
			continue
		}

		uuid := line[loc[0]:loc[1]]

		// Extract the device name — everything before the UUID's opening paren
		beforeUUID := strings.TrimRight(line[:loc[0]], " \t")
		// Trim trailing "(" from the UUID's opening paren
		name := strings.TrimRight(beforeUUID, "(")
		name = strings.TrimSpace(name)

		// Extract the state — text in the last (...) after the UUID
		afterUUID := line[loc[1]:]
		state := ""
		if lastClose := strings.LastIndex(afterUUID, ")"); lastClose >= 0 {
			if lastOpen := strings.LastIndex(afterUUID[:lastClose], "("); lastOpen >= 0 {
				state = strings.TrimSpace(afterUUID[lastOpen+1 : lastClose])
			}
		}

		// Only include booted devices
		if !strings.EqualFold(state, "booted") {
			continue
		}

		devices = append(devices, engine.Device{
			ID:         uuid,
			Name:       name,
			Platform:   engine.PlatformIOS,
			IsPhysical: false,
			IsBooted:   true,
		})
	}

	return devices
}

// ─── Physical Device Discovery via `xcrun devicectl list devices` ────────────

// devicectlDevice represents a single device from devicectl JSON output.
type devicectlDevice struct {
	Properties struct {
		Name           string `json:"name"`
		OSType         string `json:"osType"`
		OSVersion      string `json:"osVersion"`
		SerialNumber   string `json:"serialNumber"`
		ConnectionKind string `json:"connectionKind"`
		State          string `json:"state"`
	} `json:"properties"`
}

// devicectlResponse represents the top-level JSON structure from devicectl.
type devicectlResponse struct {
	Devices []devicectlDevice `json:"devices"`
	Info    struct {
		ErrorCode int    `json:"errorCode"`
		Error     string `json:"error"`
	} `json:"info"`
}

// listPhysicalDevices discovers connected physical iOS devices via `xcrun devicectl list devices`.
// Falls back gracefully if devicectl is not available or returns no devices.
func (p *iOSProvider) listPhysicalDevices() ([]engine.Device, error) {
	out, err := p.xcrunExec("devicectl", "list", "devices", "--json-output")
	if err != nil {
		// devicectl may not be available on older Xcode versions
		return nil, fmt.Errorf("devicectl not available: %w", err)
	}

	return parseDevicectlOutput(out)
}

// parseDevicectlOutput parses the JSON output of `xcrun devicectl list devices --json-output`.
//
// Expected JSON format:
//
//	{
//	  "devices": [
//	    {
//	      "properties": {
//	        "name": "My iPhone",
//	        "osType": "iOS",
//	        "osVersion": "18.2",
//	        "serialNumber": "XXXXXXXXXXXXXXXX",
//	        "connectionKind": "usb",
//	        "state": "connected"
//	      }
//	    }
//	  ],
//	  "info": { "errorCode": 0, "error": "" }
//	}
func parseDevicectlOutput(output string) ([]engine.Device, error) {
	var resp devicectlResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse devicectl JSON: %w", err)
	}

	if resp.Info.ErrorCode != 0 {
		return nil, fmt.Errorf("devicectl error: %s (code %d)", resp.Info.Error, resp.Info.ErrorCode)
	}

	var devices []engine.Device
	for _, d := range resp.Devices {
		// Only include connected iOS devices
		if d.Properties.State != "connected" {
			continue
		}
		if !strings.HasPrefix(strings.ToLower(d.Properties.OSType), "ios") {
			continue
		}

		devID := d.Properties.SerialNumber
		if devID == "" {
			devID = fmt.Sprintf("ios-%s", d.Properties.Name)
		}

		devices = append(devices, engine.Device{
			ID:         devID,
			Name:       d.Properties.Name,
			Platform:   engine.PlatformIOS,
			IsPhysical: true,
			IsBooted:   true,
		})
	}

	return devices, nil
}
