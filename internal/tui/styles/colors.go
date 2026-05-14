package styles

import "github.com/charmbracelet/lipgloss"

// Primary color palette
var (
	// Cyan — used for selection highlights and headers
	Cyan = lipgloss.Color("#00FFFF")

	// Magenta — used for charts and telemetry peaks
	Magenta = lipgloss.Color("#FF00FF")

	// Green — used for debug badges
	Green = lipgloss.Color("#00FF00")

	// Amber — used for release badges
	Amber = lipgloss.Color("#FFB000")

	// Red — used for errors and warnings
	Red = lipgloss.Color("#FF4500")

	// DimWhite — used for borders and subtle text
	DimWhite = lipgloss.Color("#888888")

	// White — primary text color
	White = lipgloss.Color("#FFFFFF")

	// DarkGray — background for panels
	DarkGray = lipgloss.Color("#1A1A1A")

	// MediumGray — secondary background
	MediumGray = lipgloss.Color("#2D2D2D")
)

// Common style shortcuts
var (
	HeaderStyle = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true).
			Padding(0, 1)

	SubHeaderStyle = lipgloss.NewStyle().
			Foreground(White).
			Bold(true)

	LabelStyle = lipgloss.NewStyle().
			Foreground(DimWhite)

	ValueStyle = lipgloss.NewStyle().
			Foreground(White).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Red).
			Bold(true)

	HighlightStyle = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true)

	TitleStyle = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true).
			Padding(0, 1)

	AppStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Margin(0, 1)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(Green).
			Padding(0, 1)

	HelpTitle = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true).
			Padding(0, 1)

	HelpSectionTitle = lipgloss.NewStyle().
				Foreground(White).
				Bold(true).
				Underline(true).
				Padding(0, 1)

	HelpKey = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true)

	HelpDesc = lipgloss.NewStyle().
			Foreground(DimWhite)

	HelpFooter = lipgloss.NewStyle().
			Foreground(Amber).
			Padding(0, 1)
)
