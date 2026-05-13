package engine

import (
	"fmt"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// RingBuffer is a fixed-capacity circular buffer for telemetry snapshots.
type RingBuffer struct {
	data    []TelemetrySnapshot
	head    int // next write position
	count   int
	capacity int
	mu      sync.RWMutex
}

// NewRingBuffer creates a ring buffer with the given capacity.
func NewRingBuffer(capacity int) *RingBuffer {
	return &RingBuffer{
		data:     make([]TelemetrySnapshot, capacity),
		capacity: capacity,
	}
}

// Push adds a snapshot to the buffer, evicting the oldest if full.
func (rb *RingBuffer) Push(s TelemetrySnapshot) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.data[rb.head] = s
	rb.head = (rb.head + 1) % rb.capacity
	if rb.count < rb.capacity {
		rb.count++
	}
}

// GetAll returns all snapshots in chronological order.
func (rb *RingBuffer) GetAll() []TelemetrySnapshot {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.count == 0 {
		return nil
	}

	result := make([]TelemetrySnapshot, rb.count)
	start := (rb.head - rb.count + rb.capacity) % rb.capacity
	for i := 0; i < rb.count; i++ {
		result[i] = rb.data[(start+i)%rb.capacity]
	}
	return result
}

// Latest returns the most recent snapshot, or nil if empty.
func (rb *RingBuffer) Latest() *TelemetrySnapshot {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.count == 0 {
		return nil
	}
	lastIdx := (rb.head - 1 + rb.capacity) % rb.capacity
	s := rb.data[lastIdx]
	return &s
}

// Count returns the number of snapshots in the buffer.
func (rb *RingBuffer) Count() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.count
}

// TickMsg is sent by the engine at each polling interval.
type TickMsg struct{}

// TelemetryMsg carries a new telemetry snapshot to the TUI.
type TelemetryMsg struct {
	Snapshot TelemetrySnapshot
	Error    error
}

// tickCmd returns a command that fires after the given duration.
func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

// Engine manages the telemetry polling loop and ring buffer.
type Engine struct {
	Provider  TelemetryProvider
	Buffer    *RingBuffer
	PID       int32
	Interval  time.Duration
	Running   bool
	mu        sync.Mutex
}

// NewEngine creates a new engine with the given provider and buffer capacity.
func NewEngine(provider TelemetryProvider, capacity int, interval time.Duration) *Engine {
	return &Engine{
		Provider: provider,
		Buffer:   NewRingBuffer(capacity),
		Interval: interval,
		PID:      -1, // no target PID yet
	}
}

// Start begins the polling loop. Returns a tea.Cmd for use with Bubble Tea.
func (e *Engine) Start() tea.Cmd {
	e.mu.Lock()
	e.Running = true
	e.mu.Unlock()
	return tickCmd(e.Interval)
}

// Stop halts the polling loop.
func (e *Engine) Stop() {
	e.mu.Lock()
	e.Running = false
	e.mu.Unlock()
}

// SetTarget sets the PID to poll.
func (e *Engine) SetTarget(pid int32) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.PID = pid
}

// Poll performs a single telemetry sample and returns a TelemetryMsg.
func (e *Engine) Poll() tea.Msg {
	e.mu.Lock()
	pid := e.PID
	provider := e.Provider
	e.mu.Unlock()

	if pid < 0 {
		return TelemetryMsg{
			Error: fmt.Errorf("no target PID set"),
		}
	}

	snapshot, err := provider.Sample(pid)
	if err != nil {
		return TelemetryMsg{Error: err}
	}

	e.Buffer.Push(*snapshot)
	return TelemetryMsg{Snapshot: *snapshot}
}

// Close releases engine resources.
func (e *Engine) Close() error {
	e.Stop()
	if e.Provider != nil {
		return e.Provider.Close()
	}
	return nil
}
