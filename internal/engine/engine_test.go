package engine

import (
	"testing"
	"time"
)

// ─── Ring Buffer Tests ───────────────────────────────────────────────────────

func TestRingBuffer_New(t *testing.T) {
	rb := NewRingBuffer(5)
	if rb == nil {
		t.Fatal("expected non-nil ring buffer")
	}
	if rb.Count() != 0 {
		t.Fatalf("expected count 0, got %d", rb.Count())
	}
	if got := rb.Latest(); got != nil {
		t.Fatal("expected nil Latest on empty buffer")
	}
	if got := rb.GetAll(); got != nil {
		t.Fatal("expected nil GetAll on empty buffer")
	}
}

func TestRingBuffer_PushAndGetAll(t *testing.T) {
	rb := NewRingBuffer(3)

	s1 := TelemetrySnapshot{Timestamp: 1, CPUPercent: 10.0, MemoryKB: 100, Threads: 5}
	s2 := TelemetrySnapshot{Timestamp: 2, CPUPercent: 20.0, MemoryKB: 200, Threads: 10}
	s3 := TelemetrySnapshot{Timestamp: 3, CPUPercent: 30.0, MemoryKB: 300, Threads: 15}

	rb.Push(s1)
	rb.Push(s2)
	rb.Push(s3)

	all := rb.GetAll()
	if len(all) != 3 {
		t.Fatalf("expected 3 snapshots, got %d", len(all))
	}
	if all[0].Timestamp != 1 || all[1].Timestamp != 2 || all[2].Timestamp != 3 {
		t.Fatalf("expected chronological order, got timestamps %v", timestamps(all))
	}
}

func TestRingBuffer_ChronologicalOrder(t *testing.T) {
	rb := NewRingBuffer(5)

	for i := int64(1); i <= 5; i++ {
		rb.Push(TelemetrySnapshot{Timestamp: i})
	}

	all := rb.GetAll()
	for i, s := range all {
		if s.Timestamp != int64(i+1) {
			t.Fatalf("expected timestamp %d at index %d, got %d", i+1, i, s.Timestamp)
		}
	}
}

func TestRingBuffer_Eviction(t *testing.T) {
	rb := NewRingBuffer(3)

	for i := int64(1); i <= 5; i++ {
		rb.Push(TelemetrySnapshot{Timestamp: i})
	}

	if rb.Count() != 3 {
		t.Fatalf("expected count 3 after eviction, got %d", rb.Count())
	}

	all := rb.GetAll()
	if len(all) != 3 {
		t.Fatalf("expected 3 items, got %d", len(all))
	}
	// After 5 pushes into capacity 3, we should have timestamps 3, 4, 5
	if all[0].Timestamp != 3 || all[1].Timestamp != 4 || all[2].Timestamp != 5 {
		t.Fatalf("expected timestamps [3,4,5] after eviction, got %v", timestamps(all))
	}
}

func TestRingBuffer_Latest(t *testing.T) {
	rb := NewRingBuffer(3)

	latest := rb.Latest()
	if latest != nil {
		t.Fatal("expected nil Latest on empty buffer")
	}

	s1 := TelemetrySnapshot{Timestamp: 1}
	s2 := TelemetrySnapshot{Timestamp: 2}

	rb.Push(s1)
	if got := rb.Latest(); got == nil || got.Timestamp != 1 {
		t.Fatalf("expected Latest timestamp 1, got %v", got)
	}

	rb.Push(s2)
	if got := rb.Latest(); got == nil || got.Timestamp != 2 {
		t.Fatalf("expected Latest timestamp 2, got %v", got)
	}
}

func TestRingBuffer_LatestAfterWrap(t *testing.T) {
	rb := NewRingBuffer(2)

	// Fill and wrap
	rb.Push(TelemetrySnapshot{Timestamp: 1})
	rb.Push(TelemetrySnapshot{Timestamp: 2})
	rb.Push(TelemetrySnapshot{Timestamp: 3})

	latest := rb.Latest()
	if latest == nil || latest.Timestamp != 3 {
		t.Fatalf("expected Latest timestamp 3 after wrap, got %v", latest)
	}
}

func TestRingBuffer_Count(t *testing.T) {
	rb := NewRingBuffer(5)

	for i := 0; i < 3; i++ {
		rb.Push(TelemetrySnapshot{Timestamp: int64(i)})
	}
	if rb.Count() != 3 {
		t.Fatalf("expected count 3, got %d", rb.Count())
	}

	// Fill beyond capacity
	for i := 3; i < 10; i++ {
		rb.Push(TelemetrySnapshot{Timestamp: int64(i)})
	}
	if rb.Count() != 5 {
		t.Fatalf("expected count 5 (capacity) after overflow, got %d", rb.Count())
	}
}

