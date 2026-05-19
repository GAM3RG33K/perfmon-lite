package views

import (
	"strings"
	"testing"
)

func TestLineCount(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", "", 0},
		{"single line", "hello", 1},
		{"single line with newline", "hello\n", 1},
		{"two lines", "hello\nworld", 2},
		{"two lines with trailing newline", "hello\nworld\n", 2},
		{"multiple lines", "a\nb\nc\nd", 4},
		{"only newlines", "\n\n", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LineCount(tt.input)
			if got != tt.want {
				t.Errorf("LineCount(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestTruncateLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLines int
		want     string
	}{
		{"empty stays empty", "", 5, ""},
		{"single line within limit", "hello", 5, "hello"},
		{"within limit", "a\nb\nc", 5, "a\nb\nc"},
		{"at limit", "a\nb\nc", 3, "a\nb\nc"},
		{"truncates", "a\nb\nc\nd\ne", 3, "a\nb\nc"},
		{"maxLines 1", "a\nb\nc", 1, "a"},
		{"maxLines 0 becomes 1", "a\nb", 0, "a"},
		{"negative maxLines becomes 1", "a\nb", -5, "a"},
		{"trailing newline stripped before count", "a\nb\n", 1, "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateLines(tt.input, tt.maxLines)
			if got != tt.want {
				t.Errorf("TruncateLines(%q, %d) = %q, want %q", tt.input, tt.maxLines, got, tt.want)
			}
		})
	}
}

func TestTruncateLines_PreservesContent(t *testing.T) {
	input := "line1\nline2\nline3\nline4\nline5"
	got := TruncateLines(input, 3)
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "line1" {
		t.Errorf("expected 'line1', got %q", lines[0])
	}
	if lines[2] != "line3" {
		t.Errorf("expected 'line3', got %q", lines[2])
	}
}
