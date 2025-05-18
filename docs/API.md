# Hephaestus API Documentation

## Overview

Hephaestus exposes its functionality through a gRPC API. This document describes the available services, their methods, and how to use them.

## Services

### HephaestusService

The main service interface for interacting with the Hephaestus system.

#### Node Management

1. **RegisterNode**
   ```protobuf
   rpc RegisterNode(RegisterNodeRequest) returns (RegisterNodeResponse)
   ```
   
   Registers a new node with the system.

   **Request**:
   ```protobuf
   message RegisterNodeRequest {
       string node_id = 1;
       string log_level = 2;
       string log_output = 3;
   }
   ```

   **Response**:
   ```protobuf
   message RegisterNodeResponse {
       string status = 1;
       string error = 2;
   }
   ```

2. **GetNodeStatus**
   ```protobuf
   rpc GetNodeStatus(NodeStatusRequest) returns (NodeStatusResponse)
   ```
   
   Retrieves the current status of a node.

   **Request**:
   ```protobuf
   message NodeStatusRequest {
       string node_identifier = 1;
   }
   ```

   **Response**:
   ```protobuf
   message NodeStatusResponse {
       string node_identifier = 1;
       string operational_status = 2;
       SystemConfiguration current_configuration = 3;
       string status_message = 4;
   }
   ```

#### Log Processing

1. **ProcessLogEntry**
   ```protobuf
   rpc ProcessLogEntry(ProcessLogEntryRequest) returns (ProcessLogEntryResponse)
   ```
   
   Processes a single log entry.

   **Request**:
   ```protobuf
   message ProcessLogEntryRequest {
       LogEntryData log_entry = 1;
   }
   ```

   **Response**:
   ```protobuf
   message ProcessLogEntryResponse {
       string status = 1;
       string error = 2;
   }
   ```

2. **StreamLogEntries**
   ```protobuf
   rpc StreamLogEntries(StreamLogEntriesRequest) returns (stream LogEntryData)
   ```
   
   Streams log entries from a node.

   **Request**:
   ```protobuf
   message StreamLogEntriesRequest {
       string node_identifier = 1;
       string log_level_filter = 2;
   }
   ```

#### Solution Management

1. **GetSolutionProposal**
   ```protobuf
   rpc GetSolutionProposal(GetSolutionProposalRequest) returns (GetSolutionProposalResponse)
   ```
   
   Generates a solution proposal for a log entry.

   **Request**:
   ```protobuf
   message GetSolutionProposalRequest {
       string node_id = 1;
       LogEntryData log_entry = 2;
   }
   ```

   **Response**:
   ```protobuf
   message GetSolutionProposalResponse {
       SolutionProposal solution = 1;
       string error = 2;
   }
   ```

2. **ValidateSolution**
   ```protobuf
   rpc ValidateSolution(ValidateSolutionRequest) returns (ValidateSolutionResponse)
   ```
   
   Validates a proposed solution.

   **Request**:
   ```protobuf
   message ValidateSolutionRequest {
       SolutionProposal solution = 1;
   }
   ```

   **Response**:
   ```protobuf
   message ValidateSolutionResponse {
       bool is_valid = 1;
       string error = 2;
   }
   ```

## Data Types

### LogEntryData
```protobuf
message LogEntryData {
    string node_identifier = 1;
    string log_level = 2;
    string log_message = 3;
    string log_timestamp = 4;
    map<string, string> log_metadata = 5;
    string error_trace = 6;
}
```

### SolutionProposal
```protobuf
message SolutionProposal {
    string solution_id = 1;
    string node_id = 2;
    LogEntryData associated_log = 3;
    string proposed_changes = 4;
    google.protobuf.Timestamp generation_time = 5;
    double confidence_score = 6;
}
```

### SystemConfiguration
```protobuf
message SystemConfiguration {
    RemoteRepositoryConfiguration remote_settings = 1;
    ModelServiceConfiguration model_settings = 2;
    LoggingConfiguration logging_settings = 3;
    string operational_mode = 4;
    RepositoryConfiguration repository_settings = 5;
}
```

## Usage Examples

### Go Client

```go
// Connect to the server
conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
if err != nil {
    log.Fatalf("did not connect: %v", err)
}
defer conn.Close()

// Create a client
client := pb.NewHephaestusServiceClient(conn)

// Register a node
registerResp, err := client.RegisterNode(ctx, &pb.RegisterNodeRequest{
    NodeId:    "test-node-1",
    LogLevel:  "info",
    LogOutput: "stdout",
})

// Process a log entry
logEntry := &pb.LogEntryData{
    NodeIdentifier: "test-node-1",
    LogLevel:      "error",
    LogMessage:    "Failed to connect to database",
    ErrorTrace:    "Error: connection refused",
}

processResp, err := client.ProcessLogEntry(ctx, &pb.ProcessLogEntryRequest{
    LogEntry: logEntry,
})
```

## Error Handling

The API uses standard gRPC error codes:

- `INVALID_ARGUMENT`: Invalid request parameters
- `NOT_FOUND`: Requested resource not found
- `ALREADY_EXISTS`: Resource already exists
- `FAILED_PRECONDITION`: Operation prerequisites not met
- `INTERNAL`: Internal server error
- `UNAVAILABLE`: Service temporarily unavailable

## Best Practices

1. **Connection Management**
   - Reuse client connections
   - Implement proper connection cleanup
   - Handle connection failures gracefully

2. **Error Handling**
   - Check error responses
   - Implement proper error handling
   - Log errors appropriately

3. **Context Usage**
   - Use context for timeouts
   - Pass context through calls
   - Handle context cancellation

4. **Security**
   - Use secure connections in production
   - Implement proper authentication
   - Validate input data 