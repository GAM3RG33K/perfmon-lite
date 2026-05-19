package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/GAM3RG33K/perfmon-lite/internal/engine"
	"github.com/GAM3RG33K/perfmon-lite/internal/export"
	"github.com/GAM3RG33K/perfmon-lite/internal/platform/mock"
)

func testModel() *Model {
	eng := engine.NewEngine(mock.NewProvider(42), 300, 0)
	eng.SetTarget(9001)
	return NewModel(eng, true, engine.PlatformMock)
}

func TestNewModel(t *testing.T) {
	eng := engine.NewEngine(mock.NewProvider(42), 300, 0)
	m := NewModel(eng, true, engine.PlatformMock)

	if m == nil {
		t.Fatal("NewModel returned nil")
	}
	if m.Engine != eng {
		t.Error("Engine not set")
	}
	if !m.Mock {
		t.Error("Mock not set")
	}
	if m.Platform != engine.PlatformMock {
		t.Errorf("expected PlatformMock, got %s", m.Platform)
	}
	if m.ActiveTab != TabDashboard {
		t.Errorf("expected TabDashboard, got %d", m.ActiveTab)
	}
	if m.Width != 80 {
		t.Errorf("expected width 80, got %d", m.Width)
	}
	if m.Height != 24 {
		t.Errorf("expected height 24, got %d", m.Height)
	}
	if m.Dashboard == nil {
		t.Error("Dashboard not initialized")
	}
	if m.TargetSelector == nil {
		t.Error("TargetSelector not initialized")
	}
	if m.Logs == nil {
		t.Error("Logs not initialized")
	}
}

func TestModel_Update_WindowSize(t *testing.T) {
	m := testModel()
	_, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if m.Width != 120 {
		t.Errorf("expected width 120, got %d", m.Width)
	}
	if m.Height != 40 {
		t.Errorf("expected height 40, got %d", m.Height)
	}
	if !m.Ready {
		t.Error("Ready not set after WindowSizeMsg")
	}
	if m.Dashboard.Width != 116 {
		t.Errorf("expected dashboard width 116, got %d", m.Dashboard.Width)
	}
}

func TestModel_Update_TabSwitching(t *testing.T) {
	m := testModel()
	m.Ready = true

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.ActiveTab != TabLogs {
		t.Errorf("expected TabLogs after Tab, got %d", m.ActiveTab)
	}

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.ActiveTab != TabDashboard {
		t.Errorf("expected TabDashboard after second Tab, got %d", m.ActiveTab)
	}

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if m.ActiveTab != TabLogs {
		t.Errorf("expected TabLogs after Right, got %d", m.ActiveTab)
	}

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if m.ActiveTab != TabDashboard {
		t.Errorf("expected TabDashboard after Left, got %d", m.ActiveTab)
	}
}

func TestModel_Update_TabJump(t *testing.T) {
	m := testModel()
	m.Ready = true

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})
	if m.ActiveTab != TabLogs {
		t.Errorf("expected TabLogs after '2', got %d", m.ActiveTab)
	}

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	if m.ActiveTab != TabDashboard {
		t.Errorf("expected TabDashboard after '1', got %d", m.ActiveTab)
	}
}

func TestModel_Update_Quit(t *testing.T) {
	m := testModel()
	m.Ready = true

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if !m.Quitting {
		t.Error("Quitting not set after 'q'")
	}
	if cmd == nil {
		t.Error("Expected tea.Quit cmd after 'q'")
	}
}

func TestModel_Update_Help(t *testing.T) {
	m := testModel()
	m.Ready = true

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	if !m.ShowHelp {
		t.Error("ShowHelp not set after '?'")
	}

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.ShowHelp {
		t.Error("ShowHelp not cleared after Esc")
	}
}

func TestModel_Update_LogScroll(t *testing.T) {
	m := testModel()
	m.Ready = true
	m.ActiveTab = TabLogs
	m.Logs.AddEntry("INFO", "test1")
	m.Logs.AddEntry("INFO", "test2")
	m.Logs.AddEntry("INFO", "test3")

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.Logs.ScrollPos != 2 {
		t.Errorf("expected scroll pos 2, got %d", m.Logs.ScrollPos)
	}

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.Logs.ScrollPos != 3 {
		t.Errorf("expected scroll pos 3, got %d", m.Logs.ScrollPos)
	}
}

