package main

import (
	"context"
	"flag"
	"log"

	"github.com/HoyeonS/hephaestus/pkg/hephaestus"
	pb "github.com/HoyeonS/hephaestus/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	useGRPC = flag.Bool("grpc", false, "Use gRPC API instead of direct client")
	addr    = flag.String("addr", "localhost:50051", "The server address (for gRPC mode)")
	repo    = flag.String("repo", "", "GitHub repository (owner/repo)")
	token   = flag.String("token", "", "GitHub token")
	aiKey   = flag.String("ai-key", "", "AI provider API key")
)

func main() {
	flag.Parse()

	if *repo == "" || *token == "" || *aiKey == "" {
		log.Fatal("repo, token, and ai-key flags are required")
	}

	if *useGRPC {
		initializeWithGRPC()
	} else {
		initializeWithClient()
	}
}

func initializeWithGRPC() {
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewHephaestusClient(conn)
	resp, err := client.Initialize(context.Background(), &pb.InitializeRequest{
		Config: &pb.Config{
			Github: &pb.GitHubConfig{
				Repository: *repo,
				Branch:    "main",
				Token:     *token,
			},
			Ai: &pb.AIConfig{
				Provider: "openai",
				ApiKey:   *aiKey,
			},
			Log: &pb.LogConfig{
				Level: "info",
			},
			Mode: "suggest",
		},
	})
	if err != nil {
		log.Fatalf("initialization failed: %v", err)
	}

	log.Printf("Node initialized with ID: %s", resp.NodeId)
}

func initializeWithClient() {
	config := hephaestus.DefaultConfig()
	config.AIProvider = "openai"
	config.AIConfig = map[string]string{
		"api_key": *aiKey,
	}

	if err := config.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	hephaestusConfig, err := config.HephaestusConfigFactoryWithDefault()
	if err != nil {
		log.Fatalf("failed to create config: %v", err)
	}

	client, err := hephaestus.New(hephaestusConfig)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	if err = client.Start(context.Background()); err != nil {
		log.Fatalf("failed to start client: %v", err)
	}

	log.Println("Client initialized successfully")
} 