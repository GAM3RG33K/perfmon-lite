package views

import (
	"strings"
	"testing"

	"github.com/GAM3RG33K/perfmon-lite/internal/engine"
)

func TestNewTargetSelectorView(t *testing.T) {
	ts := NewTargetSelectorView()
	if ts.Width != 80 {
		t.Errorf("expected width 80, got %d", ts.Width)
	}
	if ts.Height != 24 {
		t.Errorf("expected height 24, got %d", ts.Height)
	}
	if ts.SelectedDevice != 0 {
		t.Errorf("expected selected device 0, got %d", ts.SelectedDevice)
	}
	if ts.SelectedProcess != 0 {
		t.Errorf("expected selected process 0, got %d", ts.SelectedProcess)
	}
	if ts.ShowProcesses {
		t.Error("ShowProcesses should be false by default")
	}
}

func TestTargetSelectorView_Render_Narrow(t *testing.T) {
	ts := NewTargetSelectorView()
	ts.Width = 30
	output := ts.Render()

	if output != "Terminal too narrow for target selector" {
		t.Errorf("expected narrow message, got %q", output)
	}
}

func TestTargetSelectorView_Render_Empty(t *testing.T) {
	ts := NewTargetSelectorView()
	ts.Width = 80
	output := ts.Render()

	if !strings.Contains(output, "Target Selection") {
		t.Error("Render should contain 'Target Selection' header")
	}
	if !strings.Contains(output, "Devices") {
		t.Error("Render should contain 'Devices' section")
	}
	if !strings.Contains(output, "Processes") {
		t.Error("Render should contain 'Processes' section")
	}
	if !strings.Contains(output, "No devices found") {
		t.Error("Render should show 'No devices found' when empty")
	}
}

func TestTargetSelectorView_Render_WithDevices(t *testing.T) {
	ts := NewTargetSelectorView()
	ts.Width = 80
	ts.Devices = []engine.Device{
		{ID: "emulator-5554", Name: "Pixel 8", Platform: engine.PlatformAndroid, IsPhysical: true, IsBooted: true},
		{ID: "ABC-123", Name: "iPhone 15", Platform: engine.PlatformIOS, IsPhysical: false, IsBooted: true},
	}

	output := ts.Render()

	if strings.Contains(output, "No devices found") {
		t.Error("Render should not show 'No devices found' when devices exist")
	}
	if !strings.Contains(output, "Pixel 8") {
		t.Error("Render should contain device name 'Pixel 8'")
	}
	if !strings.Contains(output, "iPhone 15") {
		t.Error("Render should contain device name 'iPhone 15'")
	}
}

func TestTargetSelectorView_Render_WithProcesses(t *testing.T) {
	ts := NewTargetSelectorView()
	ts.Width = 80
	ts.Devices = []engine.Device{
		{ID: "emulator-5554", Name: "Pixel 8", Platform: engine.PlatformAndroid},
	}
	ts.Processes = []engine.AppProcess{
		{PID: 1234, PackageName: "com.example.app", BuildType: engine.BuildDebug},
		{PID: 5678, PackageName: "com.other.app", BuildType: engine.BuildRelease},
	}
	ts.ShowProcesses = true
	ts.SelectedProcess = 0

	output := ts.Render()

	if !strings.Contains(output, "com.example.app") {
		t.Error("Render should contain process package name")
	}
	if !strings.Contains(output, "PID") {
		t.Error("Render should contain 'PID'")
	}
}

func TestTargetSelectorView_Render_EmptyProcesses(t *testing.T) {
	ts := NewTargetSelectorView()
	ts.Width = 80
	ts.Devices = []engine.Device{
		{ID: "emulator-5554", Name: "Pixel 8", Platform: engine.PlatformAndroid},
	}

	output := ts.Render()

	if !strings.Contains(output, "No processes found") {
		t.Error("Render should show 'No processes found' when no processes")
	}
}
