package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/w1n/perfmon/internal/engine"
)

// ─── Helpers ────────────────────────────────────────────────────────────────

func testSnapshots() []engine.TelemetrySnapshot {
	return []engine.TelemetrySnapshot{
		{Timestamp: 1700000000, CPUPercent: 12.5, MemoryKB: 45000, Threads: 15},
		{Timestamp: 1700000001, CPUPercent: 45.2, MemoryKB: 52000, Threads: 16},
		{Timestamp: 1700000002, CPUPercent: 78.9, MemoryKB: 61000, Threads: 18},
		{Timestamp: 1700000003, CPUPercent: 30.0, MemoryKB: 48000, Threads: 15},
		{Timestamp: 1700000004, CPUPercent: 95.1, MemoryKB: 75000, Threads: 22},
	}
}

// tempOpts creates export Options backed by a test temp directory.
func tempOpts(t *testing.T, format Format) Options {
	t.Helper()
	return Options{
		Format:     format,
		OutputPath: filepath.Join(t.TempDir(), "test_export"),
		Version:    "1.0.0",
		Platform:   engine.PlatformAndroid,
		DeviceName: "Pixel_8",
		AppName:    "com.example.app",
		BuildType:  engine.BuildDebug,
	}
}

// ─── BuildExportData Tests ──────────────────────────────────────────────────

func TestBuildExportData_PopulatesFields(t *testing.T) {
	snaps := testSnapshots()
	data := BuildExportData(snaps, tempOpts(t, FormatJSON))

	if data.Schema == "" {
		t.Error("expected non-empty Schema")
	}
	if !strings.HasPrefix(data.Schema, "https://") {
		t.Errorf("expected Schema to start with https://, got %s", data.Schema)
	}

	// Metadata
	if data.Metadata.GeneratedAt == "" {
		t.Error("expected GeneratedAt to be set")
	}
	if data.Metadata.PerfmonVersion != "1.0.0" {
		t.Errorf("expected PerfmonVersion 1.0.0, got %s", data.Metadata.PerfmonVersion)
	}
	if data.Metadata.TargetPlatform != "android" {
		t.Errorf("expected TargetPlatform android, got %s", data.Metadata.TargetPlatform)
	}
	if data.Metadata.DeviceName != "Pixel_8" {
		t.Errorf("expected DeviceName Pixel_8, got %s", data.Metadata.DeviceName)
	}
	if data.Metadata.AppPackage != "com.example.app" {
		t.Errorf("expected AppPackage com.example.app, got %s", data.Metadata.AppPackage)
	}
	if data.Metadata.BuildType != engine.BuildDebug {
		t.Errorf("expected BuildType debug, got %s", data.Metadata.BuildType)
	}

	// Summary
	if data.Summary.DurationSeconds != 5 {
		t.Errorf("expected DurationSeconds 5, got %d", data.Summary.DurationSeconds)
	}
	if data.Summary.PeakCPUPercent != 95.1 {
		t.Errorf("expected PeakCPUPercent 95.1, got %.1f", data.Summary.PeakCPUPercent)
	}
	av := data.Summary.AverageCPUPercent
	if av < 52.33 || av > 52.35 {
		t.Errorf("expected AverageCPUPercent ~52.34, got %.4f", av)
	}
	if data.Summary.PeakMemoryKB != 75000 {
		t.Errorf("expected PeakMemoryKB 75000, got %d", data.Summary.PeakMemoryKB)
	}
	if data.Summary.AverageMemoryKB != 56200 {
		t.Errorf("expected AverageMemoryKB 56200, got %d", data.Summary.AverageMemoryKB)
	}
	if data.Summary.PeakThreads != 22 {
		t.Errorf("expected PeakThreads 22, got %d", data.Summary.PeakThreads)
	}

	// Telemetry
	if len(data.Telemetry) != 5 {
		t.Errorf("expected 5 telemetry samples, got %d", len(data.Telemetry))
	}
}

