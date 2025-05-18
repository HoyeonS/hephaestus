package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/HoyeonS/hephaestus/internal/config"
	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	pb "github.com/HoyeonS/hephaestus/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HephaestusServer implements the HephaestusService gRPC interface
type HephaestusServer struct {
	pb.UnimplementedHephaestusServiceServer
	
	// Dependencies
	nodeManager          hephaestus.NodeLifecycleManager
	logProcessor        hephaestus.LogProcessingService
	modelService        hephaestus.ModelServiceProvider
	remoteRepoService   hephaestus.RemoteRepositoryService
	metricsCollector    hephaestus.MetricsCollectionService
	
	// Node registry
	nodes     map[string]*hephaestus.SystemNode
	nodesMutex sync.RWMutex
}

// NewHephaestusServer creates a new instance of the Hephaestus server
func NewHephaestusServer(
	nodeManager hephaestus.NodeLifecycleManager,
	logProcessor hephaestus.LogProcessingService,
	modelService hephaestus.ModelServiceProvider,
	remoteRepoService hephaestus.RemoteRepositoryService,
	metricsCollector hephaestus.MetricsCollectionService,
) *HephaestusServer {
	return &HephaestusServer{
		nodeManager:        nodeManager,
		logProcessor:      logProcessor,
		modelService:      modelService,
		remoteRepoService: remoteRepoService,
		metricsCollector:  metricsCollector,
		nodes:            make(map[string]*hephaestus.SystemNode),
	}
}

// InitializeNode initializes a new node with the provided configuration
func (s *HephaestusServer) InitializeNode(ctx context.Context, req *pb.InitializeNodeRequest) (*pb.InitializeNodeResponse, error) {
	if req.Configuration == nil {
		return nil, status.Error(codes.InvalidArgument, "configuration is required")
	}

	// Convert proto configuration to internal configuration
	config := &hephaestus.SystemConfiguration{
		RemoteSettings:    convertRemoteConfig(req.Configuration.RemoteSettings),
		ModelSettings:     convertModelConfig(req.Configuration.ModelSettings),
		LoggingSettings:   convertLoggingConfig(req.Configuration.LoggingSettings),
		OperationalMode:   req.Configuration.OperationalMode,
		RepositorySettings: convertRepoConfig(req.Configuration.RepositorySettings),
	}

	// Create new node
	node, err := s.nodeManager.CreateSystemNode(ctx, config)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create node: %v", err)
	}

	// Store node in registry
	s.nodesMutex.Lock()
	s.nodes[node.NodeIdentifier] = node
	s.nodesMutex.Unlock()

	return &pb.InitializeNodeResponse{
		NodeIdentifier:    node.NodeIdentifier,
		OperationalStatus: string(node.CurrentStatus),
		StatusMessage:     "Node initialized successfully",
	}, nil
}

// GetNodeStatus retrieves the current status of a node
func (s *HephaestusServer) GetNodeStatus(ctx context.Context, req *pb.NodeStatusRequest) (*pb.NodeStatusResponse, error) {
	s.nodesMutex.RLock()
	node, exists := s.nodes[req.NodeIdentifier]
	s.nodesMutex.RUnlock()

	if !exists {
		return nil, status.Errorf(codes.NotFound, "node not found: %s", req.NodeIdentifier)
	}

	return &pb.NodeStatusResponse{
		NodeIdentifier:       node.NodeIdentifier,
		OperationalStatus:    string(node.CurrentStatus),
		CurrentConfiguration: convertConfigToProto(node.NodeConfig),
		StatusMessage:        "Node status retrieved successfully",
	}, nil
}

// UpdateNodeConfiguration updates the configuration of an existing node
func (s *HephaestusServer) UpdateNodeConfiguration(ctx context.Context, req *pb.UpdateNodeConfigurationRequest) (*pb.UpdateNodeConfigurationResponse, error) {
	s.nodesMutex.Lock()
	defer s.nodesMutex.Unlock()

	node, exists := s.nodes[req.NodeIdentifier]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "node not found: %s", req.NodeIdentifier)
	}

	// Convert and validate new configuration
	newConfig := &hephaestus.SystemConfiguration{
		RemoteSettings:    convertRemoteConfig(req.NewConfiguration.RemoteSettings),
		ModelSettings:     convertModelConfig(req.NewConfiguration.ModelSettings),
		LoggingSettings:   convertLoggingConfig(req.NewConfiguration.LoggingSettings),
		OperationalMode:   req.NewConfiguration.OperationalMode,
		RepositorySettings: convertRepoConfig(req.NewConfiguration.RepositorySettings),
	}

	// Update node configuration
	if err := s.nodeManager.UpdateNodeOperationalStatus(ctx, node.NodeIdentifier, hephaestus.NodeStatusInitializing); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update node status: %v", err)
	}

	node.NodeConfig = newConfig
	node.CurrentStatus = hephaestus.NodeStatusOperational

	return &pb.UpdateNodeConfigurationResponse{
		NodeIdentifier: node.NodeIdentifier,
		StatusMessage:  "Configuration updated successfully",
		Success:       true,
	}, nil
}

