# Bubble Track Server

A Go backend for analyzing social dynamics from free-text interaction reports using interdisciplinary frameworks spanning psychology, sociology, philosophy, mathematics, physics, and more — powered by Gemini AI via an agentic tool-calling loop.

## Overview

Bubble Track ingests free-text descriptions of social interactions (e.g., "Met PersonA for lunch today. PersonB joined us unexpectedly. PersonB and I have been distant lately..."), uses Gemini to extract:
- **People** mentioned (names, roles, aliases)
- **Relationships** between them (quality, strength, reciprocity)
- **Social roles** (bridge-builder, mentor, anchor, catalyst, observer, drain)

...and computes dynamic metrics from **10+ academic disciplines**.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         API Layer (Echo)                             │
│  POST /api/analyze    GET /api/graph    GET /api/people/:id/metrics   │
└──────────────────────────┬──────────────────────────────────────────────┘
                         │
┌───────────────────────▼───────────────────────────────────────────┐
│                   Application Layer                               │
│  AnalyzeUseCase  │  GraphAnalysisEngine  │  ClassificationEngine │
└──────────────────┬──────────────────────────────────────────────┘
                     │
┌───────────────────▼───────────────────────────────────────────┐
│                     Domain Layer                              │
│  Psychology  │  Sociology  │  Philosophy  │  Mathematics  │
│  Physics    │  Geography  │  Anthropology │  Economics    │
│  Neuroscience │  Communication │  History    │  NodeMetrics  │
└──────────────────────┬──────────────────────────────────────┘
                        │
┌──────────────────────▼──────────────────────────────────────┐
│                  Infrastructure Layer                        │
│  ADK Agent  │  GenAI Client  │  Qdrant Memory  │  Postgres │
│  Redis Queue  │  Redis Pub/Sub                           │
└─────────────────────────────────────────────────────────┘
```

## Tech Stack

- **Language**: Go 1.26+
- **Web Framework**: Echo v4
- **AI**: google.golang.org/genai v1.54 (Gemini)
- **Vector DB**: Qdrant (gRPC client)
- **Queue**: hibiken/asynq (Redis-backed)
- **Database**: PostgreSQL (pgx/v5)
- **Pub/Sub**: Redis

## Quick Start

```bash
# Build
go build ./...

# Run (requires PostgreSQL, Qdrant, Redis)
go run ./cmd/server

