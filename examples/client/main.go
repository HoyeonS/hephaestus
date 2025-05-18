package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	pb "github.com/HoyeonS/hephaestus/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "localhost:50051", "The server address")
	repo = flag.String("repo", "", "GitHub repository (owner/repo)")
	token = flag.String("token", "", "GitHub token")
	aiKey = flag.String("ai-key", "", "AI provider API key")
)

func main() {
	flag.Parse()

	if *repo == "" || *token == "" || *aiKey == "" {
		log.Fatal("repo, token, and ai-key flags are required")
	}

	// Set up a connection to the server
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Create client
	client := pb.NewHephaestusClient(conn)

	// Initialize Hephaestus
	ctx := context.Background()
	resp, err := client.Initialize(ctx, &pb.InitializeRequest{
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
				Level: "error",
			},
			Mode: "suggest",
		},
	})
	if err != nil {
		log.Fatalf("could not initialize: %v", err)
	}
	log.Printf("Initialized with node ID: %s", resp.NodeId)

	// Start streaming logs
	stream, err := client.StreamLogs(ctx)
	if err != nil {
		log.Fatalf("could not start streaming: %v", err)
	}

	// Send example error log
	err = stream.Send(&pb.LogEntry{
		Level:     "error",
		Message:   "Null pointer exception in user service",
		Timestamp: time.Now().Format(time.RFC3339),
		Metadata: map[string]string{
			"node_id":    resp.NodeId,
			"file":       "service/user.go",
			"line":       "42",
			"component":  "UserService",
			"operation": "GetUser",
		},
		StackTrace: `goroutine 1 [running]:
main.(*UserService).GetUser(0x0, 0x123)
	service/user.go:42 +0x123
main.main()
	main.go:15 +0x456`,
	})
	if err != nil {
		log.Fatalf("could not send log: %v", err)
	}

	// Receive and print fix
	fix, err := stream.Recv()
	if err != nil {
		log.Fatalf("could not receive fix: %v", err)
	}

	switch x := fix.Result.(type) {
	case *pb.FixResponse_SuggestedFix:
		log.Printf("Received fix suggestion:\n")
		log.Printf("File: %s\n", x.SuggestedFix.FilePath)
		log.Printf("Original code:\n%s\n", x.SuggestedFix.OriginalCode)
		log.Printf("Suggested fix:\n%s\n", x.SuggestedFix.SuggestedCode)
		log.Printf("Explanation: %s\n", x.SuggestedFix.Explanation)
	case *pb.FixResponse_PullRequest:
		log.Printf("Created pull request: %s\n", x.PullRequest.Url)
	}
} 