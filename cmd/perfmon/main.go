package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/w1n/perfmon/internal/engine"
	"github.com/w1n/perfmon/internal/export"
	"github.com/w1n/perfmon/internal/platform/android"
	iosPkg "github.com/w1n/perfmon/internal/platform/ios"
	"github.com/w1n/perfmon/internal/platform/mock"
	perfmonTui "github.com/w1n/perfmon/internal/tui"
)

// Exit codes (see docs/cli-reference.md §6)
const (
	exitSuccess       = 0
	exitGeneralError  = 1
	exitDeviceError   = 2
	exitToolError     = 3
	exitExportError   = 4
)

var version = "1.0.0"

func main() {
	// ── CLI flags ───────────────────────────────────────────────────────
	mockMode := flag.Bool("mock", false, "Run with simulated telemetry data (no device required)")
	deviceFlag := flag.String("device", "", "Specify target device (serial/UUID) or package name")
	flag.StringVar(deviceFlag, "d", "", "Shorthand for --device")
	intervalFlag := flag.Int("interval", envInt("PERFMON_POLL_INTERVAL", 1), "Polling interval in seconds (range: 1-60)")
	bufferFlag := flag.Int("buffer", envInt("PERFMON_BUFFER_SIZE", 300), "Ring buffer capacity (number of data points)")
	exportFlag := flag.String("export", "", "Export format: json, md, html")
	outputFlag := flag.String("output", envStr("PERFMON_EXPORT_DIR", "./perfmon_export"), "Output path for export file (without extension)")
	verboseFlag := flag.Bool("verbose", false, "Enable verbose logging")
	showVersion := flag.Bool("version", false, "Show version information")
	showHelp := flag.Bool("help", false, "Show help message")

	// Custom usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "perfmon v%s — Mobile Performance Monitor & Profiler\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  perfmon [flags]\n")
		fmt.Fprintf(os.Stderr, "  perfmon devices [flags]\n")
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Handle subcommands
	args := flag.Args()
	if len(args) > 0 && args[0] == "devices" {
		runDevices(args[1:], *verboseFlag)
		os.Exit(exitSuccess)
	}

	// Handle version flag
	if *showVersion {
		fmt.Printf("perfmon v%s\n", version)
		os.Exit(exitSuccess)
	}

	// Handle help flag
	if *showHelp {
		flag.Usage()
		os.Exit(exitSuccess)
	}

	// ══════════════════════════════════════════════════════════════════════
	// Provider Setup
	// ══════════════════════════════════════════════════════════════════════

	var provider engine.TelemetryProvider
	var discoveredDevices []engine.Device
	var discoveredProcesses []engine.AppProcess
	var initialPID int32
	var targetPlatform engine.Platform

	if *mockMode {
		// ── Mock mode ────────────────────────────────────────────────────
		mockProvider := mock.NewProvider(time.Now().UnixNano())
		provider = mockProvider
		initialPID = 9001
		targetPlatform = engine.PlatformMock
		discoveredDevices = []engine.Device{mock.MockDevice()}
		discoveredProcesses = []engine.AppProcess{mock.MockProcess()}

		if *verboseFlag {
			log.Printf("Starting perfmon v%s (mock mode: interval=%ds, buffer=%d)\n",
				version, *intervalFlag, *bufferFlag)
		}
	} else {
		// ── Auto-detect: discover all platforms ───────────────────────
		provider, discoveredDevices, discoveredProcesses, initialPID, targetPlatform = autoDetectProvider(*deviceFlag, *verboseFlag)
	}

	if provider == nil {
		fmt.Fprintf(os.Stderr, "No platform provider could be configured.\n")
		fmt.Fprintf(os.Stderr, "Use --mock for development:  perfmon --mock\n")
		os.Exit(exitToolError)
	}

	// ══════════════════════════════════════════════════════════════════════
	// Engine Setup
	// ══════════════════════════════════════════════════════════════════════

	eng := engine.NewEngine(provider, *bufferFlag, time.Duration(*intervalFlag)*time.Second)
	eng.SetTarget(initialPID)

	// ══════════════════════════════════════════════════════════════════════
	// Export-only mode (non-interactive)
	// ══════════════════════════════════════════════════════════════════════

	if *exportFlag != "" {
		fmt.Printf("Sampling telemetry...\n")
		sampleCount := 10
		for i := 0; i < sampleCount; i++ {
			msg := eng.Poll()
			if tm, ok := msg.(engine.TelemetryMsg); ok && tm.Error == nil {
				fmt.Printf("  Sample %d: CPU=%.1f%% Memory=%dKB Threads=%d\n",
					i+1, tm.Snapshot.CPUPercent, tm.Snapshot.MemoryKB, tm.Snapshot.Threads)
			}
			time.Sleep(time.Duration(*intervalFlag) * time.Second / 2)
		}

		snapshots := eng.Buffer.GetAll()
		if len(snapshots) == 0 {
			fmt.Println("No telemetry data collected — exiting.")
			os.Exit(exitGeneralError)
		}

		deviceName := "unknown"
		if len(discoveredDevices) > 0 {
			deviceName = discoveredDevices[0].Name
		}
		appName := "unknown"
		buildType := engine.BuildUnknown
		if initialPID > 0 {
			for _, p := range discoveredProcesses {
				if p.PID == initialPID {
					appName = p.PackageName
					buildType = p.BuildType
					break
				}
			}
		}
		if appName == "unknown" && len(discoveredProcesses) > 0 {
			appName = discoveredProcesses[0].PackageName
		}

		var format export.Format
		switch *exportFlag {
		case "json":
			format = export.FormatJSON
		case "md", "markdown":
			format = export.FormatMD
		case "html":
			format = export.FormatHTML
		default:
			fmt.Fprintf(os.Stderr, "Unsupported export format: %s (supported: json, md, html)\n", *exportFlag)
			os.Exit(exitGeneralError)
		}

		opts := export.Options{
			Format:     format,
			OutputPath: *outputFlag,
			Version:    version,
			Platform:   targetPlatform,
			DeviceName: deviceName,
			AppName:    appName,
			BuildType:  buildType,
		}

		if err := export.EnsureOutputDir(opts.OutputPath); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create output directory: %v\n", err)
			os.Exit(exitExportError)
		}

		fmt.Printf("Exporting %d data points to %s format...\n", len(snapshots), *exportFlag)
		path, err := export.Export(snapshots, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Export failed: %v\n", err)
			os.Exit(exitExportError)
		}

		fmt.Printf("Report written to: %s\n", path)
		os.Exit(exitSuccess)
	}

	// ══════════════════════════════════════════════════════════════════════
	// Interactive TUI Mode
	// ══════════════════════════════════════════════════════════════════════

	model := perfmonTui.NewModel(eng, *mockMode, targetPlatform)

	// Populate the TUI with discovered devices and processes
	model.SetTargets(discoveredDevices, discoveredProcesses)

	model.Logs.AddEntry("INFO", fmt.Sprintf("Target: %s | App: %s (PID %d)",
		targetPlatform, discoveredProcesses[0].PackageName, initialPID))
	model.Logs.AddEntry("INFO", fmt.Sprintf("Polling every %d second(s)", *intervalFlag))

	if *mockMode {
		model.Logs.AddEntry("INFO", "Mock mode — simulated telemetry data")
	} else if targetPlatform == engine.PlatformIOS {
		model.Logs.AddEntry("INFO", "iOS mode — live device telemetry")
	} else {
		model.Logs.AddEntry("INFO", "Android mode — live device telemetry")
	}

	// Start TUI with mouse support. If it fails (e.g., Windows CMD, SSH),
	// retry without mouse support.
	runTUI := func(mouse bool) error {
		opts := []tea.ProgramOption{tea.WithAltScreen()}
		if mouse {
			opts = append(opts, tea.WithMouseCellMotion())
		}
		p := tea.NewProgram(model, opts...)
		_, err := p.Run()
		return err
	}

	if err := runTUI(true); err != nil {
		if err := runTUI(false); err != nil {
			log.Fatalf("Error running TUI: %v", err)
		}
	}
}