func TestBuildExportData_EmptySnapshots(t *testing.T) {
	data := BuildExportData(nil, tempOpts(t, FormatJSON))

	if data.Metadata.GeneratedAt == "" {
		t.Error("expected GeneratedAt to be set even with empty data")
	}
	if data.Summary.DurationSeconds != 0 {
		t.Errorf("expected DurationSeconds 0, got %d", data.Summary.DurationSeconds)
	}
	if data.Telemetry == nil {
		t.Error("expected non-nil Telemetry slice")
	}
	if len(data.Telemetry) != 0 {
		t.Errorf("expected 0 telemetry samples, got %d", len(data.Telemetry))
	}
}

func TestBuildExportData_SingleSnapshot(t *testing.T) {
	snaps := []engine.TelemetrySnapshot{
		{Timestamp: 1700000000, CPUPercent: 50.0, MemoryKB: 100000, Threads: 10},
	}
	data := BuildExportData(snaps, tempOpts(t, FormatJSON))

	if data.Summary.DurationSeconds != 1 {
		t.Errorf("expected DurationSeconds 1, got %d", data.Summary.DurationSeconds)
	}
	if data.Summary.PeakCPUPercent != 50.0 {
		t.Errorf("expected PeakCPUPercent 50.0, got %.1f", data.Summary.PeakCPUPercent)
	}
	if data.Summary.PeakMemoryKB != 100000 {
		t.Errorf("expected PeakMemoryKB 100000, got %d", data.Summary.PeakMemoryKB)
	}
	if data.Summary.PeakThreads != 10 {
		t.Errorf("expected PeakThreads 10, got %d", data.Summary.PeakThreads)
	}
}

func TestBuildExportData_MetadataFromOpts(t *testing.T) {
	opts := Options{
		Version:    "2.0.0-rc1",
		Platform:   engine.PlatformIOS,
		DeviceName: "iPhone 17 Pro",
		AppName:    "com.example.iphone",
		BuildType:  engine.BuildRelease,
	}
	snaps := testSnapshots()
	data := BuildExportData(snaps, opts)

	if data.Metadata.PerfmonVersion != "2.0.0-rc1" {
		t.Errorf("expected PerfmonVersion 2.0.0-rc1, got %s", data.Metadata.PerfmonVersion)
	}
	if data.Metadata.TargetPlatform != "ios" {
		t.Errorf("expected TargetPlatform ios, got %s", data.Metadata.TargetPlatform)
	}
	if data.Metadata.DeviceName != "iPhone 17 Pro" {
		t.Errorf("expected DeviceName 'iPhone 17 Pro', got %s", data.Metadata.DeviceName)
	}
	if data.Metadata.BuildType != engine.BuildRelease {
		t.Errorf("expected BuildType release, got %s", data.Metadata.BuildType)
	}
}

func TestBuildExportData_DefaultBuildType(t *testing.T) {
	opts := Options{
		Platform: engine.PlatformMock,
		BuildType: engine.BuildUnknown,
	}
	data := BuildExportData(nil, opts)

	if data.Metadata.BuildType != engine.BuildUnknown {
		t.Errorf("expected BuildType unknown, got %s", data.Metadata.BuildType)
	}
}

// ─── ResolveOutputPath Tests ────────────────────────────────────────────────

func TestResolveOutputPath_DefaultPath(t *testing.T) {
	path := ResolveOutputPath(Options{}, nil)
	if path != "perfmon_export_0" {
		t.Errorf("expected perfmon_export_0, got %s", path)
	}

	path = ResolveOutputPath(Options{}, testSnapshots())
	if path != "perfmon_export_5" {
		t.Errorf("expected perfmon_export_5, got %s", path)
	}
}

func TestResolveOutputPath_CustomPath(t *testing.T) {
	opts := Options{OutputPath: "/tmp/my_report"}
	path := ResolveOutputPath(opts, testSnapshots())
	if path != "/tmp/my_report" {
		t.Errorf("expected /tmp/my_report, got %s", path)
	}
}

func TestResolveOutputPath_EmptySnapshotsDefault(t *testing.T) {
	path := ResolveOutputPath(Options{OutputPath: ""}, []engine.TelemetrySnapshot{})
	if path != "perfmon_export_0" {
		t.Errorf("expected perfmon_export_0, got %s", path)
	}
}

// ─── outputFilename Tests ───────────────────────────────────────────────────

