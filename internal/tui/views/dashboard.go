package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/GAM3RG33K/perfmon-lite/internal/chart"
	"github.com/GAM3RG33K/perfmon-lite/internal/engine"
	"github.com/GAM3RG33K/perfmon-lite/internal/tui/styles"
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
	chartData := chart.LimitSnapshots(snapshots, chart.MaxPoints)
	plotW := dv.Width - chartLabelWidth - 2
	if plotW < 8 {
		plotW = 8
	}
	if plotW > chart.MaxPoints {
		plotW = chart.MaxPoints
	}

	var b strings.Builder
	b.WriteString(renderHeader(latest))
	b.WriteString("\n\n")
	b.WriteString(renderCPUChart(chartData, plotW))
	b.WriteString("\n\n")
	b.WriteString(renderMemoryChart(chartData, plotW))
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

func renderCPUChart(snapshots []engine.TelemetrySnapshot, plotW int) string {
	if len(snapshots) < 2 || plotW < 4 {
		return styles.LabelStyle.Render("CPU: waiting for data...")
	}
	maxVal := chart.ComputeMax(snapshots, func(s engine.TelemetrySnapshot) float64 { return s.CPUPercent })
	if maxVal < 10 {
		maxVal = 100
	} else {
		maxVal = chart.CeilNice(maxVal * 1.2)
	}
	latest := snapshots[len(snapshots)-1]
	gauge := renderMiniGauge(latest.CPUPercent, 12, cpuChartPalette.hi)
	chart := renderBtopAreaChart(snapshots, plotW, func(s engine.TelemetrySnapshot) float64 { return s.CPUPercent }, 0, maxVal, cpuChartPalette)
	var b strings.Builder
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Center,
		styles.SubHeaderStyle.Render("CPU Utilization"),
		"  ",
		styles.ValueStyle.Render(fmt.Sprintf("%.1f%%", latest.CPUPercent)),
		" ",
		gauge,
	))
	b.WriteString("\n")
	b.WriteString(chart)
	return b.String()
}

func renderMemoryChart(snapshots []engine.TelemetrySnapshot, plotW int) string {
	if len(snapshots) < 2 || plotW < 4 {
		return styles.LabelStyle.Render("Memory: waiting for data...")
	}
	maxVal := chart.ComputeMax(snapshots, func(s engine.TelemetrySnapshot) float64 {
		return float64(s.MemoryKB) / 1024.0
	})
	if maxVal < 10 {
		maxVal = 100
	} else if maxVal > 0 {
		maxVal = chart.CeilNice(maxVal * 1.2)
	}
	latestMB := float64(snapshots[len(snapshots)-1].MemoryKB) / 1024.0
	pct := latestMB / maxVal * 100
	if pct > 100 {
		pct = 100
	}
	gauge := renderMiniGauge(pct, 12, memChartPalette.hi)
	chart := renderBtopAreaChart(snapshots, plotW, func(s engine.TelemetrySnapshot) float64 {
		return float64(s.MemoryKB) / 1024.0
	}, 0, maxVal, memChartPalette)
	var b strings.Builder
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Center,
		styles.SubHeaderStyle.Render("Memory"),
		"  ",
		styles.ValueStyle.Render(fmt.Sprintf("%.0f MB", latestMB)),
		" ",
		gauge,
	))
	b.WriteString("\n")
	b.WriteString(chart)
	return b.String()
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
