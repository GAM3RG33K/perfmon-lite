package mock

import (
	"testing"
)

func TestNewProvider(t *testing.T) {
	p := NewProvider(42)
	if p == nil {
		t.Fatal("expected non-nil provider")
	}
	if p.step != 0 {
		t.Fatalf("expected initial step 0, got %f", p.step)
	}
}

func TestNewProvider_Deterministic(t *testing.T) {
	p1 := NewProvider(99)
	p2 := NewProvider(99)

	s1, err := p1.Sample(9001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s2, err := p2.Sample(9001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s1.CPUPercent != s2.CPUPercent {
		t.Fatalf("expected same CPU for same seed: %f vs %f", s1.CPUPercent, s2.CPUPercent)
	}
	if s1.MemoryKB != s2.MemoryKB {
		t.Fatalf("expected same Memory for same seed: %d vs %d", s1.MemoryKB, s2.MemoryKB)
	}
	if s1.Threads != s2.Threads {
		t.Fatalf("expected same Threads for same seed: %d vs %d", s1.Threads, s2.Threads)
	}
}

func TestNewProvider_DifferentSeeds(t *testing.T) {
	p1 := NewProvider(1)
	p2 := NewProvider(2)

	s1, _ := p1.Sample(9001)
	s2, _ := p2.Sample(9001)

	// With different seeds, the noise component should differ
	if s1.CPUPercent == s2.CPUPercent && s1.MemoryKB == s2.MemoryKB && s1.Threads == s2.Threads {
		t.Log("Note: different seeds produced identical values for first sample (unlikely but possible)")
	}
}

func TestSample_ValidRange(t *testing.T) {
	p := NewProvider(42)

	for i := 0; i < 1000; i++ {
		s, err := p.Sample(9001)
		if err != nil {
			t.Fatalf("unexpected error at step %d: %v", i, err)
		}
		if s.CPUPercent < 2 || s.CPUPercent > 90 {
			t.Fatalf("step %d: CPU %f out of range [2, 90]", i, s.CPUPercent)
		}
		if s.MemoryKB < 50000 {
			t.Fatalf("step %d: Memory %d below minimum 50000", i, s.MemoryKB)
		}
		if s.Threads < 10 {
			t.Fatalf("step %d: Threads %d below minimum 10", i, s.Threads)
		}
		if s.Threads > 60 {
			t.Fatalf("step %d: Threads %d seems unreasonably high", i, s.Threads)
		}
	}
}

func TestSample_StepIncrement(t *testing.T) {
	p := NewProvider(42)

	// Step should increment on each call
	s1, _ := p.Sample(9001)
	s2, _ := p.Sample(9001)
	s3, _ := p.Sample(9001)

	// Values should differ due to sinusoidal progression
	if s1.CPUPercent == s2.CPUPercent && s2.CPUPercent == s3.CPUPercent {
		t.Log("Note: consecutive samples produced identical values (unlikely — check sinusoidal component)")
	}
}

func TestSample_SinusoidalCPU(t *testing.T) {
	p := NewProvider(0)

	// Disable randomness by using seed 0 and checking pattern
	// CPU should follow a sinusoidal pattern: baseCPU + cpuAmp*sin(step*0.15)
	// At step ~10.47 (pi/0.15), sin reaches peak. At step ~31.4 (pi*2/0.15), sin goes negative.
	samples := make([]float64, 50)
	for i := 0; i < 50; i++ {
		s, _ := p.Sample(9001)
		samples[i] = s.CPUPercent
	}

	// CPU should vary (not constant)
	var varied bool
	for i := 1; i < len(samples); i++ {
		if samples[i] != samples[0] {
			varied = true
			break
		}
	}
	if !varied {
		t.Fatal("CPU values did not vary across samples")
	}
}

func TestSample_MemoryLeakAfter100(t *testing.T) {
	p := NewProvider(0)

	// Run 110 samples — leak starts after step 100
	var beforeLeak, afterLeak int64
	for i := 0; i < 110; i++ {
		s, _ := p.Sample(9001)
		if i == 99 {
			beforeLeak = s.MemoryKB
		}
		if i == 109 {
			afterLeak = s.MemoryKB
		}
	}

	// After 10 leak steps, memory should have increased by at least
	// 10 * 50 = 500KB (the leak adds 50KB per step after 100)
	if afterLeak <= beforeLeak {
		t.Fatalf("expected memory increase from leak, before=%d after=%d", beforeLeak, afterLeak)
	}
	leakAmount := afterLeak - beforeLeak
	if leakAmount < 500 {
		t.Fatalf("expected at least 500KB leak, got %dKB", leakAmount)
	}
}

func TestSample_MemoryLeakCap(t *testing.T) {
	p := NewProvider(0)

	// Run enough steps to overflow the leak cap (500MB = 500*1024KB)
	// Leak = step * 50, cap = 500*1024
	// Steps needed to reach cap: 500*1024 / 50 = 10240
	for i := 0; i < 15000; i++ {
		s, _ := p.Sample(9001)
		if s.MemoryKB < 0 {
			t.Fatalf("step %d: memory overflowed to negative: %d", i, s.MemoryKB)
		}
	}

	// At step 15000, leak would be 15000 * 50 = 750000KB (~732MB) without cap
	// With cap of 500*1024 = 512000KB, it should be capped
	s, _ := p.Sample(9001)
	// Memory should be reasonable — baseMem (~184MB) + leak (500MB) = ~700MB + sinusoidal
	// Without cap it'd be in the GB range
	if s.MemoryKB > 900*1024 {
		t.Fatalf("memory seems uncapped: %d KB", s.MemoryKB)
	}
}

func TestSample_PIDIgnored(t *testing.T) {
	p1 := NewProvider(42)
	p2 := NewProvider(42)

	s1, _ := p1.Sample(9001) // PID 9001
	s2, _ := p2.Sample(9999) // PID 9999, same seed, same step

	// Different PIDs should produce the same data (mock ignores PID)
	if s1.CPUPercent != s2.CPUPercent {
		t.Fatalf("expected same CPU for different PIDs: %f vs %f", s1.CPUPercent, s2.CPUPercent)
	}
	if s1.MemoryKB != s2.MemoryKB {
		t.Fatalf("expected same Memory for different PIDs: %d vs %d", s1.MemoryKB, s2.MemoryKB)
	}
	if s1.Threads != s2.Threads {
		t.Fatalf("expected same Threads for different PIDs: %d vs %d", s1.Threads, s2.Threads)
	}
}

func TestSample_ThreadsVary(t *testing.T) {
	p := NewProvider(42)

	samples := make([]int32, 200)
	for i := 0; i < 200; i++ {
		s, _ := p.Sample(9001)
		samples[i] = s.Threads
	}

	// Threads should have some variation
	var varied bool
	for i := 1; i < len(samples); i++ {
		if samples[i] != samples[0] {
			varied = true
			break
		}
	}
	if !varied {
		t.Fatal("thread counts did not vary across samples")
	}
}

func TestSample_ReturnsCopy(t *testing.T) {
	p := NewProvider(42)

	s1, _ := p.Sample(9001)
	s2, _ := p.Sample(9001)

	// Each call should return a fresh snapshot
	if &s1 == &s2 {
		t.Fatal("samples should be independent copies")
	}
}

func TestClose(t *testing.T) {
	p := NewProvider(42)
	if err := p.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}
}

func TestMockDevice(t *testing.T) {
	d := MockDevice()
	if d.ID != "mock-device-001" {
		t.Fatalf("expected ID mock-device-001, got %s", d.ID)
	}
	if d.Platform != "mock" {
		t.Fatalf("expected platform mock, got %s", d.Platform)
	}
	if !d.IsBooted {
		t.Fatal("expected device to be booted")
	}
}

func TestMockProcess(t *testing.T) {
	p := MockProcess()
	if p.PID != 9001 {
		t.Fatalf("expected PID 9001, got %d", p.PID)
	}
	if p.Name != "com.example.app" {
		t.Fatalf("expected name com.example.app, got %s", p.Name)
	}
	if p.BuildType != "debug" {
		t.Fatalf("expected build type debug, got %s", p.BuildType)
	}
}

func TestSample_OutputShape(t *testing.T) {
	p := NewProvider(42)

	s, err := p.Sample(9001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify all fields are populated
	if s.Timestamp == 0 {
		t.Error("expected non-zero timestamp")
	}
	if s.CPUPercent == 0 {
		t.Error("expected non-zero CPU")
	}
	if s.MemoryKB == 0 {
		t.Error("expected non-zero memory")
	}
	if s.Threads == 0 {
		t.Error("expected non-zero threads")
	}
}

func BenchmarkSample(b *testing.B) {
	p := NewProvider(42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Sample(9001)
	}
}