func TestOutputFilename_AppendsExtension(t *testing.T) {
	result := outputFilename("/path/to/report", FormatJSON)
	if result != "/path/to/report.json" {
		t.Errorf("expected /path/to/report.json, got %s", result)
	}

	result = outputFilename("report", FormatMD)
	if result != "report.md" {
		t.Errorf("expected report.md, got %s", result)
	}

	result = outputFilename("./output/report", FormatHTML)
	if result != "./output/report.html" {
		t.Errorf("expected ./output/report.html, got %s", result)
	}

	result = outputFilename("report", Format("pdf"))
	if result != "report.pdf" {
		t.Errorf("expected report.pdf, got %s", result)
	}
}

// ─── EnsureOutputDir Tests ──────────────────────────────────────────────────

func TestEnsureOutputDir_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "report.json")

	if err := EnsureOutputDir(path); err != nil {
		t.Fatalf("EnsureOutputDir failed: %v", err)
	}

	// Verify the directory was created
	if _, err := os.Stat(filepath.Join(dir, "subdir")); os.IsNotExist(err) {
		t.Fatal("expected subdir to exist after EnsureOutputDir")
	}
}

func TestEnsureOutputDir_NoDirForFlatPath(t *testing.T) {
	if err := EnsureOutputDir("report.json"); err != nil {
		t.Fatalf("EnsureOutputDir failed for flat path: %v", err)
	}
}

func TestEnsureOutputDir_NestedDirectories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "c", "report.json")

	if err := EnsureOutputDir(path); err != nil {
		t.Fatalf("EnsureOutputDir failed for nested path: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "a", "b", "c")); os.IsNotExist(err) {
		t.Fatal("expected nested dirs to exist")
	}
}

func TestEnsureOutputDir_ExistingDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.json")

	// First call creates
	if err := EnsureOutputDir(path); err != nil {
		t.Fatalf("first EnsureOutputDir failed: %v", err)
	}

	// Second call should be idempotent
	if err := EnsureOutputDir(path); err != nil {
		t.Fatalf("second EnsureOutputDir failed: %v", err)
	}
}

// ─── Export Dispatcher Tests ────────────────────────────────────────────────

func TestExport_UnknownFormat(t *testing.T) {
	opts := tempOpts(t, Format("unknown"))
	_, err := Export(testSnapshots(), opts)
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
	if !strings.Contains(err.Error(), "unsupported export format") {
		t.Errorf("expected unsupported format error, got: %v", err)
	}
}

func TestExport_JSONFormat(t *testing.T) {
	opts := tempOpts(t, FormatJSON)
	path, err := Export(testSnapshots(), opts)
	if err != nil {
		t.Fatalf("Export JSON failed: %v", err)
	}
	if !strings.HasSuffix(path, ".json") {
		t.Errorf("expected .json suffix, got %s", path)
	}

	// Verify file exists and contains valid JSON
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty JSON output")
	}
	if !strings.Contains(string(data), "\"metadata\"") {
		t.Error("expected JSON to contain metadata field")
	}
	if !strings.Contains(string(data), "\"telemetry\"") {
		t.Error("expected JSON to contain telemetry field")
	}
	if !strings.Contains(string(data), "\"metrics_summary\"") {
		t.Error("expected JSON to contain metrics_summary field")
	}
}

func TestExport_JSONEmptyData(t *testing.T) {
	opts := tempOpts(t, FormatJSON)
	path, err := Export(nil, opts)
	if err != nil {
		t.Fatalf("Export JSON with empty data failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty JSON output even with empty data")
	}
}

// ─── Markdown Export Tests ──────────────────────────────────────────────────