func TestModel_Update_Refresh(t *testing.T) {
	m := testModel()
	m.Ready = true

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	if m.statusMsg == "" {
		t.Error("statusMsg not set after 'r'")
	}
}

func TestModel_Update_ExportNoData(t *testing.T) {
	m := testModel()
	m.Ready = true

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if m.showFormatPicker {
		t.Error("format picker should not open with no data")
	}
	if m.statusMsg == "" {
		t.Error("statusMsg not set when exporting with no data")
	}
}

func TestModel_Update_ExportPicker(t *testing.T) {
	m := testModel()
	m.Ready = true
	m.Engine.Buffer.Push(engine.NewTelemetrySnapshot(10, 1000, 5))
	m.Engine.Buffer.Push(engine.NewTelemetrySnapshot(20, 2000, 6))

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if !m.showFormatPicker {
		t.Error("format picker should open with data")
	}
	if m.formatPickerIdx != 0 {
		t.Errorf("expected picker idx 0, got %d", m.formatPickerIdx)
	}

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.formatPickerIdx != 1 {
		t.Errorf("expected picker idx 1, got %d", m.formatPickerIdx)
	}

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.formatPickerIdx != 0 {
		t.Errorf("expected picker idx 0, got %d", m.formatPickerIdx)
	}

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.showFormatPicker {
		t.Error("format picker should close on Esc")
	}
}

func TestModel_Update_TelemetryError(t *testing.T) {
	m := testModel()
	m.Ready = true

	msg := engine.TelemetryMsg{Error: errors.New("provider closed")}
	_, _ = m.Update(msg)

	if len(m.Logs.Entries) == 0 {
		t.Fatal("Expected error log entry")
	}
	if m.Logs.Entries[0].Level != "ERROR" {
		t.Errorf("expected ERROR level, got %s", m.Logs.Entries[0].Level)
	}
}

func TestModel_Update_TelemetrySuccess(t *testing.T) {
	m := testModel()
	m.Ready = true

	snap := engine.NewTelemetrySnapshot(75.0, 600000, 50)
	msg := engine.TelemetryMsg{Snapshot: snap}
	_, _ = m.Update(msg)

	found := false
	for _, e := range m.Logs.Entries {
		if e.Level == "TICK" && e.Message == "CPU=75.0% Mem=600000KB Threads=50" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected TICK log entry with telemetry data")
	}
}

func TestModel_Update_HighCPUAlert(t *testing.T) {
	m := testModel()
	m.Ready = true

	snap := engine.NewTelemetrySnapshot(85.0, 100000, 30)
	msg := engine.TelemetryMsg{Snapshot: snap}
	_, _ = m.Update(msg)

	found := false
	for _, e := range m.Logs.Entries {
		if e.Level == "ALERT" && e.Message == "High CPU: 85.0% (threshold: 70%)" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected ALERT log entry for high CPU")
	}
}

func TestModel_Update_HighMemoryAlert(t *testing.T) {
	m := testModel()
	m.Ready = true

	snap := engine.NewTelemetrySnapshot(10.0, 600000, 30)
	msg := engine.TelemetryMsg{Snapshot: snap}
	_, _ = m.Update(msg)

	found := false
	for _, e := range m.Logs.Entries {
		if e.Level == "ALERT" && strings.Contains(e.Message, "High RAM:") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected ALERT log entry for high RAM")
	}
}

func TestModel_View_NotReady(t *testing.T) {
	m := testModel()
	view := m.View()
	if view != "\n  Initializing perfmon..." {
		t.Errorf("unexpected view: %q", view)
	}
}

func TestModel_View_Quitting(t *testing.T) {
	m := testModel()
	m.Ready = true
	m.Quitting = true
	view := m.View()
	if view != "\n  Goodbye!\n" {
		t.Errorf("unexpected view: %q", view)
	}
}

func TestModel_View_Help(t *testing.T) {
	m := testModel()
	m.Ready = true
	m.ShowHelp = true
	view := m.View()
	if view == "" {
		t.Error("Help view should not be empty")
	}
}

