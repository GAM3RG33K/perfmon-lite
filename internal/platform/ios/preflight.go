package ios

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// FindXcrunPath searches for the xcrun binary in common locations.
// It checks:
//  1. The XCRUN_PATH environment variable (for explicit path)
//  2. The PATH environment variable
//  3. Common Xcode toolchain paths
func FindXcrunPath() (string, error) {
	// Check explicit environment variable first
	if envPath := os.Getenv("XCRUN_PATH"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath, nil
		}
	}

	// Check PATH
	if path, err := exec.LookPath("xcrun"); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("xcrun not found: check PATH or set XCRUN_PATH")
}

// XcrunVersion represents a parsed xcrun version.
type XcrunVersion struct {
	Major int
	Minor int
	Patch int
}

func (v XcrunVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// CheckVersion runs `xcrun --version` and returns the parsed version.
func CheckVersion(xcrunPath string) (XcrunVersion, error) {
	cmd := exec.Command(xcrunPath, "--version")
	out, err := cmd.Output()
	if err != nil {
		return XcrunVersion{}, fmt.Errorf("failed to get xcrun version: %w", err)
	}

	return parseVersion(string(out))
}

// parseVersion extracts the version from xcrun --version output.
// Expected format:
//
//	xcrun version 72.
func parseVersion(output string) (XcrunVersion, error) {
	output = strings.TrimSpace(output)
	// Format: "xcrun version 72."
	parts := strings.Fields(output)
	if len(parts) < 3 {
		return XcrunVersion{}, fmt.Errorf("unable to parse xcrun version from: %s", output)
	}

	// Parse the version number, stripping trailing dot
	verStr := strings.TrimRight(parts[2], ".")
	var major int
	if _, err := fmt.Sscanf(verStr, "%d", &major); err != nil {
		return XcrunVersion{}, fmt.Errorf("unable to parse version number from %q: %w", verStr, err)
	}

	return XcrunVersion{Major: major}, nil
}

// CheckXcodeSelect verifies that the active Xcode developer directory is set.
func CheckXcodeSelect() error {
	cmd := exec.Command("xcode-select", "-p")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("xcode-select failed: %s", strings.TrimSpace(string(out)))
	}
	if len(out) == 0 {
		return fmt.Errorf("xcode-select returned empty path: run 'xcode-select --install'")
	}
	return nil
}