// hintInstallXcode returns a help string for installing Xcode.
func hintInstallXcode() string {
	return `
  ── Xcode Installation Help ────────────────────────────
    macOS (App Store):  Install Xcode from the App Store
    macOS (CLI tools):  xcode-select --install
    Verify:             xcrun simctl list
  ────────────────────────────────────────────────────────
  Or run with --mock for development:  perfmon --mock
`
}

// tryAndroidProvider attempts to set up the Android provider.
// If adbPath is empty, it auto-discovers the adb binary.
func tryAndroidProvider(adbPath string, verbose bool) (engine.TelemetryProvider, []engine.Device, []engine.AppProcess, int32, engine.Platform, error) {
	if adbPath == "" {
		var err error
		adbPath, err = android.FindAdbPath()
		if err != nil {
			return nil, nil, nil, 0, "", err
		}
	}

	if verbose {
		log.Printf("ADB found at: %s", adbPath)
	}

	// Verify ADB version
	adbVer, err := android.CheckVersion(adbPath)
	if err != nil {
		return nil, nil, nil, 0, "", fmt.Errorf("ADB version check failed: %w", err)
	}
	if verbose {
		log.Printf("ADB version: %s", adbVer.String())
	}

	// Create the Android provider
	androidProvider := android.NewProvider(adbPath)

	// Discover connected devices
	devices, err := androidProvider.Discover()
	if err != nil {
		return nil, nil, nil, 0, "", fmt.Errorf("failed to discover Android devices: %w", err)
	}
	if len(devices) == 0 {
		return nil, nil, nil, 0, "", fmt.Errorf("no Android devices found")
	}

	if verbose {
		log.Printf("Found %d device(s):", len(devices))
		for _, d := range devices {
			typ := "emulator"
			if d.IsPhysical {
				typ = "physical"
			}
			log.Printf("  %s (%s) — %s", d.Name, d.ID, typ)
		}
	}

	// Auto-select the first device
	selectedDevice := devices[0]
	androidProvider.SetDevice(selectedDevice.ID)

	if verbose {
		log.Printf("Selected device: %s (%s)", selectedDevice.Name, selectedDevice.ID)
	}

	// Validate device reachability
	if err := android.ValidateDevice(adbPath, selectedDevice.ID); err != nil {
		return nil, nil, nil, 0, "", fmt.Errorf("device %s is not reachable: %w", selectedDevice.ID, err)
	}

	if verbose {
		log.Print("Device is reachable — discovering processes...")
	}

	// Discover processes on the device
	processes, err := androidProvider.MapProcesses(selectedDevice.ID)
	if err != nil {
		return nil, nil, nil, 0, "", fmt.Errorf("failed to list processes on %s: %w", selectedDevice.ID, err)
	}
	if len(processes) == 0 {
		return nil, nil, nil, 0, "", fmt.Errorf("no processes found on device %s", selectedDevice.ID)
	}

	if verbose {
		log.Printf("Found %d processes. Scanning for user applications...", len(processes))
	}

	selectedProcess := selectBestProcess(processes)
	initialPID := selectedProcess.PID

	if verbose {
		log.Printf("Selected process: %s (PID %d) [%s]",
			selectedProcess.PackageName, selectedProcess.PID, selectedProcess.BuildType)
	}

	// Detect build type for the selected process
	buildType, err := androidProvider.BuildType(selectedDevice.ID, selectedProcess.PackageName)
	if err == nil {
		selectedProcess.BuildType = buildType
		if verbose {
			log.Printf("Build type: %s", buildType)
		}
	}

	// Update the process in the list with the resolved build type
	for i := range processes {
		if processes[i].PID == selectedProcess.PID {
			processes[i].BuildType = buildType
			break
		}
	}

	return androidProvider, devices, processes, initialPID, engine.PlatformAndroid, nil
}

