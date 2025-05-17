package hephaestus

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/HoyeonS/hephaestus/internal/analyzer"
	"github.com/HoyeonS/hephaestus/internal/collector"
	"github.com/HoyeonS/hephaestus/internal/deployment"
	"github.com/HoyeonS/hephaestus/internal/generator"
	"github.com/HoyeonS/hephaestus/internal/knowledge"
	"github.com/HoyeonS/hephaestus/internal/models"
)

// HephaestusClient represents the Hephaestus client implementation
type HephaestusClient struct {
	collector  *collector.Service
	analyzer   *analyzer.Service
	generator  *generator.Service
	deployment *deployment.Service
	knowledge  *knowledge.Service

	fixChannel     chan *models.Fix
	suggestionChan chan *FixSuggestion
	errorChan      chan error

	mu     sync.RWMutex
	config *HephaestusConfig
}

// HephaestusConfig represents the client configuration
type HephaestusConfig struct {
	// Service-specific configurations
	CollectorConfig  collector.Config  `yaml:"collector"`  // Configuration for the collector service
	AnalyzerConfig   analyzer.Config   `yaml:"analyzer"`   // Configuration for the analyzer service
	GeneratorConfig  generator.Config  `yaml:"generator"`  // Configuration for the generator service
	DeploymentConfig deployment.Config `yaml:"deployment"` // Configuration for the deployment service
	KnowledgeConfig  knowledge.Config  `yaml:"knowledge"`  // Configuration for the knowledge base service

	// General settings
	LogFormat         string        `yaml:"log_format"`          // "json", "text", or "structured"
	TimeFormat        string        `yaml:"time_format"`         // time format string for parsing timestamps
	ContextTimeWindow time.Duration `yaml:"context_time_window"` // time window for collecting context around errors
	ContextBufferSize int           `yaml:"context_buffer_size"` // size of the circular buffer for context

	ErrorPatterns    map[string]string `yaml:"error_patterns"`     // map of error pattern name to regex pattern
	ErrorSeverities  map[string]int    `yaml:"error_severities"`   // map of error pattern name to severity level
	MinErrorSeverity int               `yaml:"min_error_severity"` // minimum severity level to trigger fix generation

	MaxFixAttempts int               `yaml:"max_fix_attempts"` // maximum number of fix attempts per error
	FixTimeout     time.Duration     `yaml:"fix_timeout"`      // timeout for fix generation
	AIProvider     string            `yaml:"ai_provider"`      // AI provider to use for fix generation
	AIConfig       map[string]string `yaml:"ai_config"`        // AI provider specific configuration

	KnowledgeBaseDir string `yaml:"knowledge_base_dir"` // directory to store knowledge base
	EnableLearning   bool   `yaml:"enable_learning"`    // whether to enable learning from successful fixes

	LogLevel        string   `yaml:"log_level"`         // log level (debug, info, warn, error)
	LogColorEnabled bool     `yaml:"log_color_enabled"` // enable colored log output
	LogComponents   []string `yaml:"log_components"`    // components to log (empty means all)
	LogFile         string   `yaml:"log_file"`          // log file path (empty means stdout)

	EnableMetrics   bool          `yaml:"enable_metrics"`   // whether to collect metrics
	MetricsEndpoint string        `yaml:"metrics_endpoint"` // endpoint for metrics export
	MetricsInterval time.Duration `yaml:"metrics_interval"` // interval for metrics collection
}

// FixSuggestion represents a suggested fix from Hephaestus
type FixSuggestion struct {
	ErrorID     string
	Description string
	CodeChanges []models.CodeChange
	Confidence  float64
	Metadata    map[string]interface{}
}

// New creates a new Hephaestus client
func New(config *HephaestusConfig) (Client, error) {
	collector, err := collector.New(config.CollectorConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create collector: %v", err)
	}

	analyzer, err := analyzer.New(config.AnalyzerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create analyzer: %v", err)
	}

	generator, err := generator.New(config.GeneratorConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create generator: %v", err)
	}

	deployment, err := deployment.New(config.DeploymentConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment: %v", err)
	}

	knowledge, err := knowledge.New(config.KnowledgeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create knowledge base: %v", err)
	}

	return &HephaestusClient{
		collector:      collector,
		analyzer:       analyzer,
		generator:      generator,
		deployment:     deployment,
		knowledge:      knowledge,
		fixChannel:     make(chan *models.Fix, 100),
		suggestionChan: make(chan *FixSuggestion, 100),
		errorChan:      make(chan error, 100),
		config:         config,
	}, nil
}

// Start starts the Hephaestus client
func (c *HephaestusClient) Start(ctx context.Context) error {
	// Start all services
	if err := c.collector.Start(ctx); err != nil {
		return fmt.Errorf("failed to start collector: %v", err)
	}

	if err := c.analyzer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start analyzer: %v", err)
	}

	if err := c.generator.Start(ctx); err != nil {
		return fmt.Errorf("failed to start generator: %v", err)
	}

	if err := c.deployment.Start(ctx); err != nil {
		return fmt.Errorf("failed to start deployment: %v", err)
	}

	if err := c.knowledge.Start(ctx); err != nil {
		return fmt.Errorf("failed to start knowledge base: %v", err)
	}

	// Start processing pipeline
	go c.processPipeline(ctx)

	return nil
}

