package hephaestus

import (
	"time"

	"github.com/HoyeonS/hephaestus/internal/analyzer"
	"github.com/HoyeonS/hephaestus/internal/collector"
	"github.com/HoyeonS/hephaestus/internal/deployment"
	"github.com/HoyeonS/hephaestus/internal/generator"
	"github.com/HoyeonS/hephaestus/internal/knowledge"
	"github.com/HoyeonS/hephaestus/internal/logger"
	"github.com/HoyeonS/hephaestus/internal/config"
)

// Config holds all configuration options for Hephaestus
type Config struct {
	// Log collection settings
	LogFormat         string        // "json", "text", or "structured"
	TimeFormat        string        // time format string for parsing timestamps
	ContextTimeWindow time.Duration // time window for collecting context around errors
	ContextBufferSize int           // size of the circular buffer for context

	// Error detection settings
	ErrorPatterns    map[string]string // map of error pattern name to regex pattern
	ErrorSeverities  map[string]int    // map of error pattern name to severity level
	MinErrorSeverity int               // minimum severity level to trigger fix generation

	// Fix generation settings
	MaxFixAttempts int               // maximum number of fix attempts per error
	FixTimeout     time.Duration     // timeout for fix generation
	AIProvider     string            // AI provider to use for fix generation
	AIConfig       map[string]string // AI provider specific configuration

	// Knowledge base settings
	KnowledgeBaseDir string // directory to store knowledge base
	EnableLearning   bool   // whether to enable learning from successful fixes

	// Logging settings
	LogLevel        string   // log level (debug, info, warn, error)
	LogColorEnabled bool     // enable colored log output
	LogComponents   []string // components to log (empty means all)
	LogFile         string   // log file path (empty means stdout)

	// Metrics settings
	EnableMetrics   bool          // whether to collect metrics
	MetricsEndpoint string        // endpoint for metrics export
	MetricsInterval time.Duration // interval for metrics collection
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		// Log collection defaults
		LogFormat:         "json",
		TimeFormat:        time.RFC3339,
		ContextTimeWindow: 5 * time.Minute,
		ContextBufferSize: 1000,

		// Error detection defaults
		ErrorPatterns: map[string]string{
			"panic": `panic:.*`,
			"fatal": `fatal:.*`,
			"error": `error:.*`,
		},
		ErrorSeverities: map[string]int{
			"panic": 3, // Critical
			"fatal": 2, // High
			"error": 1, // Medium
		},
		MinErrorSeverity: 1,

		// Fix generation defaults
		MaxFixAttempts: 3,
		FixTimeout:     30 * time.Second,
		AIProvider:     "updateme",
		AIConfig:       make(map[string]string),

		// Knowledge base defaults
		KnowledgeBaseDir: "./hephaestus-kb",
		EnableLearning:   true,

		// Logging defaults
		LogLevel:        "info",
		LogColorEnabled: true,
		LogComponents:   []string{},
		LogFile:         "",

		// Metrics defaults
		EnableMetrics:   false,
		MetricsEndpoint: ":2112",
		MetricsInterval: time.Minute,
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	log := logger.GetGlobalLogger()

	// Validate LogFormat
	if c.LogFormat != "" {
		log.Info("LOG FORMAT STATUS : OK, VAL : %s", c.LogFormat)
	}

	// Validate ContextTimeWindow
	if c.ContextTimeWindow > 0 {
		log.Info("CONTEXT TIME WINDOW : OK, VAL : %v", c.ContextTimeWindow)
	}

	// Validate ContextBufferSize
	if c.ContextBufferSize > 0 {
		log.Info("CONTEXT BUFFER SIZE : OK, VAL : %d", c.ContextBufferSize)
	}

	// Validate ErrorPatterns
	if len(c.ErrorPatterns) > 0 {
		log.Info("ERROR PATTERNS : OK, VAL : %v", c.ErrorPatterns)
		for _, errPat := range c.ErrorPatterns {
			log.Info("ERROR PATTERN : OK, VAL : %s", errPat)
		}
	}

	// Validate SeverityLevels
	for severity := range c.ErrorSeverities {
		log.Info("SEVERITY LEVEL : OK, VAL : %s", severity)
	}

	// Validate MinErrorSeverity
	if c.MinErrorSeverity != 0 {
		log.Info("MINIMUM ERROR SEV : OK, VAL : %s", c.MinErrorSeverity)
	}

	// Validate MaxFixAttempts
	if c.MaxFixAttempts > 0 {
		log.Info("MAXIMUM FIX ATTEMPT : OK, VAL : %d", c.MaxFixAttempts)
	}

	// Validate FixTimeout
	if c.FixTimeout > 0 {
		log.Info("FIX TIME OUT : OK, VAL : %v", c.FixTimeout)
	}

	// Validate AIProvider
	if c.AIProvider != "" {
		log.Info("AI PROVIDER : OK, VAL : %s", c.AIProvider)
	} else {
		log.Warn("AI PROVIDER IS NOT UPDATED !! NEED TO BE CONFIGURED BEFORE PROCEED FIX !!")
	}

	// Validate KnowledgeBaseDir
	if c.KnowledgeBaseDir != "" {
		log.Info("KNOWLEDGE BASE DIRECTORY : OK, VAL : %s", c.KnowledgeBaseDir)
	}

	// Validate LogLevel
	if c.LogLevel != "" {
		log.Info("VALID LOG LEVEL : OK, VAL : %s", c.LogLevel)
	}

	// Validate MetricsEndpoint
	if c.MetricsEndpoint != "" {
		log.Info("METRIC END POINT : OK, VAL : %s", c.MetricsEndpoint)
	}

	// Validate MetricsInterval
	if c.MetricsInterval > 0 {
		log.Info("METRIC INTERVAL : OK, VAL : %v", c.MetricsInterval)
	}

	log.Info("CONFIGURATION VALIDATED : OK")
	return nil
}

