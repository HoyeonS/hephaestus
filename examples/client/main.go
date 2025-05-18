package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/HoyeonS/hephaestus/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Set up a connection to the server
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Create a client
	client := pb.NewHephaestusServiceClient(conn)

	fmt.Println(client)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	fmt.Println(ctx)

	// Register a node
	nodeID := "test-node-1"
	registerResp, err := client.RegisterNode(ctx, &pb.RegisterNodeRequest{
		NodeId:    nodeID,
		LogLevel:  "info",
		LogOutput: "stdout",
	})
	if err != nil {
		log.Fatalf("could not register node: %v", err)
	}
	fmt.Printf("Node registration status: %s\n", registerResp.Status)

	// // Process a log entry
	// logEntry := &pb.LogEntryData{
	// 	NodeIdentifier: nodeID,
	// 	LogLevel:       "error",
	// 	LogMessage:     "Failed to connect to database",
	// 	ErrorTrace:     "Error: connection refused\nStack trace: ...",
	// 	LogTimestamp:   time.Now().Format(time.RFC3339),
	// 	LogMetadata: map[string]string{
	// 		"component": "database",
	// 		"service":   "user-service",
	// 	},
	// }

	// processResp, err := client.ProcessLogEntry(ctx, &pb.ProcessLogEntryRequest{
	// 	LogEntry: logEntry,
	// })
	// if err != nil {
	// 	log.Fatalf("could not process log entry: %v", err)
	// }
	// fmt.Printf("Log processing status: %s\n", processResp.Status)

	// // Get a solution proposal
	// solutionResp, err := client.GetSolutionProposal(ctx, &pb.GetSolutionProposalRequest{
	// 	NodeId:   nodeID,
	// 	LogEntry: logEntry,
	// })
	// if err != nil {
	// 	log.Fatalf("could not get solution proposal: %v", err)
	// }

	// if solutionResp.Error != "" {
	// 	fmt.Printf("Error getting solution: %s\n", solutionResp.Error)
	// } else {
	// 	fmt.Printf("Solution ID: %s\n", solutionResp.Solution.SolutionId)
	// 	fmt.Printf("Proposed changes: %s\n", solutionResp.Solution.ProposedChanges)
	// 	fmt.Printf("Confidence score: %.2f\n", solutionResp.Solution.ConfidenceScore)
	// }

	// // Validate the solution
	// if solutionResp.Solution != nil {
	// 	validateResp, err := client.ValidateSolution(ctx, &pb.ValidateSolutionRequest{
	// 		Solution: solutionResp.Solution,
	// 	})
	// 	if err != nil {
	// 		log.Fatalf("could not validate solution: %v", err)
	// 	}

	// 	if validateResp.Error != "" {
	// 		fmt.Printf("Error validating solution: %s\n", validateResp.Error)
	// 	} else {
	// 		fmt.Printf("Solution is valid: %v\n", validateResp.IsValid)
	// 	}
	// }
}
