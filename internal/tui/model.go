package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/w1n/perfmon/internal/engine"
	"github.com/w1n/perfmon/internal/export"
	"github.com/w1n/perfmon/internal/tui/styles"
	"github.com/w1n/perfmon/internal/tui/views"
)

const (
	TabDashboard = iota
	TabLogs
	TabCount
)

var tabNames = []string{"Dashboard", "System Logs"}

type Model struct {
	Engine *engine.Engine
	Mock   bool

	ActiveTab int

	Dashboard      *views.DashboardView
	TargetSelector *views.TargetSelectorView
	Logs           *views.LogsView

	Platform   engine.Platform
	Width      int
	Height     int
	Ready      bool
	Quitting   bool
	Err        error
	ShowHelp   bool
	statusMsg  string
	statusTime time.Time

	showFormatPicker bool
	formatPickerIdx  int

	AppPID   int32
	AppID    string // target app identifier from --id flag (for display when not running)
	Verbose  bool
}

func NewModel(eng *engine.Engine, mock bool, platform engine.Platform) *Model {
	return &Model{
		Engine:         eng,
		Mock:           mock,
		ActiveTab:      TabDashboard,
		Dashboard:      views.NewDashboardView(),
		TargetSelector: views.NewTargetSelectorView(),
		Logs:           views.NewLogsView(1000),
		Width:          80,
		Height:         24,
		Platform:       platform,
	}
}

func (m *Model) Init() tea.Cmd {
	m.Logs.AddEntry("INFO", "perfmon starting...")
	if m.Mock {
		m.Logs.AddEntry("INFO", "Mock mode enabled — generating simulated telemetry")
	}
	return tea.Batch(
		m.Engine.Start(),
		logCaptureCmd(m.Engine),
	)
}

type ClearStatusMsg struct{}

type LogCaptureMsg struct {
	Lines []string
}

func logCaptureCmd(eng *engine.Engine) tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		if eng == nil || eng.Provider == nil {
			return nil
		}
		lc, ok := eng.Provider.(engine.LogCapturer)
		if !ok {
			return nil
		}
		pid := eng.PID
		lines, err := lc.CaptureLogs(pid)
		if err != nil || len(lines) == 0 {
			return nil
		}
		return LogCaptureMsg{Lines: lines}
	})
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Ready = true
		m.Dashboard.Width = msg.Width - 4
		m.Dashboard.Height = msg.Height - 10
		m.Logs.Width = msg.Width - 4
		m.Logs.Height = msg.Height - 10
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case engine.TickMsg:
		return m.handleTick()

	case engine.TelemetryMsg:
		return m.handleTelemetry(msg)

	case LogCaptureMsg:
		for _, line := range msg.Lines {
			m.Logs.AddEntry("APP", line)
		}
		return m, logCaptureCmd(m.Engine)

	case ClearStatusMsg:
		m.statusMsg = ""
		return m, nil

	default:
		return m, nil
	}
}

func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.ShowHelp {
		switch msg.String() {
		case "?", "q", "esc", "enter", "ctrl+c":
			m.ShowHelp = false
		}
		return m, nil
	}

	if m.showFormatPicker {
		return m.handlePickerKey(msg)
	}

	switch msg.String() {
	case "q", "ctrl+c":
		m.Quitting = true
		m.Logs.AddEntry("INFO", "Shutting down...")
		if path, err := m.exportLogs(); err == nil {
			m.Logs.AddEntry("INFO", fmt.Sprintf("Logs saved to %s", path))
		}
		m.Engine.Close()
		return m, tea.Quit

	case "tab", "shift+tab", "left", "right":
		m.ActiveTab = (m.ActiveTab + 1) % TabCount
		return m, nil

	case "up":
		if m.ActiveTab == TabLogs {
			m.Logs.ScrollUp()
		}
		return m, nil

	case "down":
		if m.ActiveTab == TabLogs {
			m.Logs.ScrollDown()
		}
		return m, nil

	case "r":
		m.setStatus("⟳ Refreshing...")
		m.Logs.AddEntry("INFO", "Refreshing...")
		return m, nil

	case "e":
		if m.Engine.Buffer.Count() == 0 {
			m.setStatus(" No data to export yet")
			return m, nil
		}
		m.showFormatPicker = true
		m.formatPickerIdx = 0
		return m, nil

	case "E":
		return m.runExport(export.FormatMD)

	case "ctrl+e":
		return m.runExport(export.FormatHTML)

	case "?":
		m.ShowHelp = true
		return m, nil

	case "1":
		m.ActiveTab = TabDashboard
		return m, nil
	case "2":
		m.ActiveTab = TabLogs
		return m, nil
	}
	return m, nil
}

