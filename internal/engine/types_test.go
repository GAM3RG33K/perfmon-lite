package engine

import (
	"testing"
)

func TestComputeMetricsSummary_Empty(t *testing.T) {
	summary := ComputeMetricsSummary(nil)
	if summary.DurationSeconds != 0 {
		t.Fatalf("expected 0 duration, got %d", summary.DurationSeconds)
	}
	if summary.PeakCPUPercent != 0 {
		t.Fatalf("expected 0 peak CPU, got %f", summary.PeakCPUPercent)
	}
	if summary.PeakMemoryKB != 0 {
		t.Fatalf("expected 0 peak memory, got %d", summary.PeakMemoryKB)
	}
	if summary.PeakThreads != 0 {
		t.Fatalf("expected 0 peak threads, got %d", summary.PeakThreads)
	}
	if summary.AverageCPUPercent != 0 {
		t.Fatalf("expected 0 avg CPU, got %f", summary.AverageCPUPercent)
	}
	if summary.AverageMemoryKB != 0 {
		t.Fatalf("expected 0 avg memory, got %d", summary.AverageMemoryKB)
	}
}

func TestComputeMetricsSummary_SingleSnapshot(t *testing.T) {
	snapshots := []TelemetrySnapshot{
		{Timestamp: 1, CPUPercent: 55.0, MemoryKB: 1024, Threads: 20},
	}

	summary := ComputeMetricsSummary(snapshots)
	if summary.DurationSeconds != 1 {
		t.Fatalf("expected duration 1, got %d", summary.DurationSeconds)
	}
	if summary.AverageCPUPercent != 55.0 {
		t.Fatalf("expected avg CPU 55.0, got %f", summary.AverageCPUPercent)
	}
	if summary.PeakCPUPercent != 55.0 {
		t.Fatalf("expected peak CPU 55.0, got %f", summary.PeakCPUPercent)
	}
	if summary.AverageMemoryKB != 1024 {
		t.Fatalf("expected avg memory 1024, got %d", summary.AverageMemoryKB)
	}
	if summary.PeakMemoryKB != 1024 {
		t.Fatalf("expected peak memory 1024, got %d", summary.PeakMemoryKB)
	}
	if summary.PeakThreads != 20 {
		t.Fatalf("expected peak threads 20, got %d", summary.PeakThreads)
	}
}

func TestComputeMetricsSummary_MultipleSnapshots(t *testing.T) {
	snapshots := []TelemetrySnapshot{
		{Timestamp: 1, CPUPercent: 10.0, MemoryKB: 100, Threads: 5},
		{Timestamp: 2, CPUPercent: 20.0, MemoryKB: 200, Threads: 10},
		{Timestamp: 3, CPUPercent: 30.0, MemoryKB: 150, Threads: 8},
		{Timestamp: 4, CPUPercent: 40.0, MemoryKB: 300, Threads: 15},
	}

	summary := ComputeMetricsSummary(snapshots)
	if summary.DurationSeconds != 4 {
		t.Fatalf("expected duration 4, got %d", summary.DurationSeconds)
	}

	// Averages: CPU = (10+20+30+40)/4 = 25, Memory = (100+200+150+300)/4 = 187
	if summary.AverageCPUPercent != 25.0 {
		t.Fatalf("expected avg CPU 25.0, got %f", summary.AverageCPUPercent)
	}
	if summary.AverageMemoryKB != 187 {
		t.Fatalf("expected avg memory 187, got %d", summary.AverageMemoryKB)
	}

	// Peaks: CPU = 40, Memory = 300, Threads = 15
	if summary.PeakCPUPercent != 40.0 {
		t.Fatalf("expected peak CPU 40.0, got %f", summary.PeakCPUPercent)
	}
	if summary.PeakMemoryKB != 300 {
		t.Fatalf("expected peak memory 300, got %d", summary.PeakMemoryKB)
	}
	if summary.PeakThreads != 15 {
		t.Fatalf("expected peak threads 15, got %d", summary.PeakThreads)
	}
}

func TestComputeMetricsSummary_UniformData(t *testing.T) {
	snapshots := make([]TelemetrySnapshot, 10)
	for i := range snapshots {
		snapshots[i] = TelemetrySnapshot{
			Timestamp:  int64(i),
			CPUPercent: 50.0,
			MemoryKB:   1000,
			Threads:    10,
		}
	}

	summary := ComputeMetricsSummary(snapshots)
	if summary.AverageCPUPercent != 50.0 {
		t.Fatalf("expected avg CPU 50.0, got %f", summary.AverageCPUPercent)
	}
	if summary.PeakCPUPercent != 50.0 {
		t.Fatalf("expected peak CPU 50.0, got %f", summary.PeakCPUPercent)
	}
	if summary.AverageMemoryKB != 1000 {
		t.Fatalf("expected avg memory 1000, got %d", summary.AverageMemoryKB)
	}
	if summary.PeakMemoryKB != 1000 {
		t.Fatalf("expected peak memory 1000, got %d", summary.PeakMemoryKB)
	}
	if summary.PeakThreads != 10 {
		t.Fatalf("expected peak threads 10, got %d", summary.PeakThreads)
	}
}

func TestComputeMetricsSummary_FloatingPoint(t *testing.T) {
	snapshots := []TelemetrySnapshot{
		{CPUPercent: 1.5, MemoryKB: 100, Threads: 1},
		{CPUPercent: 2.5, MemoryKB: 200, Threads: 2},
	}

	summary := ComputeMetricsSummary(snapshots)
	if summary.AverageCPUPercent != 2.0 {
		t.Fatalf("expected avg CPU 2.0, got %f", summary.AverageCPUPercent)
	}
}

func TestComputeMetricsSummary_PeakMemoryZero(t *testing.T) {
	snapshots := []TelemetrySnapshot{
		{CPUPercent: 10, MemoryKB: 0, Threads: 1},
	}

	summary := ComputeMetricsSummary(snapshots)
	if summary.PeakMemoryKB != 0 {
		t.Fatalf("expected peak memory 0, got %d", summary.PeakMemoryKB)
	}
	if summary.AverageMemoryKB != 0 {
		t.Fatalf("expected avg memory 0, got %d", summary.AverageMemoryKB)
	}
}

func TestNewTelemetrySnapshot(t *testing.T) {
	s := NewTelemetrySnapshot(45.5, 2048, 12)
	if s.CPUPercent != 45.5 {
		t.Fatalf("expected CPU 45.5, got %f", s.CPUPercent)
	}
	if s.MemoryKB != 2048 {
		t.Fatalf("expected Memory 2048, got %d", s.MemoryKB)
	}
	if s.Threads != 12 {
		t.Fatalf("expected Threads 12, got %d", s.Threads)
	}
	if s.Timestamp == 0 {
		t.Fatal("expected non-zero timestamp")
	}
}
