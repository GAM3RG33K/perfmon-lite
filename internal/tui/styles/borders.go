package styles

import "github.com/charmbracelet/lipgloss"

// PanelBorder defines a rounded border style for panels.
var PanelBorder = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(DimWhite).
	Padding(1, 2)

// ActiveTabBorder defines a highlighted tab border.
var ActiveTabBorder = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(Cyan).
	Foreground(Cyan).
	Bold(true).
	Padding(0, 2)

// InactiveTabBorder defines a normal tab border.
var InactiveTabBorder = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(DimWhite).
	Foreground(DimWhite).
	Padding(0, 2)

// FooterStyle defines the persistent command footer at the bottom.
var FooterStyle = lipgloss.NewStyle().
	Foreground(DimWhite).
	Background(MediumGray).
	Padding(0, 1).
	Width(80)

// Separator is a horizontal line for dividing sections.
var Separator = lipgloss.NewStyle().
	Foreground(DimWhite).
	SetString("─").
	String()
