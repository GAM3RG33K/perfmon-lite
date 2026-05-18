package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/w1n/perfmon/internal/engine"
	"github.com/w1n/perfmon/internal/tui/styles"
)

type DashboardView struct {
	Width  int
	Height int
}

func NewDashboardView() *DashboardView {
	return &DashboardView{Width: 80, Height: 24}
}

func (dv *DashboardView) Render(snapshots []engine.TelemetrySnapshot, latest *engine.TelemetrySnapshot) string {
	if dv.Width < 40 {
		return "Terminal too narrow for dashboard"
	}
	var b strings.Builder
	b.WriteString(renderHeader(latest))
	b.WriteString("\n\n")
	b.WriteString(renderCPUChart(snapshots, dv.Width-8))
	b.WriteString("\n\n")
	b.WriteString(renderMemoryChart(snapshots, dv.Width-8))
	b.WriteString("\n\n")
	b.WriteString(renderStatsRow(snapshots))
	b.WriteString("\n")
	return b.String()
}

func renderHeader(latest *engine.TelemetrySnapshot) string {
	if latest == nil {
		return styles.HeaderStyle.Render("No data — waiting for telemetry...")
	}
	status := "●"
	statusColor := styles.Green
	if latest.CPUPercent > 70 {
		statusColor = styles.Amber
	}
	if latest.CPUPercent > 85 {
		statusColor = styles.Red
	}
	statusDot := lipgloss.NewStyle().Foreground(statusColor).SetString(status).String()
	return lipgloss.JoinHorizontal(lipgloss.Center,
		styles.LabelStyle.Render("Status:"), " ", statusDot, "  ",
		styles.ValueStyle.Render(fmt.Sprintf("CPU: %.1f%%", latest.CPUPercent)), "  ",
		styles.ValueStyle.Render(fmt.Sprintf("RAM: %s", formatBytes(latest.MemoryKB*1024))), "  ",
		styles.ValueStyle.Render(fmt.Sprintf("Threads: %d", latest.Threads)),
	)
}

// ─── Line chart ─────────────────────────────────────────────

func renderCPUChart(snapshots []engine.TelemetrySnapshot, width int) string {
	if len(snapshots) < 2 || width < 10 {
		return styles.LabelStyle.Render("CPU: waiting for data...")
	}
	maxVal := computeMax(snapshots, func(s engine.TelemetrySnapshot) float64 { return s.CPUPercent })
	if maxVal < 10 {
		maxVal = 100
	} else {
		maxVal = ceilNice(maxVal * 1.2)
	}
	chart := renderLineChart(snapshots, width, func(s engine.TelemetrySnapshot) float64 { return s.CPUPercent }, 0, maxVal, styles.Magenta)
	var b strings.Builder
	b.WriteString(styles.SubHeaderStyle.Render("CPU Utilization (%)"))
	b.WriteString("\n")
	b.WriteString(chart)
	return b.String()
}

func renderMemoryChart(snapshots []engine.TelemetrySnapshot, width int) string {
	if len(snapshots) < 2 || width < 10 {
		return styles.LabelStyle.Render("Memory: waiting for data...")
	}
	maxVal := computeMax(snapshots, func(s engine.TelemetrySnapshot) float64 {
		return float64(s.MemoryKB) / 1024.0
	})
	if maxVal < 10 {
		maxVal = 100
	} else if maxVal > 0 {
		maxVal = ceilNice(maxVal * 1.2)
	}
	chart := renderLineChart(snapshots, width, func(s engine.TelemetrySnapshot) float64 {
		return float64(s.MemoryKB) / 1024.0
	}, 0, maxVal, styles.Magenta)
	var b strings.Builder
	b.WriteString("│ ")
	b.WriteString(styles.SubHeaderStyle.Render("Memory Footprint (MB)"))
	b.WriteString("\n")
	b.WriteString(chart)
	return b.String()
}

