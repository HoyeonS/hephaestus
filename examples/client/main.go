package main

import (
	"context"
	"flag"
	"log"
	"time"

	pb "github.com/HoyeonS/hephaestus/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "localhost:50051", "The server address")
	repo = flag.String("repo", "", "GitHub repository (owner/repo)")
	token = flag.String("token", "", "GitHub token")
	aiProvider = flag.String("ai-provider", "openai", "AI provider (e.g., openai)")
	aiKey = flag.String("ai-key", "", "AI provider API key")
	mode = flag.String("mode", "suggest", "Mode (suggest or deploy)")
)

func main() {
	flag.Parse()

	// Validate flags
	if *repo == "" || *token == "" || *aiKey == "" {
		log.Fatal("repo, token, and ai-key flags are required")
	}

	// Connect to server
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewHephaestusClient(conn)
	ctx := context.Background()

	// Initialize Hephaestus node
	resp, err := client.Initialize(ctx, &pb.InitializeRequest{
		Config: &pb.Config{
			Github: &pb.GitHubConfig{
				Repository: *repo,
				Branch:    "main",
				Token:     *token,
			},
			Ai: &pb.AIConfig{
				Provider: *aiProvider,
				ApiKey:   *aiKey,
				Model:    "gpt-4", // Can be made configurable
			},
			Log: &pb.LogConfig{
				Level:          "info",
				ThresholdLevel: "error",
			},
			Mode: *mode,
		},
	})
	if err != nil {
		log.Fatalf("failed to initialize: %v", err)
	}

	log.Printf("Initialized Hephaestus node with ID: %s", resp.NodeId)

	// Start streaming logs
	stream, err := client.StreamLogs(ctx)
	if err != nil {
		log.Fatalf("failed to start streaming: %v", err)
	}

	// Start goroutine to receive solutions
	go func() {
		for {
			solution, err := stream.Recv()
			if err != nil {
				log.Printf("Error receiving solution: %v", err)
				return
			}

			switch result := solution.Result.(type) {
			case *pb.SolutionResponse_SuggestedFix:
				log.Printf("Received fix suggestion:")
				log.Printf("Solution ID: %s", result.SuggestedFix.SolutionId)
				log.Printf("Description: %s", result.SuggestedFix.Description)
				for _, change := range result.SuggestedFix.Changes {
					log.Printf("File: %s (lines %d-%d)", change.FilePath, change.LineStart, change.LineEnd)
					log.Printf("Original:\n%s", change.OriginalCode)
					log.Printf("Updated:\n%s", change.UpdatedCode)
				}
			case *pb.SolutionResponse_PullRequest:
				log.Printf("Created pull request:")
				log.Printf("URL: %s", result.PullRequest.Url)
				log.Printf("Title: %s", result.PullRequest.Title)
				log.Printf("Branch: %s", result.PullRequest.Branch)
			}
		}
	}()

	// Simulate sending error logs
	for {
		err := stream.Send(&pb.LogEntry{
			NodeId:    resp.NodeId,
			Level:     "error",
			Message:   "Null pointer exception in user service",
			Timestamp: time.Now().Format(time.RFC3339),
			Metadata: map[string]string{
				"file":      "service/user.go",
				"line":      "42",
				"component": "UserService",
			},
			StackTrace: `goroutine 1 [running]:
main.(*UserService).GetUser(0x0, 0x123)
	service/user.go:42 +0x123
main.main()
	main.go:15 +0x456`,
		})
		if err != nil {
			log.Printf("Error sending log: %v", err)
			break
		}

		time.Sleep(5 * time.Second)
	}
} 