// DeleteNode removes a node from the system
func (s *HephaestusServer) DeleteNode(ctx context.Context, req *pb.DeleteNodeRequest) (*pb.DeleteNodeResponse, error) {
	s.nodesMutex.Lock()
	defer s.nodesMutex.Unlock()

	if _, exists := s.nodes[req.NodeIdentifier]; !exists {
		return nil, status.Errorf(codes.NotFound, "node not found: %s", req.NodeIdentifier)
	}

	if err := s.nodeManager.DeleteSystemNode(ctx, req.NodeIdentifier); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete node: %v", err)
	}

	delete(s.nodes, req.NodeIdentifier)

	return &pb.DeleteNodeResponse{
		StatusMessage: "Node deleted successfully",
		Success:      true,
	}, nil
}

// Helper functions for configuration conversion
func convertRemoteConfig(config *pb.RemoteRepositoryConfiguration) hephaestus.RemoteRepositoryConfiguration {
	if config == nil {
		return hephaestus.RemoteRepositoryConfiguration{}
	}
	return hephaestus.RemoteRepositoryConfiguration{
		AuthToken:       config.AuthToken,
		RepositoryOwner: config.RepositoryOwner,
		RepositoryName:  config.RepositoryName,
		TargetBranch:    config.TargetBranch,
	}
}

func convertModelConfig(config *pb.ModelServiceConfiguration) hephaestus.ModelServiceConfiguration {
	if config == nil {
		return hephaestus.ModelServiceConfiguration{}
	}
	return hephaestus.ModelServiceConfiguration{
		ServiceProvider: config.ServiceProvider,
		ServiceAPIKey:   config.ServiceApiKey,
		ModelVersion:    config.ModelVersion,
	}
}

func convertLoggingConfig(config *pb.LoggingConfiguration) hephaestus.LoggingConfiguration {
	if config == nil {
		return hephaestus.LoggingConfiguration{}
	}
	return hephaestus.LoggingConfiguration{
		LogLevel:     config.LogLevel,
		OutputFormat: config.OutputFormat,
	}
}

func convertRepoConfig(config *pb.RepositoryConfiguration) hephaestus.RepositoryConfiguration {
	if config == nil {
		return hephaestus.RepositoryConfiguration{}
	}
	return hephaestus.RepositoryConfiguration{
		RepositoryPath: config.RepositoryPath,
		FileLimit:      int(config.FileLimit),
		FileSizeLimit:  config.FileSizeLimit,
	}
}

func convertConfigToProto(config *hephaestus.SystemConfiguration) *pb.SystemConfiguration {
	if config == nil {
		return nil
	}
	return &pb.SystemConfiguration{
		RemoteSettings: &pb.RemoteRepositoryConfiguration{
			AuthToken:       config.RemoteSettings.AuthToken,
			RepositoryOwner: config.RemoteSettings.RepositoryOwner,
			RepositoryName:  config.RemoteSettings.RepositoryName,
			TargetBranch:    config.RemoteSettings.TargetBranch,
		},
		ModelSettings: &pb.ModelServiceConfiguration{
			ServiceProvider: config.ModelSettings.ServiceProvider,
			ServiceApiKey:   config.ModelSettings.ServiceAPIKey,
			ModelVersion:    config.ModelSettings.ModelVersion,
		},
		LoggingSettings: &pb.LoggingConfiguration{
			LogLevel:     config.LoggingSettings.LogLevel,
			OutputFormat: config.LoggingSettings.OutputFormat,
		},
		OperationalMode: config.OperationalMode,
		RepositorySettings: &pb.RepositoryConfiguration{
			RepositoryPath: config.RepositorySettings.RepositoryPath,
			FileLimit:      int32(config.RepositorySettings.FileLimit),
			FileSizeLimit:  config.RepositorySettings.FileSizeLimit,
		},
	}
}