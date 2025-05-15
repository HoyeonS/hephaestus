package collector

import (
	"encoding/json"
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

// Parser handles log line parsing and error detection
type Parser struct {
	format     LogFormat
	patterns   []*regexp.Regexp
	timeFormat string
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
		format:     format,
		patterns:   compiledPatterns,
		timeFormat: timeFormat,
	}, nil
}

// ParseLine parses a log line and returns structured data
func (p *Parser) ParseLine(line string) (map[string]interface{}, error) {
	switch p.format {
	case FormatJSON:
		return p.parseJSON(line)
	case FormatText:
		return p.parseText(line)
	case FormatStructured:
		return p.parseStructured(line)
	default:
		return nil, nil
	}
}

func (p *Parser) parseJSON(line string) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(line), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (p *Parser) parseText(line string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// Try to extract timestamp
	if timestamp, rest := p.extractTimestamp(line); timestamp != nil {
		result["timestamp"] = timestamp
		line = rest
	}

	// Try to match error patterns
	for _, pattern := range p.patterns {
		if matches := pattern.FindStringSubmatch(line); matches != nil {
			for i, name := range pattern.SubexpNames() {
				if i != 0 && name != "" {
					result[name] = matches[i]
				}
			}
			break
		}
	}

	if len(result) == 0 {
		result["message"] = strings.TrimSpace(line)
	}

	return result, nil
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

func (p *Parser) extractTimestamp(line string) (*time.Time, string) {
	if p.timeFormat == "" {
		return nil, line
	}

	// Try to find timestamp at the start of the line
	if t, err := time.Parse(p.timeFormat, line[:len(p.timeFormat)]); err == nil {
		return &t, line[len(p.timeFormat):]
	}

	return nil, line
}
