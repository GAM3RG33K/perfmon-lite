package styles

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/GAM3RG33K/perfmon-lite/internal/engine"
)

// DebugBadge returns a styled [DEBUG] badge.
var DebugBadge = lipgloss.NewStyle().
	Background(Green).
	Foreground(lipgloss.Color("#000000")).
	Bold(true).
	Padding(0, 1).
	SetString("[DEBUG]").
	String()

// ReleaseBadge returns a styled [RELEASE] badge.
var ReleaseBadge = lipgloss.NewStyle().
	Background(Amber).
	Foreground(lipgloss.Color("#000000")).
	Bold(true).
	Padding(0, 1).
	SetString("[RELEASE]").
	String()

// BuildBadge returns the appropriate badge for the given build type.
func BuildBadge(bt engine.BuildType) string {
	switch bt {
	case engine.BuildDebug:
		return DebugBadge
	case engine.BuildRelease:
		return ReleaseBadge
	default:
		return lipgloss.NewStyle().
			Background(Red).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 1).
			SetString("[UNKNOWN]").
			String()
	}
}

// PlatformBadge returns a badge for the platform.
func PlatformBadge(platform engine.Platform) string {
	var bg lipgloss.Color
	switch platform {
	case engine.PlatformAndroid:
		bg = Green
	case engine.PlatformIOS:
		bg = lipgloss.Color("#555555")
	case engine.PlatformMock:
		bg = Magenta
	default:
		bg = Red
	}

	return lipgloss.NewStyle().
		Background(bg).
		Foreground(lipgloss.Color("#000000")).
		Bold(true).
		Padding(0, 1).
		SetString(string(platform)).
		String()
}
