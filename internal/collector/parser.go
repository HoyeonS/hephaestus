package collector

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// LogFormat represents supported log formats
type LogFormat int

const (
	FormatJSON LogFormat = iota
	FormatText
	FormatStructured
)

const (
	timestampLayout = "2006-01-02 15:04:05"
	timestampLen    = 19 // Length of "YYYY-MM-DD HH:MM:SS"
)

// Parser handles log line parsing and error detection
type Parser struct {
	format            LogFormat
	patterns          []*regexp.Regexp
	timeFormat        string
	timestampLayout   string
}

// NewParser creates a new log parser with the specified format
func NewParser(format LogFormat, patterns []string, timeFormat string) (*Parser, error) {
	compiledPatterns := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, err
		}
		compiledPatterns = append(compiledPatterns, re)
	}

	return &Parser{
		format:            format,
		patterns:          compiledPatterns,
		timeFormat:        timeFormat,
		timestampLayout:   timestampLayout,
	}, nil
}

// ParseLine parses a log line and returns a map containing the parsed fields
func (p *Parser) ParseLine(line string) (map[string]interface{}, error) {
	if strings.TrimSpace(line) == "" {
		return nil, fmt.Errorf("empty line")
	}

	result := make(map[string]interface{})

	// Parse text content
	text, err := p.parseText(line)
	if err != nil {
		return nil, fmt.Errorf("failed to parse text: %v", err)
	}
	result["message"] = text

	// Extract and parse timestamp
	timestamp, err := p.extractTimestamp(line)
	if err != nil {
		// If timestamp parsing fails, use current time
		result["timestamp"] = time.Now()
	} else {
		result["timestamp"] = timestamp
	}

	return result, nil
}

func (p *Parser) parseJSON(line string) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(line), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// parseText extracts the text content from the log line
func (p *Parser) parseText(line string) (string, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", fmt.Errorf("empty line")
	}

	// If line starts with timestamp, remove it
	if len(line) > timestampLen && p.isTimestamp(line[:timestampLen]) {
		line = strings.TrimSpace(line[timestampLen:])
	}

	return line, nil
}

func (p *Parser) parseStructured(line string) (map[string]interface{}, error) {
	parts := strings.Split(line, "|")
	result := make(map[string]interface{})

	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}

	return result, nil
}

// extractTimestamp attempts to extract and parse a timestamp from the log line
func (p *Parser) extractTimestamp(line string) (time.Time, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return time.Time{}, fmt.Errorf("empty line")
	}

	// Check if line is long enough to contain a timestamp
	if len(line) < timestampLen {
		return time.Time{}, fmt.Errorf("line too short for timestamp")
	}

	// Try to parse timestamp
	timestamp, err := time.Parse(p.timestampLayout, line[:timestampLen])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp format: %v", err)
	}

	return timestamp, nil
}

// isTimestamp checks if a string matches the timestamp format
func (p *Parser) isTimestamp(s string) bool {
	_, err := time.Parse(p.timestampLayout, s)
	return err == nil
}
