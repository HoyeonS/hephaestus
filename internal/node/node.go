package node

import (
	"context"
	"fmt"
	"time"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
)

// NewNode creates a new Hephaestus node
func NewNode(systemConfig *hephaestus.SystemConfiguration, clientNodeConfig *hephaestus.ClientNodeConfiguration) (*Node, error) {
	if err := hephaestus.ValidateSystemConfiguration(systemConfig); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	return &hephaestus.Node{
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
	n.status = hephaestus.NodeStatusProcessing

	// Process logs in a separate goroutine
	go func() {
		defer func() {
			n.status = hephaestus.NodeStatusOperational
		}()

		// Clear buffer after processing
		entries := make([]hephaestus.LogEntry, len(n.logBuffer))
		copy(entries, n.logBuffer)
		n.logBuffer = make([]hephaestus.LogEntry, 0)

		// Generate solution
		solution, err := n.initateSolutionFlow(entries)
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
func (n *Node) initateSolutionFlow(entries []hephaestus.LogEntry) (*hephaestus.Solution, error) {
	// TODO: Implement solution generation logic
	return &hephaestus.Solution{
		ID:          fmt.Sprintf("sol-%d", time.Now().UnixNano()),
		LogEntry:    entries[len(entries)-1],
		Description: "Generated solution",
		GeneratedAt: time.Now(),
		Confidence:  0.8,
	}, nil
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

// GetErrors returns the error channel
func (n *Node) GetErrors() <-chan error {
	return n.errorChan
}
