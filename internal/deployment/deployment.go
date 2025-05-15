package deployment

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/HoyeonS/hephaestus/internal/models"
)

// Config holds configuration for the deployment service
type Config struct {
	AutoDeploy              bool          `yaml:"auto_deploy"`
	RequireHumanApproval    bool          `yaml:"require_human_approval"`
	SandboxTimeout          time.Duration `yaml:"sandbox_timeout"`
	RollbackEnabled         bool          `yaml:"rollback_enabled"`
	MaxConcurrentDeployments int          `yaml:"max_concurrent_deployments"`
}

// Service represents the deployment service
type Service struct {
	config     Config
	inputChan  chan *models.Fix
	outputChan chan *models.Fix
	done       chan struct{}
	semaphore  chan struct{} // Limits concurrent deployments
	mu         sync.RWMutex
}

// New creates a new deployment service
func New(config Config) (*Service, error) {
	return &Service{
		config:     config,
		inputChan:  make(chan *models.Fix, 100),
		outputChan: make(chan *models.Fix, 100),
		done:       make(chan struct{}),
		semaphore:  make(chan struct{}, config.MaxConcurrentDeployments),
	}, nil
}

// Start starts the deployment service
func (s *Service) Start(ctx context.Context) error {
	// Start worker goroutines
	for i := 0; i < s.config.MaxConcurrentDeployments; i++ {
		go s.processDeployments(ctx)
	}

	return nil
}

// Stop stops the deployment service
func (s *Service) Stop() error {
	close(s.done)
	return nil
}

// GetInputChannel returns the channel for receiving fixes to deploy
func (s *Service) GetInputChannel() chan<- *models.Fix {
	return s.inputChan
}

// GetOutputChannel returns the channel for deployed fixes
func (s *Service) GetOutputChannel() <-chan *models.Fix {
	return s.outputChan
}

// processDeployments handles the deployment workflow
func (s *Service) processDeployments(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.done:
			return
		case fix := <-s.inputChan:
			if fix == nil {
				continue
			}

			// Acquire semaphore
			s.semaphore <- struct{}{}
			
			// Process fix
			result := s.deployFix(ctx, fix)
			
			// Release semaphore
			<-s.semaphore

			if result != nil {
				// Try to send result, don't block if channel is full
				select {
				case s.outputChan <- result:
				default:
					fmt.Printf("Output channel full, dropping deployed fix: %s\n", result.ID)
				}
			}
		}
	}
}

// deployFix handles the deployment of a fix
func (s *Service) deployFix(ctx context.Context, fix *models.Fix) *models.Fix {
	// Create sandbox environment
	sandbox, err := s.createSandbox(fix)
	if err != nil {
		fmt.Printf("Failed to create sandbox: %v\n", err)
		fix.UpdateStatus(models.FixFailed)
		return fix
	}
	defer s.cleanupSandbox(sandbox)

	// Test fix in sandbox
	if err := s.testInSandbox(ctx, sandbox, fix); err != nil {
		fmt.Printf("Sandbox testing failed: %v\n", err)
		fix.UpdateStatus(models.FixFailed)
		return fix
	}

	// Get human approval if required
	if s.config.RequireHumanApproval {
		approved := s.getHumanApproval(fix)
		if !approved {
			fix.UpdateStatus(models.FixFailed)
			return fix
		}
	}

	// Create backup for rollback
	if s.config.RollbackEnabled {
		if err := s.createBackup(fix); err != nil {
			fmt.Printf("Failed to create backup: %v\n", err)
			fix.UpdateStatus(models.FixFailed)
			return fix
		}
	}

	// Apply the fix
	if err := s.applyFix(fix); err != nil {
		fmt.Printf("Failed to apply fix: %v\n", err)
		if s.config.RollbackEnabled {
			s.rollback(fix)
		}
		fix.UpdateStatus(models.FixFailed)
		return fix
	}

	// Verify the fix
	if err := s.verifyFix(fix); err != nil {
		fmt.Printf("Fix verification failed: %v\n", err)
		if s.config.RollbackEnabled {
			s.rollback(fix)
		}
		fix.UpdateStatus(models.FixFailed)
		return fix
	}

	fix.UpdateStatus(models.FixVerified)
	return fix
}

