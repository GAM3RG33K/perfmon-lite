package ios

import (
	"testing"

	"github.com/GAM3RG33K/perfmon-lite/internal/engine"
)

// ─── Preflight Tests ─────────────────────────────────────────────────────────

func TestParseVersion(t *testing.T) {
	v, err := parseVersion("xcrun version 72.\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Major != 72 {
		t.Fatalf("expected major 72, got %d", v.Major)
	}
	if v.String() != "72.0.0" {
		t.Fatalf("expected '72.0.0', got '%s'", v.String())
	}
}

func TestParseVersion_NoMatch(t *testing.T) {
	_, err := parseVersion("not an xcrun version string")
	if err == nil {
		t.Fatal("expected error for invalid version string")
	}
}

func TestParseVersion_EmptyOutput(t *testing.T) {
	_, err := parseVersion("")
	if err == nil {
		t.Fatal("expected error for empty output")
	}
}

// ─── Discovery Tests ─────────────────────────────────────────────────────────

func TestParseSimctlDevices_Empty(t *testing.T) {
	devices := parseSimctlDevices("")
	if len(devices) != 0 {
		t.Fatalf("expected 0 devices, got %d", len(devices))
	}
}

func TestParseSimctlDevices_NoBooted(t *testing.T) {
	output := `-- iOS 26.0 --
    iPhone 17 Pro (ABCDEFAB-CDEF-ABCD-EFAB-CDEFABCDEFAB) (Shutdown)
    iPad Air 11-inch (M3) (00000000-0000-0000-0000-000000000000) (Shutdown)
`
	devices := parseSimctlDevices(output)
	if len(devices) != 0 {
		t.Fatalf("expected 0 devices (all shutdown), got %d", len(devices))
	}
}

func TestParseSimctlDevices_Booted(t *testing.T) {
	output := `-- iOS 26.0 --
    iPhone 17 Pro (ABCDEFAB-CDEF-ABCD-EFAB-CDEFABCDEFAB) (Booted)
    iPad Air 11-inch (M3) (00000000-0000-0000-0000-000000000000) (Shutdown)
`
	devices := parseSimctlDevices(output)
	if len(devices) != 1 {
		t.Fatalf("expected 1 booted device, got %d", len(devices))
	}
	dev := devices[0]
	if dev.ID != "ABCDEFAB-CDEF-ABCD-EFAB-CDEFABCDEFAB" {
		t.Fatalf("expected UUID, got %s", dev.ID)
	}
	if dev.Name != "iPhone 17 Pro" {
		t.Fatalf("expected name 'iPhone 17 Pro', got %s", dev.Name)
	}
	if dev.Platform != engine.PlatformIOS {
		t.Fatalf("expected platform ios, got %s", dev.Platform)
	}
	if dev.IsPhysical {
		t.Fatal("expected simulator to not be physical")
	}
	if !dev.IsBooted {
		t.Fatal("expected device to be booted")
	}
}

func TestParseSimctlDevices_MultipleBooted(t *testing.T) {
	output := `-- iOS 26.0 --
    iPhone 17 Pro (ABCDEFAB-CDEF-ABCD-EFAB-CDEFABCDEFAB) (Booted)
    iPad Air 11-inch (M3) (00000000-0000-0000-0000-000000000000) (Booted)
`
	devices := parseSimctlDevices(output)
	if len(devices) != 2 {
		t.Fatalf("expected 2 booted devices, got %d", len(devices))
	}
}

func TestParseSimctlDevices_CaseInsensitiveBooted(t *testing.T) {
	output := `-- iOS 26.0 --
    iPhone 17 Pro (ABCDEFAB-CDEF-ABCD-EFAB-CDEFABCDEFAB) (booted)
`
	devices := parseSimctlDevices(output)
	if len(devices) != 1 {
		t.Fatalf("expected 1 device (case-insensitive booted), got %d", len(devices))
	}
}

func TestParseSimctlDevices_MultipleRuntimes(t *testing.T) {
	output := `-- iOS 26.0 --
    iPhone 17 Pro (ABCDEFAB-CDEF-ABCD-EFAB-CDEFABCDEFAB) (Booted)
-- tvOS 18.0 --
    Apple TV (11111111-1111-1111-1111-111111111111) (Booted)
`
	devices := parseSimctlDevices(output)
	if len(devices) != 2 {
		t.Fatalf("expected 2 devices from multiple runtimes, got %d", len(devices))
	}
}

// ─── Devicectl Discovery Tests ───────────────────────────────────────────────

