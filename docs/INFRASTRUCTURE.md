# Infrastructure

The infrastructure layer handles external integrations.

## Repositories

### PostgreSQL

| Repository | File | Purpose |
|------------|------|---------|
| Person | `postgres_person.go` | CRUD for pessoas |
| Relationship | `postgres_relationship.go` | CRUD for relacionamentos |
| Interaction | `postgres_interaction.go` | CRUD for interacoes |
| Graph | `postgres_graph.go` | Graph queries (CTEs) |
| GraphEngine | `graph_engine.go` | Node metrics persistence |

### Connection

```go
pool, err := repository.NewPostgresPool(ctx, cfg.Postgres)
defer pool.Close()

// Auto-migrations
repository.RunMigrations(ctx, pool)
repository.AddNodeMetricsMigrations(ctx, pool)
```

### Node Metrics Table

The `node_metrics` table stores computed metrics over time:

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| person_id | VARCHAR | Foreign key to pessoa |
| user_id | VARCHAR | Owner's user ID |
| computed_at | TIMESTAMP | When computed |
| time_window | VARCHAR | "all", "30d" |
| degree | INT | Direct connections |
| degree_centrality | FLOAT | Degree centrality |
| betweenness_centrality | FLOAT | Betweenness centrality |
| closeness_centrality | FLOAT | Closeness centrality |
| pagerank | FLOAT | PageRank score |
| clustering_coef | FLOAT | Local clustering |
| community_id | VARCHAR | Assigned community |
| embeddedness | FLOAT | Internal edges ratio |
| bridge_score | FLOAT | External edges ratio |
| relational_health | FLOAT | 6-dimension health |
| ... | ... | (25+ columns total) |

## Qdrant (Vector Store)

### Memory Storage

Embeddings stored in Qdrant for semantic search:

```go
qdrantRepo := repository.NewQdrantMemoryRepository(client, cfg.Qdrant)

// Store memory
qdrantRepo.Store(ctx, userID, interactionID, embedding, metadata)

// Search memories
results, err := qdrantRepo.Search(ctx, userID, queryEmbedding, limit)
```

### Collection

- Collection: `bubble_memories`
- Vector size: 768 (Gemini embeddings)
- Payload: `{interaction_id, user_id, people, content, timestamp}`

## Redis

### Queue (Asynq)

Background job processing:

```go
// Enqueue
asynqClient, _ := queue.NewClient(cfg.Redis)
enqueuer := queue.NewEnqueuer(asynqClient)
enqueuer.EnqueueAnalyze(ctx, payload)

// Process (worker)
server := queue.NewServer(cfg.Redis, log)
mux := queue.NewServeMux()
mux.HandleFunc(queue.TypeAnalyzeInteraction, handler)
server.Run(mux)
```

### Pub/Sub (Notifier)

Real-time notifications:

```go
notifier := pubsub.NewRedisNotifier(cfg.Redis)
notifier.Notify(ctx, userID, result)
```

## GenAI Client

Wrapper around `google.golang.org/genai`:

```go
client, err := adk.NewGenAIClient(ctx, cfg.GenAI)
defer client.Close()

// Generate with function calling
resp, err := client.GenerateWithContents(ctx, contents, config)
```

## Configuration

All config via environment variables (see README.md):

```go
cfg := config.Load()
// cfg.Postgres, cfg.Qdrant, cfg.Redis, cfg.GenAI, cfg.Server
```