func TestModel_SetTargets(t *testing.T) {
	m := testModel()
	devices := []engine.Device{
		{ID: "emulator-5554", Name: "Pixel 8", Platform: engine.PlatformAndroid, IsPhysical: true},
	}
	processes := []engine.AppProcess{
		{PID: 1234, PackageName: "com.example.app", BuildType: engine.BuildDebug},
	}

	m.SetTargets(devices, processes)

	if len(m.TargetSelector.Devices) != 1 {
		t.Errorf("expected 1 device, got %d", len(m.TargetSelector.Devices))
	}
	if len(m.TargetSelector.Processes) != 1 {
		t.Errorf("expected 1 process, got %d", len(m.TargetSelector.Processes))
	}
}

func TestModel_SelectedProcess(t *testing.T) {
	m := testModel()
	m.SetTargets(nil, []engine.AppProcess{
		{PID: 100, PackageName: "com.first.app"},
		{PID: 200, PackageName: "com.second.app"},
	})

	m.AppPID = 200
	p := m.selectedProcess()
	if p == nil || p.PID != 200 {
		t.Error("Should select process with matching PID")
	}

	m.AppPID = 999
	p = m.selectedProcess()
	if p == nil || p.PID != 100 {
		t.Error("Should fall back to first process")
	}

	m.AppPID = 0
	m.TargetSelector.Processes = nil
	m.AppID = "com.test.app"
	p = m.selectedProcess()
	if p == nil || p.PackageName != "com.test.app" {
		t.Error("Should create placeholder process from AppID")
	}
}

func TestModel_SetStatus(t *testing.T) {
	m := testModel()
	m.setStatus("test message")
	if m.statusMsg != "test message" {
		t.Errorf("expected 'test message', got %q", m.statusMsg)
	}
}

func TestModel_HandleTick(t *testing.T) {
	m := testModel()
	m.Ready = true
	model, cmd := m.handleTick()

	if model != m {
		t.Error("handleTick should return same model")
	}
	if cmd == nil {
		t.Error("handleTick should return a command")
	}
}

func TestModel_HandlePickerKey_Navigation(t *testing.T) {
	m := testModel()
	m.showFormatPicker = true
	m.formatPickerIdx = 0

	_, _ = m.handlePickerKey(tea.KeyMsg{Type: tea.KeyDown})
	if m.formatPickerIdx != 1 {
		t.Errorf("expected idx 1, got %d", m.formatPickerIdx)
	}

	_, _ = m.handlePickerKey(tea.KeyMsg{Type: tea.KeyDown})
	if m.formatPickerIdx != 2 {
		t.Errorf("expected idx 2, got %d", m.formatPickerIdx)
	}

	_, _ = m.handlePickerKey(tea.KeyMsg{Type: tea.KeyDown})
	if m.formatPickerIdx != 0 {
		t.Errorf("expected idx 0 (wrap), got %d", m.formatPickerIdx)
	}

	_, _ = m.handlePickerKey(tea.KeyMsg{Type: tea.KeyUp})
	if m.formatPickerIdx != 2 {
		t.Errorf("expected idx 2 (wrap), got %d", m.formatPickerIdx)
	}
}

func TestPickerFormats(t *testing.T) {
	if len(pickerFormats) != 3 {
		t.Errorf("expected 3 formats, got %d", len(pickerFormats))
	}
	if pickerFormats[0] != export.FormatJSON {
		t.Errorf("expected JSON first, got %s", pickerFormats[0])
	}
	if pickerFormats[1] != export.FormatMD {
		t.Errorf("expected MD second, got %s", pickerFormats[1])
	}
	if pickerFormats[2] != export.FormatHTML {
		t.Errorf("expected HTML third, got %s", pickerFormats[2])
	}

	if len(pickerLabels) != 3 {
		t.Errorf("expected 3 labels, got %d", len(pickerLabels))
	}
}

func TestModel_Update_DefaultKey(t *testing.T) {
	m := testModel()
	m.Ready = true
	m.ActiveTab = TabDashboard

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("z")})
	if m.ActiveTab != TabDashboard {
		t.Error("Unknown key should not change tab")
	}
	if m.Quitting {
		t.Error("Unknown key should not quit")
	}
}
