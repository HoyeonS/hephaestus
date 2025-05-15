package health

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/HoyeonS/hephaestus/pkg/logger"
)

// ComponentStatus represents the health status of a component
type ComponentStatus struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"` // "healthy", "degraded", "failed"
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// SystemHealth represents the overall system health
type SystemHealth struct {
	Status     string            `json:"status"`
	Components []ComponentStatus `json:"components"`
	Timestamp  time.Time        `json:"timestamp"`
}

// Checker provides methods to check system health
type Checker struct {
	log            *logger.Logger
	aiEndpoint     string
	metricsPort    string
	knowledgeBase  string
	checkTimeout   time.Duration
	httpClient     *http.Client
}

// NewChecker creates a new health checker
func NewChecker(config map[string]string) *Checker {
	return &Checker{
		log:           logger.New("health", logger.INFO),
		aiEndpoint:    config["ai_endpoint"],
		metricsPort:   config["metrics_port"],
		knowledgeBase: config["kb_path"],
		checkTimeout:  5 * time.Second,
		httpClient:    &http.Client{Timeout: 5 * time.Second},
	}
}

// CheckHealth performs a comprehensive health check
func (c *Checker) CheckHealth(ctx context.Context) (*SystemHealth, error) {
	health := &SystemHealth{
		Status:     "healthy",
		Components: make([]ComponentStatus, 0),
		Timestamp:  time.Now(),
	}

	// Check AI provider connectivity
	aiStatus := c.checkAIProvider(ctx)
	health.Components = append(health.Components, aiStatus)

	// Check metrics endpoint
	metricsStatus := c.checkMetricsEndpoint(ctx)
	health.Components = append(health.Components, metricsStatus)

	// Check knowledge base
	kbStatus := c.checkKnowledgeBase(ctx)
	health.Components = append(health.Components, kbStatus)

	// Determine overall status
	for _, component := range health.Components {
		if component.Status == "failed" {
			health.Status = "failed"
			break
		} else if component.Status == "degraded" && health.Status != "failed" {
			health.Status = "degraded"
		}
	}

	return health, nil
}

// checkAIProvider checks AI provider connectivity
func (c *Checker) checkAIProvider(ctx context.Context) ComponentStatus {
	status := ComponentStatus{
		Name:      "ai_provider",
		Status:    "healthy",
		Timestamp: time.Now(),
	}

	if c.aiEndpoint == "" {
		status.Status = "failed"
		status.Message = "AI endpoint not configured"
		return status
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.aiEndpoint+"/health", nil)
	if err != nil {
		status.Status = "failed"
		status.Message = fmt.Sprintf("Failed to create request: %v", err)
		return status
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		status.Status = "failed"
		status.Message = fmt.Sprintf("Failed to connect: %v", err)
		return status
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		status.Status = "degraded"
		status.Message = fmt.Sprintf("Unhealthy response: %d", resp.StatusCode)
		return status
	}

	status.Message = "AI provider is responsive"
	return status
}

// checkMetricsEndpoint checks metrics endpoint health
func (c *Checker) checkMetricsEndpoint(ctx context.Context) ComponentStatus {
	status := ComponentStatus{
		Name:      "metrics",
		Status:    "healthy",
		Timestamp: time.Now(),
	}

	if c.metricsPort == "" {
		status.Status = "failed"
		status.Message = "Metrics port not configured"
		return status
	}

	url := fmt.Sprintf("http://localhost:%s/metrics", c.metricsPort)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		status.Status = "failed"
		status.Message = fmt.Sprintf("Failed to create request: %v", err)
		return status
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		status.Status = "failed"
		status.Message = fmt.Sprintf("Failed to connect: %v", err)
		return status
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		status.Status = "degraded"
		status.Message = fmt.Sprintf("Unhealthy response: %d", resp.StatusCode)
		return status
	}

	status.Message = "Metrics endpoint is responsive"
	return status
}

// checkKnowledgeBase checks knowledge base health
func (c *Checker) checkKnowledgeBase(ctx context.Context) ComponentStatus {
	status := ComponentStatus{
		Name:      "knowledge_base",
		Status:    "healthy",
		Timestamp: time.Now(),
	}

	if c.knowledgeBase == "" {
		status.Status = "failed"
		status.Message = "Knowledge base path not configured"
		return status
	}

	// Check if directory exists and is writable
	info, err := os.Stat(c.knowledgeBase)
	if err != nil {
		status.Status = "failed"
		status.Message = fmt.Sprintf("Failed to access knowledge base: %v", err)
		return status
	}

	if !info.IsDir() {
		status.Status = "failed"
		status.Message = "Knowledge base path is not a directory"
		return status
	}

	// Try to create a test file
	testFile := filepath.Join(c.knowledgeBase, ".test")
	f, err := os.Create(testFile)
	if err != nil {
		status.Status = "degraded"
		status.Message = fmt.Sprintf("Knowledge base is not writable: %v", err)
		return status
	}
	f.Close()
	os.Remove(testFile)

	status.Message = "Knowledge base is accessible and writable"
	return status
}

// PingComponents pings all components quickly
func (c *Checker) PingComponents(ctx context.Context) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.checkTimeout)
	defer cancel()

	// Check all components concurrently
	errCh := make(chan error, 3)
	
	go func() {
		if status := c.checkAIProvider(ctx); status.Status == "failed" {
			errCh <- fmt.Errorf("AI provider check failed: %s", status.Message)
			return
		}
		errCh <- nil
	}()

	go func() {
		if status := c.checkMetricsEndpoint(ctx); status.Status == "failed" {
			errCh <- fmt.Errorf("Metrics check failed: %s", status.Message)
			return
		}
		errCh <- nil
	}()

	go func() {
		if status := c.checkKnowledgeBase(ctx); status.Status == "failed" {
			errCh <- fmt.Errorf("Knowledge base check failed: %s", status.Message)
			return
		}
		errCh <- nil
	}()

	// Collect results
	for i := 0; i < 3; i++ {
		if err := <-errCh; err != nil {
			return fmt.Errorf("health check failed: %v", err)
		}
	}

	return nil
} 