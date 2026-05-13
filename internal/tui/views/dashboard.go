package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"

	"github.com/w1n/perfmon/internal/engine"
	"github.com/w1n/perfmon/internal/tui/styles"
)

// DashboardView renders the main telemetry dashboard.
type DashboardView struct {
	Width  int
	Height int
}

// NewDashboardView creates a new dashboard view.
func NewDashboardView() *DashboardView {
	return &DashboardView{
		Width:  80,
		Height: 24,
	}
}

// Render draws the dashboard with current telemetry data.
func (dv *DashboardView) Render(snapshots []engine.TelemetrySnapshot, latest *engine.TelemetrySnapshot) string {
	if dv.Width < 40 {
		return "Terminal too narrow for dashboard"
	}

	var b strings.Builder

	// Header section
	b.WriteString(renderHeader(latest))
	b.WriteString("\n\n")

	// Charts section
	b.WriteString(renderCPUChart(snapshots, dv.Width-8))
	b.WriteString("\n\n")
	b.WriteString(renderMemoryChart(snapshots, dv.Width-8))
	b.WriteString("\n\n")

	// Stats row
	b.WriteString(renderStatsRow(snapshots))
	b.WriteString("\n")

	return wordwrap.String(b.String(), dv.Width)
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

	cpuStr := fmt.Sprintf("CPU: %.1f%%", latest.CPUPercent)
	memStr := fmt.Sprintf("RAM: %s", formatBytes(latest.MemoryKB*1024))
	thrStr := fmt.Sprintf("Threads: %d", latest.Threads)

	return lipgloss.JoinHorizontal(lipgloss.Center,
		styles.LabelStyle.Render("Status:"),
		" ",
		statusDot,
		"  ",
		styles.ValueStyle.Render(cpuStr),
		"  ",
		styles.ValueStyle.Render(memStr),
		"  ",
		styles.ValueStyle.Render(thrStr),
	)
}

func renderCPUChart(snapshots []engine.TelemetrySnapshot, width int) string {
	if len(snapshots) < 2 || width < 10 {
		return styles.LabelStyle.Render("CPU: waiting for data...")
	}

	title := styles.SubHeaderStyle.Render("CPU Utilization (%)")
	chart := renderSparkline(snapshots, width, func(s engine.TelemetrySnapshot) float64 {
		return s.CPUPercent
	}, 0, 100)

	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(chart)
	return b.String()
}

func renderMemoryChart(snapshots []engine.TelemetrySnapshot, width int) string {
	if len(snapshots) < 2 || width < 10 {
		return styles.LabelStyle.Render("Memory: waiting for data...")
	}

	title := styles.SubHeaderStyle.Render("Memory Footprint (MB)")
	chart := renderSparkline(snapshots, width, func(s engine.TelemetrySnapshot) float64 {
		return float64(s.MemoryKB) / 1024.0 // Convert KB to MB
	}, 0, 500)

	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(chart)
	return b.String()
}

func renderSparkline(snapshots []engine.TelemetrySnapshot, width int, valueFn func(engine.TelemetrySnapshot) float64, minVal, maxVal float64) string {
	if len(snapshots) < 2 || width < 5 {
		return ""
	}

	// Determine how many points we can display
	numPoints := len(snapshots)
	if numPoints > width-4 {
		numPoints = width - 4
	}
	// Sample evenly across available data
	step := float64(len(snapshots)) / float64(numPoints)

	// Build vertical bars
	height := 6
	canvas := make([][]rune, height)
	for i := range canvas {
		canvas[i] = []rune(strings.Repeat(" ", numPoints))
	}

	maxY := maxVal
	minY := minVal
	rangeY := maxY - minY
	if rangeY == 0 {
		rangeY = 1
	}

	for i := 0; i < numPoints; i++ {
		idx := int(float64(i) * step)
		if idx >= len(snapshots) {
			idx = len(snapshots) - 1
		}
		val := valueFn(snapshots[idx])
		barHeight := int(((val - minY) / rangeY) * float64(height-1))
		if barHeight < 0 {
			barHeight = 0
		}
		if barHeight >= height {
			barHeight = height - 1
		}

		// Draw from bottom up
		for row := height - 1; row >= height-1-barHeight; row-- {
			canvas[row][i] = '█'
		}
	}

	var b strings.Builder
	for row := 0; row < height; row++ {
		// Y-axis labels
		switch row {
		case 0:
			b.WriteString(fmt.Sprintf("%4.0f ┤", maxY))
		case height - 1:
			b.WriteString(fmt.Sprintf("%4.0f └", minY))
		default:
			b.WriteString("    │")
		}
		colored := string(canvas[row])
		// Color the line
		if row == height-1 {
			b.WriteString(lipgloss.NewStyle().Foreground(styles.Cyan).Render(colored))
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(styles.Magenta).Render(colored))
		}
		b.WriteString("\n")
	}
	// X-axis
	b.WriteString(strings.Repeat(" ", 5))
	b.WriteString(styles.LabelStyle.Render(strings.Repeat("─", numPoints-1)))
	b.WriteString("\n")

	// Time labels
	b.WriteString(strings.Repeat(" ", 5))
	timeLabel := styles.LabelStyle.Render(fmt.Sprintf("← %ds ago", len(snapshots)))
	b.WriteString(timeLabel)
	b.WriteString("\n")

	return b.String()
}

func renderStatsRow(snapshots []engine.TelemetrySnapshot) string {
	if len(snapshots) == 0 {
		return ""
	}

	summary := engine.ComputeMetricsSummary(snapshots)

	peakCPU := fmt.Sprintf("Peak CPU: %.1f%%", summary.PeakCPUPercent)
	peakRAM := fmt.Sprintf("Peak RAM: %s", formatBytes(summary.PeakMemoryKB*1024))
	avgCPU := fmt.Sprintf("Avg CPU: %.1f%%", summary.AverageCPUPercent)
	avgRAM := fmt.Sprintf("Avg RAM: %s", formatBytes(summary.AverageMemoryKB*1024))
	peakThr := fmt.Sprintf("Peak Threads: %d", summary.PeakThreads)
	duration := fmt.Sprintf("Duration: %ds", summary.DurationSeconds)

	items := []string{
		styles.LabelStyle.Render(peakCPU),
		styles.LabelStyle.Render(peakRAM),
		styles.LabelStyle.Render(avgCPU),
		styles.LabelStyle.Render(avgRAM),
		styles.LabelStyle.Render(peakThr),
		styles.LabelStyle.Render(duration),
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
