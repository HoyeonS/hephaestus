package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/HoyeonS/hephaestus/internal/github"
	"github.com/HoyeonS/hephaestus/internal/repository"
	pb "github.com/HoyeonS/hephaestus/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Service implements the Hephaestus gRPC service
type Service struct {
	pb.UnimplementedHephaestusServer
	repositories map[string]*repository.VirtualRepository
	githubClient *github.Client
	mu          sync.RWMutex
}

// NewService creates a new Hephaestus service
func NewService() *Service {
	return &Service{
		repositories: make(map[string]*repository.VirtualRepository),
	}
}

// Initialize creates a new Hephaestus instance with the given configuration
func (s *Service) Initialize(ctx context.Context, req *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	if req.Config == nil {
		return nil, status.Error(codes.InvalidArgument, "configuration is required")
	}

	// Create GitHub client
	githubClient, err := github.NewClient(
		req.Config.Github.Token,
		req.Config.Github.Repository,
	)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to create GitHub client: %v", err)
	}

	// Create virtual repository
	vRepo, err := githubClient.FetchRepository(ctx, req.Config.Github.Branch)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch repository: %v", err)
	}

	// Store configuration
	vRepo.Configuration = &repository.Configuration{
		GitHub: repository.GitHubConfig{
			Repository: req.Config.Github.Repository,
			Branch:    req.Config.Github.Branch,
			Token:     req.Config.Github.Token,
		},
		AI: repository.AIConfig{
			Provider: req.Config.Ai.Provider,
			APIKey:   req.Config.Ai.ApiKey,
		},
		Log: repository.LogConfig{
			Level: req.Config.Log.Level,
		},
		Mode: req.Config.Mode,
	}

	// Store repository
	s.mu.Lock()
	s.repositories[vRepo.ID] = vRepo
	s.githubClient = githubClient
	s.mu.Unlock()

	return &pb.InitializeResponse{
		Status:  "success",
		Message: "Repository initialized successfully",
		NodeId:  vRepo.ID,
	}, nil
}

// StreamLogs handles the bidirectional streaming of logs and fixes
func (s *Service) StreamLogs(stream pb.Hephaestus_StreamLogsServer) error {
	// Get first message to get node ID
	firstLog, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.Internal, "failed to receive first log: %v", err)
	}

	// Get repository
	s.mu.RLock()
	repo, ok := s.repositories[firstLog.Metadata["node_id"]]
	s.mu.RUnlock()
	if !ok {
		return status.Errorf(codes.NotFound, "repository not found for node ID: %s", firstLog.Metadata["node_id"])
	}

	// Create channels for communication
	logChan := make(chan *pb.LogEntry, 100)
	errChan := make(chan error, 1)

	// Start goroutine to receive logs
	go func() {
		for {
			log, err := stream.Recv()
			if err != nil {
				errChan <- err
				return
			}
			logChan <- log
		}
	}()

	// Process logs and generate fixes
	for {
		select {
		case err := <-errChan:
			return status.Errorf(codes.Internal, "error receiving logs: %v", err)

		case logEntry := <-logChan:
			// Check if log level meets threshold
			if !isLogLevelMet(logEntry.Level, repo.Configuration.Log.Level) {
				continue
			}

			// Generate fix
			fix, err := s.generateFix(stream.Context(), repo, logEntry)
			if err != nil {
				log.Printf("Error generating fix: %v", err)
				continue
			}

			// Send fix response
			resp := &pb.FixResponse{
				Status:  "success",
				Message: "Fix generated successfully",
			}

			if repo.Configuration.Mode == "suggest" {
				resp.Result = &pb.FixResponse_SuggestedFix{
					SuggestedFix: &pb.SuggestedFix{
						FilePath:      fix.FilePath,
						OriginalCode:  fix.OriginalCode,
						SuggestedCode: fix.SuggestedCode,
						Explanation:   fix.Explanation,
					},
				}
			} else {
				// Create pull request
				pr, err := s.createPullRequest(stream.Context(), repo, fix)
				if err != nil {
					log.Printf("Error creating pull request: %v", err)
					continue
				}

				resp.Result = &pb.FixResponse_PullRequest{
					PullRequest: &pb.PullRequest{
						Url:    *pr.HTMLURL,
						Title:  *pr.Title,
						Branch: *pr.Head.Ref,
					},
				}
			}

			if err := stream.Send(resp); err != nil {
				return status.Errorf(codes.Internal, "error sending fix: %v", err)
			}

		case <-stream.Context().Done():
			return nil
		}
	}
}

// Helper types and functions

type Fix struct {
	FilePath      string
	OriginalCode  string
	SuggestedCode string
	Explanation   string
}

func (s *Service) generateFix(ctx context.Context, repo *repository.VirtualRepository, logEntry *pb.LogEntry) (*Fix, error) {
	// TODO: Implement AI-based fix generation
	// This is a placeholder that should be replaced with actual AI integration
	return &Fix{
		FilePath:      "example.go",
		OriginalCode:  "func example() error { return nil }",
		SuggestedCode: "func example() error { return fmt.Errorf(\"implemented\") }",
		Explanation:   "Added error implementation",
	}, nil
}

func (s *Service) createPullRequest(ctx context.Context, repo *repository.VirtualRepository, fix *Fix) (*github.PullRequest, error) {
	changes := map[string]string{
		fix.FilePath: fix.SuggestedCode,
	}

	branch := fmt.Sprintf("fix/%d", time.Now().Unix())
	title := "Fix: Automated code improvement"
	body := fix.Explanation

	return s.githubClient.CreatePullRequest(ctx, branch, title, body, changes)
}

func isLogLevelMet(currentLevel, thresholdLevel string) bool {
	levels := map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
		"fatal": 4,
	}

	currentValue, ok1 := levels[currentLevel]
	thresholdValue, ok2 := levels[thresholdLevel]

	if !ok1 || !ok2 {
		return false
	}

	return currentValue >= thresholdValue
} 