package views

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestColorToRGB(t *testing.T) {
	tests := []struct {
		color lipgloss.Color
		r, g, b int
	}{
		{lipgloss.Color("#FF0000"), 255, 0, 0},
		{lipgloss.Color("#00FF00"), 0, 255, 0},
		{lipgloss.Color("#0000FF"), 0, 0, 255},
		{lipgloss.Color("#FFFFFF"), 255, 255, 255},
		{lipgloss.Color("#000000"), 0, 0, 0},
		{lipgloss.Color("#123456"), 0x12, 0x34, 0x56},
	}

	for _, tt := range tests {
		r, g, b := colorToRGB(tt.color)
		if r != tt.r || g != tt.g || b != tt.b {
			t.Errorf("colorToRGB(%q) = (%d, %d, %d), want (%d, %d, %d)",
				tt.color, r, g, b, tt.r, tt.g, tt.b)
		}
	}
}

func TestColorToRGB_Invalid(t *testing.T) {
	r, g, b := colorToRGB(lipgloss.Color("invalid"))
	if r != 136 || g != 136 || b != 136 {
		t.Errorf("expected default (136,136,136) for invalid, got (%d,%d,%d)", r, g, b)
	}

	r, g, b = colorToRGB(lipgloss.Color("#GG"))
	if r != 136 || g != 136 || b != 136 {
		t.Errorf("expected default (136,136,136) for short hex, got (%d,%d,%d)", r, g, b)
	}
}

func TestLerpColor(t *testing.T) {
	a := lipgloss.Color("#000000")
	b := lipgloss.Color("#FFFFFF")

	r, g, bVal := colorToRGB(lerpColor(a, b, 0.0))
	if r != 0 || g != 0 || bVal != 0 {
		t.Errorf("lerp at 0.0 should be black, got (%d,%d,%d)", r, g, bVal)
	}

	r, g, bVal = colorToRGB(lerpColor(a, b, 1.0))
	if r != 255 || g != 255 || bVal != 255 {
		t.Errorf("lerp at 1.0 should be white, got (%d,%d,%d)", r, g, bVal)
	}

	r, g, bVal = colorToRGB(lerpColor(a, b, 0.5))
	if r != 127 || g != 127 || bVal != 127 {
		t.Errorf("lerp at 0.5 should be gray, got (%d,%d,%d)", r, g, bVal)
	}
}

func TestLerpColor_Clamping(t *testing.T) {
	a := lipgloss.Color("#000000")
	b := lipgloss.Color("#FFFFFF")

	r, g, bVal := colorToRGB(lerpColor(a, b, -0.5))
	if r != 0 || g != 0 || bVal != 0 {
		t.Errorf("lerp at -0.5 should clamp to black, got (%d,%d,%d)", r, g, bVal)
	}

	r, g, bVal = colorToRGB(lerpColor(a, b, 1.5))
	if r != 255 || g != 255 || bVal != 255 {
		t.Errorf("lerp at 1.5 should clamp to white, got (%d,%d,%d)", r, g, bVal)
	}
}
