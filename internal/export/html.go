package export

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/w1n/perfmon/internal/chart"
	"github.com/w1n/perfmon/internal/engine"
)

//go:embed templates/style.css
var cssContent string

//go:embed templates/chart.js
var chartJSContent string

//go:embed templates/logo.svg
var logoSVG string

// htmlTemplate is the embedded HTML report template.
const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>perfmon Telemetry Report — {{ .Metadata.DeviceName }}</title>
<style>
{{ .CSS }}
</style>
</head>
<body>
<div class="container">
  <!-- Header -->
  <div class="header">
    <div class="header-logo">{{ .Logo }}</div>
    <div class="header-text">
      <div class="header-title">perfmon Telemetry Report</div>
      <div class="header-meta">{{ .Metadata.GeneratedAt }}</div>
    </div>
  </div>

  <!-- Metadata Cards -->
  <div class="card-grid">
    <div class="card">
      <div class="card-label">Target Device</div>
      <div class="card-value">{{ .Metadata.DeviceName }}</div>
    </div>
    <div class="card">
      <div class="card-label">Platform</div>
      <div class="card-value platform-{{ .Metadata.TargetPlatform }}">{{ .Metadata.TargetPlatform }}</div>
    </div>
    <div class="card">
      <div class="card-label">Application</div>
      <div class="card-value">{{ .Metadata.AppPackage }}</div>
    </div>
    <div class="card">
      <div class="card-label">Build Type</div>
      <div class="card-value badge-{{ .Metadata.BuildType }}">{{ .Metadata.BuildType }}</div>
    </div>
    <div class="card">
      <div class="card-label">Duration</div>
      <div class="card-value">{{ .Summary.DurationSeconds }} samples</div>
    </div>
    <div class="card">
      <div class="card-label">Tool Version</div>
      <div class="card-value">{{ .Metadata.PerfmonVersion }}</div>
    </div>
  </div>

  <!-- Summary Stats -->
  <div class="section-title">Metrics Summary</div>
  <div class="card-grid">
    <div class="stat-card">
      <div class="stat-value cpu-color">{{ printf "%.1f" .Summary.PeakCPUPercent }}%</div>
      <div class="stat-label">Peak CPU</div>
    </div>
    <div class="stat-card">
      <div class="stat-value cpu-color">{{ printf "%.1f" .Summary.AverageCPUPercent }}%</div>
      <div class="stat-label">Average CPU</div>
    </div>
    <div class="stat-card">
      <div class="stat-value mem-color">{{ formatBytesHTML .Summary.PeakMemoryKB }}</div>
      <div class="stat-label">Peak Memory</div>
    </div>
    <div class="stat-card">
      <div class="stat-value mem-color">{{ formatBytesHTML .Summary.AverageMemoryKB }}</div>
      <div class="stat-label">Average Memory</div>
    </div>
    <div class="stat-card">
      <div class="stat-value thr-color">{{ .Summary.PeakThreads }}</div>
      <div class="stat-label">Peak Threads</div>
    </div>
  </div>

  <!-- CPU Chart -->
  <div class="section-title">CPU Utilization (%)</div>
  <div class="chart-container">
    <svg viewBox="0 0 {{ .ChartWidth }} 200" preserveAspectRatio="none" class="chart-svg" role="img" aria-label="CPU utilization over time">
      <!-- Grid lines -->
      <line x1="0" y1="0" x2="{{ .ChartWidth }}" y2="0" stroke="#333" stroke-width="1"/>
      <line x1="0" y1="50" x2="{{ .ChartWidth }}" y2="50" stroke="#333" stroke-width="0.5" stroke-dasharray="4,4"/>
      <line x1="0" y1="100" x2="{{ .ChartWidth }}" y2="100" stroke="#333" stroke-width="0.5" stroke-dasharray="4,4"/>
      <line x1="0" y1="150" x2="{{ .ChartWidth }}" y2="150" stroke="#333" stroke-width="0.5" stroke-dasharray="4,4"/>
      <line x1="0" y1="200" x2="{{ .ChartWidth }}" y2="200" stroke="#333" stroke-width="1"/>
      <!-- CPU Line -->
      <polyline fill="none" stroke="#38bdf8" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"
        points="{{ .CPULinePoints }}"/>
      <polygon fill="url(#cpuGradient)" opacity="0.45"
        points="0,200 {{ .CPULinePoints }} {{ .ChartWidth }},200"/>
      <defs>
        <linearGradient id="cpuGradient" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#38bdf8" stop-opacity="0.7"/>
          <stop offset="100%" stop-color="#172554" stop-opacity="0.1"/>
        </linearGradient>
      </defs>
    </svg>
    <div class="chart-labels">
      <span>0%</span>
      <span>100%</span>
    </div>
  </div>

  <!-- Memory Chart -->
  <div class="section-title">Memory Footprint (MB)</div>
  <div class="chart-container">
    <svg viewBox="0 0 {{ .ChartWidth }} 200" preserveAspectRatio="none" class="chart-svg" role="img" aria-label="Memory usage over time">
      <!-- Grid lines -->
      <line x1="0" y1="0" x2="{{ .ChartWidth }}" y2="0" stroke="#333" stroke-width="1"/>
      <line x1="0" y1="50" x2="{{ .ChartWidth }}" y2="50" stroke="#333" stroke-width="0.5" stroke-dasharray="4,4"/>
      <line x1="0" y1="100" x2="{{ .ChartWidth }}" y2="100" stroke="#333" stroke-width="0.5" stroke-dasharray="4,4"/>
      <line x1="0" y1="150" x2="{{ .ChartWidth }}" y2="150" stroke="#333" stroke-width="0.5" stroke-dasharray="4,4"/>
      <line x1="0" y1="200" x2="{{ .ChartWidth }}" y2="200" stroke="#333" stroke-width="1"/>
      <!-- Memory Line -->
      <polyline fill="none" stroke="#8b5cf6" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"
        points="{{ .MemLinePoints }}"/>
      <polygon fill="url(#memGradient)" opacity="0.45"
        points="0,200 {{ .MemLinePoints }} {{ .ChartWidth }},200"/>
      <defs>
        <linearGradient id="memGradient" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#8b5cf6" stop-opacity="0.7"/>
          <stop offset="100%" stop-color="#2e1065" stop-opacity="0.1"/>
        </linearGradient>
      </defs>
    </svg>
    <div class="chart-labels">
      <span>0 MB</span>
      <span>500 MB</span>
    </div>
  </div>

  <!-- Threads Chart -->
  <div class="section-title">Thread Count</div>
  <div class="chart-container">
    <svg viewBox="0 0 {{ .ChartWidth }} 200" preserveAspectRatio="none" class="chart-svg" role="img" aria-label="Thread count over time">
      <!-- Grid lines -->
      <line x1="0" y1="0" x2="{{ .ChartWidth }}" y2="0" stroke="#333" stroke-width="1"/>
      <line x1="0" y1="50" x2="{{ .ChartWidth }}" y2="50" stroke="#333" stroke-width="0.5" stroke-dasharray="4,4"/>
      <line x1="0" y1="100" x2="{{ .ChartWidth }}" y2="100" stroke="#333" stroke-width="0.5" stroke-dasharray="4,4"/>
      <line x1="0" y1="150" x2="{{ .ChartWidth }}" y2="150" stroke="#333" stroke-width="0.5" stroke-dasharray="4,4"/>
      <line x1="0" y1="200" x2="{{ .ChartWidth }}" y2="200" stroke="#333" stroke-width="1"/>
      <!-- Threads Line -->
      <polyline fill="none" stroke="#00FF00" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
        points="{{ .ThrLinePoints }}"/>
      <polygon fill="url(#thrGradient)" opacity="0.3"
        points="0,200 {{ .ThrLinePoints }} {{ .ChartWidth }},200"/>
      <defs>
        <linearGradient id="thrGradient" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#00FF00" stop-opacity="0.6"/>
          <stop offset="100%" stop-color="#00FF00" stop-opacity="0.05"/>
        </linearGradient>
      </defs>
    </svg>
    <div class="chart-labels">
      <span>0</span>
      <span>max</span>
    </div>
  </div>

  <!-- Telemetry Table -->
  <div class="section-title">Raw Telemetry (last 200 samples)</div>
  <div class="table-container">
    <table>
      <thead>
        <tr>
          <th>#</th>
          <th>Timestamp</th>
          <th>CPU (%)</th>
          <th>Memory (KB)</th>
          <th>Threads</th>
          <th>Stack</th>
        </tr>
      </thead>
      <tbody>
      {{- $len := len .Telemetry }}{{ $start := sub $len 200 }}{{ if lt $start 0 }}{{ $start = 0 }}{{ end }}
      {{- range $i, $s := slice $.Telemetry $start }}
        <tr{{ if modcheck $i 2 }} class="alt"{{ end }}>
          <td>{{ add $start $i 1 }}</td>
          <td>{{ formatUnix $s.Timestamp }}</td>
          <td{{ if gt $s.CPUPercent 70.0 }} class="alert"{{ end }}>{{ printf "%.1f" $s.CPUPercent }}</td>
          <td>{{ $s.MemoryKB }}</td>
          <td>{{ $s.Threads }}</td>
          <td>{{ if $s.Stack }}<span class="stack-badge">✓</span>{{ else }}—{{ end }}</td>
        </tr>
      {{- end }}
      </tbody>
    </table>
  </div>

  <!-- Captured App Logs -->
  <div class="section-title">Captured App Logs</div>
  {{- if .Logs }}
  <div class="stack-section">
    <pre class="stack-pre">{{ range .Logs }}{{ . }}
{{- end }}</pre>
  </div>
  {{- else }}
  <p class="chart-labels">No app logs were captured during this session.</p>
  {{- end }}

  <!-- Stack Traces -->
  {{- $hasStack := false }}{{ range .Telemetry }}{{ if .Stack }}{{ $hasStack = true }}{{ end }}{{ end }}
  {{- if $hasStack }}
  <div class="section-title">Stack Traces (High CPU Samples)</div>
  <div class="stack-section">
  {{- $len := len .Telemetry }}
  {{- range $i, $s := .Telemetry }}
  {{- if $s.Stack }}
    <details class="stack-details">
      <summary class="stack-summary">Sample {{ inc $i }} — CPU {{ printf "%.1f" $s.CPUPercent }}% at {{ formatUnix $s.Timestamp }}</summary>
      <pre class="stack-pre">{{ $s.Stack }}</pre>
    </details>
  {{- end }}
  {{- end }}
  </div>
  {{- end }}

  <div class="footer">
    Generated by <a href="https://perfmon.qzz.io">perfmon</a> v{{ .Metadata.PerfmonVersion }}
  </div>
