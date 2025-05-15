package models

import (
	"time"
)

// FixStatus represents the current status of a fix
type FixStatus string

const (
	FixPending    FixStatus = "pending"
	FixInProgress FixStatus = "in_progress"
	FixApplied    FixStatus = "applied"
	FixFailed     FixStatus = "failed"
	FixRollback   FixStatus = "rollback"
	FixVerified   FixStatus = "verified"
)

// FixStrategy represents the type of fix strategy used
type FixStrategy string

const (
	NullCheck          FixStrategy = "null_check"
	ExceptionHandling  FixStrategy = "exception_handling"
	ResourceCleanup    FixStrategy = "resource_cleanup"
	TypeConversion     FixStrategy = "type_conversion"
	ConditionInversion FixStrategy = "condition_inversion"
	CodeRefactoring    FixStrategy = "code_refactoring"
)

// Fix represents a generated fix for an error
type Fix struct {
	ID            string      `json:"id"`
	ErrorID       string      `json:"error_id"`           // Reference to the error being fixed
	Strategy      FixStrategy `json:"strategy"`           // Strategy used to generate the fix
	Status        FixStatus   `json:"status"`            // Current status of the fix
	CreatedAt     time.Time   `json:"created_at"`        // When the fix was generated
	AppliedAt     *time.Time  `json:"applied_at"`        // When the fix was applied (if applicable)
	VerifiedAt    *time.Time  `json:"verified_at"`       // When the fix was verified (if applicable)
	CodeChanges   []CodeChange `json:"code_changes"`      // List of code changes to apply
	TestResults   []TestResult `json:"test_results"`      // Results of verification tests
	Confidence    float64     `json:"confidence"`        // Confidence score of the fix (0-1)
	Description   string      `json:"description"`       // Human-readable description of the fix
	Metadata      map[string]interface{} `json:"metadata"` // Additional fix-specific metadata
	RollbackData  *RollbackData `json:"rollback_data,omitempty"` // Data needed for rollback
}

// CodeChange represents a single code change in a fix
type CodeChange struct {
	FilePath    string `json:"file_path"`    // Path to the file to modify
	StartLine   int    `json:"start_line"`   // Starting line number
	EndLine     int    `json:"end_line"`     // Ending line number
	OldCode     string `json:"old_code"`     // Original code
	NewCode     string `json:"new_code"`     // Modified code
	ChangeType  string `json:"change_type"`  // Type of change (add, modify, delete)
	Description string `json:"description"`   // Description of the change
}

// TestResult represents the result of a verification test
type TestResult struct {
	Name        string    `json:"name"`         // Name of the test
	Status      string    `json:"status"`       // Pass/Fail/Error
	Duration    float64   `json:"duration"`     // Test duration in seconds
	Output      string    `json:"output"`       // Test output/logs
	ExecutedAt  time.Time `json:"executed_at"`  // When the test was run
	ErrorDetail string    `json:"error_detail"` // Details if test failed
}

// RollbackData contains information needed to rollback a fix
type RollbackData struct {
	Backup      map[string]string `json:"backup"`       // Backup of modified files
	CheckPoint  string           `json:"check_point"`  // System state checkpoint
	ValidateCmd string           `json:"validate_cmd"` // Command to validate rollback
}

// NewFix creates a new Fix instance
func NewFix(errorID string, strategy FixStrategy) *Fix {
	return &Fix{
		ID:          generateUUID(),
		ErrorID:     errorID,
		Strategy:    strategy,
		Status:      FixPending,
		CreatedAt:   time.Now(),
		CodeChanges: make([]CodeChange, 0),
		TestResults: make([]TestResult, 0),
		Metadata:    make(map[string]interface{}),
	}
}

// AddCodeChange adds a new code change to the fix
func (f *Fix) AddCodeChange(change CodeChange) {
	f.CodeChanges = append(f.CodeChanges, change)
}

// AddTestResult adds a test result to the fix
func (f *Fix) AddTestResult(result TestResult) {
	f.TestResults = append(f.TestResults, result)
}

// UpdateStatus updates the fix status and relevant timestamps
func (f *Fix) UpdateStatus(status FixStatus) {
	f.Status = status
	now := time.Now()
	
	switch status {
	case FixApplied:
		f.AppliedAt = &now
	case FixVerified:
		f.VerifiedAt = &now
	}
}

// SetRollbackData sets the rollback data for the fix
func (f *Fix) SetRollbackData(backup map[string]string, checkPoint, validateCmd string) {
	f.RollbackData = &RollbackData{
		Backup:      backup,
		CheckPoint:  checkPoint,
		ValidateCmd: validateCmd,
	}
}

// IsSuccessful returns whether the fix was successfully applied and verified
func (f *Fix) IsSuccessful() bool {
	return f.Status == FixVerified
}
