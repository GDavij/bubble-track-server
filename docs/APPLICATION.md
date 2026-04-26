# Application Layer

Orchestrates business logic using domain and infrastructure.

## Use Cases

### AnalyzeUseCase
```go
type AnalyzeUseCase struct {
    agent          *adk.Agent
    interactionRepo domain.InteractionRepository
    personRepo     domain.PersonRepository
    relationshipRepo domain.RelationshipRepository
    memoryRepo    domain.MemoryRepository
    embedder      domain.Embedder
    notifier     domain.AnalysisNotifier
    enqueuer     *queue.Enqueuer
    logger       *slog.Logger
}
```

**Methods:**
- `Submit(ctx, req AnalyzeRequest) -> (AnalyzeResponse, error)` — Enqueues async job, returns immediately
- `ProcessJob(ctx, payload) error` — Worker handler, calls agent and persists results

**Flow:**
1. Validate input
2. Create Interacao record (Pending)
3. Enqueue Asynq task
4. Return job ID (202 Accepted)
5. Worker processes via Agent loop
6. Persist people/relationships
7. Update Interacao status

### GraphAnalysisEngine
```go
type GraphAnalysisEngine struct {
    engine          *repository.GraphEngine
    personRepo      domain.PersonRepository
    relationshipRepo domain.RelationshipRepository
    logger          *slog.Logger
}
```

**Methods:**
- `ComputeAllMetrics(ctx, userID) -> (map[string]*NodeMetrics, error)` — All node metrics
- `ComputeGraphSnapshot(ctx, userID) -> *GraphSnapshot` — Graph-level metrics
- `GetMetricHistory(ctx, personID, metricName, limit) -> []MetricPoint`
- `PersistMetrics(ctx, metrics)` — Persists to node_metrics table

### ClassificationEngine
```go
type ClassificationEngine struct{}
```

**Methods:**
- `ClassifyRole(m *NodeMetrics) -> *InterdisciplinaryRoleScore` — Single person role
- `ClassifyAllRoles(metrics map[string]*NodeMetrics) -> map[string]*InterdisciplinaryRoleScore`

**Role Scoring (6 roles):**
- Bridge — Structural holes, connects groups
- Mentor — High influence, growth-oriented
- Anchor — Central, stabilizes network
- Catalyst — Change driver
- Observer — Peripheral info gatherer
- Drain — Energy depleting

### AggregationEngine
```go
type AggregationEngine struct {
    graphEngine *repository.GraphEngine
}
```

**Methods:**
- `AggregateProfile(m *NodeMetrics, all map[string]*NodeMetrics) -> *AggregatedProfile`
- `AggregateAll(metrics) -> []*AggregatedProfile`

**Output includes:**
- Graph position (rank, percentiles)
- Behavioral patterns (trend, stability)
- Profile summary

## Queues

### Task Types
- `analyze_interaction` — Process interaction text

### Payload
```go
type AnalyzePayload struct {
    InteractionID string
    UserID         string
    RawText        string
}
```

## Worker

```go
mux := queue.NewServeMux()
mux.HandleFunc(queue.TypeAnalyzeInteraction, handler)
server := queue.NewServer(cfg.Redis, log)
server.Run(mux)
```