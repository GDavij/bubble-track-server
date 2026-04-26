# Domain Modules

All business logic lives in `/internal/domain/`. This is where the interdisciplinary computation happens.

## Module Overview

| Module | Discipline | Key Functions |
|--------|------------|--------------|
| `attachment.go` | Psychology | Bowlby attachment classification |
| `social_exchange.go` | Psychology | Thibaut & Kelley cost/benefit |
| `sociology.go` | Sociology | Bourdieu, Granovetter, Dunbar |
| `philosophy.go` | Philosophy | Sartre, Aristotle, Gilligan |
| `mathematics.go` | Mathematics | Graph algorithms, PageRank |
| `physics.go` | Physics | Thermodynamics, resilience |
| `geography.go` | Geography | Gravity model |
| `anthropology.go` | Anthropology | Mauss gift economy |
| `economics.go` | Economics | Game theory |
| `neuroscience.go` | Neuroscience | Mirror neurons |
| `communication.go` | Communication | Watzlawick axioms |
| `history.go` | History | Path dependency |
| `node_metrics.go` | Core | Node/Graph snapshots |

## Core Entities

### Pessoa (Person)
```go
type Pessoa struct {
    ID        string
    UserID    string
    Name     string
    Aliases  []string
    Role     string
    CreatedAt time.Time
}
```

### Relacionamento (Relationship)
```go
type Relacionamento struct {
    ID              string
    UserID          string
    SourcePersonID  string
    TargetPersonID  string
    Quality        Quality  // positive, negative, neutral
    Strength       float64 // [0,1]
    ReciprocityIndex float64 // [0,1]
}
```

### Interacao (Interaction)
```go
type Interacao struct {
    ID         string
    UserID     string
    RawText    string
    Status     JobStatus
}
```

## Key Interfaces

```go
type PersonRepository interface {
    Create(ctx, p *Pessoa) error
    GetByID(ctx, id string) (*Pessoa, error)
    GetByUserID(ctx, userID string) ([]Pessoa, error)
    Update(ctx, p *Pessoa) error
    Upsert(ctx, p *Pessoa) error
}

type RelationshipRepository interface {
    Create(ctx, r *Relacionamento) error
    GetByID(ctx, id string) (*Relacionamento, error)
    GetByPeople(ctx, sourceID, targetID string) (*Relacionamento, error)
    GetByUser(ctx, userID string) ([]Relacionamento, error)
    Update(ctx, r *Relacionamento) error
    Upsert(ctx, r *Relacionamento) error
}

type InteractionRepository interface {
    Create(ctx, i *Interacao) error
    GetByID(ctx, id string) (*Interacao, error)
    UpdateStatus(ctx, id string, status JobStatus) error
}

type MemoryRepository interface {
    Store(ctx, userID, interactionID string, embedding []float32, metadata) error
    Search(ctx, userID string, queryEmbedding []float32, limit int) ([]MemoryResult, error)
}
```

## Node Metrics

### NodeMetrics
Computed dynamic metrics for each person node:

```go
type NodeMetrics struct {
    PersonID         string
    UserID          string
    ComputedAt       time.Time
    TimeWindow      string  // "all", "30d", etc.
    Degree          int     // Direct connections
    
    Centrality      CentralityScores
    Community      CommunityMetrics
    RelationalHealth RelationshipHealth
    SocialCapital  SocialCapital
    Attachment    AttachmentProfile
    HumanistScore  HumanistProfile
    SocialExchange SocialExchangeProfile
    
    TrendDirection TrendDirection
}
```

### CentralityScores
```go
type CentralityScores struct {
    Degree         float64
    Betweenness    float64
    Closeness      float64
    Eigenvector   float64
    PageRank      float64
    ClusteringCoef float64
}
```

## Error Types

```go
type NotFoundError struct {
    Resource string
    ID       string
}

type ValidationError struct {
    Field   string
    Message string
}

type AlreadyExistsError struct {
    Resource string
    ID       string
}
```

## Utility Functions

- `clamp(v, min, max float64) float64` — Constrain value to range [min, max]
- All compute functions return values clamped to [0, 1]