</div>
</body>
</html>`

// htmlRenderer holds the data passed to the HTML template.
type htmlRenderer struct {
	ExportData
	CSS          string
	ChartJS      string
	ChartWidth   int
	CPULinePoints string
	MemLinePoints string
	ThrLinePoints string
	Logo         string
}

// ExportHTML writes the export data as a standalone HTML file.
func ExportHTML(data ExportData, snapshots []engine.TelemetrySnapshot, opts Options) (string, error) {
	outPath := outputFilename(opts.OutputPath, FormatHTML)

	f, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	chartWidth := 800
	if len(snapshots) < 2 {
		chartWidth = 200
	}

	cpuPoints := buildSVGLinePoints(snapshots, chartWidth, 200, 0, 100, func(s engine.TelemetrySnapshot) float64 {
		return s.CPUPercent
	})

	// Determine max memory for scaling
	var maxMem float64
	for _, s := range snapshots {
		m := float64(s.MemoryKB) / 1024.0
		if m > maxMem {
			maxMem = m
		}
	}
	if maxMem < 100 {
		maxMem = 100
	}

	memPoints := buildSVGLinePoints(snapshots, chartWidth, 200, 0, maxMem, func(s engine.TelemetrySnapshot) float64 {
		return float64(s.MemoryKB) / 1024.0
	})

	// Determine max threads for scaling
	var maxThr float64
	for _, s := range snapshots {
		if float64(s.Threads) > maxThr {
			maxThr = float64(s.Threads)
		}
	}
	if maxThr < 10 {
		maxThr = 10
	}

	thrPoints := buildSVGLinePoints(snapshots, chartWidth, 200, 0, maxThr, func(s engine.TelemetrySnapshot) float64 {
		return float64(s.Threads)
	})

	funcMap := template.FuncMap{
		"add":            func(a, b, c int) int { return a + b + c - 2 },
		"inc":            func(a int) int { return a + 1 },
		"sub":            func(a, b int) int { return a - b },
		"modcheck":       func(a, b int) bool { return a%b == 0 },
		"slice":          func(s []engine.TelemetrySnapshot, start int) []engine.TelemetrySnapshot { if start >= len(s) { return nil }; return s[start:] },
		"formatBytesHTML": func(kb int64) string {
			if kb < 1024 {
				return fmt.Sprintf("%d KB", kb)
			}
			return fmt.Sprintf("%.1f MB", float64(kb)/1024.0)
		},
		"formatUnix": func(ts int64) string {
			t := time.Unix(ts, 0)
			return t.Format("15:04:05")
		},
	}

	tmpl, err := template.New("html").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return "", fmt.Errorf("template parse: %w", err)
	}

	renderer := htmlRenderer{
		ExportData:    data,
		CSS:           cssContent,
		ChartJS:       chartJSContent,
		ChartWidth:    chartWidth,
		CPULinePoints: cpuPoints,
		MemLinePoints: memPoints,
		ThrLinePoints: thrPoints,
		Logo:          logoSVG,
	}

	if err := tmpl.Execute(f, renderer); err != nil {
		return "", fmt.Errorf("template execute: %w", err)
	}

	return outPath, nil
}

func buildSVGLinePoints(snapshots []engine.TelemetrySnapshot, width, height int, dataMin, dataMax float64, valueFn func(engine.TelemetrySnapshot) float64) string {
	if len(snapshots) == 0 {
		return ""
	}
	if len(snapshots) == 1 {
		rangeY := dataMax - dataMin
		if rangeY == 0 {
			rangeY = 1
		}
		val := valueFn(snapshots[0])
		y := float64(height) - ((val-dataMin)/rangeY)*float64(height)
		if y < 0 {
			y = 0
		}
		if y > float64(height) {
			y = float64(height)
		}
		return fmt.Sprintf("0.0,%.1f", y)
	}

	rawN := len(snapshots)
	if rawN > chart.MaxPoints {
		rawN = chart.MaxPoints
	}
	raw := chart.SampleSnapshots(snapshots, rawN, valueFn)
	smooth := chart.SmoothSeries(raw, width)

	rangeY := dataMax - dataMin
	if rangeY == 0 {
		rangeY = 1
	}

	var step float64
	if width == 1 {
		step = 0
	} else {
		step = float64(width) / float64(width-1)
	}

	var b strings.Builder
	for i, val := range smooth {
		x := float64(i) * step
		y := float64(height) - ((val-dataMin)/rangeY)*float64(height)
		if y < 0 {
			y = 0
		}
		if y > float64(height) {
			y = float64(height)
		}
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(fmt.Sprintf("%.1f,%.1f", x, y))
	}
	return b.String()
}


