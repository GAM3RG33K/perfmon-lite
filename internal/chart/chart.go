package chart

import (
	"fmt"
	"math"
	"strings"

	"github.com/w1n/perfmon/internal/engine"
)

const (
	MaxPoints       = 100
	DefaultHeight   = 10
	DefaultWidth    = 60
	DefaultLabelW   = 6
)

var BlockSymbolsUp = []rune{
	' ', '▗', '▗', '▐', '▐',
	'▖', '▄', '▄', '▟', '▟',
	'▖', '▄', '▄', '▟', '▟',
	'▌', '▙', '▙', '█', '█',
	'▌', '▙', '▙', '█', '█',
}

func CatmullRom(p0, p1, p2, p3, t float64) float64 {
	t2 := t * t
	t3 := t2 * t
	return 0.5 * ((2 * p1) +
		(-p0+p2)*t +
		(2*p0-5*p1+4*p2-p3)*t2 +
		(-p0+3*p1-3*p2+p3)*t3)
}

func SmoothSeries(vals []float64, samples int) []float64 {
	n := len(vals)
	if n == 0 {
		return nil
	}
	if n == 1 {
		out := make([]float64, samples)
		for i := range out {
			out[i] = vals[0]
		}
		return out
	}
	out := make([]float64, samples)
	for i := 0; i < samples; i++ {
		t := float64(i) / float64(samples-1) * float64(n-1)
		i1 := int(math.Floor(t))
		if i1 >= n-1 {
			out[i] = vals[n-1]
			continue
		}
		frac := t - float64(i1)
		p0, p1, p2, p3 := vals[i1], vals[i1+1], vals[i1+1], vals[i1+1]
		if i1 > 0 {
			p0 = vals[i1-1]
		}
		if i1+2 < n {
			p3 = vals[i1+2]
		}
		out[i] = CatmullRom(p0, p1, p2, p3, frac)
	}
	return out
}

func SampleSnapshots(snapshots []engine.TelemetrySnapshot, n int, valueFn func(engine.TelemetrySnapshot) float64) []float64 {
	vals := make([]float64, n)
	if len(snapshots) == 1 {
		for i := range vals {
			vals[i] = valueFn(snapshots[0])
		}
		return vals
	}
	for i := 0; i < n; i++ {
		idx := int(float64(i) * float64(len(snapshots)-1) / float64(n-1))
		vals[i] = valueFn(snapshots[idx])
	}
	return vals
}

