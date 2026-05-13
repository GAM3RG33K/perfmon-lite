package mock

import (
	"math"
	"math/rand"

	"github.com/w1n/perfmon/internal/engine"
)

// Provider generates simulated telemetry data for TUI development.
type Provider struct {
	rng       *rand.Rand
	step      float64
	baseCPU   float64
	baseMem   int64
	baseThr   int32
	cpuAmp    float64
	memAmp    int64
	thrAmp    int32
}

// NewProvider creates a new mock provider with the given seed for reproducibility.
func NewProvider(seed int64) *Provider {
	src := rand.NewSource(seed)
	return &Provider{
		rng:     rand.New(src),
		step:    0,
		baseCPU: 30.0,
		baseMem: 180 * 1024, // 180 MB in KB
		baseThr: 38,
		cpuAmp:  25.0,
		memAmp:  40 * 1024, // 40 MB in KB
		thrAmp:  8,
	}
}

// Sample generates a mock telemetry snapshot with sinusoidal variation.
func (p *Provider) Sample(pid int32) (*engine.TelemetrySnapshot, error) {
	p.step++

	// CPU: sinusoidal wave between 5% and 80%
	cpu := p.baseCPU + p.cpuAmp*math.Sin(p.step*0.15)
	cpu += p.rng.Float64()*8 - 4 // add noise
	if cpu < 2 {
		cpu = 2
	}
	if cpu > 90 {
		cpu = 90
	}

	// Memory: slower sinusoidal wave with slight upward trend
	memRatio := 0.5 + 0.5*math.Sin(p.step*0.05)
	memKB := p.baseMem + int64(float64(p.memAmp)*memRatio)
	memKB += int64(p.rng.Float64()*float64(p.memAmp)*0.2 - float64(p.memAmp)*0.1)
	if p.step > 100 {
		// gradual memory increase to simulate a leak (capped at 500MB)
		leak := int64(p.step * 50)
		if leak > 500*1024 {
			leak = 500 * 1024
		}
		memKB += leak
	}
	if memKB < 50000 {
		memKB = 50000
	}

	// Threads: varies slightly around base
	threads := p.baseThr + int32(float64(p.thrAmp)*math.Sin(p.step*0.1))
	threads += int32(p.rng.Intn(5) - 2)
	if threads < 10 {
		threads = 10
	}

	snapshot := engine.NewTelemetrySnapshot(cpu, memKB, threads)
	return &snapshot, nil
}

// Close is a no-op for the mock provider.
func (p *Provider) Close() error {
	return nil
}

// MockDevice returns a mock device entry for --mock mode.
func MockDevice() engine.Device {
	return engine.Device{
		ID:         "mock-device-001",
		Name:       "Mock Device (Simulated)",
		Platform:   engine.PlatformMock,
		IsPhysical: false,
		IsBooted:   true,
	}
}

// MockProcess returns a mock process entry for --mock mode.
func MockProcess() engine.AppProcess {
	return engine.AppProcess{
		PID:         9001,
		Name:        "com.example.app",
		PackageName: "com.example.app",
		BuildType:   engine.BuildDebug,
	}
}
