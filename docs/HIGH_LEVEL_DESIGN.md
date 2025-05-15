# Hephaestus High-Level Design

## Overall System Architecture

```mermaid
graph TB
    subgraph "Log Sources"
        A1[Application Logs]
        A2[System Logs]
        A3[Error Reports]
    end

    subgraph "Kafka Layer"
        K1[Kafka Cluster]
        K2[ZooKeeper Ensemble]
        T1[Raw Logs Topic]
        T2[Structured Logs Topic]
        T3[Error Events Topic]
    end

    subgraph "Processing Layer"
        P1[Log Collector Service]
        P2[Log Processor Service]
        P3[Error Analyzer Service]
    end

    subgraph "MCP Layer"
        M1[Model Composer]
        M2[Model Executor]
        M3[Solution Validator]
    end

    subgraph "Model Pool"
        MP1[Syntax Analyzer]
        MP2[Logic Analyzer]
        MP3[Code Fixer]
        MP4[Test Generator]
        MP5[Doc Generator]
    end

    subgraph "Deployment Layer"
        D1[Sandbox Environment]
        D2[Deployment Manager]
        D3[Rollback Handler]
    end

    %% Log flow
    A1 & A2 & A3 --> K1
    K1 --> T1
    T1 --> P1
    P1 --> T2
    T2 --> P2
    P2 --> T3
    T3 --> P3

    %% MCP flow
    P3 --> M1
    M1 --> M2
    M2 --> MP1 & MP2 & MP3 & MP4 & MP5
    MP1 & MP2 & MP3 & MP4 & MP5 --> M3
    M3 --> D1
    D1 --> D2
    D2 --> D3

    %% Kafka management
    K2 <--> K1

    style K1 fill:#e1f5fe,stroke:#0288d1
    style K2 fill:#e1f5fe,stroke:#0288d1
    style M1 fill:#f3e5f5,stroke:#7b1fa2
    style M2 fill:#f3e5f5,stroke:#7b1fa2
    style M3 fill:#f3e5f5,stroke:#7b1fa2
```

## Log Processing Pipeline

```mermaid
sequenceDiagram
    participant LS as Log Sources
    participant KC as Kafka Cluster
    participant LC as Log Collector
    participant LP as Log Processor
    participant EA as Error Analyzer
    participant MC as Model Composer

    LS->>KC: Send raw logs
    KC->>LC: Consume raw logs
    LC->>KC: Produce structured logs
    KC->>LP: Consume structured logs
    LP->>KC: Produce error events
    KC->>EA: Consume error events
    EA->>MC: Send error analysis
```

## Model Composition Flow

```mermaid
stateDiagram-v2
    [*] --> ErrorAnalysis
    ErrorAnalysis --> SyntaxAnalysis
    ErrorAnalysis --> LogicAnalysis
    
    SyntaxAnalysis --> SyntaxFix
    LogicAnalysis --> LogicFix
    
    SyntaxFix --> UnitTests
    LogicFix --> IntegrationTests
    
    UnitTests --> ValidationPhase
    IntegrationTests --> ValidationPhase
    
    ValidationPhase --> Deployment
    ValidationPhase --> ErrorAnalysis: Validation Failed
    
    Deployment --> [*]
```

## Deployment Pipeline

```mermaid
graph LR
    subgraph "Validation"
        V1[Unit Tests]
        V2[Integration Tests]
        V3[Performance Tests]
    end

    subgraph "Sandbox"
        S1[Test Environment]
        S2[Load Testing]
    end

    subgraph "Deployment"
        D1[Staging]
        D2[Production]
        D3[Rollback Handler]
    end

    V1 & V2 & V3 --> S1
    S1 --> S2
    S2 --> D1
    D1 --> D2
    D2 --> D3
    D3 -.-> D1: Rollback if needed

    style V1 fill:#e8f5e9,stroke:#2e7d32
    style V2 fill:#e8f5e9,stroke:#2e7d32
    style V3 fill:#e8f5e9,stroke:#2e7d32
    style D1 fill:#fff3e0,stroke:#ef6c00
    style D2 fill:#fff3e0,stroke:#ef6c00
```

## Monitoring and Metrics

```mermaid
graph TB
    subgraph "System Metrics"
        SM1[Log Ingestion Rate]
        SM2[Processing Latency]
        SM3[Error Detection Rate]
    end

    subgraph "Model Metrics"
        MM1[Model Performance]
        MM2[Fix Success Rate]
        MM3[Generation Time]
    end

    subgraph "Business Metrics"
        BM1[Fix Accuracy]
        BM2[System Uptime]
        BM3[Cost per Fix]
    end

    subgraph "Alerting"
        A1[Alert Manager]
        A2[Notification System]
    end

    SM1 & SM2 & SM3 --> A1
    MM1 & MM2 & MM3 --> A1
    BM1 & BM2 & BM3 --> A1
    A1 --> A2

    style SM1 fill:#f3e5f5,stroke:#7b1fa2
    style MM1 fill:#e1f5fe,stroke:#0288d1
    style BM1 fill:#e8f5e9,stroke:#2e7d32
    style A1 fill:#fff3e0,stroke:#ef6c00
```

These diagrams provide a visual representation of:
1. Overall system architecture showing all components and their interactions
2. Log processing pipeline sequence
3. Model composition and state transitions
4. Deployment pipeline with validation steps
5. Monitoring and metrics structure

Each component is color-coded for better visualization:
- Blue: Kafka and infrastructure components
- Purple: MCP and model components
- Green: Validation components
- Orange: Deployment components 