func LimitSnapshots(s []engine.TelemetrySnapshot, n int) []engine.TelemetrySnapshot {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

func ScaleToPercent(vals []float64, minV, maxV float64) []int {
	out := make([]int, len(vals))
	rng := maxV - minV
	if rng == 0 {
		rng = 1
	}
	for i, v := range vals {
		pct := (v - minV) / rng * 100
		if pct < 0 {
			pct = 0
		}
		if pct > 100 {
			pct = 100
		}
		out[i] = int(math.Round(pct))
	}
	return out
}

func graphLevel(value, curHigh, curLow int, mod float64) int {
	if value >= curHigh {
		return 4
	}
	if value <= curLow {
		return 0
	}
	if curHigh == curLow {
		return 0
	}
	lv := int(math.Round(float64(value-curLow)*4/float64(curHigh-curLow) + mod))
	if lv < 0 {
		lv = 0
	}
	if lv > 4 {
		lv = 4
	}
	return lv
}

func BtopGraphRows(data []int, width, height int, symbols []rune) []string {
	if len(data) == 0 || width < 1 || height < 1 {
		return nil
	}

	padded := make([]int, width)
	start := 0
	if len(data) > width {
		start = len(data) - width
	}
	copy(padded[width-(len(data)-start):], data[start:])

	rows := make([][]rune, height)
	for h := 0; h < height; h++ {
		rows[h] = make([]rune, width)
	}

	mod := 0.1
	if height == 1 {
		mod = 0.3
	}

	last := 0
	for x := 0; x < width; x++ {
		cur := padded[x]
		for horizon := 0; horizon < height; horizon++ {
			curHigh := 100
			curLow := 0
			if height > 1 {
				curHigh = int(math.Round(100.0 * float64(height-horizon) / float64(height)))
				curLow = int(math.Round(100.0 * float64(height-(horizon+1)) / float64(height)))
			}
			prevLv := graphLevel(last, curHigh, curLow, mod)
			curLv := graphLevel(cur, curHigh, curLow, mod)
			if prevLv+curLv == 0 {
				rows[horizon][x] = ' '
			} else {
				rows[horizon][x] = symbols[prevLv*5+curLv]
			}
		}
		last = cur
	}

	lines := make([]string, height)
	for horizon := 0; horizon < height; horizon++ {
		var b strings.Builder
		for x := 0; x < width; x++ {
			ch := rows[horizon][x]
			if ch == ' ' {
				if height > 1 && (horizon == height/4 || horizon == height/2 || horizon == 3*height/4) {
					b.WriteRune('·')
				} else {
					b.WriteRune(' ')
				}
			} else {
				b.WriteRune(ch)
			}
		}
		lines[horizon] = b.String()
	}
	return lines
}

func CeilNice(v float64) float64 {
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

func ComputeMax(snapshots []engine.TelemetrySnapshot, fn func(engine.TelemetrySnapshot) float64) float64 {
	var max float64
	for _, s := range snapshots {
		if v := fn(s); v > max {
			max = v
		}
	}
	return max
}

type AreaConfig struct {
	Title    string
	Width    int
	Height   int
	MinV     float64
	MaxV     float64
	Unit     string
	Window   int
}

func PreparePercentData(snapshots []engine.TelemetrySnapshot, plotW int, valueFn func(engine.TelemetrySnapshot) float64, minV, maxV float64) []int {
	rawN := len(snapshots)
	if rawN > MaxPoints {
		rawN = MaxPoints
	}
	raw := SampleSnapshots(snapshots, rawN, valueFn)
	smooth := SmoothSeries(raw, plotW)
	return ScaleToPercent(smooth, minV, maxV)
}

func RenderAreaChart(snapshots []engine.TelemetrySnapshot, plotW int, valueFn func(engine.TelemetrySnapshot) float64, cfg AreaConfig) string {
	if len(snapshots) < 2 || plotW < 8 {
		return ""
	}
	height := cfg.Height
	if height < 1 {
		height = DefaultHeight
	}
	labelW := DefaultLabelW

	data := PreparePercentData(snapshots, plotW, valueFn, cfg.MinV, cfg.MaxV)
	graphLines := BtopGraphRows(data, plotW, height, BlockSymbolsUp)

	var b strings.Builder
	b.WriteString(cfg.Title)
	b.WriteString("\n")

	for horizon := 0; horizon < height; horizon++ {
		switch horizon {
		case 0:
			b.WriteString(fmt.Sprintf("%*.*f │", labelW-1, 0, cfg.MaxV))
		case height / 2:
			mid := (cfg.MaxV + cfg.MinV) / 2
			b.WriteString(fmt.Sprintf("%*.*f │", labelW-1, 0, mid))
		default:
			b.WriteString(fmt.Sprintf("%*s │", labelW-1, ""))
		}
		b.WriteString(graphLines[horizon])
		b.WriteString("\n")
	}

	b.WriteString(strings.Repeat(" ", labelW))
	b.WriteString("└" + strings.Repeat("─", plotW))
	b.WriteString("\n")

	window := cfg.Window
	if window <= 0 {
		window = MaxPoints
	}
	leftLabel := fmt.Sprintf("%ds ago", window)
	gap := plotW - len(leftLabel) - 3
	if gap < 1 {
		gap = 1
	}
	b.WriteString(strings.Repeat(" ", labelW))
	b.WriteString(leftLabel + strings.Repeat(" ", gap) + "now")
	if cfg.Unit != "" {
		b.WriteString("  (" + cfg.Unit + ")")
	}
	b.WriteString("\n")
	return b.String()
}

func RenderCPUChart(snapshots []engine.TelemetrySnapshot, plotW int) string {
	if len(snapshots) < 2 {
		return ""
	}
	maxVal := ComputeMax(snapshots, func(s engine.TelemetrySnapshot) float64 { return s.CPUPercent })
	if maxVal < 10 {
		maxVal = 100
	} else {
		maxVal = CeilNice(maxVal * 1.2)
	}
	latest := snapshots[len(snapshots)-1].CPUPercent
	return RenderAreaChart(snapshots, plotW, func(s engine.TelemetrySnapshot) float64 { return s.CPUPercent }, AreaConfig{
		Title:  fmt.Sprintf("CPU Utilization  %.1f%%", latest),
		Width:  plotW,
		Height: DefaultHeight,
		MinV:   0,
		MaxV:   maxVal,
		Unit:   "%",
		Window: MaxPoints,
	})
}

func RenderMemoryChart(snapshots []engine.TelemetrySnapshot, plotW int) string {
	if len(snapshots) < 2 {
		return ""
	}
	maxVal := ComputeMax(snapshots, func(s engine.TelemetrySnapshot) float64 {
		return float64(s.MemoryKB) / 1024.0
	})
	if maxVal < 10 {
		maxVal = 100
	} else if maxVal > 0 {
		maxVal = CeilNice(maxVal * 1.2)
	}
	latestMB := float64(snapshots[len(snapshots)-1].MemoryKB) / 1024.0
	return RenderAreaChart(snapshots, plotW, func(s engine.TelemetrySnapshot) float64 {
		return float64(s.MemoryKB) / 1024.0
	}, AreaConfig{
		Title:  fmt.Sprintf("Memory  %.0f MB", latestMB),
		Width:  plotW,
		Height: DefaultHeight,
		MinV:   0,
		MaxV:   maxVal,
		Unit:   "MB",
		Window: MaxPoints,
	})
}

func RenderSessionCharts(snapshots []engine.TelemetrySnapshot, plotW int) string {
	if len(snapshots) < 2 {
		return ""
	}
	if plotW < 8 {
		plotW = 8
	}
	var b strings.Builder
	b.WriteString(RenderCPUChart(snapshots, plotW))
	b.WriteString("\n")
	b.WriteString(RenderMemoryChart(snapshots, plotW))
	return b.String()
}

func MiniGauge(pct float64, width int) string {
	if width < 4 {
		width = 4
	}
	filled := int(pct / 100 * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}
