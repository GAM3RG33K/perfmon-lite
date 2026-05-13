package android

import (
	"fmt"
	"testing"

	"github.com/w1n/perfmon/internal/engine"
)

// ─── Discovery Tests ────────────────────────────────────────────────────────

func TestParseDevicesOutput_Empty(t *testing.T) {
	devices := parseDevicesOutput("")
	if len(devices) != 0 {
		t.Fatalf("expected 0 devices, got %d", len(devices))
	}
}

func TestParseDevicesOutput_HeaderOnly(t *testing.T) {
	devices := parseDevicesOutput("List of devices attached\n")
	if len(devices) != 0 {
		t.Fatalf("expected 0 devices, got %d", len(devices))
	}
}

func TestParseDevicesOutput_Emulator(t *testing.T) {
	output := `List of devices attached
emulator-5554          device product:sdk_gphone16k_arm64 model:pixel_8 device:emu64a transport_id:5
`
	devices := parseDevicesOutput(output)
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}
	dev := devices[0]
	if dev.ID != "emulator-5554" {
		t.Fatalf("expected ID emulator-5554, got %s", dev.ID)
	}
	if dev.Platform != engine.PlatformAndroid {
		t.Fatalf("expected platform android, got %s", dev.Platform)
	}
	if dev.IsPhysical {
		t.Fatal("expected emulator to not be physical")
	}
	if !dev.IsBooted {
		t.Fatal("expected device to be booted")
	}
	if dev.Name != "pixel_8" {
		t.Fatalf("expected name pixel_8, got %s", dev.Name)
	}
}

func TestParseDevicesOutput_PhysicalDevice(t *testing.T) {
	output := `List of devices attached
RF8M21M6DEF          device product:husky model:Pixel_8 device:shusky transport_id:1
`
	devices := parseDevicesOutput(output)
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}
	dev := devices[0]
	if dev.ID != "RF8M21M6DEF" {
		t.Fatalf("expected ID RF8M21M6DEF, got %s", dev.ID)
	}
	if !dev.IsPhysical {
		t.Fatal("expected physical device to be physical")
	}
	if dev.Name != "Pixel_8" {
		t.Fatalf("expected name Pixel_8, got %s", dev.Name)
	}
}

func TestParseDevicesOutput_MultipleDevices(t *testing.T) {
	output := `List of devices attached
emulator-5554          device product:sdk_gphone16k_arm64 model:pixel_8 device:emu64a transport_id:5
RF8M21M6DEF            device product:husky model:Pixel_8 device:shusky transport_id:1
`
	devices := parseDevicesOutput(output)
	if len(devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(devices))
	}
}

func TestParseDevicesOutput_OfflineDevice(t *testing.T) {
	output := `List of devices attached
emulator-5554          offline
`
	devices := parseDevicesOutput(output)
	if len(devices) != 0 {
		t.Fatalf("expected 0 devices (offline), got %d", len(devices))
	}
}

func TestParseDevicesOutput_UnauthorizedDevice(t *testing.T) {
	output := `List of devices attached
RF8M21M6DEF            unauthorized
`
	devices := parseDevicesOutput(output)
	if len(devices) != 0 {
		t.Fatalf("expected 0 devices (unauthorized), got %d", len(devices))
	}
}

func TestParseDevicesOutput_NoModel(t *testing.T) {
	output := `List of devices attached
emulator-5554          device product:sdk_gphone16k_arm64 transport_id:5
`
	devices := parseDevicesOutput(output)
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}
	if devices[0].Name != "" {
		t.Fatalf("expected empty name, got %s", devices[0].Name)
	}
}

func TestIsEmulator(t *testing.T) {
	tests := []struct {
		serial  string
		isEmu   bool
		desc    string
	}{
		{"emulator-5554", true, "standard emulator port"},
		{"emulator-5556", true, "alternative emulator port"},
		{"RF8M21M6DEF", false, "physical device serial"},
		{"0123456789ABCDEF", false, "physical device serial hex"},
		{"127.0.0.1:5555", false, "TCP/IP connected device"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := isEmulator(tt.serial)
			if got != tt.isEmu {
				t.Fatalf("isEmulator(%q) = %v, want %v", tt.serial, got, tt.isEmu)
			}
		})
	}
}

