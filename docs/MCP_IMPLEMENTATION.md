# Model Context Protocol (MCP) Implementation Guide

## Table of Contents
1. [Overview](#1-overview)
2. [Core Components](#2-core-components)
3. [Implementation Steps](#3-implementation-steps)
4. [Integration Points](#4-integration-points)
5. [Best Practices](#5-best-practices)
6. [Examples](#6-examples)

## 1. Overview

### 1.1. What is MCP?
Model Context Protocol (MCP) is a standardized approach for managing context and state in AI/ML model interactions. It ensures consistent handling of:
- Context management
- State transitions
- Model inputs/outputs
- Error handling
- Version control

### 1.2. Benefits for Hephaestus
- Standardized model interactions
- Improved context management
- Better error tracing
- Version compatibility
- Enhanced debugging capabilities

## 2. Core Components

### 2.1. Context Manager
```go
// internal/mcp/context.go
type ContextManager struct {
    ModelVersion string
    ContextID    string
    Metadata     map[string]interface{}
    State        *State
    History      []Interaction
}

type State struct {
    CurrentPhase    string
    PreviousPhase   string
    Variables       map[string]interface{}
    LastUpdateTime  time.Time
}

type Interaction struct {
    Timestamp   time.Time
    Input       string
    Output      string
    ModelID     string
    ContextID   string
    Metadata    map[string]interface{}
}
```

### 2.2. Protocol Handler
```go
// internal/mcp/protocol.go
type ProtocolHandler struct {
    contextManager *ContextManager
    modelClient    *ModelClient
    config        *Config
}

func (h *ProtocolHandler) ProcessRequest(ctx context.Context, req *Request) (*Response, error) {
    // Validate context
    if err := h.validateContext(req.Context); err != nil {
        return nil, fmt.Errorf("invalid context: %w", err)
    }

    // Prepare model input
    input := h.prepareModelInput(req)

    // Process through model
    output, err := h.modelClient.Process(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("model processing failed: %w", err)
    }

    // Update context
    h.contextManager.UpdateState(output)

    return h.prepareResponse(output), nil
}
```

## 3. Implementation Steps

### 3.1. Context Integration
```go
// internal/service/analyzer.go
type Analyzer struct {
    mcp         *mcp.ProtocolHandler
    modelClient *model.Client
}

func (a *Analyzer) AnalyzeLog(ctx context.Context, log string) (*Solution, error) {
    // Create MCP context
    mcpCtx := &mcp.Context{
        ID: uuid.New().String(),
        Metadata: map[string]interface{}{
            "type": "log_analysis",
            "timestamp": time.Now(),
        },
    }

    // Prepare request with context
    req := &mcp.Request{
        Context: mcpCtx,
        Input: log,
        Config: a.getAnalysisConfig(),
    }

    // Process through MCP
    resp, err := a.mcp.ProcessRequest(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("mcp processing failed: %w", err)
    }

    return a.convertToSolution(resp), nil
}
```

### 3.2. State Management
```go
// internal/mcp/state.go
type StateManager struct {
    store    StateStore
    emitter  EventEmitter
}

func (sm *StateManager) TransitionState(ctx context.Context, from, to string, data interface{}) error {
    // Validate transition
    if !sm.isValidTransition(from, to) {
        return ErrInvalidTransition
    }

    // Create state transition
    transition := &StateTransition{
        From:      from,
        To:        to,
        Timestamp: time.Now(),
        Data:      data,
    }

    // Store transition
    if err := sm.store.SaveTransition(ctx, transition); err != nil {
        return fmt.Errorf("failed to save transition: %w", err)
    }

    // Emit event
    sm.emitter.Emit(StateTransitionEvent{
        Transition: transition,
    })

    return nil
}
```

## 4. Integration Points

### 4.1. API Integration
```go
// internal/api/handlers/mcp_handler.go
func (h *MCPHandler) Handle(w http.ResponseWriter, r *http.Request) {
    var req MCPRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "invalid request")
        return
    }

    // Create or retrieve context
    ctx, err := h.contextManager.GetContext(req.ContextID)
    if err != nil {
        ctx = h.contextManager.NewContext()
    }

    // Process request
    resp, err := h.protocolHandler.ProcessRequest(r.Context(), &mcp.Request{
        Context: ctx,
        Input:   req.Input,
        Config:  req.Config,
    })
    if err != nil {
        respondError(w, http.StatusInternalServerError, "processing failed")
        return
    }

    respondJSON(w, http.StatusOK, resp)
}
```

### 4.2. Model Integration
```go
// internal/model/client.go
type ModelClient struct {
    mcpHandler *mcp.ProtocolHandler
    config     *Config
}

func (c *ModelClient) Process(ctx context.Context, input interface{}) (interface{}, error) {
    // Prepare MCP context
    mcpCtx := c.mcpHandler.PrepareContext(ctx)

    // Process with model
    output, err := c.processWithModel(mcpCtx, input)
    if err != nil {
        return nil, fmt.Errorf("model processing failed: %w", err)
    }

    // Update MCP context
    if err := c.mcpHandler.UpdateContext(mcpCtx, output); err != nil {
        return nil, fmt.Errorf("context update failed: %w", err)
    }

    return output, nil
}
```

## 5. Best Practices

### 5.1. Context Management
- Always maintain context throughout the request lifecycle
- Include relevant metadata in context
- Implement proper context cleanup
- Use context versioning

### 5.2. Error Handling
```go
// internal/mcp/errors.go
type MCPError struct {
    Code       string
    Message    string
    Context    *Context
    Timestamp  time.Time
}

func (e *MCPError) Error() string {
    return fmt.Sprintf("[%s] %s (Context: %s)", e.Code, e.Message, e.Context.ID)
}

func HandleMCPError(err error) *Response {
    if mcpErr, ok := err.(*MCPError); ok {
        return &Response{
            Status:  "error",
            Code:    mcpErr.Code,
            Message: mcpErr.Message,
            Context: mcpErr.Context,
        }
    }
    return &Response{
        Status:  "error",
        Code:    "UNKNOWN_ERROR",
        Message: err.Error(),
    }
}
```

## 6. Examples

### 6.1. Basic Usage
```go
// examples/mcp_basic.go
func main() {
    // Initialize MCP handler
    handler := mcp.NewProtocolHandler(config)

    // Create context
    ctx := handler.NewContext()

    // Process request
    req := &mcp.Request{
        Context: ctx,
        Input:   "Sample input",
    }

    resp, err := handler.ProcessRequest(context.Background(), req)
    if err != nil {
        log.Fatalf("Processing failed: %v", err)
    }

    fmt.Printf("Response: %+v\n", resp)
}
```

### 6.2. Advanced Usage
```go
// examples/mcp_advanced.go
func ProcessWithContext(input string) (*Solution, error) {
    // Initialize components
    handler := mcp.NewProtocolHandler(config)
    analyzer := NewAnalyzer(handler)

    // Create context with metadata
    ctx := handler.NewContext(mcp.WithMetadata(map[string]interface{}{
        "source": "log_analysis",
        "version": "1.0",
        "timestamp": time.Now(),
    }))

    // Process with state tracking
    req := &mcp.Request{
        Context: ctx,
        Input:   input,
        Config: &mcp.Config{
            TrackState: true,
            KeepHistory: true,
        },
    }

    // Process and handle response
    resp, err := analyzer.ProcessWithMCP(context.Background(), req)
    if err != nil {
        return nil, fmt.Errorf("processing failed: %w", err)
    }

    return convertToSolution(resp), nil
}
```

### 6.3. Configuration Example
```yaml
# config/mcp.yaml
mcp:
  version: "1.0"
  features:
    context_tracking: true
    state_management: true
    history_tracking: true
  timeouts:
    context_ttl: 3600s
    state_ttl: 1800s
  storage:
    type: "redis"
    config:
      host: "localhost"
      port: 6379
  logging:
    level: "info"
    format: "json"
``` 