package views

import (
	"strings"
	"testing"
)

func TestNewLogsView(t *testing.T) {
	lv := NewLogsView(500)
	if lv.Width != 80 {
		t.Errorf("expected width 80, got %d", lv.Width)
	}
	if lv.Height != 24 {
		t.Errorf("expected height 24, got %d", lv.Height)
	}
	if lv.MaxEntries != 500 {
		t.Errorf("expected max 500, got %d", lv.MaxEntries)
	}
	if lv.ScrollPos != 0 {
		t.Errorf("expected scroll pos 0, got %d", lv.ScrollPos)
	}
}

func TestLogsView_AddEntry(t *testing.T) {
	lv := NewLogsView(1000)
	lv.AddEntry("INFO", "test message")

	if len(lv.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(lv.Entries))
	}
	if lv.Entries[0].Level != "INFO" {
		t.Errorf("expected level INFO, got %s", lv.Entries[0].Level)
	}
	if lv.Entries[0].Message != "test message" {
		t.Errorf("expected message 'test message', got %s", lv.Entries[0].Message)
	}
	if lv.Entries[0].Timestamp.IsZero() {
		t.Error("timestamp should be set")
	}
	if lv.ScrollPos != 1 {
		t.Errorf("expected scroll pos 1, got %d", lv.ScrollPos)
	}
}

func TestLogsView_AddEntry_MaxEntries(t *testing.T) {
	lv := NewLogsView(3)
	lv.AddEntry("INFO", "msg1")
	lv.AddEntry("INFO", "msg2")
	lv.AddEntry("INFO", "msg3")
	lv.AddEntry("INFO", "msg4")

	if len(lv.Entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(lv.Entries))
	}
	if lv.Entries[0].Message != "msg2" {
		t.Errorf("expected first entry 'msg2', got %s", lv.Entries[0].Message)
	}
	if lv.Entries[2].Message != "msg4" {
		t.Errorf("expected last entry 'msg4', got %s", lv.Entries[2].Message)
	}
}

func TestLogsView_ScrollUp(t *testing.T) {
	lv := NewLogsView(1000)
	lv.AddEntry("INFO", "msg1")
	lv.AddEntry("INFO", "msg2")

	lv.ScrollUp()
	if lv.ScrollPos != 1 {
		t.Errorf("expected scroll pos 1, got %d", lv.ScrollPos)
	}

	lv.ScrollUp()
	if lv.ScrollPos != 0 {
		t.Errorf("expected scroll pos 0, got %d", lv.ScrollPos)
	}

	lv.ScrollUp()
	if lv.ScrollPos != 0 {
		t.Error("ScrollUp should not go below 0")
	}
}

func TestLogsView_ScrollDown(t *testing.T) {
	lv := NewLogsView(1000)
	lv.AddEntry("INFO", "msg1")
	lv.AddEntry("INFO", "msg2")

	lv.ScrollPos = 0
	lv.ScrollDown()
	if lv.ScrollPos != 1 {
		t.Errorf("expected scroll pos 1, got %d", lv.ScrollPos)
	}

	lv.ScrollDown()
	if lv.ScrollPos != 2 {
		t.Errorf("expected scroll pos 2, got %d", lv.ScrollPos)
	}

	lv.ScrollDown()
	if lv.ScrollPos != 2 {
		t.Error("ScrollDown should not exceed entry count")
	}
}

func TestLogsView_Render_Empty(t *testing.T) {
	lv := NewLogsView(1000)
	lv.Width = 80
	output := lv.Render()

	if !strings.Contains(output, "System Logs") {
		t.Error("Render should contain 'System Logs' header")
	}
	if !strings.Contains(output, "No log entries yet") {
		t.Error("Render should show 'No log entries yet' when empty")
	}
}

func TestLogsView_Render_Narrow(t *testing.T) {
	lv := NewLogsView(1000)
	lv.Width = 30
	output := lv.Render()

	if output != "Terminal too narrow for logs" {
		t.Errorf("expected narrow message, got %q", output)
	}
}

func TestLogsView_Render_WithEntries(t *testing.T) {
	lv := NewLogsView(1000)
	lv.Width = 80
	lv.Height = 24
	lv.AddEntry("INFO", "test info")
	lv.AddEntry("ERROR", "test error")

	output := lv.Render()

	if !strings.Contains(output, "INFO") {
		t.Error("Render should contain INFO level")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("Render should contain ERROR level")
	}
	if !strings.Contains(output, "test info") {
		t.Error("Render should contain message text")
	}
}

func TestLogsView_Render_LongMessage(t *testing.T) {
	lv := NewLogsView(1000)
	lv.Width = 60
	lv.Height = 24
	longMsg := strings.Repeat("x", 200)
	lv.AddEntry("INFO", longMsg)

	output := lv.Render()
	if strings.Contains(output, longMsg) {
		t.Error("Long message should be truncated")
	}
	if strings.Contains(output, "xxx...") || strings.Contains(output, "xxx") {
		// Truncation indicator present
	}
}
