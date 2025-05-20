package node

import (
	"context"
	"testing"
	"time"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLogBuffer is a mock implementation of LogBuffer
type MockLogBuffer struct {
	mock.Mock
}

func (m *MockLogBuffer) Add(entry hephaestus.LogEntry) {
	m.Called(entry)
}

func (m *MockLogBuffer) GetEntries() []hephaestus.LogEntry {
	args := m.Called()
	return args.Get(0).([]hephaestus.LogEntry)
}

// MockThresholdMonitor is a mock implementation of ThresholdMonitor
type MockThresholdMonitor struct {
	mock.Mock
}

func (m *MockThresholdMonitor) Check() bool {
	args := m.Called()
	return args.Bool(0)
}

// MockSolutionGenerator is a mock implementation of SolutionGenerator
type MockSolutionGenerator struct {
	mock.Mock
}

func (m *MockSolutionGenerator) Generate(entries []hephaestus.LogEntry) (*hephaestus.Solution, error) {
	args := m.Called(entries)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*hephaestus.Solution), args.Error(1)
}

func TestNewNode(t *testing.T) {
	tests := []struct {
		name    string
		config  *hephaestus.SystemConfiguration
		wantErr bool
	}{
		{
			name: "valid configuration",
			config: &hephaestus.SystemConfiguration{
				LogSettings: hephaestus.LogSettings{
					ThresholdLevel:  "error",
					ThresholdCount:  3,
					ThresholdWindow: 5 * time.Minute,
				},
				Mode: "suggest",
			},
			wantErr: false,
		},
		{
			name:    "nil configuration",
			config:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := NewNode(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, node)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, node)
			}
		})
	}
}

func TestNode_ProcessLog(t *testing.T) {
	mockBuffer := new(MockLogBuffer)
	mockMonitor := new(MockThresholdMonitor)
	mockGenerator := new(MockSolutionGenerator)

	node := &Node{
		buffer:   mockBuffer,
		monitor:  mockMonitor,
		generator: mockGenerator,
		config: &hephaestus.SystemConfiguration{
			Mode: "suggest",
		},
	}

	tests := []struct {
		name    string
		entry   hephaestus.LogEntry
		setup   func()
		wantErr bool
	}{
		{
			name: "successful log processing",
			entry: hephaestus.LogEntry{
				Timestamp:   time.Now(),
				Level:       "error",
				Message:     "test error",
				ProcessedAt: time.Now(),
			},
			setup: func() {
				mockBuffer.On("Add", mock.Anything).Return()
				mockMonitor.On("Check").Return(false)
			},
			wantErr: false,
		},
		{
			name: "threshold met - solution generated",
			entry: hephaestus.LogEntry{
				Timestamp:   time.Now(),
				Level:       "error",
				Message:     "test error",
				ProcessedAt: time.Now(),
			},
			setup: func() {
				mockBuffer.On("Add", mock.Anything).Return()
				mockMonitor.On("Check").Return(true)
				mockBuffer.On("GetEntries").Return([]hephaestus.LogEntry{})
				mockGenerator.On("Generate", mock.Anything).Return(&hephaestus.Solution{}, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := node.ProcessLog(tt.entry)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNode_Start(t *testing.T) {
	node := &Node{
		config: &hephaestus.SystemConfiguration{
			Mode: "suggest",
		},
	}

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "successful start",
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "cancelled context",
			ctx:     func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := node.Start(tt.ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNode_GetSolutions(t *testing.T) {
	node := &Node{
		solutions: make(chan *hephaestus.Solution, 1),
	}

	// Test channel is returned
	solutions := node.GetSolutions()
	assert.NotNil(t, solutions)
}

func TestNode_GetErrors(t *testing.T) {
	node := &Node{
		errors: make(chan error, 1),
	}

	// Test channel is returned
	errors := node.GetErrors()
	assert.NotNil(t, errors)
} 