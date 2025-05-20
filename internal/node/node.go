package node

import (
	"context"
	"fmt"
	"time"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
)

// Node represents a Hephaestus node
type Node struct {
	systemConfig     *hephaestus.SystemConfiguration
	clientNodeConfig *hephaestus.ClientNodeConfiguration
	status           hephaestus.NodeStatus
	// Log processing
	logBuffer     []hephaestus.LogEntry
	lastProcessed time.Time

	// Solution processing
	solutionChan chan *hephaestus.Solution
	errorChan    chan error
}

// NewNode creates a new Hephaestus node
func NewNode(systemConfig *hephaestus.SystemConfiguration, clientNodeConfig *hephaestus.ClientNodeConfiguration) (*Node, error) {
	if err := hephaestus.ValidateSystemConfiguration(systemConfig); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	return &Node{
		systemConfig:     systemConfig,
		clientNodeConfig: clientNodeConfig,
		status:           hephaestus.NodeStatusInitializing,
		logBuffer:        make([]hephaestus.LogEntry, 0),
		solutionChan:     make(chan *hephaestus.Solution, 100),
		errorChan:        make(chan error, 100),
		lastProcessed:    time.Now(),
	}, nil
}

// Start initializes and starts the node
func (n *Node) Start(ctx context.Context) error {
	n.status = hephaestus.NodeStatusOperational
	// Start log processing
	go n.processLogs(ctx)

	return nil
}

// Stop gracefully stops the node
func (n *Node) Stop(ctx context.Context) error {
	n.status = hephaestus.NodeStatusError

	// Close channels
	close(n.solutionChan)
	close(n.errorChan)

	return nil
}

// ProcessLog processes a new log entry
func (n *Node) ProcessLog(entry hephaestus.LogEntry) error {

	// Check if log chunk exceeded then remove the old logs
	if len(n.logBuffer) == n.systemConfig.LimitConfiguration.LogChunkLimit {
		n.logBuffer = n.logBuffer[1:]
	}
	// Add to buffer
	n.logBuffer = append(n.logBuffer, entry)

	// Check if we need to process logs
	if n.shouldProcessLogs(entry) {
		return n.triggerLogProcessing()
	}

	return nil
}

// shouldProcessLogs checks if we should process logs based on threshold
func (n *Node) shouldProcessLogs(entry hephaestus.LogEntry) bool {
	// need to implement logic to cover higher cases
	return n.clientNodeConfig.LogProcessingConfiguration.ThresholdLevel == entry.Level
}

// triggerLogProcessing triggers log processing
func (n *Node) triggerLogProcessing() error {
	n.statusLock.Lock()
	n.status = hephaestus.NodeStatusProcessing
	n.statusLock.Unlock()

	// Process logs in a separate goroutine
	go func() {
		defer func() {
			n.statusLock.Lock()
			n.status = hephaestus.NodeStatusOperational
			n.statusLock.Unlock()
		}()

		// Clear buffer after processing
		n.logBufferLock.Lock()
		entries := make([]hephaestus.LogEntry, len(n.logBuffer))
		copy(entries, n.logBuffer)
		n.logBuffer = make([]hephaestus.LogEntry, 0)
		n.logBufferLock.Unlock()

		// Generate solution
		solution, err := n.generateSolution(entries)
		if err != nil {
			n.errorChan <- fmt.Errorf("failed to generate solution: %v", err)
			return
		}

		// Send solution for processing
		n.solutionChan <- solution
	}()

	return nil
}

// generateSolution generates a solution based on log entries
func (n *Node) generateSolution(entries []hephaestus.LogEntry) (*hephaestus.Solution, error) {
	// TODO: Implement solution generation logic
	return &hephaestus.Solution{
		ID:          fmt.Sprintf("sol-%d", time.Now().UnixNano()),
		LogEntry:    entries[len(entries)-1],
		Description: "Generated solution",
		GeneratedAt: time.Now(),
		Confidence:  0.8,
	}, nil
}

// processLogs continuously processes logs
func (n *Node) processLogs(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if n.shouldProcessLogs() {
				if err := n.triggerLogProcessing(); err != nil {
					n.errorChan <- fmt.Errorf("failed to trigger log processing: %v", err)
				}
			}
		}
	}
}

// processSolutions processes generated solutions
func (n *Node) processSolutions(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case solution := <-n.solutionChan:
			if err := n.handleSolution(solution); err != nil {
				n.errorChan <- fmt.Errorf("failed to handle solution: %v", err)
			}
		}
	}
}

// handleSolution handles a generated solution based on mode
func (n *Node) handleSolution(solution *hephaestus.Solution) error {
	switch n.config.OperationalMode {
	case "suggest":
		return n.handleSuggestMode(solution)
	case "deploy":
		return n.handleDeployMode(solution)
	default:
		return fmt.Errorf("unknown operational mode: %s", n.config.OperationalMode)
	}
}

// handleSuggestMode handles solution in suggest mode
func (n *Node) handleSuggestMode(solution *hephaestus.Solution) error {
	fmt.Printf("[Hephaestus] Solution generated: %s\n", solution.Description)
	return nil
}

// handleDeployMode handles solution in deploy mode
func (n *Node) handleDeployMode(solution *hephaestus.Solution) error {
	// TODO: Implement remote repository PR creation
	return fmt.Errorf("deploy mode not implemented yet")
}

// GetStatus returns the current node status
func (n *Node) GetStatus() hephaestus.NodeStatus {
	n.statusLock.RLock()
	defer n.statusLock.RUnlock()
	return n.status
}

// GetErrors returns the error channel
func (n *Node) GetErrors() <-chan error {
	return n.errorChan
}
