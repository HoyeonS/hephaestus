package collector_test

import (
	"testing"
	"time"

	"github.com/HoyeonS/hephaestus/internal/collector"
	"github.com/HoyeonS/hephaestus/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetErrorChannel(t *testing.T) {
	tests := []struct {
		name           string
		bufferSize     int
		expectedBuffer int
	}{
		{
			name:           "default buffer size",
			bufferSize:     10,
			expectedBuffer: 10,
		},
		{
			name:           "large buffer size",
			bufferSize:     1000,
			expectedBuffer: 1000,
		},
		{
			name:           "minimum buffer size",
			bufferSize:     1,
			expectedBuffer: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create collector
			config := collector.Config{
				LogPaths:        []string{"test.log"},
				PollingInterval: 100 * time.Millisecond,
				BufferSize:      tt.bufferSize,
			}
			c, err := collector.New(config)
			require.NoError(t, err)

			// Get error channel
			errorChan := c.GetErrorChannel()
			require.NotNil(t, errorChan)

			// Test channel properties
			assert.Equal(t, tt.expectedBuffer, cap(errorChan))

			// Test channel is unique per call
			errorChan2 := c.GetErrorChannel()
			assert.Equal(t, errorChan, errorChan2, "should return same channel for multiple calls")
		})
	}
}

func TestGetErrorChannel_SendReceive(t *testing.T) {
	// Create collector
	config := collector.Config{
		LogPaths:        []string{"test.log"},
		PollingInterval: 100 * time.Millisecond,
		BufferSize:      2,
	}
	c, err := collector.New(config)
	require.NoError(t, err)

	// Get error channel
	errorChan := c.GetErrorChannel()
	require.NotNil(t, errorChan)

	// Test sending and receiving
	testError := &models.Error{
		Message:  "test error",
		Source:   "test.log",
		Severity: collector.SeverityHigh,
	}

	// Send error in goroutine to avoid blocking
	go func() {
		errorChan <- testError
	}()

	// Receive error
	select {
	case receivedError := <-errorChan:
		assert.Equal(t, testError, receivedError)
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for error")
	}
}

func TestGetErrorChannel_BufferBehavior(t *testing.T) {
	// Create collector with small buffer
	config := collector.Config{
		LogPaths:        []string{"test.log"},
		PollingInterval: 100 * time.Millisecond,
		BufferSize:      2,
	}
	c, err := collector.New(config)
	require.NoError(t, err)

	// Get error channel
	errorChan := c.GetErrorChannel()
	require.NotNil(t, errorChan)

	// Create test errors
	testErrors := []*models.Error{
		{
			Message:  "error 1",
			Source:   "test.log",
			Severity: collector.SeverityHigh,
		},
		{
			Message:  "error 2",
			Source:   "test.log",
			Severity: collector.SeverityHigh,
		},
	}

	// Fill buffer
	for _, err := range testErrors {
		errorChan <- err
	}

	// Verify buffer is full
	select {
	case err := <-errorChan:
		assert.Equal(t, testErrors[0], err)
	default:
		t.Error("should be able to receive from full buffer")
	}

	// Try to send to full buffer
	done := make(chan bool)
	go func() {
		errorChan <- &models.Error{
			Message:  "error 3",
			Source:   "test.log",
			Severity: collector.SeverityHigh,
		}
		done <- true
	}()

	// Verify send blocks
	select {
	case <-done:
		t.Error("should block on full buffer")
	case <-time.After(100 * time.Millisecond):
		// Expected behavior
	}

	// Drain channel
	for range testErrors {
		<-errorChan
	}
} 