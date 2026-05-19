package views

import (
	"strings"
	"testing"

	"github.com/GAM3RG33K/perfmon-lite/internal/engine"
)

func TestNewDashboardView(t *testing.T) {
	dv := NewDashboardView()
	if dv.Width != 80 {
		t.Errorf("expected width 80, got %d", dv.Width)
	}
	if dv.Height != 24 {
		t.Errorf("expected height 24, got %d", dv.Height)
	}
}

func TestDashboardView_Render_Narrow(t *testing.T) {
	dv := NewDashboardView()
	dv.Width = 30
	output := dv.Render(nil, nil)

	if output != "Terminal too narrow for dashboard" {
		t.Errorf("expected narrow message, got %q", output)
	}
}

func TestDashboardView_Render_NoData(t *testing.T) {
	dv := NewDashboardView()
	dv.Width = 80
	output := dv.Render(nil, nil)

	if !strings.Contains(output, "No data") {
		t.Error("Render should show 'No data' message when no telemetry")
	}
	if !strings.Contains(output, "waiting for telemetry") {
		t.Error("Render should show 'waiting for telemetry' message")
	}
}

func TestDashboardView_Render_WithLatest(t *testing.T) {
	dv := NewDashboardView()
	dv.Width = 80
	latest := engine.NewTelemetrySnapshot(45.0, 200000, 35)
	output := dv.Render([]engine.TelemetrySnapshot{latest}, &latest)

	if strings.Contains(output, "No data") {
		t.Error("Render should not show 'No data' when telemetry exists")
	}
	if !strings.Contains(output, "CPU") {
		t.Error("Render should contain CPU")
	}
	if !strings.Contains(output, "RAM") {
		t.Error("Render should contain RAM")
	}
	if !strings.Contains(output, "Threads") {
		t.Error("Render should contain Threads")
	}
}

func TestDashboardView_Render_WithSnapshots(t *testing.T) {
	dv := NewDashboardView()
	dv.Width = 80
	snaps := []engine.TelemetrySnapshot{
		engine.NewTelemetrySnapshot(10.0, 100000, 20),
		engine.NewTelemetrySnapshot(20.0, 150000, 25),
		engine.NewTelemetrySnapshot(30.0, 200000, 30),
	}
	latest := snaps[len(snaps)-1]
	output := dv.Render(snaps, &latest)

	if !strings.Contains(output, "CPU Utilization") {
		t.Error("Render should contain 'CPU Utilization'")
	}
	if !strings.Contains(output, "Memory") {
		t.Error("Render should contain 'Memory'")
	}
	if !strings.Contains(output, "Peak CPU") {
		t.Error("Render should contain 'Peak CPU' in stats row")
	}
	if !strings.Contains(output, "Peak RAM") {
		t.Error("Render should contain 'Peak RAM' in stats row")
	}
	if !strings.Contains(output, "Duration") {
		t.Error("Render should contain 'Duration' in stats row")
	}
}

func TestDashboardView_Render_HighCPUStatusColor(t *testing.T) {
	dv := NewDashboardView()
	dv.Width = 80
	latest := engine.NewTelemetrySnapshot(90.0, 100000, 20)
	output := dv.Render([]engine.TelemetrySnapshot{latest}, &latest)

	if !strings.Contains(output, "CPU") {
		t.Error("Render should contain CPU info")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		got := formatBytes(tt.bytes)
		if got != tt.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}
