package main

import (
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

var version = "1.0.0"

func main() {
	// CLI flags
	mockMode := flag.Bool("mock", false, "Run with simulated telemetry data (no device required)")
	iOSMode := flag.Bool("ios", false, "Force iOS mode (use xcrun instead of ADB)")
	intervalFlag := flag.Int("interval", 1, "Polling interval in seconds (range: 1-60)")
	bufferFlag := flag.Int("buffer", 300, "Ring buffer capacity (number of data points)")
	exportFlag := flag.String("export", "", "Export format: json, md, html, pdf")
	outputFlag := flag.String("output", "./perfmon_export", "Output path for export file (without extension)")
	verboseFlag := flag.Bool("verbose", false, "Enable verbose logging")
	showVersion := flag.Bool("version", false, "Show version information")
	showHelp := flag.Bool("help", false, "Show help message")

	// Custom usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "perfmon v%s — Mobile Performance Monitor & Profiler\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  perfmon [flags]\n")
		fmt.Fprintf(os.Stderr, "  perfmon devices [flags]\n")
		fmt.Fprintf(os.Stderr, "  perfmon export <format> [flags]\n")
		fmt.Fprintf(os.Stderr, "  perfmon version [flags]\n")
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("perfmon v%s\n", version)
		os.Exit(0)
	}

	// Handle help flag
	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// Validate interval
	if *intervalFlag < 1 || *intervalFlag > 60 {
		log.Fatalf("--interval must be between 1 and 60 seconds, got %d", *intervalFlag)
	}

	// Validate buffer
	if *bufferFlag < 10 {
		log.Fatalf("--buffer must be at least 10, got %d", *bufferFlag)
	}

	if *verboseFlag {
		log.SetFlags(log.Ltime | log.Lmicroseconds)
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
	} else if *iOSMode {
		// ── iOS mode (forced via --ios flag) ────────────────────────────
		provider, discoveredDevices, discoveredProcesses, initialPID, targetPlatform = setupiOSProvider(*verboseFlag)
	} else {
		// ── Auto-detect: try Android first ────────────────────────────
		adbPath, adbErr := android.FindAdbPath()
		if adbErr != nil {
			wizardResult := runPreflightWizard()
			switch wizardResult {
			case "mock":
				mockProvider := mock.NewProvider(time.Now().UnixNano())
				provider = mockProvider
				initialPID = 9001
				targetPlatform = engine.PlatformMock
				discoveredDevices = []engine.Device{mock.MockDevice()}
				discoveredProcesses = []engine.AppProcess{mock.MockProcess()}
			case "retry":
		provider, discoveredDevices, discoveredProcesses, initialPID, targetPlatform, _ = tryAndroidProvider(adbPath, *verboseFlag)
		case "quit":
				os.Exit(0)
			default:
				// Fall through to iOS
			}

			// If wizard didn't set up a provider, try iOS
			if provider == nil {
				if *verboseFlag {
					log.Printf("Android not available — trying iOS...")
				}
				provider, discoveredDevices, discoveredProcesses, initialPID, targetPlatform = setupiOSProvider(*verboseFlag)
			}
		} else {
			provider, discoveredDevices, discoveredProcesses, initialPID, targetPlatform, _ = tryAndroidProvider(adbPath, *verboseFlag)
		}
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
			time.Sleep(time.Duration(*intervalFlag) * time.Second / 10)
		}

		snapshots := eng.Buffer.GetAll()
		if len(snapshots) == 0 {
			fmt.Println("No telemetry data collected — exiting.")
			os.Exit(1)
		}

		deviceName := "unknown"
		if len(discoveredDevices) > 0 {
			deviceName = discoveredDevices[0].Name
		}
		appName := "unknown"
		buildType := engine.BuildUnknown
		if len(discoveredProcesses) > 0 {
			appName = discoveredProcesses[0].PackageName
			buildType = discoveredProcesses[0].BuildType
		}

		var format export.Format
		switch *exportFlag {
		case "json":
			format = export.FormatJSON
		case "md", "markdown":
			format = export.FormatMD
		case "html":
			format = export.FormatHTML
		case "pdf":
			format = export.FormatPDF
		default:
			fmt.Fprintf(os.Stderr, "Unsupported export format: %s (supported: json, md, html, pdf)\n", *exportFlag)
			os.Exit(1)
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
			os.Exit(1)
		}

		fmt.Printf("Exporting %d data points to %s format...\n", len(snapshots), *exportFlag)
		path, err := export.Export(snapshots, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Export failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Report written to: %s\n", path)
		os.Exit(0)
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

	// Build options
	teaOpts := []tea.ProgramOption{
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	}

	p := tea.NewProgram(model, teaOpts...)

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running TUI: %v", err)
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
// adbPath should be resolved via android.FindAdbPath() first.
func tryAndroidProvider(adbPath string, verbose bool) (engine.TelemetryProvider, []engine.Device, []engine.AppProcess, int32, engine.Platform, error) {
	if verbose {
		log.Print("Looking for ADB...")
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

// setupiOSProvider creates and configures the iOS provider.
func setupiOSProvider(verbose bool) (engine.TelemetryProvider, []engine.Device, []engine.AppProcess, int32, engine.Platform) {
	if verbose {
		log.Print("Looking for xcrun...")
	}

	xcrunPath, err := iosPkg.FindXcrunPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "iOS Error: %v\n", err)
		fmt.Fprint(os.Stderr, hintInstallXcode())
		os.Exit(1)
	}

	if verbose {
		log.Printf("xcrun found at: %s", xcrunPath)
	}

	// Verify xcrun version
	xcrunVer, err := iosPkg.CheckVersion(xcrunPath)
	if err != nil {
		log.Fatalf("xcrun version check failed: %v", err)
	}
	if verbose {
		log.Printf("xcrun version: %s", xcrunVer.String())
	}

	// Check xcode-select
	if err := iosPkg.CheckXcodeSelect(); err != nil {
		log.Fatalf("Xcode not configured: %v", err)
	}

	// Create the iOS provider
	iOSProvider := iosPkg.NewProvider(xcrunPath)

	// Discover iOS devices and simulators
	devices, err := iOSProvider.Discover()
	if err != nil {
		log.Fatalf("Failed to discover iOS devices: %v", err)
	}
	if len(devices) == 0 {
		log.Fatalf("No iOS devices or simulators found. Boot a simulator, connect a device, or use --mock for development.")
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
		log.Fatalf("Failed to list processes on %s: %v", selectedDevice.ID, err)
	}
	if len(processes) == 0 {
		log.Fatalf("No processes found on device %s", selectedDevice.ID)
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

	return iOSProvider, devices, processes, initialPID, engine.PlatformIOS
}

// selectBestProcess picks the most interesting process from a list of processes.
// Preference order:
//  1. A process whose name/package contains a dot (user app, not a system daemon)
//  2. The first non-kernel process
//  3. The first process overall
func selectBestProcess(processes []engine.AppProcess) engine.AppProcess {
	// First pass: find user apps (names with a dot)
	var userApps []engine.AppProcess
	for _, p := range processes {
		if strings.Contains(p.PackageName, ".") &&
			!strings.HasPrefix(p.PackageName, "android.") &&
			!strings.HasPrefix(p.PackageName, "com.apple.") {
			userApps = append(userApps, p)
		}
	}
	if len(userApps) > 0 {
		return userApps[0]
	}

	// Fallback: first non-kernel process
	if len(processes) > 0 {
		return processes[0]
	}

	// Should never get here, but return empty
	return engine.AppProcess{}
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

// isCommandAvailable checks if a command exists in PATH.
func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// hintInstallADB returns a help string for installing ADB.
func hintInstallADB() string {
	return `
  ── ADB Installation Help ──────────────────────────────
    macOS (Homebrew):    brew install android-platform-tools
    macOS (Manual):      brew install --cask android-sdk
    Linux:               sudo apt install adb
                         sudo pacman -S android-tools
    Verify:              adb devices -l
  ────────────────────────────────────────────────────────
  Or run with --mock for development:  perfmon --mock
`
}

// listDevices lists connected devices.
func listDevices(useMock bool) {
	if useMock {
		fmt.Println("Available Devices:")
		fmt.Println("──────────────────────────────")
		dev := mock.MockDevice()
		fmt.Printf("  %s (%s) — %s\n", dev.Name, dev.ID, dev.Platform)
		fmt.Println()
		fmt.Println("Processes:")
		fmt.Println("──────────────────────────────")
		proc := mock.MockProcess()
		fmt.Printf("  PID %d — %s [%s]\n", proc.PID, proc.PackageName, proc.BuildType)
		return
	}

	adbPath, err := android.FindAdbPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ADB not found: %v\n", err)
		fmt.Fprintf(os.Stderr, "Use --mock for simulated output.\n")
		return
	}

	provider := android.NewProvider(adbPath)
	devices, err := provider.Discover()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to discover devices: %v\n", err)
		return
	}

	if len(devices) == 0 {
		fmt.Println("No Android devices found.")
		fmt.Println("Connect a device or use --mock for simulated output.")
		return
	}

	fmt.Println("Available Devices:")
	fmt.Println("──────────────────────────────")
	for _, d := range devices {
		typ := "emulator"
		if d.IsPhysical {
			typ = "physical"
		}
		fmt.Printf("  %s (%s) — %s, %s\n", d.Name, d.ID, d.Platform, typ)
	}
}

// exportData exports telemetry data (stub for now).
func exportData(eng *engine.Engine, args []string, output string) {
	format := "json"
	if len(args) > 1 {
		format = args[1]
	}

	snapshots := eng.Buffer.GetAll()
	fmt.Printf("Exporting %d telemetry data points in %s format...\n", len(snapshots), format)
	fmt.Println("Full export subsystem coming in Phase 4.")
	fmt.Printf("Output would be written to: %s.%s\n", output, format)
}
