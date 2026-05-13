package engine

// DeviceDiscovery is implemented by platform providers to discover connected devices.
type DeviceDiscovery interface {
	// Discover returns a list of connected/booted devices.
	Discover() ([]Device, error)
}

// ProcessMapper is implemented by platform providers to map running processes.
type ProcessMapper interface {
	// MapProcesses returns running processes for the given device.
	MapProcesses(deviceID string) ([]AppProcess, error)

	// BuildType detects whether the given package/bundle is debug or release.
	BuildType(deviceID, packageName string) (BuildType, error)
}

// TelemetryProvider is implemented by platform providers to sample metrics.
type TelemetryProvider interface {
	// Sample collects a single telemetry snapshot for the given PID.
	Sample(pid int32) (*TelemetrySnapshot, error)

	// Close releases any resources held by the provider.
	Close() error
}

// PlatformProvider combines all platform-specific interfaces.
type PlatformProvider interface {
	DeviceDiscovery
	ProcessMapper
	TelemetryProvider
}