// renderLineChart draws an ASCII line chart with filled area underneath.
// Uses Unicode block chars: █ (filled), ▄ (lower half), ▀ (upper half), ░ (light shade)
func renderLineChart(snapshots []engine.TelemetrySnapshot, width int, valueFn func(engine.TelemetrySnapshot) float64, minVal, maxVal float64, lineColor lipgloss.Color) string {
	if len(snapshots) < 2 || width < 8 {
		return ""
	}

	numPoints := len(snapshots)
	chWidth := width / 2 * 2 // ensure even for half-block chars
	if chWidth > numPoints*2 {
		chWidth = numPoints * 2
	}
	if chWidth < 4 {
		chWidth = 4
	}

	step := float64(len(snapshots)) / float64(chWidth)
	rangeY := maxVal - minVal
	if rangeY == 0 {
		rangeY = 1
	}

	// Collect sample points (each row is 2 chars with half-block rendering)
	height := 6
	sampledVals := make([]float64, chWidth)
	for i := 0; i < chWidth; i++ {
		idx := int(float64(i) * step)
		if idx >= len(snapshots) {
			idx = len(snapshots) - 1
		}
		sampledVals[i] = valueFn(snapshots[idx])
	}

	// Build the chart from top to bottom
	var b strings.Builder
	// Y-axis label width
	labelW := 5

	for row := 0; row < height; row++ {
		switch row {
		case 0:
			b.WriteString(fmt.Sprintf("%*.*f ┤", labelW-1, 0, maxVal))
		case height - 1:
			b.WriteString(fmt.Sprintf("%*.*f └", labelW-1, 0, minVal))
		default:
			b.WriteString(fmt.Sprintf("%*s │", labelW-1, ""))
		}

		for i := 0; i < chWidth; i++ {
			val := sampledVals[i]
			normalized := (val - minVal) / rangeY
			cellTop := float64(height-1-row) / float64(height-1)
			cellBottom := float64(height-2-row) / float64(height-1)

			var ch rune
			if normalized >= cellTop {
				// Fully filled
				ch = '█'
			} else if normalized > cellBottom {
				// Half filled (lower half block)
				ch = '▄'
			} else {
				// Empty
				ch = ' '
			}
			styled := lipgloss.NewStyle().Foreground(lineColor).SetString(string(ch)).String()
			b.WriteString(styled)
		}
		b.WriteString("\n")
	}

	// X-axis
	b.WriteString(strings.Repeat(" ", labelW))
	b.WriteString(styles.LabelStyle.Render("└" + strings.Repeat("─", chWidth-1)))
	b.WriteString("\n")

	// Time label
	b.WriteString(strings.Repeat(" ", labelW))
	b.WriteString(styles.LabelStyle.Render(fmt.Sprintf("← %ds ago", len(snapshots))))
	b.WriteString("\n")

	return b.String()
}

func computeMax(snapshots []engine.TelemetrySnapshot, fn func(engine.TelemetrySnapshot) float64) float64 {
	var max float64
	for _, s := range snapshots {
		if v := fn(s); v > max {
			max = v
		}
	}
	return max
}

func ceilNice(v float64) float64 {
	if v < 1 {
		return 1
	}
	mag := 1.0
	for v >= 10.0 {
		v /= 10.0
		mag *= 10.0
	}
	switch {
	case v <= 1:
		v = 1
	case v <= 2:
		v = 2
	case v <= 5:
		v = 5
	default:
		v = 10
	}
	return v * mag
}

func renderStatsRow(snapshots []engine.TelemetrySnapshot) string {
	if len(snapshots) == 0 {
		return ""
	}
	summary := engine.ComputeMetricsSummary(snapshots)
	items := []string{
		styles.LabelStyle.Render(fmt.Sprintf("Peak CPU: %.1f%%", summary.PeakCPUPercent)),
		styles.LabelStyle.Render(fmt.Sprintf("Peak RAM: %s", formatBytes(summary.PeakMemoryKB*1024))),
		styles.LabelStyle.Render(fmt.Sprintf("Avg CPU: %.1f%%", summary.AverageCPUPercent)),
		styles.LabelStyle.Render(fmt.Sprintf("Avg RAM: %s", formatBytes(summary.AverageMemoryKB*1024))),
		styles.LabelStyle.Render(fmt.Sprintf("Peak Threads: %d", summary.PeakThreads)),
		styles.LabelStyle.Render(fmt.Sprintf("Duration: %ds", summary.DurationSeconds)),
	}
	return lipgloss.JoinHorizontal(lipgloss.Center,
		strings.Join(items, "  │  "),
	)
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
