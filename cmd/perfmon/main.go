package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/w1n/perfmon/internal/engine"
	"github.com/w1n/perfmon/internal/platform/android"
	"github.com/w1n/perfmon/internal/platform/mock"
	perfmonTui "github.com/w1n/perfmon/internal/tui"
)

const version = "1.0.0"

func main() {
	// CLI flags
	mockMode := flag.Bool("mock", false, "Run with simulated telemetry data (no device required)")
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
	} else {
		// ── Android mode ────────────────────────────────────────────────
		if *verboseFlag {
			log.Print("Looking for ADB...")
		}

		adbPath, err := android.FindAdbPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ADB not found: %v\n", err)
		fmt.Fprint(os.Stderr, hintInstallADB())
		os.Exit(1)
	}

		if *verboseFlag {
			log.Printf("ADB found at: %s", adbPath)
		}

		// Verify ADB version
		adbVer, err := android.CheckVersion(adbPath)
		if err != nil {
			log.Fatalf("ADB version check failed: %v", err)
		}
		if *verboseFlag {
			log.Printf("ADB version: %s", adbVer.String())
		}

		// Create the Android provider
		androidProvider := android.NewProvider(adbPath)

		// Discover connected devices
		devices, err := androidProvider.Discover()
		if err != nil {
			log.Fatalf("Failed to discover Android devices: %v", err)
		}
		if len(devices) == 0 {
			log.Fatalf("No Android devices found. Connect a device or use --mock for development.")
		}

		if *verboseFlag {
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
		discoveredDevices = devices

		if *verboseFlag {
			log.Printf("Selected device: %s (%s)", selectedDevice.Name, selectedDevice.ID)
		}

		// Validate device reachability
		if err := android.ValidateDevice(adbPath, selectedDevice.ID); err != nil {
			log.Fatalf("Device %s is not reachable: %v", selectedDevice.ID, err)
		}

		if *verboseFlag {
			log.Print("Device is reachable — discovering processes...")
		}

		// Discover processes on the device
		processes, err := androidProvider.MapProcesses(selectedDevice.ID)
		if err != nil {
			log.Fatalf("Failed to list processes on %s: %v", selectedDevice.ID, err)
		}
		if len(processes) == 0 {
			log.Fatalf("No processes found on device %s", selectedDevice.ID)
		}

		if *verboseFlag {
			log.Printf("Found %d processes. Scanning for user applications...", len(processes))
		}

		// Select the most interesting process — prefer a user app with a package name
		// containing a dot (indicating a Java/Kotlin app), falling back to the first
		// non-system process, then the first process overall.
		selectedProcess := selectBestProcess(processes)
		initialPID = selectedProcess.PID

		if *verboseFlag {
			log.Printf("Selected process: %s (PID %d) [%s]",
				selectedProcess.PackageName, selectedProcess.PID, selectedProcess.BuildType)
		}

		// Detect build type for the selected process
		buildType, err := androidProvider.BuildType(selectedDevice.ID, selectedProcess.PackageName)
		if err == nil {
			selectedProcess.BuildType = buildType
			if *verboseFlag {
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
		discoveredProcesses = processes

		provider = androidProvider
		targetPlatform = engine.PlatformAndroid
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

		fmt.Printf("Exporting to %s format at %s...\n", *exportFlag, *outputFlag)
		fmt.Println("Export subsystem coming in Phase 4.")
		os.Exit(0)
	}

	// ══════════════════════════════════════════════════════════════════════
	// Interactive TUI Mode
	// ══════════════════════════════════════════════════════════════════════

	model := perfmonTui.NewModel(eng, *mockMode)

	// Populate the TUI with discovered devices and processes
	model.SetTargets(discoveredDevices, discoveredProcesses)

	model.Logs.AddEntry("INFO", fmt.Sprintf("Target: %s | App: %s (PID %d)",
		targetPlatform, discoveredProcesses[0].PackageName, initialPID))
	model.Logs.AddEntry("INFO", fmt.Sprintf("Polling every %d second(s)", *intervalFlag))

	if *mockMode {
		model.Logs.AddEntry("INFO", "Mock mode — simulated telemetry data")
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

// selectBestProcess picks the most interesting process from a list of Android processes.
// Preference order:
//  1. A process whose package name contains a dot (user app, not a native daemon)
//  2. The first non-kernel process
//  3. The first process overall
func selectBestProcess(processes []engine.AppProcess) engine.AppProcess {
	// First pass: find user apps (package names with a dot)
	var userApps []engine.AppProcess
	for _, p := range processes {
		if strings.Contains(p.PackageName, ".") && !strings.HasPrefix(p.PackageName, "android.") {
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