// Stop stops the Hephaestus client
func (c *HephaestusClient) Stop(ctx context.Context) error {
	if err := c.collector.Stop(); err != nil {
		return fmt.Errorf("failed to stop collector: %v", err)
	}

	if err := c.analyzer.Stop(); err != nil {
		return fmt.Errorf("failed to stop analyzer: %v", err)
	}

	if err := c.generator.Stop(); err != nil {
		return fmt.Errorf("failed to stop generator: %v", err)
	}

	if err := c.deployment.Stop(); err != nil {
		return fmt.Errorf("failed to stop deployment: %v", err)
	}

	if err := c.knowledge.Stop(); err != nil {
		return fmt.Errorf("failed to stop knowledge base: %v", err)
	}

	return nil
}

// GetFixChannel returns the channel for receiving applied fixes
func (c *HephaestusClient) GetFixChannel() <-chan *models.Fix {
	return c.fixChannel
}

// GetSuggestionChannel returns the channel for receiving fix suggestions
func (c *HephaestusClient) GetSuggestionChannel() <-chan *FixSuggestion {
	return c.suggestionChan
}

// GetErrorChannel returns the channel for receiving errors
func (c *HephaestusClient) GetErrorChannel() <-chan error {
	return c.errorChan
}

// processPipeline processes the error detection and fix pipeline
func (c *HephaestusClient) processPipeline(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-c.collector.GetErrorChannel():
			if err == nil {
				continue
			}

			// Send error to analyzer
			c.analyzer.GetInputChannel() <- err

		case analyzedErr := <-c.analyzer.GetOutputChannel():
			if analyzedErr == nil {
				continue
			}

			// Send analyzed error to generator
			c.generator.GetInputChannel() <- analyzedErr

		case fix := <-c.generator.GetOutputChannel():
			if fix == nil {
				continue
			}

			// Check if we should apply the fix directly or suggest it
			if c.config.AIProvider == "direct" {
				c.deployment.GetInputChannel() <- fix
			} else {
				// Convert fix to suggestion
				suggestion := &FixSuggestion{
					ErrorID:     fix.ErrorID,
					Description: fix.Description,
					CodeChanges: fix.CodeChanges,
					Confidence:  fix.Confidence,
					Metadata:    fix.Metadata,
				}

				select {
				case c.suggestionChan <- suggestion:
				default:
					c.errorChan <- fmt.Errorf("suggestion channel full, dropping suggestion for error %s", fix.ErrorID)
				}
			}

		case deployedFix := <-c.deployment.GetOutputChannel():
			if deployedFix == nil {
				continue
			}

			// Send deployed fix to knowledge base
			c.knowledge.GetInputChannel() <- deployedFix

			// Send fix to client
			select {
			case c.fixChannel <- deployedFix:
			default:
				c.errorChan <- fmt.Errorf("fix channel full, dropping fix for error %s", deployedFix.ErrorID)
			}
		}
	}
}

// MonitorReader starts monitoring a reader for errors
func (c *HephaestusClient) MonitorReader(ctx context.Context, reader io.Reader, source string) error {
	// Implementation would monitor the reader for errors
	return fmt.Errorf("monitor reader not implemented")
}

// MonitorCommand starts monitoring a command's output for errors
func (c *HephaestusClient) MonitorCommand(ctx context.Context, name string, args ...string) (<-chan *Error, error) {
	// Implementation would monitor command output for errors
	return nil, fmt.Errorf("monitor command not implemented")
}

// AddErrorPattern adds a new error pattern to detect
func (c *HephaestusClient) AddErrorPattern(pattern string, severity int) error {
	// Implementation would add a new error pattern
	return fmt.Errorf("add error pattern not implemented")
}

// RemoveErrorPattern removes an error pattern
func (c *HephaestusClient) RemoveErrorPattern(pattern string) error {
	// Implementation would remove an error pattern
	return fmt.Errorf("remove error pattern not implemented")
}

// GetMetrics returns current metrics
func (c *HephaestusClient) GetMetrics() (*Metrics, error) {
	// Implementation would return current metrics
	return nil, fmt.Errorf("get metrics not implemented")
}

// Ping performs a quick connectivity check to all components
func (c *HephaestusClient) Ping(ctx context.Context) error {
	// Implementation would check connectivity
	return fmt.Errorf("ping not implemented")
}

// CheckHealth performs a comprehensive health check of all components
func (c *HephaestusClient) CheckHealth(ctx context.Context) (*SystemHealth, error) {
	// Implementation would perform health check
	return nil, fmt.Errorf("health check not implemented")
}

// TestConnectivity performs a basic connectivity test
func (c *HephaestusClient) TestConnectivity(ctx context.Context) error {
	// Implementation would test connectivity
	return fmt.Errorf("connectivity test not implemented")
}
