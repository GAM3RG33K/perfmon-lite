package android

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// DefaultAdbPaths lists common locations for the adb binary.
var DefaultAdbPaths = []string{
	"adb", // in PATH
}

// FindAdbPath searches for the adb binary in common locations.
// It checks:
//  1. The ADB_SYSTEM_PATH / PERFMON_ADB_PATH environment variable
//  2. The PATH environment variable
//  3. $ANDROID_HOME/platform-tools/adb
//  4. $ANDROID_SDK_ROOT/platform-tools/adb
//  5. ~/Library/Android/sdk/platform-tools/adb (macOS)
//  6. ~/Android/Sdk/platform-tools/adb (Linux)
//  7. %LOCALAPPDATA%/Android/Sdk/platform-tools/adb.exe (Windows)
func FindAdbPath() (string, error) {
	// Check explicit environment variables first
	for _, env := range []string{"PERFMON_ADB_PATH", "ADB_SYSTEM_PATH"} {
		if envPath := os.Getenv(env); envPath != "" {
			if _, err := os.Stat(envPath); err == nil {
				return envPath, nil
			}
		}
	}

	// Check PATH
	if path, err := exec.LookPath("adb"); err == nil {
		return path, nil
	}

	// Check ANDROID_HOME / SDK paths
	searchDirs := []string{
		os.Getenv("ANDROID_HOME"),
		os.Getenv("ANDROID_SDK_ROOT"),
	}

	// Add default platform-specific SDK paths
	home, _ := os.UserHomeDir()
	if home != "" {
		searchDirs = append(searchDirs,
			filepath.Join(home, "Library", "Android", "sdk"),  // macOS
			filepath.Join(home, "Android", "Sdk"),              // Linux
			filepath.Join(home, "AppData", "Local", "Android", "Sdk"), // Windows
		)
	}

	// Try adb (unix) and adb.exe (windows)
	for _, dir := range searchDirs {
		if dir == "" {
			continue
		}
		for _, bin := range []string{"adb", "adb.exe"} {
			adbPath := filepath.Join(dir, "platform-tools", bin)
			if _, err := os.Stat(adbPath); err == nil {
				return adbPath, nil
			}
		}
	}

	return "", fmt.Errorf("adb not found: check PATH or set ANDROID_HOME")
}

// AdbVersion represents a parsed ADB version.
type AdbVersion struct {
	Major int
	Minor int
	Patch int
}

func (v AdbVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// versionRegex parses ADB version strings like "Version 36.0.0-13206524".
var versionRegex = regexp.MustCompile(`Version\s+(\d+)\.(\d+)\.(\d+)`)

// CheckVersion runs `adb version` and returns the parsed version.
func CheckVersion(adbPath string) (AdbVersion, error) {
	cmd := exec.Command(adbPath, "version")
	out, err := cmd.Output()
	if err != nil {
		return AdbVersion{}, fmt.Errorf("failed to get adb version: %w", err)
	}

	return ParseAdbVersion(string(out))
}

// ParseAdbVersion extracts the version from the adb version output.
//
// Expected format:
//
//	Android Debug Bridge version 1.0.41
//	Version 36.0.0-13206524
//	Installed as /path/to/adb
func ParseAdbVersion(output string) (AdbVersion, error) {
	matches := versionRegex.FindStringSubmatch(output)
	if matches == nil || len(matches) < 4 {
		return AdbVersion{}, fmt.Errorf("unable to parse adb version from: %s", strings.TrimSpace(output))
	}

	major := parseIntOrZero(matches[1])
	minor := parseIntOrZero(matches[2])
	patch := parseIntOrZero(matches[3])

	return AdbVersion{Major: major, Minor: minor, Patch: patch}, nil
}

func parseIntOrZero(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

// ValidateDevice checks that the specified device is reachable and responsive.
func ValidateDevice(adbPath, deviceID string) error {
	cmd := exec.Command(adbPath, "-s", deviceID, "shell", "echo", "ok")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("device %s is not reachable: %s", deviceID, strings.TrimSpace(string(out)))
	}
	if strings.TrimSpace(string(out)) != "ok" {
		return fmt.Errorf("device %s returned unexpected response", deviceID)
	}
	return nil
}