// autoDetectProvider discovers the best available platform (Android → iOS → mock).
// If a deviceID is provided, it selects that specific device.
func autoDetectProvider(deviceID string, verbose bool) (
	provider engine.TelemetryProvider,
	devices []engine.Device,
	processes []engine.AppProcess,
	pid int32,
	platform engine.Platform,
) {
	// Try Android first
	var androidErr error
	if aProv, aDevs, aProcs, aPID, aPlat, aErr := tryAndroidProvider("", verbose); aErr == nil {
		androidErr = nil
		if deviceID == "" || matchDevice(aDevs, deviceID) {
			provider, devices, processes, pid, platform = aProv, aDevs, aProcs, aPID, aPlat
			return
		}
	} else {
		androidErr = aErr
	}

	// Try iOS
	if iProv, iDevs, iProcs, iPID, iPlat, iErr := tryiOSProvider(verbose); iErr == nil {
		if deviceID == "" || matchDevice(iDevs, deviceID) {
			provider, devices, processes, pid, platform = iProv, iDevs, iProcs, iPID, iPlat
			return
		}
	} else if androidErr != nil && verbose {
		log.Printf("iOS also unavailable: %v", iErr)
	}

	// Device was specified but not found on any platform
	if deviceID != "" {
		fmt.Fprintf(os.Stderr, "Device %q not found on any platform.\n", deviceID)
		fmt.Fprintf(os.Stderr, "Use --mock for development:  perfmon --mock\n")
		os.Exit(exitDeviceError)
	}

	// Run the pre-flight wizard
	if verbose && androidErr != nil {
		log.Printf("Android not available: %v", androidErr)
	}

	wizardResult := runPreflightWizard()
	switch wizardResult {
	case "mock":
		mockProv := mock.NewProvider(time.Now().UnixNano())
		provider, pid, platform = mockProv, 9001, engine.PlatformMock
		devices = []engine.Device{mock.MockDevice()}
		processes = []engine.AppProcess{mock.MockProcess()}
		return
	case "retry":
		if aProv, aDevs, aProcs, aPID, aPlat, aErr := tryAndroidProvider("", verbose); aErr == nil {
			provider, devices, processes, pid, platform = aProv, aDevs, aProcs, aPID, aPlat
			return
		} else if verbose {
			log.Printf("Android still unavailable: %v", aErr)
		}
		// iOS fallback
		if iProv, iDevs, iProcs, iPID, iPlat, iErr := tryiOSProvider(verbose); iErr == nil {
			provider, devices, processes, pid, platform = iProv, iDevs, iProcs, iPID, iPlat
			return
		} else if verbose {
			log.Printf("iOS also unavailable: %v", iErr)
		}
	case "quit":
		os.Exit(exitSuccess)
	}

	fmt.Fprintf(os.Stderr, "No platform provider could be configured.\n")
	fmt.Fprintf(os.Stderr, "Use --mock for development:  perfmon --mock\n")
	os.Exit(exitToolError)
	return
}

