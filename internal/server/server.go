package server

import (
	"context"
	"fmt"
	"net"

	"github.com/HoyeonS/hephaestus/internal/logger"
	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	pb "github.com/HoyeonS/hephaestus/proto"
	"google.golang.org/grpc"
)

// Server implements the HephaestusService gRPC server
type Server struct {
	pb.UnimplementedHephaestusServiceServer
	clientNode Node
}

// NewServer creates a new instance of the HephaestusService server
func NewServer(nodeManager hephaestus.NodeManager, modelService hephaestus.ModelService, metricsCollector hephaestus.MetricsCollectionService) *Server {
	return &Server{
		nodeManager:      nodeManager,
		modelService:     modelService,
		metricsCollector: metricsCollector,
	}
}

// Start starts the gRPC server
func (s *Server) Start(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterHephaestusServiceServer(server, s)

	logger.Info(context.Background(), "Starting gRPC server", logger.Field("address", address))
	if err := server.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}

// RegisterNode registers a new node with the system
func (s *Server) RegisterNode(ctx context.Context, req *pb.RegisterNodeRequest) (*pb.RegisterNodeResponse, error) {
	logger.Info(ctx, "Registering node", logger.Field("node_id", req.NodeId))

	node := &hephaestus.Node{
		NodeID: req.NodeId,
	}

	if err := s.nodeManager.RegisterNode(ctx, node); err != nil {
		logger.Error(ctx, "Failed to register node", logger.Field("error", err))
		return &pb.RegisterNodeResponse{
			Status: "error",
			Error:  err.Error(),
		}, nil
	}

	return &pb.RegisterNodeResponse{
		Status: "success",
	}, nil
}

// ProcessLogEntry processes a log entry from a node
func (s *Server) ProcessLogEntry(ctx context.Context, req *pb.ProcessLogEntryRequest) (*pb.ProcessLogEntryResponse, error) {
	logger.Info(ctx, "Processing log entry", logger.Field("node_id", req.LogEntry.NodeId))

	logEntry := &hephaestus.LogEntryData{
		NodeIdentifier: req.LogEntry.NodeId,
		LogMessage:     req.LogEntry.Message,
		LogLevel:       req.LogEntry.LogLevel,
		ErrorTrace:     req.LogEntry.ErrorTrace,
		Timestamp:      req.LogEntry.Timestamp.AsTime(),
	}

	if err := s.modelService.ProcessLogEntry(ctx, logEntry); err != nil {
		logger.Error(ctx, "Failed to process log entry", logger.Field("error", err))
		return &pb.ProcessLogEntryResponse{
			Status: "error",
			Error:  err.Error(),
		}, nil
	}

	return &pb.ProcessLogEntryResponse{
		Status: "success",
	}, nil
}

// GetSolutionProposal generates a solution proposal for a log entry
func (s *Server) GetSolutionProposal(ctx context.Context, req *pb.GetSolutionProposalRequest) (*pb.GetSolutionProposalResponse, error) {
	logger.Info(ctx, "Generating solution proposal", logger.Field("node_id", req.NodeId))

	logEntry := &hephaestus.LogEntryData{
		NodeIdentifier: req.LogEntry.NodeId,
		LogMessage:     req.LogEntry.Message,
		LogLevel:       req.LogEntry.LogLevel,
		ErrorTrace:     req.LogEntry.ErrorTrace,
		Timestamp:      req.LogEntry.Timestamp.AsTime(),
	}

	solution, err := s.modelService.GenerateSolutionProposal(ctx, logEntry)
	if err != nil {
		logger.Error(ctx, "Failed to generate solution proposal", logger.Field("error", err))
		return &pb.GetSolutionProposalResponse{
			Error: err.Error(),
		}, nil
	}

	return &pb.GetSolutionProposalResponse{
		Solution: &pb.SolutionProposal{
			SolutionId:      solution.SolutionID,
			NodeId:          solution.NodeIdentifier,
			AssociatedLog:   req.LogEntry,
			ProposedChanges: solution.ProposedChanges,
			GenerationTime:  solution.GenerationTime.AsTime(),
			ConfidenceScore: solution.ConfidenceScore,
		},
	}, nil
}

// ValidateSolution validates a solution proposal
func (s *Server) ValidateSolution(ctx context.Context, req *pb.ValidateSolutionRequest) (*pb.ValidateSolutionResponse, error) {
	logger.Info(ctx, "Validating solution", logger.Field("solution_id", req.Solution.SolutionId))

	solution := &hephaestus.ProposedSolution{
		SolutionID:      req.Solution.SolutionId,
		NodeIdentifier:  req.Solution.NodeId,
		ProposedChanges: req.Solution.ProposedChanges,
		GenerationTime:  req.Solution.GenerationTime.AsTime(),
		ConfidenceScore: req.Solution.ConfidenceScore,
	}

	if err := s.modelService.ValidateSolutionProposal(ctx, solution); err != nil {
		logger.Error(ctx, "Failed to validate solution", logger.Field("error", err))
		return &pb.ValidateSolutionResponse{
			IsValid: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.ValidateSolutionResponse{
		IsValid: true,
	}, nil
}
