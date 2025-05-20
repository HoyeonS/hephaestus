package log

import (
	"testing"
	"time"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	"github.com/stretchr/testify/assert"
)

func TestLogBuffer_Add(t *testing.T) {
	buffer := NewLogBuffer(&hephaestus.SystemConfiguration{
		LogSettings: hephaestus.LogSettings{
			ThresholdWindow: 5 * time.Minute,
		},
	})

	tests := []struct {
		name  string
		entry hephaestus.LogEntry
	}{
		{
			name: "add error log",
			entry: hephaestus.LogEntry{
				Timestamp:   time.Now(),
				Level:       "error",
				Message:     "test error",
				ProcessedAt: time.Now(),
			},
		},
		{
			name: "add info log",
			entry: hephaestus.LogEntry{
				Timestamp:   time.Now(),
				Level:       "info",
				Message:     "test info",
				ProcessedAt: time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer.Add(tt.entry)
			entries := buffer.GetEntries()
			assert.Contains(t, entries, tt.entry)
		})
	}
}

func TestLogBuffer_GetEntries(t *testing.T) {
	buffer := NewLogBuffer(&hephaestus.SystemConfiguration{
		LogSettings: hephaestus.LogSettings{
			ThresholdWindow: 5 * time.Minute,
		},
	})

	// Add multiple entries
	entries := []hephaestus.LogEntry{
		{
			Timestamp:   time.Now(),
			Level:       "error",
			Message:     "error 1",
			ProcessedAt: time.Now(),
		},
		{
			Timestamp:   time.Now(),
			Level:       "error",
			Message:     "error 2",
			ProcessedAt: time.Now(),
		},
		{
			Timestamp:   time.Now(),
			Level:       "info",
			Message:     "info 1",
			ProcessedAt: time.Now(),
		},
	}

	for _, entry := range entries {
		buffer.Add(entry)
	}

	// Test GetEntries
	retrieved := buffer.GetEntries()
	assert.Equal(t, len(entries), len(retrieved))
	for _, entry := range entries {
		assert.Contains(t, retrieved, entry)
	}
}

func TestLogBuffer_Cleanup(t *testing.T) {
	buffer := NewLogBuffer(&hephaestus.SystemConfiguration{
		LogSettings: hephaestus.LogSettings{
			ThresholdWindow: 5 * time.Minute,
		},
	})

	// Add old entries
	oldTime := time.Now().Add(-6 * time.Minute)
	oldEntry := hephaestus.LogEntry{
		Timestamp:   oldTime,
		Level:       "error",
		Message:     "old error",
		ProcessedAt: oldTime,
	}
	buffer.Add(oldEntry)

	// Add recent entries
	recentTime := time.Now()
	recentEntry := hephaestus.LogEntry{
		Timestamp:   recentTime,
		Level:       "error",
		Message:     "recent error",
		ProcessedAt: recentTime,
	}
	buffer.Add(recentEntry)

	// Cleanup old entries
	buffer.Cleanup()

	// Verify only recent entries remain
	entries := buffer.GetEntries()
	assert.NotContains(t, entries, oldEntry)
	assert.Contains(t, entries, recentEntry)
}

func TestLogBuffer_ConcurrentAccess(t *testing.T) {
	buffer := NewLogBuffer(&hephaestus.SystemConfiguration{
		LogSettings: hephaestus.LogSettings{
			ThresholdWindow: 5 * time.Minute,
		},
	})

	// Test concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			entry := hephaestus.LogEntry{
				Timestamp:   time.Now(),
				Level:       "error",
				Message:     "concurrent error",
				ProcessedAt: time.Now(),
			}
			buffer.Add(entry)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all entries were added
	entries := buffer.GetEntries()
	assert.Equal(t, 10, len(entries))
}

func TestLogBuffer_ThresholdWindow(t *testing.T) {
	tests := []struct {
		name           string
		window         time.Duration
		entries        []hephaestus.LogEntry
		expectedCount  int
	}{
		{
			name:   "entries within window",
			window: 5 * time.Minute,
			entries: []hephaestus.LogEntry{
				{
					Timestamp:   time.Now(),
					Level:       "error",
					Message:     "error 1",
					ProcessedAt: time.Now(),
				},
				{
					Timestamp:   time.Now().Add(-2 * time.Minute),
					Level:       "error",
					Message:     "error 2",
					ProcessedAt: time.Now().Add(-2 * time.Minute),
				},
			},
			expectedCount: 2,
		},
		{
			name:   "entries outside window",
			window: 5 * time.Minute,
			entries: []hephaestus.LogEntry{
				{
					Timestamp:   time.Now(),
					Level:       "error",
					Message:     "error 1",
					ProcessedAt: time.Now(),
				},
				{
					Timestamp:   time.Now().Add(-6 * time.Minute),
					Level:       "error",
					Message:     "error 2",
					ProcessedAt: time.Now().Add(-6 * time.Minute),
				},
			},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := NewLogBuffer(&hephaestus.SystemConfiguration{
				LogSettings: hephaestus.LogSettings{
					ThresholdWindow: tt.window,
				},
			})

			for _, entry := range tt.entries {
				buffer.Add(entry)
			}

			buffer.Cleanup()
			entries := buffer.GetEntries()
			assert.Equal(t, tt.expectedCount, len(entries))
		})
	}
} 