// createSandbox creates a sandbox environment for testing fixes
func (s *Service) createSandbox(fix *models.Fix) (string, error) {
	// Create temporary directory for sandbox
	sandboxDir, err := os.MkdirTemp("", "hephaestus-sandbox-*")
	if err != nil {
		return "", fmt.Errorf("failed to create sandbox directory: %v", err)
	}

	// Copy relevant files to sandbox
	for _, change := range fix.CodeChanges {
		srcPath := change.FilePath
		dstPath := filepath.Join(sandboxDir, filepath.Base(srcPath))

		if err := copyFile(srcPath, dstPath); err != nil {
			os.RemoveAll(sandboxDir)
			return "", fmt.Errorf("failed to copy file to sandbox: %v", err)
		}
	}

	return sandboxDir, nil
}

// testInSandbox tests a fix in the sandbox environment
func (s *Service) testInSandbox(ctx context.Context, sandbox string, fix *models.Fix) error {
	// Create context with timeout
	testCtx, cancel := context.WithTimeout(ctx, s.config.SandboxTimeout)
	defer cancel()

	// Apply fix in sandbox
	if err := s.applyFixInSandbox(sandbox, fix); err != nil {
		return fmt.Errorf("failed to apply fix in sandbox: %v", err)
	}

	// Run tests
	for _, test := range fix.TestResults {
		cmd := exec.CommandContext(testCtx, "go", "test", "-v", "./...")
		cmd.Dir = sandbox

		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("test failed: %v\nOutput: %s", err, output)
		}

		// Record test results
		test.Status = "pass"
		test.Output = string(output)
		test.ExecutedAt = time.Now()
	}

	return nil
}

// getHumanApproval gets approval for a fix from a human operator
func (s *Service) getHumanApproval(fix *models.Fix) bool {
	// Implementation would integrate with an approval system
	// This is a placeholder that always returns true
	return true
}

// createBackup creates a backup for rollback
func (s *Service) createBackup(fix *models.Fix) error {
	backup := make(map[string]string)

	for _, change := range fix.CodeChanges {
		content, err := os.ReadFile(change.FilePath)
		if err != nil {
			return fmt.Errorf("failed to read file for backup: %v", err)
		}
		backup[change.FilePath] = string(content)
	}

	fix.SetRollbackData(backup, time.Now().Format(time.RFC3339), "go test ./...")
	return nil
}

// applyFix applies a fix to the actual codebase
func (s *Service) applyFix(fix *models.Fix) error {
	for _, change := range fix.CodeChanges {
		if err := applyCodeChange(change); err != nil {
			return fmt.Errorf("failed to apply code change: %v", err)
		}
	}

	fix.UpdateStatus(models.FixApplied)
	return nil
}

// verifyFix verifies that a fix works in production
func (s *Service) verifyFix(fix *models.Fix) error {
	// Run verification tests
	cmd := exec.Command("go", "test", "-v", "./...")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("verification failed: %v\nOutput: %s", err, output)
	}

	return nil
}

// rollback reverts a fix
func (s *Service) rollback(fix *models.Fix) error {
	if fix.RollbackData == nil {
		return fmt.Errorf("no rollback data available")
	}

	// Restore files from backup
	for path, content := range fix.RollbackData.Backup {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to restore file %s: %v", path, err)
		}
	}

	// Verify rollback
	cmd := exec.Command("sh", "-c", fix.RollbackData.ValidateCmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rollback validation failed: %v", err)
	}

	fix.UpdateStatus(models.FixRollback)
	return nil
}

// cleanupSandbox removes the sandbox environment
func (s *Service) cleanupSandbox(sandbox string) {
	os.RemoveAll(sandbox)
}

// Helper functions

func copyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, content, 0644)
}

func applyCodeChange(change models.CodeChange) error {
	// Implementation would apply the code change to the file
	// This is a placeholder
	return nil
}

// applyFixInSandbox applies a fix in the sandbox environment
func (s *Service) applyFixInSandbox(sandbox string, fix *models.Fix) error {
	for _, change := range fix.CodeChanges {
		dstPath := filepath.Join(sandbox, filepath.Base(change.FilePath))
		if err := applyCodeChange(models.CodeChange{
			FilePath: dstPath,
			Content:  change.Content,
			Type:     change.Type,
		}); err != nil {
			return fmt.Errorf("failed to apply change in sandbox: %v", err)
		}
	}
	return nil
} 