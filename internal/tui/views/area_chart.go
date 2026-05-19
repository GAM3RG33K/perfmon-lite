package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/w1n/perfmon/internal/chart"
	"github.com/w1n/perfmon/internal/engine"
	"github.com/w1n/perfmon/internal/tui/styles"
)

const chartLabelWidth = 6

type chartPalette struct {
	line lipgloss.Color
	hi   lipgloss.Color
	mid  lipgloss.Color
	lo   lipgloss.Color
}

var (
	cpuChartPalette = chartPalette{
		line: lipgloss.Color("#e0f2fe"),
		hi:   lipgloss.Color("#38bdf8"),
		mid:  lipgloss.Color("#2563eb"),
		lo:   lipgloss.Color("#172554"),
	}
	memChartPalette = chartPalette{
		line: lipgloss.Color("#ede9fe"),
		hi:   lipgloss.Color("#a78bfa"),
		mid:  lipgloss.Color("#7c3aed"),
		lo:   lipgloss.Color("#2e1065"),
	}
)

func lerpColor(a, b lipgloss.Color, t float64) lipgloss.Color {
	ar, ag, ab := colorToRGB(a)
	br, bg, bb := colorToRGB(b)
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	r := int(float64(ar) + float64(br-ar)*t)
	g := int(float64(ag) + float64(bg-ag)*t)
	bv := int(float64(ab) + float64(bb-ab)*t)
	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, g, bv))
}

func colorToRGB(c lipgloss.Color) (int, int, int) {
	s := strings.TrimPrefix(string(c), "#")
	if len(s) != 6 {
		return 136, 136, 136
	}
	var r, g, b int
	if _, err := fmt.Sscanf(s, "%02x%02x%02x", &r, &g, &b); err != nil {
		return 136, 136, 136
	}
	return r, g, b
}

func renderColoredGraph(data []int, plotW, height int, palette chartPalette) []string {
	plain := chart.BtopGraphRows(data, plotW, height, chart.BlockSymbolsUp)
	gridStyle := lipgloss.NewStyle().Foreground(styles.ChartGrid)
	lines := make([]string, len(plain))
	for horizon, row := range plain {
		t := float64(horizon) / float64(height-1)
		rowStyle := lipgloss.NewStyle().Foreground(lerpColor(palette.hi, palette.lo, t))
		var b strings.Builder
		for _, ch := range row {
			if ch == '·' {
				b.WriteString(gridStyle.Render("·"))
			} else if ch == ' ' {
				b.WriteString(" ")
			} else {
				b.WriteString(rowStyle.SetString(string(ch)).String())
			}
		}
		lines[horizon] = b.String()
	}
	return lines
}

func renderMiniGauge(pct float64, width int, color lipgloss.Color) string {
	g := chart.MiniGauge(pct, width)
	filled := strings.Count(g, "█")
	fg := lipgloss.NewStyle().Foreground(color)
	bg := lipgloss.NewStyle().Foreground(styles.ChartGrid)
	return fg.Render(strings.Repeat("█", filled)) + bg.Render(strings.Repeat("░", len(g)-filled))
}

func renderBtopAreaChart(
	snapshots []engine.TelemetrySnapshot,
	plotW int,
	valueFn func(engine.TelemetrySnapshot) float64,
	minV, maxV float64,
	palette chartPalette,
) string {
	if len(snapshots) < 2 || plotW < 8 {
		return ""
	}

	height := chart.DefaultHeight
	data := chart.PreparePercentData(snapshots, plotW, valueFn, minV, maxV)
	graphLines := renderColoredGraph(data, plotW, height, palette)

	var b strings.Builder
	labelW := chartLabelWidth
	gridStyle := lipgloss.NewStyle().Foreground(styles.ChartGrid)

	for horizon := 0; horizon < height; horizon++ {
		switch horizon {
		case 0:
			b.WriteString(fmt.Sprintf("%*.*f │", labelW-1, 0, maxV))
		case height / 2:
			mid := (maxV + minV) / 2
			b.WriteString(fmt.Sprintf("%*.*f │", labelW-1, 0, mid))
		default:
			b.WriteString(fmt.Sprintf("%*s │", labelW-1, ""))
		}
		b.WriteString(graphLines[horizon])
		b.WriteString("\n")
	}

	b.WriteString(strings.Repeat(" ", labelW))
	b.WriteString(gridStyle.Render("└" + strings.Repeat("─", plotW)))
	b.WriteString("\n")

	leftLabel := fmt.Sprintf("%ds ago", chart.MaxPoints)
	gap := plotW - len(leftLabel) - 3
	if gap < 1 {
		gap = 1
	}
	b.WriteString(strings.Repeat(" ", labelW))
	b.WriteString(styles.LabelStyle.Render(leftLabel + strings.Repeat(" ", gap) + "now"))
	b.WriteString("\n")
	return b.String()
}
