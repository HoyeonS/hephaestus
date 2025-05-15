package hephaestus

import (
	"context"
	"fmt"
	"sync"

	"github.com/HoyeonS/hephaestus/internal/analyzer"
	"github.com/HoyeonS/hephaestus/internal/collector"
	"github.com/HoyeonS/hephaestus/internal/deployment"
	"github.com/HoyeonS/hephaestus/internal/generator"
	"github.com/HoyeonS/hephaestus/internal/knowledge"
	"github.com/HoyeonS/hephaestus/internal/models"
)

// Client represents the Hephaestus client
type Client struct {
	collector  *collector.Service
	analyzer   *analyzer.Service
	generator  *generator.Service
	deployment *deployment.Service
	knowledge  *knowledge.Service
	
	fixChannel     chan *models.Fix
	suggestionChan chan *FixSuggestion
	errorChan     chan error
	
	mu     sync.RWMutex
	config *Config
}

// Config represents the client configuration
type Config struct {
	CollectorConfig  collector.Config  `yaml:"collector"`
	AnalyzerConfig   analyzer.Config   `yaml:"analyzer"`
	GeneratorConfig  generator.Config  `yaml:"generator"`
	DeploymentConfig deployment.Config `yaml:"deployment"`
	KnowledgeConfig  knowledge.Config  `yaml:"knowledge"`
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
func New(config *Config) (*Client, error) {
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

	return &Client{
		collector:      collector,
		analyzer:       analyzer,
		generator:      generator,
		deployment:     deployment,
		knowledge:      knowledge,
		fixChannel:     make(chan *models.Fix, 100),
		suggestionChan: make(chan *FixSuggestion, 100),
		errorChan:     make(chan error, 100),
		config:        config,
	}, nil
}

// Start starts the Hephaestus client
func (c *Client) Start(ctx context.Context) error {
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
func (c *Client) Stop() error {
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
func (c *Client) GetFixChannel() <-chan *models.Fix {
	return c.fixChannel
}

// GetSuggestionChannel returns the channel for receiving fix suggestions
func (c *Client) GetSuggestionChannel() <-chan *FixSuggestion {
	return c.suggestionChan
}

// GetErrorChannel returns the channel for receiving errors
func (c *Client) GetErrorChannel() <-chan error {
	return c.errorChan
}

// processPipeline processes the error detection and fix pipeline
func (c *Client) processPipeline(ctx context.Context) {
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
			if c.config.GeneratorConfig.AIModel.FixMode == "direct" {
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