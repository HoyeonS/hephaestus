package knowledge

import (
	"time"
)

// Documentation manages error and fix documentation
type Documentation struct {
	records map[string]*Record
}

// Record represents a documented error and its fix
type Record struct {
	ErrorType    string
	Description  string
	Solution     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewDocumentation creates a new documentation manager
func NewDocumentation() *Documentation {
	return &Documentation{
		records: make(map[string]*Record),
	}
}

// AddRecord adds a new documentation record
func (d *Documentation) AddRecord(errorType string, description string, solution string) *Record {
	record := &Record{
		ErrorType:    errorType,
		Description:  description,
		Solution:     solution,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	d.records[errorType] = record
	return record
}