# Test
go test ./...
```

## API Endpoints

### Core
- `POST /api/analyze` — Submit interaction text for analysis (returns job ID)
- `GET /api/graph` — Get user's social graph
- `GET /api/health` — Health check

### People & Analysis
- `GET /api/people` — List all people
- `GET /api/people/:id` — Person detail with classification
- `GET /api/people/:id/metrics` — Dynamic node metrics
- `GET /api/people/:id/history/:metric` — Metric history over time

### Aggregate Analysis
- `GET /api/analysis/roles` — All role classifications
- `GET /api/analysis/profiles` — All aggregated profiles
- `GET /api/analysis/graph/snapshot` — Graph snapshot metrics
- `GET /api/relationships` — All relationships
- `GET /api/relationships/:id/health` — Relationship health

## Interdisciplinary Analysis

| Discipline | Framework | Key Metrics |
|-----------|-----------|------------|
| **Psychology** | Bowlby Attachment, Social Exchange (Thibaut & Kelley) | Attachment style, Satisfaction, Investment |
| **Sociology** | Bourdieu Capital, Granovetter Weak Ties, Dunbar | Bonding/Bridging capital, 5/15/50/150 layers |
| **Philosophy** | Sartre Existentialism, Aristotle Virtue, Gilligan Care | Agency, Authenticity, Phronesis |
| **Mathematics** | Graph Theory, PageRank, Louvain | Degree/Betweenness/Closeness centrality |
| **Physics** | Thermodynamics, Complex Systems | Social entropy, Phase transitions |
| **Geography** | Gravity Model, Distance Decay | Proximity, Mobility range |
| **Anthropology** | Mauss Gift Economy, Fictive Kin | Obligation load, Credit/Debt |
| **Economics** | Game Theory, Prisoner's Dilemma | Cooperation rate, Nash equilibrium |
| **Neuroscience** | Mirror Neurons, Hormones | Oxytocin/Dopamine proxies |
| **Communication** | Watzlawick Axioms, Narrative | Message entropy, Coherence |
| **History** | Path Dependency, Howe-Strauss | Lock-in, Generational cycles |

## Project Structure

```
/workspaces/bubble-track-server
├── cmd/server/main.go           # Entry point, DI wiring
├── internal/
│   ├── api/
│   │   ├── handler.go          # HTTP handlers
│   │   └── middleware.go       # Auth middleware
│   ├── application/
│   │   ├── analyze.go         # AnalyzeUseCase
│   │   ├── graph_analysis.go # GraphAnalysisEngine
│   │   ├── classification.go # ClassificationEngine
│   │   └── aggregation.go   # AggregationEngine
│   ├── domain/
│   │   ├── interfaces.go     # Repository interfaces
│   │   ├── person.go       # Pessoa entity
│   │   ├── relationship.go # Relacionamento entity
│   │   ├── interaction.go # Interacao entity
│   │   ├── graph.go       # Graph types
│   │   ├── node_metrics.go # NodeMetrics, GraphSnapshot
│   │   ├── errors.go     # Error types, clamp()
│   │   ├── attachment.go # Psychology: Bowlby
│   │   ├── social_exchange.go # Psychology: Thibaut & Kelley
│   │   ├── sociology.go  # Bourdieu, Granovetter, Dunbar
│   │   ├── philosophy.go # Sartre, Aristotle, Gilligan
│   │   ├── mathematics.go # Graph algorithms
│   │   ├── physics.go    # Thermodynamics
│   │   ├── geography.go # Gravity model
│   │   ├── anthropology.go # Mauss, Fictive kin
│   │   ├── economics.go # Game theory
│   │   ├── neuroscience.go # Mirror neurons
│   │   ├── communication.go # Watzlawick
│   │   └── history.go   # Path dependency
│   ├── infrastructure/
│   │   ├── adk/
│   │   │   ├── agent.go      # Agent loop (max 10 iterations)
│   │   │   ├── client.go    # GenAI wrapper
│   │   │   ├── prompts.go   # 10-discipline system prompt
│   │   │   ├── tools.go    # Tool definitions
│   │   │   └── executor.go # Tool executor
│   │   ├── repository/
│   │   │   ├── postgres_person.go
│   │   │   ├── postgres_relationship.go
│   │   │   ├── postgres_interaction.go
│   │   │   ├── postgres_graph.go # Graph queries
│   │   │   ├── graph_engine.go # Node metrics persistence
│   │   │   └── qdrant.go    # Memory vector store
│   │   ├── queue/
│   │   │   ├── tasks.go   # Task definitions
│   │   │   ├── client.go # Asynq client
│   │   │   └── mux.go   # Server mux
│   │   └── pubsub/
│   │       └── notifier.go # Redis pub/sub
│   ├── config/
│   │   └── config.go
│   ├── logger/
│   │   └── logger.go
│   └── mock/
│       └── user.go # Mock auth provider
├── go.mod
├── go.sum
└── .devcontainer/
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|------------|
| `POSTGRES_HOST` | postgres.database | PostgreSQL host |
| `POSTGRES_PORT` | 5432 | PostgreSQL port |
| `POSTGRES_USER` | bubble | Database user |
| `POSTGRES_PASSWORD` | — | Database password |
| `POSTGRES_DB` | bubble | Database name |
| `QDRANT_HOST` | qdrant.database | Qdrant host |
| `QDRANT_PORT` | 6334 | Qdrant gRPC port |
| `REDIS_HOST` | redis-stack.stack | Redis host |
| `REDIS_PORT` | 6379 | Redis port |
| `GENAI_API_KEY` | — | Gemini API key |
| `SERVER_PORT` | 8080 | HTTP server port |
| `SERVER_ENV` | development | production |

## Testing

```bash
# Run all tests
go test ./...

# Run specific package
go test ./internal/domain/...

# Run with coverage
go test -cover ./...
```

## License

MIT