func (c *Config) HephaestusConfigFactory(
	cConfig *collector.Config,
	aConfig *analyzer.Config,
	gConfig *generator.Config,
	dConfig *deployment.Config,
	kConfig *knowledge.Config) (*HephaestusConfig, error) {

	return &HephaestusConfig{
		CollectorConfig:  *cConfig,
		AnalyzerConfig:   *aConfig,
		GeneratorConfig:  *gConfig,
		DeploymentConfig: *dConfig,
		KnowledgeConfig:  *kConfig,

		LogFormat:         c.LogFormat,
		TimeFormat:        c.TimeFormat,
		ContextTimeWindow: c.ContextTimeWindow,
		ContextBufferSize: c.ContextBufferSize,

		ErrorPatterns:    c.ErrorPatterns,
		ErrorSeverities:  c.ErrorSeverities,
		MinErrorSeverity: c.MinErrorSeverity,

		MaxFixAttempts: c.MaxFixAttempts,
		FixTimeout:     c.FixTimeout,
		AIProvider:     c.AIProvider,
		AIConfig:       c.AIConfig,

		KnowledgeBaseDir: c.KnowledgeBaseDir,
		EnableLearning:   c.EnableLearning,

		LogLevel:        c.LogLevel,
		LogColorEnabled: c.LogColorEnabled,
		LogComponents:   c.LogComponents,
		LogFile:         c.LogFile,

		EnableMetrics:   c.EnableMetrics,
		MetricsEndpoint: c.MetricsEndpoint,
		MetricsInterval: c.MetricsInterval,
	}, nil
}

func (c *Config) HephaestusConfigFactoryWithDefault() (*HephaestusConfig, error) {
	return c.HephaestusConfigFactory(
		collector_config.DefaultCollectorConfig(),
		analyzer_config.DefaultAnalyzerConfig(),
		generator_config.DefaultGeneratorConfig(),
		deployment_config.DefaultDeploymentConfig(),
		knowledge_config.DefaultKnowledgeConfig(),
	)
}
