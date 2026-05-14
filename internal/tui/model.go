package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/w1n/perfmon/internal/engine"
	"github.com/w1n/perfmon/internal/export"
	"github.com/w1n/perfmon/internal/tui/styles"
	"github.com/w1n/perfmon/internal/tui/views"
)

// Tab identifiers
const (
	TabDashboard = iota
	TabThreads
	TabLogs
	TabCount
)

var tabNames = []string{"Dashboard", "Threads/Procs", "System Logs"}

// Model is the root Bubble Tea model for the perfmon TUI.
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

	// Export format picker state
	showFormatPicker bool
	formatPickerIdx  int
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
	return m.Engine.Start()
}

// ClearStatusMsg is sent after a short delay to clear the status bar.
type ClearStatusMsg struct{}

func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return ClearStatusMsg{}
	})
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Ready = true
		m.Dashboard.Width = msg.Width - 4
		m.Dashboard.Height = msg.Height - 8
		m.TargetSelector.Width = msg.Width - 4
		m.TargetSelector.Height = msg.Height - 8
		m.Logs.Width = msg.Width - 4
		m.Logs.Height = msg.Height - 8
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case engine.TickMsg:
		return m.handleTick()

	case engine.TelemetryMsg:
		return m.handleTelemetry(msg)

	case ClearStatusMsg:
		m.statusMsg = ""
		return m, nil

	default:
		return m, nil
	}
}

func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If help is open, only handle escape/enter to close
	if m.ShowHelp {
		switch msg.String() {
		case "?", "q", "esc", "enter", "ctrl+c":
			m.ShowHelp = false
		}
		return m, nil
	}

	// If format picker is open, handle navigation
	if m.showFormatPicker {
		return m.handlePickerKey(msg)
	}

	switch msg.String() {

	case "q", "ctrl+c":
		m.Quitting = true
		m.Logs.AddEntry("INFO", "Shutting down...")
		m.Engine.Close()
		return m, tea.Quit

	case "tab":
		m.ActiveTab = (m.ActiveTab + 1) % TabCount
		return m, nil

	case "shift+tab":
		m.ActiveTab = (m.ActiveTab - 1 + TabCount) % TabCount
		return m, nil

	case "left":
		m.ActiveTab = (m.ActiveTab - 1 + TabCount) % TabCount
		return m, nil

	case "right":
		m.ActiveTab = (m.ActiveTab + 1) % TabCount
		return m, nil

	case "up":
		switch m.ActiveTab {
		case TabLogs:
			m.Logs.ScrollUp()
		case TabThreads:
			if m.TargetSelector.ShowProcesses && m.TargetSelector.SelectedProcess > 0 {
				m.TargetSelector.SelectedProcess--
			} else if !m.TargetSelector.ShowProcesses && m.TargetSelector.SelectedDevice > 0 {
				m.TargetSelector.SelectedDevice--
			}
		}
		return m, nil

	case "down":
		switch m.ActiveTab {
		case TabLogs:
			m.Logs.ScrollDown()
		case TabThreads:
			if m.TargetSelector.ShowProcesses && m.TargetSelector.SelectedProcess < len(m.TargetSelector.Processes)-1 {
				m.TargetSelector.SelectedProcess++
			} else if !m.TargetSelector.ShowProcesses && m.TargetSelector.SelectedDevice < len(m.TargetSelector.Devices)-1 {
				m.TargetSelector.SelectedDevice++
			}
		}
		return m, nil

	case "enter":
		if m.ActiveTab == TabThreads && len(m.TargetSelector.Processes) > 0 {
			m.TargetSelector.ShowProcesses = true
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
		m.ActiveTab = TabThreads
		return m, nil
	case "3":
		m.ActiveTab = TabLogs
		return m, nil
	}

	return m, nil
}

// handlePickerKey processes keys during export format selection.
var pickerFormats = []export.Format{export.FormatJSON, export.FormatMD, export.FormatHTML}
var pickerLabels = []string{"JSON", "Markdown", "HTML"}

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

// runExport exports data and shows result in the status bar.
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

