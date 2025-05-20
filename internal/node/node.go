package node

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
)

// Node represents a Hephaestus node
type Node struct {
	config     *hephaestus.SystemConfiguration
	status     hephaestus.NodeStatus
	statusLock sync.RWMutex

	// Log processing
	logBuffer     []hephaestus.LogEntry
	logBufferLock sync.RWMutex
	lastProcessed time.Time

	// Solution processing
	solutionChan chan *hephaestus.Solution
	errorChan    chan error
}

// NewNode creates a new Hephaestus node
func NewNode(config *hephaestus.SystemConfiguration) (*Node, error) {
	if err := hephaestus.ValidateSystemConfiguration(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	return &Node{
		config:        config,
		status:        hephaestus.NodeStatusInitializing,
		logBuffer:     make([]hephaestus.LogEntry, 0),
		solutionChan:  make(chan *hephaestus.Solution, 100),
		errorChan:     make(chan error, 100),
		lastProcessed: time.Now(),
	}, nil
}

// Start initializes and starts the node
func (n *Node) Start(ctx context.Context) error {
	n.statusLock.Lock()
	n.status = hephaestus.NodeStatusOperational
	n.statusLock.Unlock()

	// Start log processing
	go n.processLogs(ctx)

	// Start solution processing
	go n.processSolutions(ctx)

	return nil
}

// Stop gracefully stops the node
func (n *Node) Stop(ctx context.Context) error {
	n.statusLock.Lock()
	n.status = hephaestus.NodeStatusError
	n.statusLock.Unlock()

	// Close channels
	close(n.solutionChan)
	close(n.errorChan)

	return nil
}

// ProcessLog processes a new log entry
func (n *Node) ProcessLog(entry hephaestus.LogEntry) error {
	n.logBufferLock.Lock()
	defer n.logBufferLock.Unlock()

	// Add to buffer
	n.logBuffer = append(n.logBuffer, entry)

	// Check if we need to process logs
	if n.shouldProcessLogs() {
		return n.triggerLogProcessing()
	}

	return nil
}

// shouldProcessLogs checks if we should process logs based on threshold
func (n *Node) shouldProcessLogs() bool {
	thresholdCount := 0
	windowStart := time.Now().Add(-n.config.LogSettings.ThresholdWindow)

	for _, entry := range n.logBuffer {
		if entry.Timestamp.After(windowStart) && entry.Level == n.config.LogSettings.ThresholdLevel {
			thresholdCount++
		}
	}

	return thresholdCount >= n.config.LogSettings.ThresholdCount
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