func (m *Model) handlePickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "left":
		m.formatPickerIdx = (m.formatPickerIdx - 1 + len(pickerFormats)) % len(pickerFormats)
		return m, nil
	case "down", "right":
		m.formatPickerIdx = (m.formatPickerIdx + 1) % len(pickerFormats)
		return m, nil
	case "enter":
		m.showFormatPicker = false
		return m.runExport(pickerFormats[m.formatPickerIdx])
	case "esc", "q", "ctrl+c":
		m.showFormatPicker = false
		m.setStatus(" Export cancelled")
		return m, nil
	}
	return m, nil
}

var pickerFormats = []export.Format{export.FormatJSON, export.FormatMD, export.FormatHTML}
var pickerLabels = []string{"JSON", "Markdown", "HTML"}

func (m *Model) handleTick() (tea.Model, tea.Cmd) {
	return m, tea.Batch(
		func() tea.Msg { return m.Engine.Poll() },
		m.Engine.Start(),
	)
}

func (m *Model) handleTelemetry(msg engine.TelemetryMsg) (tea.Model, tea.Cmd) {
	if msg.Error != nil {
		m.Logs.AddEntry("ERROR", msg.Error.Error())
		return m, nil
	}

	s := msg.Snapshot
	// Log each tick's CPU/RAM to system logs
	m.Logs.AddEntry("TICK",
		fmt.Sprintf("CPU=%.1f%% Mem=%dKB Threads=%d", s.CPUPercent, s.MemoryKB, s.Threads))

	// High CPU alert
	if s.CPUPercent > 70 {
		m.Logs.AddEntry("ALERT",
			fmt.Sprintf("High CPU: %.1f%% (threshold: 70%%)", s.CPUPercent))
		if s.Stack != "" {
			m.Logs.AddEntry("STACK",
				fmt.Sprintf("Stack for PID %d:\n%s", m.AppPID, s.Stack))
		}
	}

	// High memory alert
	memMB := float64(s.MemoryKB) / 1024
	if memMB > 500 {
		m.Logs.AddEntry("ALERT",
			fmt.Sprintf("High RAM: %.0f MB (threshold: 500 MB)", memMB))
	}

	if m.statusMsg != "" && time.Since(m.statusTime) > 3*time.Second {
		m.statusMsg = ""
	}
	return m, nil
}

func (m *Model) View() string {
	if !m.Ready {
		return "\n  Initializing perfmon..."
	}
	if m.Quitting {
		return "\n  Goodbye!\n"
	}
	if m.Err != nil {
		return fmt.Sprintf("\n  Error: %v\n", m.Err)
	}
	if m.ShowHelp {
		return m.renderHelp()
	}
	if m.showFormatPicker {
		return m.renderFormatPicker()
	}
	return m.renderMainView()
}

func (m *Model) selectedProcess() *engine.AppProcess {
	if m.AppPID > 0 {
		for i := range m.TargetSelector.Processes {
			if m.TargetSelector.Processes[i].PID == m.AppPID {
				return &m.TargetSelector.Processes[i]
			}
		}
	}
	if len(m.TargetSelector.Processes) > 0 {
		return &m.TargetSelector.Processes[0]
	}
	if m.AppID != "" {
		return &engine.AppProcess{
			PackageName: m.AppID,
			BuildType:   engine.BuildUnknown,
			Name:        m.AppID,
		}
	}
	return nil
}

func (m *Model) setStatus(msg string) {
	m.statusMsg = msg
	m.statusTime = time.Now()
}