// ─── Process Parsing Tests ──────────────────────────────────────────────────

func TestParsePsOutput_Empty(t *testing.T) {
	procs := parsePsOutput("")
	if len(procs) != 0 {
		t.Fatalf("expected 0 processes, got %d", len(procs))
	}
}

func TestParsePsOutput_HeaderOnly(t *testing.T) {
	output := `USER           PID  PPID        VSZ    RSS WCHAN            ADDR S NAME
`
	procs := parsePsOutput(output)
	if len(procs) != 0 {
		t.Fatalf("expected 0 processes, got %d", len(procs))
	}
}

func TestParsePsOutput_AppProcess(t *testing.T) {
	output := `USER           PID  PPID        VSZ    RSS WCHAN            ADDR S NAME
u0_a123        4567  1234    1234567    89123 0                   0 S com.example.app
`
	procs := parsePsOutput(output)
	if len(procs) != 1 {
		t.Fatalf("expected 1 process, got %d", len(procs))
	}
	proc := procs[0]
	if proc.PID != 4567 {
		t.Fatalf("expected PID 4567, got %d", proc.PID)
	}
	if proc.Name != "com.example.app" {
		t.Fatalf("expected name com.example.app, got %s", proc.Name)
	}
	if proc.PackageName != "com.example.app" {
		t.Fatalf("expected package com.example.app, got %s", proc.PackageName)
	}
	if proc.BuildType != engine.BuildUnknown {
		t.Fatalf("expected BuildUnknown, got %s", proc.BuildType)
	}
}

func TestParsePsOutput_KernelThreadsFiltered(t *testing.T) {
	output := `USER           PID  PPID        VSZ    RSS WCHAN            ADDR S NAME
root             1     0   10934576    768 0                   0 S init
root             2     0          0      0 0                   0 S [kthreadd]
root             3     2          0      0 0                   0 I [kworker/0:0]
u0_a123        4567  1234    1234567    89123 0                   0 S com.example.app
`
	procs := parsePsOutput(output)
	if len(procs) != 2 {
		t.Fatalf("expected 2 processes (init + app), got %d", len(procs))
	}
	// First should be init (PID 1)
	if procs[0].PID != 1 {
		t.Fatalf("expected first process PID 1 (init), got %d", procs[0].PID)
	}
	// Second should be the app
	if procs[1].PID != 4567 {
		t.Fatalf("expected second process PID 4567, got %d", procs[1].PID)
	}
}

func TestParsePsOutput_MultipleApps(t *testing.T) {
	output := `USER           PID  PPID        VSZ    RSS WCHAN            ADDR S NAME
u0_a100        1000   500    2000000   100000 0                   0 S com.android.chrome
u0_a200        2000   500    1500000    80000 0                   0 S com.example.app
u0_a50         3000   500     500000    30000 0                   0 S com.android.systemui
`
	procs := parsePsOutput(output)
	if len(procs) != 3 {
		t.Fatalf("expected 3 processes, got %d", len(procs))
	}
	if procs[1].PID != 2000 || procs[1].Name != "com.example.app" {
		t.Fatalf("invalid process at index 1: PID=%d Name=%s", procs[1].PID, procs[1].Name)
	}
}

// ─── Build Type Detection Tests ─────────────────────────────────────────────

func TestParseBuildType_Debuggable(t *testing.T) {
	output := `Package [com.example.app] (abcdef12):
    userId=10123
    flags=[0x1] DEBUGGABLE
    versionCode=1 targetSdk=34
`
	bt := parseBuildType(output)
	if bt != engine.BuildDebug {
		t.Fatalf("expected BuildDebug, got %s", bt)
	}
}

func TestParseBuildType_Release(t *testing.T) {
	output := `Package [com.example.app] (abcdef12):
    userId=10123
    flags=[0x0]
    versionCode=1 targetSdk=34
`
	bt := parseBuildType(output)
	if bt != engine.BuildRelease {
		t.Fatalf("expected BuildRelease, got %s", bt)
	}
}

func TestParseBuildType_DebuggableAltFormat(t *testing.T) {
	output := `Package [com.example.app]:
    flags=0x1
    pkgFlags=DEBUGGABLE
`
	bt := parseBuildType(output)
	if bt != engine.BuildDebug {
		t.Fatalf("expected BuildDebug, got %s", bt)
	}
}