// setStatus sets a status message that auto-clears after 3 seconds.
func (m *Model) setStatus(msg string) {
	m.statusMsg = msg
	m.statusTime = time.Now()
}

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
	// Auto-clear status after 3 seconds
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

	var b strings.Builder

	// Title bar
	title := fmt.Sprintf(" perfmon v1.0.0 %s", styles.PlatformBadge(m.Platform))
	b.WriteString(styles.TitleStyle.Render(title))
	b.WriteString("\n")

	// Target info bar
	devices := m.TargetSelector.Devices
	processes := m.TargetSelector.Processes
	if len(devices) > 0 && len(processes) > 0 {
		info := fmt.Sprintf(" Target: %s  │  App: %s %s",
			devices[0].Name,
			processes[0].PackageName,
			styles.BuildBadge(processes[0].BuildType),
		)
		b.WriteString(styles.LabelStyle.Render(info))
		b.WriteString("\n")
	}

	// Status bar (brief notifications)
	if m.statusMsg != "" {
		statusLine := styles.StatusBarStyle.Render(m.statusMsg)
		b.WriteString(statusLine)
		b.WriteString("\n")
	}

	// Separator
	sep := strings.Repeat("─", m.Width-2)
	b.WriteString(styles.LabelStyle.Render(sep))
	b.WriteString("\n")

	// Tabs
	b.WriteString(m.renderTabs())
	b.WriteString("\n")

	// Body
	var body string
	bodyWidth := m.Width - 4
	if bodyWidth < 10 {
		bodyWidth = 10
	}
	switch m.ActiveTab {
	case TabDashboard:
		snapshots := m.Engine.Buffer.GetAll()
		latest := m.Engine.Buffer.Latest()
		body = m.Dashboard.Render(snapshots, latest)
	case TabThreads:
		body = m.TargetSelector.Render()
	case TabLogs:
		body = m.Logs.Render()
	}
	b.WriteString(styles.PanelBorder.Width(bodyWidth).Render(body))
	b.WriteString("\n")

	// Footer
	b.WriteString(m.renderFooter())

	return b.String()
}

// renderFormatPicker shows a modal overlay for selecting export format.
func (m *Model) renderFormatPicker() string {
	var b strings.Builder
	b.WriteString("\n\n\n")
	b.WriteString(styles.HelpTitle.Render("  Select export format"))
	b.WriteString("\n\n")

	for i, label := range pickerLabels {
		prefix := "  "
		cursor := " "
		if i == m.formatPickerIdx {
			prefix = "▸ "
			cursor = "▸"
			_ = cursor
		}
		line := prefix + label
		if i == m.formatPickerIdx {
			b.WriteString(styles.HighlightStyle.Render(line))
		} else {
			b.WriteString(styles.LabelStyle.Render(line))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpFooter.Render("  ↑/↓ navigate  Enter select  Esc cancel"))
	return b.String()
}

// renderHelp shows a full-screen help overlay.
func (m *Model) renderHelp() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(styles.HelpTitle.Render("  perfmon — Help"))
	b.WriteString("\n\n")

	sections := []struct {
		title string
		keys  [][2]string
	}{
		{
			"Navigation",
			 [][2]string{
				{"↑/↓", "Navigate lists"},
				{"←/→", "Switch tabs"},
				{"Tab", "Cycle forward"},
				{"Shift+Tab", "Cycle backward"},
				{"1-3", "Jump to tab"},
			},
		},
		{
			"Actions",
			[][2]string{
				{"Enter", "Select item / show processes"},
				{"e", "Export to JSON"},
				{"Shift+E", "Export to Markdown"},
				{"Ctrl+E", "Export to HTML"},
				{"r", "Refresh device list"},
			},
		},
		{
			"General",
			[][2]string{
				{"?", "Toggle this help"},
				{"q / Ctrl+C", "Quit"},
			},
		},
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
		"[↑/↓] Navigate",
		"[←/→] Tabs",
		"[TAB] Switch",
		"[e] Export",
		"[r] Refresh",
		"[?] Help",
		"[q] Quit",
	}
	footerWidth := m.Width - 2
	if footerWidth < 20 {
		footerWidth = 20
	}
	return styles.FooterStyle.Width(footerWidth).Render(strings.Join(hints, "  "))
}

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
	if len(m.TargetSelector.Processes) > 0 {
		appName = m.TargetSelector.Processes[0].PackageName
		buildType = m.TargetSelector.Processes[0].BuildType
	}

	opts := export.Options{
		Format:     formatType,
		OutputPath: "",
		Version:    "1.0.0",
		Platform:   m.Platform,
		DeviceName: deviceName,
		AppName:    appName,
		BuildType:  buildType,
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

func (m *Model) SetTargets(devices []engine.Device, processes []engine.AppProcess) {
	m.TargetSelector.Devices = devices
	m.TargetSelector.Processes = processes
}