func (m *Model) runExport(formatType export.Format) (tea.Model, tea.Cmd) {
	path, err := m.exportCurrentData(formatType)
	if err != nil {
		m.Logs.AddEntry("ERROR", fmt.Sprintf("Export failed: %v", err))
		m.setStatus(styles.ErrorStyle.Render("✗ Export failed"))
	} else {
		m.Logs.AddEntry("INFO", fmt.Sprintf("Exported to %s", path))
		shortPath := path
		if len(shortPath) > 40 {
			shortPath = "..." + shortPath[len(shortPath)-37:]
		}
		m.setStatus(fmt.Sprintf("✓ Exported to %s", shortPath))
	}
	return m, nil
}

// ─── Main view with bottom log panel ────────────────────────

func (m *Model) renderMainView() string {
	var b strings.Builder

	// Title
	title := fmt.Sprintf(" perfmon v0.0.1 %s", styles.PlatformBadge(m.Platform))
	b.WriteString(styles.TitleStyle.Render(title))
	b.WriteString("\n")

	// Target info
	devices := m.TargetSelector.Devices
	app := m.selectedProcess()
	if len(devices) > 0 && app != nil {
		info := fmt.Sprintf(" Target: %s  │  App: %s %s", devices[0].Name, app.PackageName, styles.BuildBadge(app.BuildType))
		b.WriteString(styles.LabelStyle.Render(info))
		b.WriteString("\n")
	}

	// Status bar
	if m.statusMsg != "" {
		b.WriteString(styles.StatusBarStyle.Render(m.statusMsg))
		b.WriteString("\n")
	}

	// Separator
	b.WriteString(styles.LabelStyle.Render(strings.Repeat("─", m.Width-2)))
	b.WriteString("\n")

	// Tabs
	b.WriteString(m.renderTabs())
	b.WriteString("\n")

	// Main body area — dashboard on top, logs at bottom
	bodyWidth := m.Width - 4
	if bodyWidth < 10 {
		bodyWidth = 10
	}

	// Dashboard section (top ~60% of available height)
	var body string
	switch m.ActiveTab {
	case TabDashboard:
		snapshots := m.Engine.Buffer.GetAll()
		latest := m.Engine.Buffer.Latest()
		body = m.Dashboard.Render(snapshots, latest)
	case TabLogs:
		body = m.Logs.Render()
	}

	b.WriteString(styles.PanelBorder.Width(bodyWidth).Render(body))
	b.WriteString("\n")

	// Bottom log panel (always visible as a compact console)
	b.WriteString(styles.LabelStyle.Render(strings.Repeat("─", m.Width-2)))
	b.WriteString("\n")
	logContent := m.Logs.RenderRecent(5)
	b.WriteString(styles.PanelBorder.Width(bodyWidth).Render(logContent))
	b.WriteString("\n")

	// Footer
	b.WriteString(m.renderFooter())

	return b.String()
}

func (m *Model) renderTabs() string {
	var tabs []string
	for i, name := range tabNames {
		if i == m.ActiveTab {
			tabs = append(tabs, styles.ActiveTabBorder.Render(name))
		} else {
			tabs = append(tabs, styles.InactiveTabBorder.Render(name))
		}
	}
	tabLine := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	return styles.PanelBorder.Width(m.Width - 4).Render(tabLine)
}

func (m *Model) renderFooter() string {
	hints := []string{
		"[↑/↓] Scroll", "[TAB] Switch", "[e] Export",
		"[r] Refresh", "[?] Help", "[q] Quit",
	}
	footerWidth := m.Width - 2
	if footerWidth < 20 {
		footerWidth = 20
	}
	return styles.FooterStyle.Width(footerWidth).Render(strings.Join(hints, "  "))
}