func TestParseBuildType_EmptyOutput(t *testing.T) {
	bt := parseBuildType("")
	if bt != engine.BuildUnknown {
		t.Fatalf("expected BuildUnknown for empty output, got %s", bt)
	}
}

// ─── Telemetry Parsing Tests ────────────────────────────────────────────────

func TestParseCPU_Found(t *testing.T) {
	topOutput := `Tasks: 200 total, 1 running, 199 sleeping
Mem: 6000000k total, 3000000k used
Swap: 0k total, 0k used
100%cpu 12%user 5%nice 30%sys 53%idle 0%iow 0%irq 0%sirq 0%host
  PID   USER   PR  NI  VIRT  RES   SHR  S  %CPU  %MEM   TIME+   ARGS
 4567   u0_a   20   0  1.2G  120M  80M  S  12.5   2.3    0:15.34  com.example.app
`
	cpu := parseCPU(topOutput, 4567)
	if cpu < 12.0 || cpu > 13.0 {
		t.Fatalf("expected CPU ~12.5, got %f", cpu)
	}
}

func TestParseCPU_NotFound(t *testing.T) {
	topOutput := `  PID   USER   PR  NI  VIRT  RES   SHR  S  %CPU  %MEM   TIME+   ARGS
 4567   u0_a   20   0  1.2G  120M  80M  S  12.5   2.3    0:15.34  com.example.app
`
	cpu := parseCPU(topOutput, 9999)
	if cpu != -1 {
		t.Fatalf("expected -1 for unknown PID, got %f", cpu)
	}
}

func TestParseCPU_EmptyOutput(t *testing.T) {
	cpu := parseCPU("", 1234)
	if cpu != -1 {
		t.Fatalf("expected -1 for empty output, got %f", cpu)
	}
}

func TestParseCPU_ZeroCPU(t *testing.T) {
	topOutput := `  PID   USER   PR  NI  VIRT  RES   SHR  S  %CPU  %MEM   TIME+   ARGS
 4567   u0_a   20   0  1.2G  120M  80M  S   0.0   2.3    0:15.34  com.example.app
`
	cpu := parseCPU(topOutput, 4567)
	if cpu != 0.0 {
		t.Fatalf("expected CPU 0.0, got %f", cpu)
	}
}

func TestParseCPU_VariedColumnPositions(t *testing.T) {
	// Some top versions may have slightly different spacing
	topOutput := `  PID   USER   PR  NI  VIRT  RES   SHR  S  %CPU  %MEM   TIME+   ARGS
 4567   u0_a   20   0  1.2G  120M  80M  S   8.2   2.3    0:15.34  com.example.app
`
	cpu := parseCPU(topOutput, 4567)
	if cpu < 8.0 || cpu > 8.5 {
		t.Fatalf("expected CPU ~8.2, got %f", cpu)
	}
}

func TestParseVmRSS_Found(t *testing.T) {
	statusOutput := `Name:   com.example.app
Umask:  0077
State:  S (sleeping)
VmRSS:     120480 kB
Threads:	15
`
	mem := parseVmRSS(statusOutput)
	if mem != 120480 {
		t.Fatalf("expected VmRSS 120480, got %d", mem)
	}
}

func TestParseVmRSS_NotFound(t *testing.T) {
	mem := parseVmRSS("Name: test\nState: S\n")
	if mem != -1 {
		t.Fatalf("expected -1, got %d", mem)
	}
}

func TestParseVmRSS_EmptyOutput(t *testing.T) {
	mem := parseVmRSS("")
	if mem != -1 {
		t.Fatalf("expected -1, got %d", mem)
	}
}

func TestParseVmRSS_Zero(t *testing.T) {
	statusOutput := `VmRSS:     0 kB
`
	mem := parseVmRSS(statusOutput)
	if mem != 0 {
		t.Fatalf("expected 0, got %d", mem)
	}
}

func TestParseThreads_Found(t *testing.T) {
	statusOutput := `Name:   com.example.app
Threads:	15
`
	threads := parseThreads(statusOutput)
	if threads != 15 {
		t.Fatalf("expected 15 threads, got %d", threads)
	}
}

func TestParseThreads_NotFound(t *testing.T) {
	threads := parseThreads("Name: test\nState: S\n")
	if threads != -1 {
		t.Fatalf("expected -1, got %d", threads)
	}
}