// matchDevice checks if any device in the list matches the given ID or name.
func matchDevice(devices []engine.Device, id string) bool {
	for _, d := range devices {
		if d.ID == id || d.Name == id {
			return true
		}
	}
	return false
}

// tryiOSProvider attempts to set up the iOS provider.
// Returns an error if xcrun is unavailable or no devices/processes are found.
func tryiOSProvider(verbose bool) (engine.TelemetryProvider, []engine.Device, []engine.AppProcess, int32, engine.Platform, error) {
	if verbose {
		log.Print("Looking for xcrun...")
	}

	xcrunPath, err := iosPkg.FindXcrunPath()
	if err != nil {
		return nil, nil, nil, 0, "", fmt.Errorf("xcrun not found: %w", err)
	}

	if verbose {
		log.Printf("xcrun found at: %s", xcrunPath)
	}

	// Verify xcrun version
	xcrunVer, err := iosPkg.CheckVersion(xcrunPath)
	if err != nil {
		return nil, nil, nil, 0, "", fmt.Errorf("xcrun version check failed: %w", err)
	}
	if verbose {
		log.Printf("xcrun version: %s", xcrunVer.String())
	}

	// Check xcode-select
	if err := iosPkg.CheckXcodeSelect(); err != nil {
		return nil, nil, nil, 0, "", fmt.Errorf("Xcode not configured: %w", err)
	}

	// Create the iOS provider
	iOSProvider := iosPkg.NewProvider(xcrunPath)

	// Discover iOS devices and simulators
	devices, err := iOSProvider.Discover()
	if err != nil {
		return nil, nil, nil, 0, "", fmt.Errorf("failed to discover iOS devices: %w", err)
	}
	if len(devices) == 0 {
		return nil, nil, nil, 0, "", errors.New("no iOS devices or simulators found")
	}

	if verbose {
		log.Printf("Found %d iOS device(s):", len(devices))
		for _, d := range devices {
			typ := "simulator"
			if d.IsPhysical {
				typ = "physical"
			}
			log.Printf("  %s (%s) — %s", d.Name, d.ID, typ)
		}
	}

	// Cache devices for later lookups
	iOSProvider.CacheDevices(devices)

	// Auto-select the first device
	selectedDevice := devices[0]
	iOSProvider.SetDevice(selectedDevice.ID)

	if verbose {
		log.Printf("Selected device: %s (%s)", selectedDevice.Name, selectedDevice.ID)
	}

	// Discover processes
	processes, err := iOSProvider.MapProcesses(selectedDevice.ID)
	if err != nil {
		return nil, nil, nil, 0, "", fmt.Errorf("failed to list processes on %s: %w", selectedDevice.ID, err)
	}
	if len(processes) == 0 {
		return nil, nil, nil, 0, "", fmt.Errorf("no processes found on device %s", selectedDevice.ID)
	}

	if verbose {
		log.Printf("Found %d processes.", len(processes))
	}

	// Select the best process
	selectedProcess := selectBestProcess(processes)
	initialPID := selectedProcess.PID

	if verbose {
		log.Printf("Selected process: %s (PID %d) [%s]",
			selectedProcess.PackageName, selectedProcess.PID, selectedProcess.BuildType)
	}

	// Detect build type
	buildType, err := iOSProvider.BuildType(selectedDevice.ID, selectedProcess.PackageName)
	if err == nil {
		selectedProcess.BuildType = buildType
		if verbose {
			log.Printf("Build type: %s", buildType)
		}
	}

	// Update the process list with resolved build type
	for i := range processes {
		if processes[i].PID == selectedProcess.PID {
			processes[i].BuildType = buildType
			break
		}
	}

	return iOSProvider, devices, processes, initialPID, engine.PlatformIOS, nil
}