func (m *Model) renderHelp() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(styles.HelpTitle.Render("  perfmon — Help"))
	b.WriteString("\n\n")

	sections := []struct {
		title string
		keys  [][2]string
	}{
		{"Navigation", [][2]string{
			{"↑/↓", "Scroll logs / navigate lists"},
			{"←/→ / Tab", "Switch tabs"},
			{"1-2", "Jump to tab"},
		}},
		{"Actions", [][2]string{
			{"e", "Open export format picker"},
			{"Shift+E", "Export directly to Markdown"},
			{"Ctrl+E", "Export directly to HTML"},
			{"r", "Refresh device list"},
		}},
		{"General", [][2]string{
			{"?", "Toggle this help"},
			{"q / Ctrl+C", "Quit"},
		}},
	}

	for _, section := range sections {
		b.WriteString(styles.HelpSectionTitle.Render("  " + section.title))
		b.WriteString("\n")
		for _, pair := range section.keys {
			key := styles.HelpKey.Render("    " + pair[0])
			desc := styles.HelpDesc.Render(pair[1])
			b.WriteString(fmt.Sprintf("%s  %s\n", key, desc))
		}
		b.WriteString("\n")
	}

	b.WriteString(styles.HelpFooter.Render("  Press [?], Esc, Enter, or q to close help"))
	return b.String()
}

func (m *Model) renderFormatPicker() string {
	baseView := m.renderMainView()
	var modal strings.Builder
	modal.WriteString("\n")
	modal.WriteString(styles.HighlightStyle.Render("  Select export format"))
	modal.WriteString("\n\n")
	for i, label := range pickerLabels {
		if i == m.formatPickerIdx {
			modal.WriteString(styles.HighlightStyle.Render("  ▸ " + label))
		} else {
			modal.WriteString(styles.LabelStyle.Render("    " + label))
		}
		modal.WriteString("\n")
	}
	modal.WriteString("\n")
	modal.WriteString(styles.HelpFooter.Render("  ↑/↓  Enter  Esc"))

	lines := strings.Split(baseView, "\n")
	modalLines := strings.Split(strings.TrimRight(modal.String(), "\n"), "\n")
	startY := len(lines)/2 - len(modalLines)/2
	if startY < 0 {
		startY = 0
	}
	var result strings.Builder
	for i, line := range lines {
		if i >= startY && i < startY+len(modalLines) {
			result.WriteString(modalLines[i-startY])
		} else {
			result.WriteString(line)
		}
		result.WriteString("\n")
	}
	return result.String()
}

// exportCurrentData exports the current ring buffer to the given format.
func (m *Model) exportCurrentData(formatType export.Format) (string, error) {
	snapshots := m.Engine.Buffer.GetAll()
	if len(snapshots) == 0 {
		return "", fmt.Errorf("no telemetry data to export")
	}
	deviceName := "unknown"
	if len(m.TargetSelector.Devices) > 0 {
		deviceName = m.TargetSelector.Devices[0].Name
	}
	appName := "unknown"
	buildType := engine.BuildUnknown
	if app := m.selectedProcess(); app != nil {
		appName = app.PackageName
		buildType = app.BuildType
	}
	var logLines []string
	for _, e := range m.Logs.Entries {
		logLines = append(logLines, fmt.Sprintf("[%s] %s", e.Level, e.Message))
	}
	opts := export.Options{
		Format:     formatType,
		OutputPath: "",
		Version:    "1.0.0",
		Platform:   m.Platform,
		DeviceName: deviceName,
		AppName:    appName,
		BuildType:  buildType,
		Logs:       logLines,
	}
	opts.OutputPath = export.ResolveOutputPath(opts, snapshots)
	if err := export.EnsureOutputDir(opts.OutputPath); err != nil {
		return "", fmt.Errorf("creating output directory: %w", err)
	}
	path, err := export.Export(snapshots, opts)
	if err != nil {
		return "", fmt.Errorf("export: %w", err)
	}
	return path, nil
}

// exportLogs writes captured logs to a .log file in the current directory.
func (m *Model) exportLogs() (string, error) {
	path := fmt.Sprintf("perfmon_logs_%d.log", time.Now().Unix())
	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("creating log file: %w", err)
	}
	defer f.Close()

	for _, e := range m.Logs.Entries {
		line := fmt.Sprintf("[%s] [%s] %s\n", e.Timestamp.Format("15:04:05"), e.Level, e.Message)
		if _, err := f.WriteString(line); err != nil {
			return "", fmt.Errorf("writing log file: %w", err)
		}
	}
	return path, nil
}

func (m *Model) SetTargets(devices []engine.Device, processes []engine.AppProcess) {
	m.TargetSelector.Devices = devices
	m.TargetSelector.Processes = processes
}
