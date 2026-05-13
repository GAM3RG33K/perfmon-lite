package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/w1n/perfmon/internal/engine"
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

// tab names displayed in the UI
var tabNames = []string{
	"Dashboard",
	"Threads/Procs",
	"System Logs",
}

// Model is the root Bubble Tea model for the perfmon TUI.
type Model struct {
	// Engine
	Engine *engine.Engine
	Mock   bool

	// Tabs
	ActiveTab int

	// Views
	Dashboard      *views.DashboardView
	TargetSelector *views.TargetSelectorView
	Logs           *views.LogsView

	// Layout
	Width  int
	Height int

	// App state
	Ready    bool
	Quitting bool
	Err      error
}

// NewModel creates a new TUI model.
func NewModel(eng *engine.Engine, mock bool) *Model {
	return &Model{
		Engine:         eng,
		Mock:           mock,
		ActiveTab:      TabDashboard,
		Dashboard:      views.NewDashboardView(),
		TargetSelector: views.NewTargetSelectorView(),
		Logs:           views.NewLogsView(1000),
		Width:          80,
		Height:         24,
	}
}

// Init initializes the model and returns the initial commands.
func (m *Model) Init() tea.Cmd {
	m.Logs.AddEntry("INFO", "perfmon starting...")

	if m.Mock {
		m.Logs.AddEntry("INFO", "Mock mode enabled — generating simulated telemetry")
	}

	return m.Engine.Start()
}

// Update handles all incoming messages and returns the updated model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Ready = true

		// Update view dimensions
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

	default:
		return m, nil
	}
}

// handleKeyMsg processes keyboard input.
func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {

	case "q", "ctrl+c":
		m.Quitting = true
		m.Logs.AddEntry("INFO", "Shutting down...")
		m.Engine.Close()
		return m, tea.Quit

	case "tab":
		m.ActiveTab = (m.ActiveTab + 1) % TabCount

	case "shift+tab":
		m.ActiveTab = (m.ActiveTab - 1 + TabCount) % TabCount

	case "left":
		m.ActiveTab = (m.ActiveTab - 1 + TabCount) % TabCount

	case "right":
		m.ActiveTab = (m.ActiveTab + 1) % TabCount

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

	case "enter":
		if m.ActiveTab == TabThreads && len(m.TargetSelector.Processes) > 0 {
			m.TargetSelector.ShowProcesses = true
		}

	case "r":
		m.Logs.AddEntry("INFO", "Refreshing...")

	case "e":
		m.Logs.AddEntry("INFO", "Export triggered — press 'e' in a future build")

	case "?":
		m.Logs.AddEntry("INFO", "Help: [TAB] Switch Tabs  [↑/↓] Navigate  [Enter] Select  [e] Export  [r] Refresh  [q] Quit")

	case "1":
		m.ActiveTab = TabDashboard
	case "2":
		m.ActiveTab = TabThreads
	case "3":
		m.ActiveTab = TabLogs
	}

	return m, nil
}

// handleTick processes a polling tick from the engine.
func (m *Model) handleTick() (tea.Model, tea.Cmd) {
	return m, tea.Batch(
		func() tea.Msg { return m.Engine.Poll() },
		m.Engine.Start(), // schedule next tick
	)
}

// handleTelemetry processes a telemetry message from the engine.
func (m *Model) handleTelemetry(msg engine.TelemetryMsg) (tea.Model, tea.Cmd) {
	if msg.Error != nil {
		m.Logs.AddEntry("ERROR", msg.Error.Error())
		return m, nil
	}

	return m, nil
}

// View renders the complete TUI.
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

	var b strings.Builder

	// Title bar
	title := fmt.Sprintf(" perfmon v1.0.0 %s", styles.PlatformBadge(engine.PlatformMock))
	b.WriteString(styles.TitleStyle.Render(title))
	b.WriteString("\n")

	// Target info bar
	if m.Mock {
		dev := m.TargetSelector.Devices
		proc := m.TargetSelector.Processes
		if len(dev) > 0 && len(proc) > 0 {
			info := fmt.Sprintf(" Target: %s  │  App: %s %s",
				dev[0].Name,
				proc[0].PackageName,
				styles.BuildBadge(proc[0].BuildType),
			)
			b.WriteString(styles.LabelStyle.Render(info))
			b.WriteString("\n")
		}
	}

	// Separator
	separator := strings.Repeat("─", m.Width-2)
	b.WriteString(styles.LabelStyle.Render(separator))
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

// renderTabs draws the tab bar dynamically sized to the terminal width.
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

// renderFooter draws the command footer dynamically sized to the terminal width.
func (m *Model) renderFooter() string {
	hints := []string{
		"[↑/↓] Navigate",
		"[←/→] Tabs",
		"[TAB] Switch",
		"[Enter] Select",
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

// SetTargets populates the device and process data into the TUI views.
// Used both in mock mode and when connected to a real device.
func (m *Model) SetTargets(devices []engine.Device, processes []engine.AppProcess) {
	m.TargetSelector.Devices = devices
	m.TargetSelector.Processes = processes
}