func TestRingBuffer_PushConcurrent(t *testing.T) {
	rb := NewRingBuffer(100)
	done := make(chan struct{})

	go func() {
		for i := 0; i < 50; i++ {
			rb.Push(TelemetrySnapshot{Timestamp: int64(i)})
		}
		done <- struct{}{}
	}()
	go func() {
		for i := 50; i < 100; i++ {
			rb.Push(TelemetrySnapshot{Timestamp: int64(i)})
		}
		done <- struct{}{}
	}()

	<-done
	<-done

	if rb.Count() != 100 {
		t.Fatalf("expected count 100 after concurrent pushes, got %d", rb.Count())
	}
}

func TestRingBuffer_GetAllConcurrent(t *testing.T) {
	rb := NewRingBuffer(50)
	for i := 0; i < 25; i++ {
		rb.Push(TelemetrySnapshot{Timestamp: int64(i)})
	}

	done := make(chan struct{})
	go func() {
		for i := 25; i < 50; i++ {
			rb.Push(TelemetrySnapshot{Timestamp: int64(i)})
		}
		done <- struct{}{}
	}()

	// Read while writes happen concurrently
	all := rb.GetAll()
	if len(all) == 0 {
		t.Error("expected non-empty GetAll during concurrent writes")
	}

	<-done
}

func TestRingBuffer_SingleElement(t *testing.T) {
	rb := NewRingBuffer(1)

	rb.Push(TelemetrySnapshot{Timestamp: 1})
	if rb.Count() != 1 {
		t.Fatalf("expected count 1, got %d", rb.Count())
	}

	all := rb.GetAll()
	if len(all) != 1 || all[0].Timestamp != 1 {
		t.Fatalf("expected [1], got %v", timestamps(all))
	}

	// Push second element, first should be evicted
	rb.Push(TelemetrySnapshot{Timestamp: 2})
	all = rb.GetAll()
	if len(all) != 1 || all[0].Timestamp != 2 {
		t.Fatalf("expected [2] after eviction, got %v", timestamps(all))
	}
}

func TestRingBuffer_GetAllDoesNotMutate(t *testing.T) {
	rb := NewRingBuffer(3)
	rb.Push(TelemetrySnapshot{Timestamp: 1})
	rb.Push(TelemetrySnapshot{Timestamp: 2})

	// Call GetAll multiple times, should return consistent results
	first := rb.GetAll()
	second := rb.GetAll()

	if len(first) != len(second) {
		t.Fatalf("expected same length, got %d and %d", len(first), len(second))
	}
	for i := range first {
		if first[i].Timestamp != second[i].Timestamp {
			t.Fatalf("index %d: expected %d, got %d", i, first[i].Timestamp, second[i].Timestamp)
		}
	}
}

// ─── Engine Tests ────────────────────────────────────────────────────────────

// mockProviderStub implements TelemetryProvider for testing.
type mockProviderStub struct {
	sampleFn func(int32) (*TelemetrySnapshot, error)
	closeFn  func() error
}

func (m *mockProviderStub) Sample(pid int32) (*TelemetrySnapshot, error) {
	if m.sampleFn != nil {
		return m.sampleFn(pid)
	}
	return &TelemetrySnapshot{Timestamp: time.Now().Unix(), CPUPercent: 50, MemoryKB: 1000, Threads: 10}, nil
}

func (m *mockProviderStub) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}

func TestEngine_New(t *testing.T) {
	p := &mockProviderStub{}
	e := NewEngine(p, 300, time.Second)

	if e.Provider != p {
		t.Fatal("expected provider to be set")
	}
	if e.PID != -1 {
		t.Fatalf("expected PID -1, got %d", e.PID)
	}
	if e.Interval != time.Second {
		t.Fatalf("expected interval 1s, got %v", e.Interval)
	}
	if e.Running {
		t.Fatal("expected engine to not be running initially")
	}
	if e.Buffer == nil || e.Buffer.Count() != 0 {
		t.Fatal("expected empty ring buffer")
	}
}

func TestEngine_StartStop(t *testing.T) {
	p := &mockProviderStub{}
	e := NewEngine(p, 300, time.Second)

	cmd := e.Start()
	if !e.Running {
		t.Fatal("expected engine to be running after Start")
	}
	if cmd == nil {
		t.Fatal("expected non-nil tea.Cmd from Start")
	}

	e.Stop()
	if e.Running {
		t.Fatal("expected engine to be stopped after Stop")
	}
}