// selectBestProcess picks the most interesting process from a list of processes.
// Preference order:
//  1. A process whose name/package contains a dot (user app, not a system daemon)
//  2. The first non-kernel process
//  3. The first process overall
func selectBestProcess(processes []engine.AppProcess) engine.AppProcess {
	// Pass 1: real user apps — 3+ domain parts, not a known system prefix
	var userApps []engine.AppProcess
	for _, p := range processes {
		name := p.PackageName
		if strings.Count(name, ".") < 2 {
			continue // needs at least com.example.app (3 parts)
		}
		// Skip known system/reserved prefixes
		if hasAnyPrefix(name, []string{
			"android.", "com.android.", "com.google.", "com.google.android.",
			"com.apple.", "com.samsung.", "com.qualcomm.",
			"media.", "system.", "zygote",
			"UIKitApplication:com.apple.",
		}) {
			continue
		}
		userApps = append(userApps, p)
	}
	if len(userApps) > 0 {
		// Prefer non-com apps (e.g. in.thetatva.tatva over com.instagram.android)
		// since user's own apps often use custom domains
		for _, app := range userApps {
			if !strings.HasPrefix(app.PackageName, "com.") {
				return app
			}
		}
		return userApps[0]
	}

	// Pass 2: any process with 2+ domain parts
	for _, p := range processes {
		if strings.Count(p.PackageName, ".") >= 1 {
			return p
		}
	}

	// Fallback: first process overall
	if len(processes) > 0 {
		return processes[0]
	}

	return engine.AppProcess{}
}

func hasAnyPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

// runPreflightWizard displays an interactive setup wizard when ADB is not found.
// Returns the user's choice: "mock", "retry", "quit", or "" (fall through to iOS).
func runPreflightWizard() string {
	fmt.Print("\n")
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║           perfmon — Pre-flight Setup Wizard             ║")
	fmt.Println("╠══════════════════════════════════════════════════════════╣")
	fmt.Println("║  ADB (Android Debug Bridge) was not found.              ║")
	fmt.Println("║  ADB is required to profile Android devices.            ║")
	fmt.Println("║                                                        ║")
	fmt.Println("║  Common install methods:                               ║")
	if isCommandAvailable("brew") {
		fmt.Println("║    brew install android-platform-tools                  ║")
	}
	if isCommandAvailable("apt-get") {
		fmt.Println("║    sudo apt install adb                                 ║")
	}
	fmt.Println("║    https://developer.android.com/studio/releases/       ║")
	fmt.Println("║              platform-tools                              ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	for {
		fmt.Println("What would you like to do?")
		if isCommandAvailable("brew") {
			fmt.Println("  1) Install ADB via Homebrew (recommended)")
		}
		fmt.Println("  2) I've installed ADB — retry detection")
		fmt.Println("  3) Skip — use mock mode (no device needed)")
		fmt.Println("  4) Skip — try iOS mode (macOS only)")
		fmt.Println("  5) Quit")
		fmt.Print("Choice [1-5]: ")

		var choice string
		fmt.Scanln(&choice)

		switch choice {
		case "1":
			if isCommandAvailable("brew") {
				fmt.Println("\nInstalling ADB via Homebrew...")
				cmd := exec.Command("brew", "install", "android-platform-tools")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "Installation failed: %v\n\n", err)
					continue
				}
				fmt.Println("\n✓ ADB installed successfully! Retrying...")
				return "retry"
			}
			fmt.Println("Homebrew not available on this system.")
			continue
		case "2":
			fmt.Println("\nRetrying ADB detection...")
			return "retry"
		case "3":
			fmt.Println("\nStarting in mock mode...")
			return "mock"
		case "4":
			fmt.Println("\nSkipping to iOS mode...")
			return ""
		case "5":
			fmt.Println("Goodbye.")
			return "quit"
		default:
			fmt.Println("Invalid choice — please enter a number 1-5.")
		}
	}
}

