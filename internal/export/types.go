package export

import (
	"fmt"
	"time"

	"github.com/w1n/perfmon/internal/engine"
)

// Format represents an export output format.
type Format string

const (
	FormatJSON Format = "json"
	FormatMD   Format = "md"
	FormatHTML Format = "html"
	FormatPDF  Format = "pdf"
)

// Options controls the export behavior.
type Options struct {
	Format      Format
	OutputPath  string // Full output path (without extension if empty)
	Verbose     bool
	Version     string
	Platform    engine.Platform
	DeviceName  string
	AppName     string
	BuildType   engine.BuildType
	Logs        []string // captured app logs (logcat/log stream)
}

// ExportData bundles all data needed to render an export file.
type ExportData struct {
	Schema    string                    `json:"$schema"`
	Metadata  ExportMetadata            `json:"metadata"`
	Summary   engine.MetricsSummary     `json:"metrics_summary"`
	Telemetry []engine.TelemetrySnapshot `json:"telemetry"`
	Logs      []string                  `json:"logs,omitempty"` // captured app logs
}

// ExportMetadata describes the session and target.
type ExportMetadata struct {
	GeneratedAt    string           `json:"generated_at"`
	PerfmonVersion string           `json:"perfmon_version"`
	TargetPlatform string           `json:"target_platform"`
	DeviceName     string           `json:"device_name"`
	AppPackage     string           `json:"app_package"`
	BuildType      engine.BuildType `json:"build_type"`
}

// BuildExportData assembles an ExportData from engine snapshots and metadata.
func BuildExportData(snapshots []engine.TelemetrySnapshot, opts Options) ExportData {
	telemetry := snapshots
	if telemetry == nil {
		telemetry = []engine.TelemetrySnapshot{}
	}
	logs := opts.Logs
	if logs == nil {
		logs = []string{}
	}
	return ExportData{
		Schema: fmt.Sprintf("https://perfmon.qzz.io/schemas/export-v1.json"),
		Metadata: ExportMetadata{
			GeneratedAt:    time.Now().UTC().Format(time.RFC3339),
			PerfmonVersion: opts.Version,
			TargetPlatform: string(opts.Platform),
			DeviceName:     opts.DeviceName,
			AppPackage:     opts.AppName,
			BuildType:      opts.BuildType,
		},
		Summary:   engine.ComputeMetricsSummary(snapshots),
		Telemetry: telemetry,
		Logs:      logs,
	}
}

// outputFilename returns the full output filename including extension.
func outputFilename(basePath string, format Format) string {
	return fmt.Sprintf("%s.%s", basePath, string(format))
}