func TestEngine_SetTarget(t *testing.T) {
	p := &mockProviderStub{}
	e := NewEngine(p, 300, time.Second)

	if e.PID != -1 {
		t.Fatalf("expected initial PID -1, got %d", e.PID)
	}

	e.SetTarget(42)
	if e.PID != 42 {
		t.Fatalf("expected PID 42, got %d", e.PID)
	}
}

func TestEngine_Poll(t *testing.T) {
	p := &mockProviderStub{
		sampleFn: func(pid int32) (*TelemetrySnapshot, error) {
			return &TelemetrySnapshot{
				Timestamp:  1234567890,
				CPUPercent: 42.5,
				MemoryKB:   2048,
				Threads:    16,
			}, nil
		},
	}
	e := NewEngine(p, 300, time.Second)
	e.SetTarget(42)

	msg := e.Poll()
	tm, ok := msg.(TelemetryMsg)
	if !ok {
		t.Fatalf("expected TelemetryMsg, got %T", msg)
	}
	if tm.Error != nil {
		t.Fatalf("unexpected error: %v", tm.Error)
	}
	if tm.Snapshot.CPUPercent != 42.5 {
		t.Fatalf("expected CPU 42.5, got %f", tm.Snapshot.CPUPercent)
	}
	if tm.Snapshot.MemoryKB != 2048 {
		t.Fatalf("expected Memory 2048, got %d", tm.Snapshot.MemoryKB)
	}
	if tm.Snapshot.Threads != 16 {
		t.Fatalf("expected Threads 16, got %d", tm.Snapshot.Threads)
	}
	if e.Buffer.Count() != 1 {
		t.Fatalf("expected 1 snapshot in buffer, got %d", e.Buffer.Count())
	}
}

func TestEngine_Poll_NoTarget(t *testing.T) {
	p := &mockProviderStub{}
	e := NewEngine(p, 300, time.Second)
	// PID is -1 (unset)

	msg := e.Poll()
	tm, ok := msg.(TelemetryMsg)
	if !ok {
		t.Fatalf("expected TelemetryMsg, got %T", msg)
	}
	if tm.Error == nil {
		t.Fatal("expected error for unset PID")
	}
	if e.Buffer.Count() != 0 {
		t.Fatalf("expected 0 snapshots in buffer after error, got %d", e.Buffer.Count())
	}
}

func TestEngine_Poll_ProviderError(t *testing.T) {
	p := &mockProviderStub{
		sampleFn: func(pid int32) (*TelemetrySnapshot, error) {
			return nil, &mockProviderError{"provider failed"}
		},
	}
	e := NewEngine(p, 300, time.Second)
	e.SetTarget(42)

	msg := e.Poll()
	tm, ok := msg.(TelemetryMsg)
	if !ok {
		t.Fatalf("expected TelemetryMsg, got %T", msg)
	}
	if tm.Error == nil {
		t.Fatal("expected provider error")
	}
	if e.Buffer.Count() != 0 {
		t.Fatalf("expected 0 snapshots in buffer after error, got %d", e.Buffer.Count())
	}
}

type mockProviderError struct {
	msg string
}

func (m *mockProviderError) Error() string {
	return m.msg
}

func TestEngine_Close(t *testing.T) {
	closed := false
	p := &mockProviderStub{
		closeFn: func() error {
			closed = true
			return nil
		},
	}
	e := NewEngine(p, 300, time.Second)

	if err := e.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}
	if !closed {
		t.Fatal("expected provider Close to be called")
	}
	if e.Running {
		t.Fatal("expected engine to be stopped after Close")
	}
}

func TestEngine_Close_NilProvider(t *testing.T) {
	e := NewEngine(nil, 300, time.Second)
	if err := e.Close(); err != nil {
		t.Fatalf("expected no error closing nil provider, got: %v", err)
	}
}

func TestEngine_Poll_ConcurrentSafety(t *testing.T) {
	p := &mockProviderStub{}
	e := NewEngine(p, 100, time.Second)
	e.SetTarget(42)

	done := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			e.Poll()
		}
		done <- struct{}{}
	}()
	go func() {
		for i := 0; i < 50; i++ {
			e.Poll()
		}
		done <- struct{}{}
	}()
	<-done
	<-done

	if e.Buffer.Count() != 100 {
		t.Fatalf("expected 100 snapshots, got %d", e.Buffer.Count())
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func timestamps(snapshots []TelemetrySnapshot) []int64 {
	ts := make([]int64, len(snapshots))
	for i, s := range snapshots {
		ts[i] = s.Timestamp
	}
	return ts
}
