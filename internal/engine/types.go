package engine

import "time"

// Platform represents the target mobile platform.
type Platform string

const (
	PlatformAndroid Platform = "android"
	PlatformIOS     Platform = "ios"
	PlatformMock    Platform = "mock"
)

// BuildType represents whether an app is built for debug or release.
type BuildType string

const (
	BuildDebug   BuildType = "debug"
	BuildRelease BuildType = "release"
	BuildUnknown BuildType = "unknown"
)

// Device represents a connected mobile device or emulator/simulator.
type Device struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Platform Platform `json:"platform"`
	IsPhysical bool   `json:"is_physical"`
	IsBooted  bool   `json:"is_booted"`
}

// AppProcess represents a running application process on a device.
type AppProcess struct {
	PID       int32     `json:"pid"`
	Name      string    `json:"name"`
	PackageName string  `json:"package_name"`
	BuildType BuildType `json:"build_type"`
}

// TelemetrySnapshot contains a single sample of performance metrics.
type TelemetrySnapshot struct {
	Timestamp  int64   `json:"timestamp"`
	CPUPercent float64 `json:"cpu"`
	MemoryKB   int64   `json:"memory_kb"`
	Threads    int32   `json:"threads"`
}

// NewTelemetrySnapshot creates a new telemetry snapshot with the current time.
func NewTelemetrySnapshot(cpu float64, memoryKB int64, threads int32) TelemetrySnapshot {
	return TelemetrySnapshot{
		Timestamp:  time.Now().Unix(),
		CPUPercent: cpu,
		MemoryKB:   memoryKB,
		Threads:    threads,
	}
}

// MetricsSummary aggregates telemetry data for export.
type MetricsSummary struct {
	DurationSeconds  int     `json:"duration_seconds"`
	PeakMemoryKB     int64   `json:"peak_memory_kb"`
	AverageCPUPercent float64 `json:"average_cpu_percentage"`
	PeakCPUPercent   float64 `json:"peak_cpu_percentage"`
	AverageMemoryKB  int64   `json:"average_memory_kb"`
	PeakThreads      int32   `json:"peak_threads"`
}

// ComputeMetricsSummary computes summary statistics from a slice of snapshots.
func ComputeMetricsSummary(snapshots []TelemetrySnapshot) MetricsSummary {
	if len(snapshots) == 0 {
		return MetricsSummary{}
	}

	var totalCPU float64
	var totalMemory int64
	var peakCPU float64
	var peakMemory int64
	var peakThreads int32

	for _, s := range snapshots {
		totalCPU += s.CPUPercent
		totalMemory += s.MemoryKB

		if s.CPUPercent > peakCPU {
			peakCPU = s.CPUPercent
		}
		if s.MemoryKB > peakMemory {
			peakMemory = s.MemoryKB
		}
		if s.Threads > peakThreads {
			peakThreads = s.Threads
		}
	}

	n := len(snapshots)
	return MetricsSummary{
		DurationSeconds:   n,
		PeakMemoryKB:      peakMemory,
		AverageCPUPercent: totalCPU / float64(n),
		PeakCPUPercent:    peakCPU,
		AverageMemoryKB:   totalMemory / int64(n),
		PeakThreads:       peakThreads,
	}
}
