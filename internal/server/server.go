package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/HoyeonS/hephaestus/internal/github"
	"github.com/HoyeonS/hephaestus/internal/ai"
	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	pb "github.com/HoyeonS/hephaestus/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the Hephaestus gRPC service
type Server struct {
	pb.UnimplementedHephaestusServer
	nodes map[string]*hephaestus.HephaestusNode
	repos map[string]*hephaestus.VirtualRepository
	mu    sync.RWMutex

	// Channels for each node's log processing
	logStreams map[string]chan *pb.LogEntry
}

// NewServer creates a new Hephaestus server
func NewServer() *Server {
	return &Server{
		nodes:      make(map[string]*hephaestus.HephaestusNode),
		repos:      make(map[string]*hephaestus.VirtualRepository),
		logStreams: make(map[string]chan *pb.LogEntry),
	}
}

// Initialize implements the Initialize RPC method
func (s *Server) Initialize(ctx context.Context, req *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	// Convert proto config to internal config
	config := &hephaestus.Config{
		GitHub: hephaestus.GitHubConfig{
			Repository: req.Config.Github.Repository,
			Branch:    req.Config.Github.Branch,
			Token:     req.Config.Github.Token,
		},
		AI: hephaestus.AIConfig{
			Provider: req.Config.Ai.Provider,
			APIKey:   req.Config.Ai.ApiKey,
			Model:    req.Config.Ai.Model,
		},
		Log: hephaestus.LogConfig{
			Level:          req.Config.Log.Level,
			ThresholdLevel: req.Config.Log.ThresholdLevel,
		},
		Mode: req.Config.Mode,
	}

	// Validate configuration
	if err := hephaestus.ValidateConfig(config); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid configuration: %v", err)
	}

	// Create GitHub client
	githubClient, err := github.NewClient(config.GitHub.Token, config.GitHub.Repository)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create GitHub client: %v", err)
	}

	// Fetch repository and create virtual repository
	vRepo, err := githubClient.FetchRepository(ctx, config.GitHub.Branch)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch repository: %v", err)
	}

	// Generate IDs
	nodeID := uuid.New().String()
	repoID := uuid.New().String()

	// Create node
	node := &hephaestus.HephaestusNode{
		ID:            nodeID,
		Configuration: config,
		RepositoryID:  repoID,
		Status:        "active",
		CreatedAt:     time.Now(),
		LastActive:    time.Now(),
	}

	// Store node and repository
	s.mu.Lock()
	s.nodes[nodeID] = node
	s.repos[repoID] = vRepo
	s.logStreams[nodeID] = make(chan *pb.LogEntry, 100) // Buffer for log processing
	s.mu.Unlock()

	return &pb.InitializeResponse{
		Status:  "success",
		Message: "Node initialized successfully",
		NodeId:  nodeID,
	}, nil
}

// StreamLogs implements the StreamLogs RPC method
func (s *Server) StreamLogs(stream pb.Hephaestus_StreamLogsServer) error {
	// Get first message to identify the node
	firstLog, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.Internal, "failed to receive first log: %v", err)
	}

	nodeID := firstLog.NodeId
	
	// Verify node exists
	s.mu.RLock()
	node, exists := s.nodes[nodeID]
	logChan := s.logStreams[nodeID]
	s.mu.RUnlock()

	if !exists {
		return status.Errorf(codes.NotFound, "node not found: %s", nodeID)
	}

	// Create AI client based on configuration
	aiClient, err := ai.NewClient(node.Configuration.AI)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to create AI client: %v", err)
	}

	// Start goroutine to receive logs
	errChan := make(chan error, 1)
	go func() {
		for {
			log, err := stream.Recv()
			if err != nil {
				errChan <- err
				return
			}

			// Check if log level meets threshold
			if !isLogLevelMet(log.Level, node.Configuration.Log.ThresholdLevel) {
				continue
			}

			select {
			case logChan <- log:
			default:
				// Channel full, log processing too slow
				errChan <- status.Error(codes.ResourceExhausted, "log processing backed up")
				return
			}
		}
	}()

	// Process logs and generate solutions
	for {
		select {
		case err := <-errChan:
			return err
		case log := <-logChan:
			// Update last active timestamp
			s.mu.Lock()
			node.LastActive = time.Now()
			s.mu.Unlock()

			// Generate solution using AI
			solution, err := aiClient.GenerateSolution(stream.Context(), log, s.repos[node.RepositoryID])
			if err != nil {
				continue // Skip this log if solution generation fails
			}

			// Prepare response based on mode
			response := &pb.SolutionResponse{
				Status:  "success",
				Message: "Solution generated",
			}

			if node.Configuration.Mode == "suggest" {
				// Create suggested fix
				suggestedFix := &pb.SuggestedFix{
					SolutionId:  solution.ID,
					Description: solution.Description,
					Changes:     make([]*pb.CodeChange, len(solution.Changes)),
				}

				for i, change := range solution.Changes {
					suggestedFix.Changes[i] = &pb.CodeChange{
						FilePath:     change.FilePath,
						OriginalCode: change.OriginalCode,
						UpdatedCode:  change.UpdatedCode,
						LineStart:    int32(change.LineStart),
						LineEnd:      int32(change.LineEnd),
					}
				}

				response.Result = &pb.SolutionResponse_SuggestedFix{
					SuggestedFix: suggestedFix,
				}
			} else {
				// Create pull request
				pr, err := githubClient.CreatePullRequest(stream.Context(), solution)
				if err != nil {
					continue // Skip if PR creation fails
				}

				response.Result = &pb.SolutionResponse_PullRequest{
					PullRequest: &pb.PullRequest{
						Url:    pr.URL,
						Title:  pr.Title,
						Branch: pr.Branch,
					},
				}
			}

			if err := stream.Send(response); err != nil {
				return status.Errorf(codes.Internal, "failed to send solution: %v", err)
			}
		}
	}
}

// GetNode implements the GetNode RPC method
func (s *Server) GetNode(ctx context.Context, req *pb.GetNodeRequest) (*pb.GetNodeResponse, error) {
	s.mu.RLock()
	node, exists := s.nodes[req.NodeId]
	s.mu.RUnlock()

	if !exists {
		return nil, status.Errorf(codes.NotFound, "node not found: %s", req.NodeId)
	}

	return &pb.GetNodeResponse{
		Id:            node.ID,
		Configuration: convertConfigToProto(node.Configuration),
		RepositoryId:  node.RepositoryID,
		Status:        node.Status,
		CreatedAt:     node.CreatedAt.Format(time.RFC3339),
		LastActive:    node.LastActive.Format(time.RFC3339),
	}, nil
}

// Helper functions

func isLogLevelMet(logLevel, thresholdLevel string) bool {
	levels := map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
	}

	return levels[logLevel] >= levels[thresholdLevel]
}

func convertConfigToProto(config *hephaestus.Config) *pb.Config {
	return &pb.Config{
		Github: &pb.GitHubConfig{
			Repository: config.GitHub.Repository,
			Branch:    config.GitHub.Branch,
			Token:     config.GitHub.Token,
		},
		Ai: &pb.AIConfig{
			Provider: config.AI.Provider,
			ApiKey:   config.AI.APIKey,
			Model:    config.AI.Model,
		},
		Log: &pb.LogConfig{
			Level:          config.Log.Level,
			ThresholdLevel: config.Log.ThresholdLevel,
		},
		Mode: config.Mode,
	}
} 