package views

import "strings"

func LineCount(s string) int {
	if s == "" {
		return 0
	}
	return len(strings.Split(strings.TrimRight(s, "\n"), "\n"))
}

func TruncateLines(s string, maxLines int) string {
	if maxLines < 1 {
		maxLines = 1
	}
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) <= maxLines {
		return strings.Join(lines, "\n")
	}
	return strings.Join(lines[:maxLines], "\n")
}
