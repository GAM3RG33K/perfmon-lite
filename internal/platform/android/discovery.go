package android

import (
	"strings"

	"github.com/GAM3RG33K/perfmon-lite/internal/engine"
)

// Discover returns a list of connected Android devices by parsing the
// output of `adb devices -l`.
//
// Expected output format:
//
//	List of devices attached
//	emulator-5554          device product:sdk_gphone16k_arm64 model:pixel_8 device:emu64a transport_id:5
//	RF8M21M6DEF            device product:husky model:Pixel_8 device:shusky transport_id:1
func (p *ADBProvider) Discover() ([]engine.Device, error) {
	out, err := p.adbExec("devices", "-l")
	if err != nil {
		return nil, err
	}

	return parseDevicesOutput(out), nil
}

// parseDevicesOutput parses the raw output of `adb devices -l` into device entries.
func parseDevicesOutput(output string) []engine.Device {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return nil
	}

	var devices []engine.Device
	for _, line := range lines[1:] { // skip header line
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		dev := parseDeviceLine(line)
		if dev != nil {
			devices = append(devices, *dev)
		}
	}
	return devices
}

// parseDeviceLine parses a single device line from `adb devices -l`.
//
// Formats accepted:
//
//	<serial> <state> product:<p> model:<m> device:<d> transport_id:<n>
//	<serial> <state> usb:<vendor> product:<p> model:<m> device:<d>
func parseDeviceLine(line string) *engine.Device {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return nil
	}

	serial := fields[0]
	state := fields[1]

	if state != "device" {
		return nil
	}

	dev := &engine.Device{
		ID:         serial,
		Platform:   engine.PlatformAndroid,
		IsBooted:   true,
		IsPhysical: !isEmulator(serial),
	}

	// Parse key=value properties from remaining fields
	for _, f := range fields[2:] {
		if strings.HasPrefix(f, "model:") {
			dev.Name = strings.TrimPrefix(f, "model:")
			dev.Name = strings.TrimSpace(dev.Name)
		}
	}

	return dev
}

// isEmulator returns true if the device serial indicates an emulator.
// Emulator serials typically start with "emulator-" (e.g., emulator-5554).
func isEmulator(serial string) bool {
	return strings.HasPrefix(serial, "emulator-")
}
