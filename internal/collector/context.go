package collector

import (
	"container/ring"
	"sync"
	"time"
)

// LogContext represents the context around a log entry
type LogContext struct {
	Before []map[string]interface{}
	After  []map[string]interface{}
}

// ContextManager maintains a circular buffer of recent log entries
type ContextManager struct {
	buffer     *ring.Ring
	size       int
	mu         sync.RWMutex
	timeWindow time.Duration
}

// NewContextManager creates a new context manager with specified buffer size
func NewContextManager(size int, timeWindow time.Duration) *ContextManager {
	return &ContextManager{
		buffer:     ring.New(size),
		size:       size,
		timeWindow: timeWindow,
	}
}

// AddEntry adds a new log entry to the context buffer
func (cm *ContextManager) AddEntry(entry map[string]interface{}) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.buffer.Value = entry
	cm.buffer = cm.buffer.Next()
}

// GetContext retrieves context around a specific timestamp
func (cm *ContextManager) GetContext(timestamp time.Time, beforeLines, afterLines int) *LogContext {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	context := &LogContext{
		Before: make([]map[string]interface{}, 0, beforeLines),
		After:  make([]map[string]interface{}, 0, afterLines),
	}

	// Create a slice of all entries within the time window
	entries := make([]map[string]interface{}, 0, cm.size)
	current := cm.buffer
	for i := 0; i < cm.size; i++ {
		if entry, ok := current.Value.(map[string]interface{}); ok && entry != nil {
			if ts, ok := entry["timestamp"].(time.Time); ok {
				if ts.Sub(timestamp) <= cm.timeWindow {
					entries = append(entries, entry)
				}
			}
		}
		current = current.Next()
	}

	// Find the target entry and collect context
	for i, entry := range entries {
		if ts, ok := entry["timestamp"].(time.Time); ok {
			if ts.Equal(timestamp) {
				// Collect before context
				start := max(0, i-beforeLines)
				context.Before = entries[start:i]

				// Collect after context
				end := min(len(entries), i+afterLines+1)
				if i+1 < end {
					context.After = entries[i+1:end]
				}
				break
			}
		}
	}

	return context
}

// Clear empties the context buffer
func (cm *ContextManager) Clear() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.buffer = ring.New(cm.size)
}

// GetTimeWindow returns the configured time window
func (cm *ContextManager) GetTimeWindow() time.Duration {
	return cm.timeWindow
}

// SetTimeWindow updates the time window for context collection
func (cm *ContextManager) SetTimeWindow(window time.Duration) {
	cm.timeWindow = window
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