func TestParseThreads_EmptyOutput(t *testing.T) {
	threads := parseThreads("")
	if threads != -1 {
		t.Fatalf("expected -1, got %d", threads)
	}
}

func TestParseThreads_ManyThreads(t *testing.T) {
	statusOutput := `Threads:	256
`
	threads := parseThreads(statusOutput)
	if threads != 256 {
		t.Fatalf("expected 256 threads, got %d", threads)
	}
}

// ─── Preflight Tests ────────────────────────────────────────────────────────

func TestParseAdbVersion(t *testing.T) {
	output := "Android Debug Bridge version 1.0.41\nVersion 36.0.0-13206524\nInstalled as /usr/local/bin/adb\n"
	v, err := ParseAdbVersion(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Major != 36 {
		t.Fatalf("expected major 36, got %d", v.Major)
	}
	if v.Minor != 0 {
		t.Fatalf("expected minor 0, got %d", v.Minor)
	}
	if v.Patch != 0 {
		t.Fatalf("expected patch 0, got %d", v.Patch)
	}
}

func TestParseAdbVersion_NoMatch(t *testing.T) {
	_, err := ParseAdbVersion("not an adb version string")
	if err == nil {
		t.Fatal("expected error for invalid version string")
	}
}

func TestParseAdbVersion_EmptyOutput(t *testing.T) {
	_, err := ParseAdbVersion("")
	if err == nil {
		t.Fatal("expected error for empty output")
	}
}

func TestAdbVersion_String(t *testing.T) {
	v := AdbVersion{Major: 36, Minor: 0, Patch: 0}
	if v.String() != "36.0.0" {
		t.Fatalf("expected '36.0.0', got '%s'", v.String())
	}
}

// ─── Provider Tests ─────────────────────────────────────────────────────────

func TestNewProvider(t *testing.T) {
	p := NewProvider("/path/to/adb")
	if p.AdbPath != "/path/to/adb" {
		t.Fatalf("expected adb path /path/to/adb, got %s", p.AdbPath)
	}
	if p.DeviceID != "" {
		t.Fatalf("expected empty device ID, got %s", p.DeviceID)
	}
}

func TestSetDevice(t *testing.T) {
	p := NewProvider("/path/to/adb")
	p.SetDevice("emulator-5554")
	if p.DeviceID != "emulator-5554" {
		t.Fatalf("expected device ID emulator-5554, got %s", p.DeviceID)
	}
}

func TestAdbCommand_NoDevice(t *testing.T) {
	p := NewProvider("/path/to/adb")
	cmd := p.adbCommand("devices", "-l")
	if cmd.Path != "/path/to/adb" {
		t.Fatalf("expected path /path/to/adb, got %s", cmd.Path)
	}
}

func TestAdbCommand_WithDevice(t *testing.T) {
	p := NewProvider("/path/to/adb")
	p.SetDevice("emulator-5554")
	cmd := p.adbCommand("shell", "echo", "test")

	// Should include -s emulator-5554 before the args
	args := cmd.Args
	if len(args) < 3 || args[1] != "-s" || args[2] != "emulator-5554" {
		t.Fatalf("expected args to include -s emulator-5554, got %v", args)
	}
}

func TestClose(t *testing.T) {
	p := NewProvider("/path/to/adb")
	if err := p.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ─── Interface Compliance ───────────────────────────────────────────────────

func TestProviderImplementsInterfaces(t *testing.T) {
	// Compile-time checks
	var _ engine.PlatformProvider = (*ADBProvider)(nil)
	var _ engine.DeviceDiscovery = (*ADBProvider)(nil)
	var _ engine.ProcessMapper = (*ADBProvider)(nil)
	var _ engine.TelemetryProvider = (*ADBProvider)(nil)
}

// ─── AdbError Tests ─────────────────────────────────────────────────────────

func TestAdbError(t *testing.T) {
	err := &AdbError{
		Args:   []string{"shell", "ls"},
		Output: "device not found",
		Err:    fmt.Errorf("exit status 1"),
	}
	errStr := err.Error()
	if errStr == "" {
		t.Fatal("expected non-empty error message")
	}
	if err.Unwrap() == nil {
		t.Fatal("expected unwrapped error")
	}
}
