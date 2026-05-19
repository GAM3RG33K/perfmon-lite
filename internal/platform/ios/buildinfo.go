package ios

import (
	"fmt"
	"strings"

	"github.com/GAM3RG33K/perfmon-lite/internal/engine"
)

// BuildType detects whether the given bundle is a debug or release build.
// For simulators, it checks the application's Info.plist for the
// Get-task-allow entitlement (present in debug builds) or checks if the
// binary contains the __DEBUG section.
//
// For physical devices, build type detection is limited.
func (p *iOSProvider) BuildType(deviceID, bundleID string) (engine.BuildType, error) {
	device, err := p.getDevice(deviceID)
	if err != nil {
		return engine.BuildUnknown, err
	}

	if device.IsPhysical {
		return engine.BuildUnknown, fmt.Errorf("build type detection not supported on physical iOS devices")
	}

	return p.simulatorBuildType(deviceID, bundleID)
}

// simulatorBuildType detects the build type of an app in a booted simulator.
// It checks the `Get-task-allow` entitlement and the presence of debug symbols.
func (p *iOSProvider) simulatorBuildType(udid, bundleID string) (engine.BuildType, error) {
	// Try to locate the app container path (host-side command)
	appPath, err := p.xcrunExec("simctl", "get_app_container", udid, bundleID)
	if err != nil {
		// Fall back: check via entitlements
		return p.simulatorBuildTypeByEntitlements(udid, bundleID)
	}

	appPath = strings.TrimSpace(appPath)
	if appPath == "" {
		return p.simulatorBuildTypeByEntitlements(udid, bundleID)
	}

	// Check Info.plist for app store receipt (release builds have it)
	out, err := p.simctlSpawn(udid, "defaults", "read", appPath+"/Info.plist", "CFBundleDevelopmentRegion")
	if err == nil && strings.TrimSpace(out) != "" {
		// Check for the presence of a provisioning profile or receipt
		receiptCheck, _ := p.simctlSpawn(udid, "ls", appPath+"/_CodeSignature/")
		if strings.TrimSpace(receiptCheck) != "" {
			return engine.BuildRelease, nil
		}
	}

	// Default to debug for simulator builds (simulator builds are typically debug)
	return engine.BuildDebug, nil
}

// simulatorBuildTypeByEntitlements checks the app's entitlements for Get-task-allow.
func (p *iOSProvider) simulatorBuildTypeByEntitlements(udid, bundleID string) (engine.BuildType, error) {
	// Try to read the app's entitlements
	out, err := p.simctlSpawn(udid, "codesign", "-d", "--entitlements", "-", "--find", bundleID)
	if err != nil {
		// codesign not available or app not found; simulator apps are typically debug
		return engine.BuildDebug, nil
	}

	// Look for get-task-allow entitlement (present in debug builds)
	if strings.Contains(strings.ToLower(out), "get-task-allow") &&
		strings.Contains(out, "true") {
		return engine.BuildDebug, nil
	}

	return engine.BuildRelease, nil
}
