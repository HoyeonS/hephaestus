package solution

import (
	"testing"
	"time"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	"github.com/stretchr/testify/assert"
)

func TestSolutionGenerator_Generate(t *testing.T) {
	generator := NewSolutionGenerator(&hephaestus.SystemConfiguration{
		Mode: "suggest",
	})

	tests := []struct {
		name    string
		entries []hephaestus.LogEntry
		want    *hephaestus.Solution
		wantErr bool
	}{
		{
			name: "generate solution from error logs",
			entries: []hephaestus.LogEntry{
				{
					Timestamp:   time.Now(),
					Level:       "error",
					Message:     "database connection failed",
					Context:     map[string]interface{}{"database": "main"},
					ProcessedAt: time.Now(),
				},
				{
					Timestamp:   time.Now(),
					Level:       "error",
					Message:     "database connection failed",
					Context:     map[string]interface{}{"database": "main"},
					ProcessedAt: time.Now(),
				},
			},
			wantErr: false,
		},
		{
			name:    "empty entries",
			entries: []hephaestus.LogEntry{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			solution, err := generator.Generate(tt.entries)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, solution)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, solution)
				assert.NotEmpty(t, solution.ID)
				assert.NotEmpty(t, solution.Description)
				assert.NotEmpty(t, solution.CodeChanges)
				assert.NotZero(t, solution.Confidence)
				assert.Equal(t, tt.entries[len(tt.entries)-1], solution.LogEntry)
			}
		})
	}
}

func TestSolutionGenerator_AnalyzePatterns(t *testing.T) {
	generator := NewSolutionGenerator(&hephaestus.SystemConfiguration{
		Mode: "suggest",
	})

	tests := []struct {
		name    string
		entries []hephaestus.LogEntry
		want    []string
	}{
		{
			name: "detect repeated error pattern",
			entries: []hephaestus.LogEntry{
				{
					Timestamp:   time.Now(),
					Level:       "error",
					Message:     "database connection failed",
					Context:     map[string]interface{}{"database": "main"},
					ProcessedAt: time.Now(),
				},
				{
					Timestamp:   time.Now(),
					Level:       "error",
					Message:     "database connection failed",
					Context:     map[string]interface{}{"database": "main"},
					ProcessedAt: time.Now(),
				},
			},
			want: []string{"repeated_error", "database_connection"},
		},
		{
			name: "detect sequence pattern",
			entries: []hephaestus.LogEntry{
				{
					Timestamp:   time.Now(),
					Level:       "error",
					Message:     "authentication failed",
					Context:     map[string]interface{}{"user": "test"},
					ProcessedAt: time.Now(),
				},
				{
					Timestamp:   time.Now(),
					Level:       "error",
					Message:     "database connection failed",
					Context:     map[string]interface{}{"database": "main"},
					ProcessedAt: time.Now(),
				},
			},
			want: []string{"error_sequence", "authentication_database"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns := generator.analyzePatterns(tt.entries)
			assert.Equal(t, tt.want, patterns)
		})
	}
}

func TestSolutionGenerator_GenerateChanges(t *testing.T) {
	generator := NewSolutionGenerator(&hephaestus.SystemConfiguration{
		Mode: "suggest",
	})

	tests := []struct {
		name     string
		patterns []string
		want     []hephaestus.CodeChange
	}{
		{
			name:     "generate changes for database error",
			patterns: []string{"database_connection"},
			want: []hephaestus.CodeChange{
				{
					File:    "database.go",
					Type:    "add",
					Content: "// Add connection retry logic",
				},
			},
		},
		{
			name:     "generate changes for authentication error",
			patterns: []string{"authentication"},
			want: []hephaestus.CodeChange{
				{
					File:    "auth.go",
					Type:    "modify",
					Content: "// Add error handling for authentication",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := generator.generateChanges(tt.patterns)
			assert.Equal(t, tt.want, changes)
		})
	}
}

func TestSolutionGenerator_CalculateConfidence(t *testing.T) {
	generator := NewSolutionGenerator(&hephaestus.SystemConfiguration{
		Mode: "suggest",
	})

	tests := []struct {
		name      string
		patterns  []string
		changes   []hephaestus.CodeChange
		wantMin   float64
		wantMax   float64
	}{
		{
			name:      "high confidence for known pattern",
			patterns:  []string{"database_connection"},
			changes:   []hephaestus.CodeChange{{Type: "add", Content: "// Add retry logic"}},
			wantMin:   0.8,
			wantMax:   1.0,
		},
		{
			name:      "low confidence for unknown pattern",
			patterns:  []string{"unknown_pattern"},
			changes:   []hephaestus.CodeChange{{Type: "add", Content: "// Generic fix"}},
			wantMin:   0.0,
			wantMax:   0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := generator.calculateConfidence(tt.patterns, tt.changes)
			assert.GreaterOrEqual(t, confidence, tt.wantMin)
			assert.LessOrEqual(t, confidence, tt.wantMax)
		})
	}
}

func TestSolutionGenerator_GenerateDescription(t *testing.T) {
	generator := NewSolutionGenerator(&hephaestus.SystemConfiguration{
		Mode: "suggest",
	})

	tests := []struct {
		name     string
		patterns []string
		want     string
	}{
		{
			name:     "description for database error",
			patterns: []string{"database_connection"},
			want:     "Fix database connection issues by adding retry logic",
		},
		{
			name:     "description for authentication error",
			patterns: []string{"authentication"},
			want:     "Improve authentication error handling",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			description := generator.generateDescription(tt.patterns)
			assert.Equal(t, tt.want, description)
		})
	}
} 