func TestParseDevicectlOutput_Empty(t *testing.T) {
	devices, err := parseDevicectlOutput(`{"devices":[],"info":{"errorCode":0,"error":""}}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 0 {
		t.Fatalf("expected 0 devices, got %d", len(devices))
	}
}

func TestParseDevicectlOutput_Connected(t *testing.T) {
	output := `{
		"devices": [
			{
				"properties": {
					"name": "My iPhone",
					"osType": "iOS",
					"osVersion": "18.2",
					"serialNumber": "ABCD1234",
					"connectionKind": "usb",
					"state": "connected"
				}
			}
		],
		"info": {"errorCode": 0, "error": ""}
	}`
	devices, err := parseDevicectlOutput(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}
	dev := devices[0]
	if dev.ID != "ABCD1234" {
		t.Fatalf("expected serial ABCD1234, got %s", dev.ID)
	}
	if dev.Name != "My iPhone" {
		t.Fatalf("expected name 'My iPhone', got %s", dev.Name)
	}
	if !dev.IsPhysical {
		t.Fatal("expected physical device")
	}
	if !dev.IsBooted {
		t.Fatal("expected device to be booted")
	}
}

func TestParseDevicectlOutput_NotConnected(t *testing.T) {
	output := `{
		"devices": [
			{
				"properties": {
					"name": "My iPhone",
					"osType": "iOS",
					"osVersion": "18.2",
					"serialNumber": "ABCD1234",
					"connectionKind": "usb",
					"state": "disconnected"
				}
			}
		],
		"info": {"errorCode": 0, "error": ""}
	}`
	devices, err := parseDevicectlOutput(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 0 {
		t.Fatalf("expected 0 devices (disconnected), got %d", len(devices))
	}
}

func TestParseDevicectlOutput_NonIOS(t *testing.T) {
	output := `{
		"devices": [
			{
				"properties": {
					"name": "Apple Watch",
					"osType": "watchOS",
					"serialNumber": "WATCH001",
					"state": "connected"
				}
			}
		],
		"info": {"errorCode": 0, "error": ""}
	}`
	devices, err := parseDevicectlOutput(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 0 {
		t.Fatalf("expected 0 devices (non-iOS), got %d", len(devices))
	}
}

func TestParseDevicectlOutput_NoSerial(t *testing.T) {
	output := `{
		"devices": [
			{
				"properties": {
					"name": "My iPhone",
					"osType": "iOS",
					"state": "connected"
				}
			}
		],
		"info": {"errorCode": 0, "error": ""}
	}`
	devices, err := parseDevicectlOutput(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("expected 1 device (with generated ID), got %d", len(devices))
	}
	if devices[0].ID != "ios-My iPhone" {
		t.Fatalf("expected fallback ID 'ios-My iPhone', got %s", devices[0].ID)
	}
}

func TestParseDevicectlOutput_Error(t *testing.T) {
	output := `{"devices":[],"info":{"errorCode":-1,"error":"no devices available"}}`
	_, err := parseDevicectlOutput(output)
	if err == nil {
		t.Fatal("expected error for non-zero error code")
	}
}

func TestParseDevicectlOutput_InvalidJSON(t *testing.T) {
	_, err := parseDevicectlOutput("not json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// ─── Process Mapping Tests ───────────────────────────────────────────────────

func TestParseLaunchctlList_Empty(t *testing.T) {
	procs := parseLaunchctlList("")
	if len(procs) != 0 {
		t.Fatalf("expected 0 processes, got %d", len(procs))
	}
}

func TestParseLaunchctlList_HeaderOnly(t *testing.T) {
	output := "PID\tStatus\tLabel\n"
	procs := parseLaunchctlList(output)
	if len(procs) != 0 {
		t.Fatalf("expected 0 processes, got %d", len(procs))
	}
}

func TestParseLaunchctlList_RunningProcesses(t *testing.T) {
	output := `PID	Status	Label
1	0	com.apple.launchd
1234	0	com.apple.springboard
5678	0	com.example.app
`
	procs := parseLaunchctlList(output)
	if len(procs) != 3 {
		t.Fatalf("expected 3 processes, got %d", len(procs))
	}
	if procs[0].PID != 1 {
		t.Fatalf("expected PID 1, got %d", procs[0].PID)
	}
	if procs[2].PID != 5678 {
		t.Fatalf("expected PID 5678, got %d", procs[2].PID)
	}
	if procs[2].PackageName != "com.example.app" {
		t.Fatalf("expected package 'com.example.app', got %s", procs[2].PackageName)
	}
	if procs[2].Name != "app" {
		t.Fatalf("expected name 'app' (last component), got %s", procs[2].Name)
	}
}

func TestParseLaunchctlList_SkipNotRunning(t *testing.T) {
	output := `PID	Status	Label
1	0	com.apple.launchd
-	78	com.apple.notrunning
1234	0	com.example.app
`
	procs := parseLaunchctlList(output)
	if len(procs) != 2 {
		t.Fatalf("expected 2 processes (skipping not-running), got %d", len(procs))
	}
}

func TestParsePSOutput_Empty(t *testing.T) {
	procs := parsePSOutput("")
	if len(procs) != 0 {
		t.Fatalf("expected 0 processes, got %d", len(procs))
	}
}

func TestParsePSOutput_HeaderOnly(t *testing.T) {
	output := "PID\tCOMM\n"
	procs := parsePSOutput(output)
	if len(procs) != 0 {
		t.Fatalf("expected 0 processes, got %d", len(procs))
	}
}

func TestParsePSOutput_Processes(t *testing.T) {
	output := `PID	COMM
1	/sbin/launchd
1234	/System/Library/CoreServices/SpringBoard.app/SpringBoard
5678	/var/containers/Bundle/Application/ABCD/Example.app/Example
`
	procs := parsePSOutput(output)
	if len(procs) != 3 {
		t.Fatalf("expected 3 processes, got %d", len(procs))
	}
	if procs[0].PID != 1 || procs[0].Name != "launchd" {
		t.Fatalf("expected launchd (PID 1), got %s (PID %d)", procs[0].Name, procs[0].PID)
	}
	if procs[2].PID != 5678 || procs[2].Name != "Example" {
		t.Fatalf("expected Example app, got %s (PID %d)", procs[2].Name, procs[2].PID)
	}
}

// ─── Telemetry Parsing Tests ─────────────────────────────────────────────────
// ps -p <pid> -o pid=,%cpu=,rss=,comm=  →  no header, space-separated

func TestParsePSSample_Found(t *testing.T) {
	output := "1234 12.5 46080 /path/to/ExampleApp\n"
	snapshot := parsePSSample(output, 1234)
	if snapshot == nil {
		t.Fatal("expected snapshot, got nil")
	}
	if snapshot.CPUPercent < 12.0 || snapshot.CPUPercent > 13.0 {
		t.Fatalf("expected CPU ~12.5, got %f", snapshot.CPUPercent)
	}
	if snapshot.MemoryKB != 46080 {
		t.Fatalf("expected 46080 KB, got %d", snapshot.MemoryKB)
	}
}

func TestParsePSSample_NotFound(t *testing.T) {
	output := "9999 12.5 46080 /path/to/Other\n"
	snapshot := parsePSSample(output, 1234)
	if snapshot != nil {
		t.Fatal("expected nil for unknown PID")
	}
}

func TestParsePSSample_Empty(t *testing.T) {
	snapshot := parsePSSample("", 1234)
	if snapshot != nil {
		t.Fatal("expected nil for empty output")
	}
}

func TestParsePSSample_MissingFields(t *testing.T) {
	output := "1234 12.5\n" // missing RSS and COMM
	snapshot := parsePSSample(output, 1234)
	if snapshot != nil {
		t.Fatal("expected nil for missing fields")
	}
}

// ─── Launch Label Cleaning Tests ─────────────────────────────────────────────

func TestCleanLaunchLabel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"UIKitApplication:in.thetatva.tatva[7140][rb-legacy]", "in.thetatva.tatva"},
		{"com.apple.springboard", "com.apple.springboard"},
		{"UIKitApplication:com.apple.Spotlight[eccb][rb-legacy]", "com.apple.Spotlight"},
		{"", ""},
	}

	for _, tt := range tests {
		got := cleanLaunchLabel(tt.input)
		if got != tt.want {
			t.Errorf("cleanLaunchLabel(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// ─── Provider Tests ─────────────────────────────────────────────────────────

func TestNewProvider(t *testing.T) {
	p := NewProvider("/usr/bin/xcrun")
	if p.XcrunPath != "/usr/bin/xcrun" {
		t.Fatalf("expected path /usr/bin/xcrun, got %s", p.XcrunPath)
	}
	if p.DeviceID != "" {
		t.Fatalf("expected empty device ID, got %s", p.DeviceID)
	}
}

func TestSetDevice(t *testing.T) {
	p := NewProvider("/usr/bin/xcrun")
	p.SetDevice("XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX")
	if p.DeviceID != "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX" {
		t.Fatalf("expected device ID, got %s", p.DeviceID)
	}
}

func TestClose(t *testing.T) {
	p := NewProvider("/usr/bin/xcrun")
	if err := p.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ─── Interface Compliance ───────────────────────────────────────────────────

func TestProviderImplementsInterfaces(t *testing.T) {
	var _ engine.PlatformProvider = (*iOSProvider)(nil)
	var _ engine.DeviceDiscovery = (*iOSProvider)(nil)
	var _ engine.ProcessMapper = (*iOSProvider)(nil)
	var _ engine.TelemetryProvider = (*iOSProvider)(nil)
}
