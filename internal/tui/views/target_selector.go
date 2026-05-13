package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/w1n/perfmon/internal/engine"
	"github.com/w1n/perfmon/internal/tui/styles"
)

// TargetSelectorView renders the device and process selection list.
type TargetSelectorView struct {
	Width        int
	Height       int
	Devices      []engine.Device
	Processes    []engine.AppProcess
	SelectedDevice int
	SelectedProcess int
	ShowProcesses  bool
}

// NewTargetSelectorView creates a new target selector view.
func NewTargetSelectorView() *TargetSelectorView {
	return &TargetSelectorView{
		Width:  80,
		Height: 24,
	}
}

// Render draws the target selector.
func (ts *TargetSelectorView) Render() string {
	if ts.Width < 40 {
		return "Terminal too narrow for target selector"
	}

	var b strings.Builder

	b.WriteString(styles.HeaderStyle.Render("Target Selection"))
	b.WriteString("\n\n")

	// Device list section
	b.WriteString(styles.SubHeaderStyle.Render("Devices"))
	b.WriteString("\n")
	if len(ts.Devices) == 0 {
		b.WriteString(styles.LabelStyle.Render("  No devices found. Connect a device or use --mock.\n"))
	} else {
		for i, d := range ts.Devices {
			prefix := "  "
			if i == ts.SelectedDevice {
				prefix = "▸ "
			}

			platformBadge := styles.PlatformBadge(d.Platform)
			deviceType := "emulator"
			if d.IsPhysical {
				deviceType = "physical"
			}
			status := "●"
			statusColor := styles.Green
			if !d.IsBooted {
				status = "○"
				statusColor = styles.DimWhite
			}
			statusDot := lipgloss.NewStyle().Foreground(statusColor).SetString(status).String()

			line := fmt.Sprintf("%s%s %s %s (%s) %s",
				prefix,
				statusDot,
				d.Name,
				d.ID,
				deviceType,
				platformBadge,
			)

			if i == ts.SelectedDevice {
				line = styles.HighlightStyle.Render(line)
			}

			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	// Process list section
	b.WriteString("\n")
	b.WriteString(styles.SubHeaderStyle.Render("Processes"))
	b.WriteString("\n")
	if len(ts.Processes) == 0 {
		if len(ts.Devices) > 0 {
			b.WriteString(styles.LabelStyle.Render("  No processes found. Select a device first.\n"))
		} else {
			b.WriteString(styles.LabelStyle.Render("  Waiting for device selection...\n"))
		}
	} else {
		for i, p := range ts.Processes {
			prefix := "  "
			if i == ts.SelectedProcess && ts.ShowProcesses {
				prefix = "▸ "
			}

			buildBadge := styles.BuildBadge(p.BuildType)
			line := fmt.Sprintf("%sPID %-6d %-40s %s",
				prefix,
				p.PID,
				p.PackageName,
				buildBadge,
			)

			if i == ts.SelectedProcess && ts.ShowProcesses {
				line = styles.HighlightStyle.Render(line)
			}

			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	return b.String()
}
