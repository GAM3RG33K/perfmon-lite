package chart

import (
	"strings"
	"testing"

	"github.com/GAM3RG33K/perfmon-lite/internal/engine"
)

func TestRenderCPUChart_Empty(t *testing.T) {
	if got := RenderCPUChart(nil, 40); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestRenderCPUChart_Basic(t *testing.T) {
	snaps := []engine.TelemetrySnapshot{
		{CPUPercent: 0},
		{CPUPercent: 50},
		{CPUPercent: 100},
	}
	got := RenderCPUChart(snaps, 40)
	if got == "" {
		t.Fatal("expected chart output")
	}
	if !strings.Contains(got, "CPU Utilization") {
		t.Error("missing title")
	}
	if !strings.ContainsAny(got, "█▄▗▟") {
		t.Error("expected block chart characters")
	}
}

func TestBtopGraphRows_Height(t *testing.T) {
	data := []int{0, 25, 50, 75, 100}
	rows := BtopGraphRows(data, 20, DefaultHeight, BlockSymbolsUp)
	if len(rows) != DefaultHeight {
		t.Fatalf("expected %d rows, got %d", DefaultHeight, len(rows))
	}
}