func TestExport_MarkdownFormat(t *testing.T) {
	opts := tempOpts(t, FormatMD)
	path, err := Export(testSnapshots(), opts)
	if err != nil {
		t.Fatalf("Export Markdown failed: %v", err)
	}
	if !strings.HasSuffix(path, ".md") {
		t.Errorf("expected .md suffix, got %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	content := string(data)

	// Check key sections
	if !strings.Contains(content, "perfmon Telemetry Report") {
		t.Error("expected report title")
	}
	if !strings.Contains(content, "Metrics Summary") {
		t.Error("expected Metrics Summary section")
	}
	if !strings.Contains(content, "CPU Utilization") {
		t.Error("expected CPU Utilization section")
	}
	if !strings.Contains(content, "Memory Footprint") {
		t.Error("expected Memory Footprint section")
	}
	if !strings.Contains(content, "Telemetry Data") {
		t.Error("expected Telemetry Data section")
	}
	if !strings.Contains(content, "com.example.app") {
		t.Error("expected app package name in output")
	}
	if !strings.Contains(content, "Pixel_8") {
		t.Error("expected device name in output")
	}
	if !strings.Contains(content, "debug") {
		t.Error("expected build type in output")
	}
	if !strings.Contains(content, "1.0.0") {
		t.Error("expected version in output")
	}
	// Check some data values
	if !strings.Contains(content, "95.1%") {
		t.Error("expected peak CPU 95.1% in output")
	}
	if !strings.Contains(content, "52.3") { // average CPU
		t.Error("expected average CPU in output")
	}
}

func TestExport_MarkdownEmptyData(t *testing.T) {
	opts := tempOpts(t, FormatMD)
	path, err := Export(nil, opts)
	if err != nil {
		t.Fatalf("Export Markdown with empty data failed: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty Markdown output")
	}
}

func TestExport_MarkdownSingleSnapshot(t *testing.T) {
	snaps := []engine.TelemetrySnapshot{
		{Timestamp: 1700000000, CPUPercent: 25.0, MemoryKB: 30000, Threads: 8},
	}
	opts := tempOpts(t, FormatMD)
	path, err := Export(snaps, opts)
	if err != nil {
		t.Fatalf("Export Markdown with single snapshot failed: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty Markdown output")
	}
	// Should have the table row with our data
	content := string(data)
	if !strings.Contains(content, "25.0") {
		t.Error("expected CPU value in output")
	}
}

// ─── HTML Export Tests ──────────────────────────────────────────────────────

func TestExport_HTMLFormat(t *testing.T) {
	opts := tempOpts(t, FormatHTML)
	path, err := Export(testSnapshots(), opts)
	if err != nil {
		t.Fatalf("Export HTML failed: %v", err)
	}
	if !strings.HasSuffix(path, ".html") {
		t.Errorf("expected .html suffix, got %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	content := string(data)

	// Check document structure
	if !strings.Contains(content, "<!DOCTYPE html>") {
		t.Error("expected DOCTYPE declaration")
	}
	if !strings.Contains(content, "<html") {
		t.Error("expected html tag")
	}
	if !strings.Contains(content, "</html>") {
		t.Error("expected closing html tag")
	}
	if !strings.Contains(content, "<style>") {
		t.Error("expected style tag")
	}

	// Check content sections
	if !strings.Contains(content, "perfmon Telemetry Report") {
		t.Error("expected report title")
	}
	if !strings.Contains(content, "Metrics Summary") {
		t.Error("expected Metrics Summary section")
	}
	if !strings.Contains(content, "CPU Utilization") {
		t.Error("expected CPU section")
	}
	if !strings.Contains(content, "Memory Footprint") {
		t.Error("expected Memory section")
	}
	if !strings.Contains(content, "Thread Count") {
		t.Error("expected Thread Count section")
	}
	if !strings.Contains(content, "Raw Telemetry") {
		t.Error("expected Raw Telemetry section")
	}

	// Check SVG charts present
	if !strings.Contains(content, "<svg") {
		t.Error("expected SVG elements")
	}
	if !strings.Contains(content, "<polyline") {
		t.Error("expected polyline elements")
	}
	if !strings.Contains(content, "<polygon") {
		t.Error("expected polygon fill elements")
	}

	// Check data values
	if !strings.Contains(content, "Pixel_8") {
		t.Error("expected device name in output")
	}
	if !strings.Contains(content, "com.example.app") {
		t.Error("expected app name in output")
	}
}

func TestExport_HTMLEmptyData(t *testing.T) {
	opts := tempOpts(t, FormatHTML)
	path, err := Export(nil, opts)
	if err != nil {
		t.Fatalf("Export HTML with empty data failed: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty HTML output")
	}
}

func TestExport_HTMLSingleSnapshot(t *testing.T) {
	snaps := []engine.TelemetrySnapshot{
		{Timestamp: 1700000000, CPUPercent: 50.0, MemoryKB: 100000, Threads: 10},
	}
	opts := tempOpts(t, FormatHTML)
	path, err := Export(snaps, opts)
	if err != nil {
		t.Fatalf("Export HTML with single snapshot failed: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	content := string(data)
	// Single snapshot should still produce valid SVG (with smaller chartWidth)
	if !strings.Contains(content, "<svg") {
		t.Error("expected SVG even with single snapshot")
	}
}

// ─── PDF Export Tests ───────────────────────────────────────────────────────

func TestExport_PDFFormat(t *testing.T) {
	opts := tempOpts(t, FormatPDF)
	path, err := Export(testSnapshots(), opts)
	if err != nil {
		t.Fatalf("Export PDF failed: %v", err)
	}
	if !strings.HasSuffix(path, ".pdf") {
		t.Errorf("expected .pdf suffix, got %s", path)
	}

	stat, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat output: %v", err)
	}
	if stat.Size() == 0 {
		t.Fatal("expected non-empty PDF output")
	}

	// PDF magic bytes: %PDF-
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	if len(data) < 5 || string(data[:5]) != "%PDF-" {
		t.Errorf("expected PDF magic bytes '%%PDF-', got %q", string(data[:min(len(data), 5)]))
	}
}

func TestExport_PDFEmptyData(t *testing.T) {
	opts := tempOpts(t, FormatPDF)
	path, err := Export(nil, opts)
	if err != nil {
		t.Fatalf("Export PDF with empty data failed: %v", err)
	}
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat output: %v", err)
	}
	if stat.Size() == 0 {
		t.Fatal("expected non-empty PDF output")
	}
}

func TestExport_PDFSingleSnapshot(t *testing.T) {
	snaps := []engine.TelemetrySnapshot{
		{Timestamp: 1700000000, CPUPercent: 25.0, MemoryKB: 30000, Threads: 8},
	}
	opts := tempOpts(t, FormatPDF)
	path, err := Export(snaps, opts)
	if err != nil {
		t.Fatalf("Export PDF with single snapshot failed: %v", err)
	}
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat output: %v", err)
	}
	if stat.Size() == 0 {
		t.Fatal("expected non-empty PDF output")
	}

	// Verify it's a valid PDF
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	if string(data[:5]) != "%PDF-" {
		t.Errorf("expected PDF magic bytes, got %q", string(data[:5]))
	}
}

// ─── buildSVGLinePoints Tests ───────────────────────────────────────────────

func TestBuildSVGLinePoints_Empty(t *testing.T) {
	result := buildSVGLinePoints(nil, 800, 200, 0, 100, func(s engine.TelemetrySnapshot) float64 {
		return s.CPUPercent
	})
	if result != "" {
		t.Errorf("expected empty string for no snapshots, got %q", result)
	}
}

func TestBuildSVGLinePoints_SinglePoint(t *testing.T) {
	snaps := []engine.TelemetrySnapshot{
		{CPUPercent: 50.0},
	}
	result := buildSVGLinePoints(snaps, 800, 200, 0, 100, func(s engine.TelemetrySnapshot) float64 {
		return s.CPUPercent
	})
	// Single point: x=0, y=100 (50% of 200 from bottom)
	expected := "0.0,100.0"
	if result != expected {
		t.Errorf("expected %q for single point, got %q", expected, result)
	}
}

func TestBuildSVGLinePoints_MultiplePoints(t *testing.T) {
	snaps := []engine.TelemetrySnapshot{
		{CPUPercent: 0.0},
		{CPUPercent: 50.0},
		{CPUPercent: 100.0},
	}
	result := buildSVGLinePoints(snaps, 400, 200, 0, 100, func(s engine.TelemetrySnapshot) float64 {
		return s.CPUPercent
	})

	// width=400, n=3, step=200
	// Point 0: x=0.0,   val=0   => y=200-0   =200.0
	// Point 1: x=200.0, val=50  => y=200-100 =100.0
	// Point 2: x=400.0, val=100 => y=200-200 =0.0
	expected := "0.0,200.0 200.0,100.0 400.0,0.0"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestBuildSVGLinePoints_Clamping(t *testing.T) {
	snaps := []engine.TelemetrySnapshot{
		{CPUPercent: -10.0},
		{CPUPercent: 150.0},
	}
	result := buildSVGLinePoints(snaps, 400, 200, 0, 100, func(s engine.TelemetrySnapshot) float64 {
		return s.CPUPercent
	})

	// Width=400, n=2, step=400
	// Point 0: x=0.0,   val=-10  clamped to 0  => y=200
	// Point 1: x=400.0, val=150  clamped to 100 => y=0
	expected := "0.0,200.0 400.0,0.0"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestBuildSVGLinePoints_ZeroRange(t *testing.T) {
	snaps := []engine.TelemetrySnapshot{
		{CPUPercent: 50.0},
		{CPUPercent: 50.0},
	}
	result := buildSVGLinePoints(snaps, 400, 200, 50, 50, func(s engine.TelemetrySnapshot) float64 {
		return s.CPUPercent
	})
	// rangeY = 0, defaults to 1
	// Point 0: x=0.0,   val=50 => (50-50)/1=0 => y=200-0=200
	// Point 1: x=400.0, val=50 => (50-50)/1=0 => y=200-0=200
	expected := "0.0,200.0 400.0,200.0"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// ─── buildASCIISparkline Tests ──────────────────────────────────────────────

func TestBuildASCIISparkline_Empty(t *testing.T) {
	result := buildASCIISparkline(nil, func(s engine.TelemetrySnapshot) float64 {
		return s.CPUPercent
	}, 0, 100)
	if result != "" {
		t.Errorf("expected empty string for nil snapshots, got %q", result)
	}
}

func TestBuildASCIISparkline_SingleSnapshot(t *testing.T) {
	result := buildASCIISparkline([]engine.TelemetrySnapshot{{CPUPercent: 50.0}},
		func(s engine.TelemetrySnapshot) float64 { return s.CPUPercent }, 0, 100)
	if result != "" {
		t.Errorf("expected empty string for single snapshot, got %q", result)
	}
}

func TestBuildASCIISparkline_Basic(t *testing.T) {
	snaps := []engine.TelemetrySnapshot{
		{CPUPercent: 0.0},
		{CPUPercent: 50.0},
		{CPUPercent: 100.0},
	}
	result := buildASCIISparkline(snaps,
		func(s engine.TelemetrySnapshot) float64 { return s.CPUPercent }, 0, 100)

	lines := strings.Split(strings.TrimRight(result, "\n"), "\n")
	if len(lines) != 6 {
		t.Errorf("expected 6 rows, got %d", len(lines))
	}

	rowLen := len([]rune(lines[0]))
	for i, line := range lines {
		if len([]rune(line)) != rowLen {
			t.Errorf("row %d has %d runes, expected %d", i, len([]rune(line)), rowLen)
		}
	}

	// Should contain fill characters
	if !strings.ContainsAny(lines[5], "▓█▄") {
		t.Error("expected line chart to have fill characters at bottom row for 100% value")
	}
}

func TestBuildASCIISparkline_ZeroRange(t *testing.T) {
	snaps := []engine.TelemetrySnapshot{
		{CPUPercent: 75.0},
		{CPUPercent: 75.0},
		{CPUPercent: 75.0},
	}
	result := buildASCIISparkline(snaps,
		func(s engine.TelemetrySnapshot) float64 { return s.CPUPercent }, 75, 75)

	lines := strings.Split(strings.TrimRight(result, "\n"), "\n")
	if len(lines) != 6 {
		t.Errorf("expected 6 rows, got %d", len(lines))
	}
	// With zero range, normalized value is 0.0, so bottom row should have █ (0% → row 5)
	_ = lines
}

// ─── formatPDFBytes Tests ───────────────────────────────────────────────────

func TestFormatPDFBytes(t *testing.T) {
	tests := []struct {
		kb   int64
		want string
	}{
		{0, "0 KB"},
		{512, "512 KB"},
		{1023, "1023 KB"},
		{1024, "1.0 MB"},
		{1536, "1.5 MB"},
		{10240, "10.0 MB"},
		{1048576, "1024.0 MB"},
	}

	for _, tt := range tests {
		got := formatPDFBytes(tt.kb)
		if got != tt.want {
			t.Errorf("formatPDFBytes(%d) = %q, want %q", tt.kb, got, tt.want)
		}
	}
}