// runDevices implements the `perfmon devices` subcommand.
// Flags: --json, --platform, --build-info
func runDevices(args []string, verbose bool) {
	var jsonOut, buildInfo bool
	platformFilter := "all"

	fs := flag.NewFlagSet("devices", flag.ExitOnError)
	fs.BoolVar(&jsonOut, "json", false, "Output as JSON")
	fs.StringVar(&platformFilter, "platform", "all", "Filter: android, ios, all")
	fs.BoolVar(&buildInfo, "build-info", false, "Show build type (Debug/Release)")
	fs.Parse(args)

	if verbose {
		log.Printf("devices: json=%v platform=%s build-info=%v", jsonOut, platformFilter, buildInfo)
	}

	// Collect devices from available platforms
	type deviceEntry struct {
		Device    engine.Device    `json:"device"`
		Processes []engine.AppProcess `json:"processes,omitempty"`
	}
	var entries []deviceEntry

	// Android
	if platformFilter == "all" || platformFilter == "android" {
		adbPath, err := android.FindAdbPath()
		if err == nil {
			prov := android.NewProvider(adbPath)
			if devices, err := prov.Discover(); err == nil {
				for _, d := range devices {
					entry := deviceEntry{Device: d}
					if buildInfo {
						prov.SetDevice(d.ID)
						if procs, err := prov.MapProcesses(d.ID); err == nil {
							entry.Processes = procs
						}
					}
					entries = append(entries, entry)
				}
			}
		}
	}

	// iOS
	if platformFilter == "all" || platformFilter == "ios" {
		xcrunPath, err := iosPkg.FindXcrunPath()
		if err == nil {
			prov := iosPkg.NewProvider(xcrunPath)
			if devices, err := prov.Discover(); err == nil {
				prov.CacheDevices(devices)
				for _, d := range devices {
					entry := deviceEntry{Device: d}
					if buildInfo {
						prov.SetDevice(d.ID)
						if procs, err := prov.MapProcesses(d.ID); err == nil {
							entry.Processes = procs
						}
					}
					entries = append(entries, entry)
				}
			}
		}
	}

	// Output
	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(entries); err != nil {
			fmt.Fprintf(os.Stderr, "JSON output error: %v\n", err)
		}
		return
	}

	if len(entries) == 0 {
		fmt.Println("No devices found.")
		fmt.Println("Connect a device or use --mock for development.")
		return
	}

	// Group by platform for display
	fmt.Println("Available Devices:")
	fmt.Println("──────────────────────────────")
	for _, e := range entries {
		typ := "emulator"
		if e.Device.IsPhysical {
			typ = "physical"
		}
		fmt.Printf("  • %s  %s  (%s, %s)\n", e.Device.ID, e.Device.Name, e.Device.Platform, typ)
		if buildInfo && len(e.Processes) > 0 {
			for _, p := range e.Processes {
				fmt.Printf("      PID %d — %s [%s]\n", p.PID, p.PackageName, p.BuildType)
			}
		}
	}
}

// isCommandAvailable checks if a command exists in PATH.
func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// envInt reads an integer from an environment variable, falling back to defaultVal.
func envInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return defaultVal
}

// envStr reads a string from an environment variable, falling back to defaultVal.
func envStr(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
