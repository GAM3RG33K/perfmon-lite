package export

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/GAM3RG33K/perfmon-lite/internal/engine"
)

// Export dispatches to the appropriate format generator and returns the output path.
func Export(snapshots []engine.TelemetrySnapshot, opts Options) (string, error) {
	data := BuildExportData(snapshots, opts)

	switch opts.Format {
	case FormatJSON:
		return ExportJSON(data, opts)
	case FormatMD:
		return ExportMarkdown(data, snapshots, opts)
	case FormatHTML:
		return ExportHTML(data, snapshots, opts)
	case FormatPDF:
		return ExportPDF(data, snapshots, opts)
	default:
		return "", fmt.Errorf("unsupported export format: %s", opts.Format)
	}
}

// ResolveOutputPath returns the full output base path (without extension).
// If opts.OutputPath is empty, it generates a timestamped default path.
func ResolveOutputPath(opts Options, snapshots []engine.TelemetrySnapshot) string {
	if opts.OutputPath != "" {
		return opts.OutputPath
	}
	ts := time.Now().Format("2006-01-02_150405")
	return fmt.Sprintf("perfmon_export_%s", ts)
}

// EnsureOutputDir creates the parent directory for the output file if needed.
func EnsureOutputDir(path string) error {
	dir := dirName(path)
	if dir == "" || dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func dirName(path string) string {
	return filepath.Dir(path)
}
