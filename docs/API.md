# API Reference

The API layer exposes HTTP endpoints via Echo v4 with JSON request/response bodies.

## Authentication

All endpoints (except `/auth/*`, `/api/health`, and `/ws`) require a JWT Bearer token in the `Authorization` header.

### JWT Authentication

```
Authorization: Bearer <access_token>
```

- **Algorithm:** HS256
- **Access token lifetime:** 15 minutes
- **Refresh token lifetime:** 7 days
- **Claims:** `sub` (user ID), `email`, `name`, `iat`, `exp`

### Auth Flow

```
1. POST /auth/register → creates account → returns access + refresh tokens
2. POST /auth/login    → validates credentials → returns access + refresh tokens
3. Use access token in Authorization header for all API requests
4. POST /auth/refresh  → exchange refresh token for new access + refresh tokens
5. POST /auth/logout   → revokes all refresh tokens for the user
```

---

## Auth Endpoints

### POST /auth/register

Create a new account.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123",
  "display_name": "My Name"
}
```

**Validation:**
- `email`: required, valid format
- `password`: required, min 8 characters
- `display_name`: required, min 1 character

**Response (201 Created):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "dGhpcyBpcyBhIHJlZn...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "display_name": "My Name"
  }
}
```

**Errors:**
| Status | When |
|--------|------|
| 400 | Invalid request body or validation failure |
| 409 | Email already registered |

---

### POST /auth/login

Authenticate with email and password.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "dGhpcyBpcyBhIHJlZn...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "display_name": "My Name"
  }
}
```

**Errors:**
| Status | When |
|--------|------|
| 400 | Invalid request body |
| 401 | Invalid email or password |

---

### POST /auth/refresh

Exchange a valid refresh token for new token pair.

**Request:**
```json
{
  "refresh_token": "dGhpcyBpcyBhIHJlZn..."
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "bmV3IHJlZnJlc2ggdG9r...",
  "token_type": "Bearer",
  "expires_in": 900
}
```

**Errors:**
| Status | When |
|--------|------|
| 400 | Invalid request body |
| 401 | Invalid or revoked refresh token |

---

### POST /auth/logout

Revoke all refresh tokens for the authenticated user.

**Request:** (requires Bearer token)
```json
{}
```

**Response (200 OK):**
```json
{
  "message": "logged out successfully"
}
```

---

### GET /auth/me

Get current authenticated user profile.

**Request:** (requires Bearer token)

**Response (200 OK):**
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "display_name": "My Name"
}
```

## Routes

```go
func (h *Handler) RegisterRoutes(e *echo.Echo) {
    api := e.Group("/api")

    // Core
    api.POST("/analyze", h.Analyze)
    api.GET("/graph", h.GetGraph)
    api.GET("/graph/full", h.GetFullGraph)
    api.GET("/health", h.Health)
    api.GET("/analysis/stream", h.AnalysisStream)
    api.GET("/memories", h.GetMemories)
    api.POST("/chat", h.SendChatMessage)
    api.GET("/chat", h.GetChatMessages)
    api.GET("/states", h.GetEmotionalTimeline)
    api.GET("/states/self", h.GetSelfStates)

    // People
    people := api.Group("/people")
    people.GET("", h.ListPeople)
    people.GET("/:id", h.GetPersonDetail)
    people.GET("/:id/metrics", h.GetPersonMetrics)
    people.POST("/:id/classify", h.ClassifySocialRole)

    // Analysis
    analysis := api.Group("/analysis")
    analysis.GET("/roles", h.GetAllRoles)
    analysis.GET("/profiles", h.GetAllProfiles)
    analysis.GET("/graph/snapshot", h.GetGraphSnapshot)

    // Relationships
    relationships := api.Group("/relationships")
    relationships.GET("/:id/health", h.GetRelationshipHealth)
}
```

---

## Core Endpoints

### POST /api/analyze

Submit interaction text for multi-disciplinary social analysis. Returns immediately with a job ID; processing is async via Redis queue.

**Request:**
```json
{
  "text": "Met PersonA for lunch today..."
}
```

**Response (202 Accepted):**
```json
{
  "interaction_id": "uuid",
  "job_id": "asynq:uuid",
  "status": "pending"
}
```

**Pipeline:** `Handler → AnalyzeUseCase.Submit → Redis Queue → Worker → Agent Loop → Persist Results → Notify`

---

### GET /api/graph

Get the user's complete social graph with nodes, edges, and stats.

**Response (200 OK):**
```json
{
  "user_id": "test-user",
  "nodes": [
    {
      "id": "person-uuid",
      "display_name": "PersonA",
      "social_role": "bridge",
      "current_mood": "happy",
      "current_energy": 0.8,
      "interaction_count": 5
    }
  ],
  "edges": [
    {
      "source": "user-uuid",
      "target": "person-uuid",
      "quality": "nourishing",
      "strength": 0.85,
      "source_weight": 0.9,
      "target_weight": 0.8,
      "protocol": "deep",
      "reciprocity_index": 0.75
    }
  ],
  "stats": {
    "total_people": 10,
    "total_relationships": 15,
    "avg_reciprocity": 0.72,
    "bridge_count": 3,
    "strongest_connection": "PersonA ↔ PersonB"
  }
}
```

---

### GET /api/graph/full

Get the complete social graph with all computed data — nodes, edges, metrics, role classifications, aggregated profiles, person states, and emotional timeline. Single-request alternative to calling multiple endpoints separately.

**Response (200 OK):**
```json
{
  "graph": {
    "user_id": "test-user",
    "nodes": [
      {
        "id": "person-uuid",
        "display_name": "PersonA",
        "social_role": "bridge",
        "current_mood": "happy",
        "current_energy": 0.8,
        "interaction_count": 5
      }
    ],
    "edges": [
      {
        "source": "user-uuid",
        "target": "person-uuid",
        "quality": "nourishing",
        "strength": 0.85,
        "source_weight": 0.9,
        "target_weight": 0.8,
        "protocol": "deep",
        "reciprocity_index": 0.75
      }
    ],
    "stats": {
      "total_people": 10,
      "total_relationships": 15,
      "avg_reciprocity": 0.72,
      "bridge_count": 3,
      "strongest_connection": "PersonA ↔ PersonB"
    }
  },
  "metrics": {
    "person-uuid": {
      "degree": 5,
      "centrality_degree": 0.5,
      "centrality_betweenness": 0.3,
      "relational_health_overall": 0.75,
      "social_capital_total": 0.8
    }
  },
  "roles": {
    "person-uuid": {
      "primary_role": "bridge",
      "confidence": 0.85,
      "scores": { "bridge": 0.85, "mentor": 0.3 }
    }
  },
  "profiles": {
    "person-uuid": {
      "rank": 2,
      "percentile": 85,
      "trend": "strengthening",
      "stability": 0.8,
      "summary": "Strong bridge between social groups..."
    }
  },
  "person_states": {
    "person-uuid": [
      {
        "mood": "happy",
        "energy": 0.8,
        "valence": 0.6,
        "context": "gathering",
        "created_at": "2025-01-15T10:30:00Z"
      }
    ]
  },
  "timeline": [
    {
      "person_name": "PersonA",
      "mood": "happy",
      "energy": 0.8,
      "valence": 0.6,
      "context": "gathering",
      "created_at": "2025-01-15T10:30:00Z"
    }
  ]
}
```

---

### GET /api/health

Health check endpoint.

**Response (200 OK):**
```json
{
  "status": "ok"
}
```

---

### GET /api/analysis/stream

Server-Sent Events (SSE) stream for real-time analysis results. Keeps connection alive and pushes analysis events as they complete.

**Response:** `text/event-stream`

```
data: {"people_extracted": [...], "summary": "...", ...}

data: {"people_extracted": [...], "summary": "...", ...}
```

---

### GET /api/memories

Search interaction memories using RAG (Retrieval-Augmented Generation) with semantic vector search.

**Query Params:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `query` | string | required | Search query text |
| `limit` | int | 10 | Max results |
| `segment` | string | | Filter by segment |
| `session` | string | | Filter by session |
| `person` | string | | Filter by person name |

**Response (200 OK):**
```json
{
  "memories": [
    {
      "content": "interaction text...",
      "score": 0.89,
      "metadata": {"people": ["PersonA"], "interaction_id": "uuid"}
    }
  ]
}
```

---

## Chat Endpoints

### POST /api/chat

Send a chat message. Also triggers analysis pipeline on the message content.

**Request:**
```json
{
  "sender": "You",
  "content": "Tell me about my relationship with PersonA",
  "session_id": "optional-session-id"
}
```

**Response (201 Created):**
```json
{
  "message": {
    "id": "uuid",
    "user_id": "test-user",
    "sender": "You",
    "content": "Tell me about my relationship with PersonA",
    "is_user": true,
    "session_id": "optional-session-id",
    "created_at": "2025-01-15T10:30:00Z"
  },
  "analysis_job": {
    "interaction_id": "uuid",
    "job_id": "asynq:uuid",
    "status": "pending"
  }
}
```

---

### GET /api/chat

Get chat message history.

**Query Params:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `session_id` | string | | Filter by session |
| `limit` | int | 50 | Max messages |

**Response (200 OK):**
```json
{
  "messages": [
    {
      "id": "uuid",
      "sender": "You",
      "content": "Hello!",
      "is_user": true,
      "created_at": "2025-01-15T10:30:00Z"
    }
  ]
}
```

---

## Emotional State Endpoints

### GET /api/states

Get emotional state timeline across all people.

**Query Params:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `limit` | int | 50 | Max results (max 100) |

**Response (200 OK):**
```json
{
  "states": [
    {
      "person_id": "uuid",
      "person_name": "PersonA",
      "mood": "happy",
      "energy": 0.8,
      "valence": 0.6,
      "context": "gathering",
      "trigger": "positive interaction",
      "created_at": "2025-01-15T10:30:00Z"
    }
  ],
  "count": 25
}
```

---

### GET /api/states/self

Get emotional states for the user's own self-reflection entries.

**Query Params:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `limit` | int | 50 | Max results (max 100) |

**Response (200 OK):**
```json
{
  "states": [
    {
      "mood": "hopeful",
      "energy": 0.7,
      "valence": 0.5,
      "context": "reflection",
      "created_at": "2025-01-15T10:30:00Z"
    }
  ],
  "count": 10
}
```

---

## People Endpoints

### GET /api/people

List all people from the user's social graph.

**Response (200 OK):** Array of `GraphNode` objects (see GET /api/graph nodes).

---

### GET /api/people/:id

Get person detail with computed metrics, role classification, and aggregated profile.

**Response (200 OK):**
```json
{
  "person": {
    "person_id": "uuid",
    "centrality": {
      "degree": 0.5,
      "betweenness": 0.3,
      "closeness": 0.7,
      "eigenvector": 0.4,
      "pagerank": 0.15,
      "clustering_coef": 0.6
    },
    "relational_health": { "overall": 0.75 },
    "social_capital": { "bonding": 0.8, "bridging": 0.6 },
    "attachment": { "style": "secure", "score": 0.7 }
  },
  "classification": {
    "primary_role": "bridge",
    "confidence": 0.85,
    "scores": {
      "bridge": 0.85,
      "mentor": 0.3,
      "anchor": 0.6,
      "catalyst": 0.2,
      "observer": 0.1,
      "drain": 0.05
    }
  },
  "profile": {
    "rank": 2,
    "percentile": 85,
    "trend": "strengthening",
    "stability": 0.8,
    "summary": "Strong bridge between social groups..."
  }
}
```

---

### GET /api/people/:id/metrics

Get computed NodeMetrics for a specific person. Includes all 10+ discipline computations.

**Response (200 OK):** Full `NodeMetrics` struct with centrality, community, health, capital, attachment, exchange, and trend data.

---

### POST /api/people/:id/classify

Classify a person's social role based on their computed metrics.

**Response (200 OK):**
```json
{
  "person_id": "uuid",
  "role": {
    "primary_role": "bridge",
    "confidence": 0.85,
    "scores": { "bridge": 0.85, "mentor": 0.3, "anchor": 0.6 }
  }
}
```

---

## Analysis Endpoints

### GET /api/analysis/roles

Get role classifications for all people.

**Response (200 OK):**
```json
{
  "person-uuid-1": { "primary_role": "bridge", "confidence": 0.85 },
  "person-uuid-2": { "primary_role": "anchor", "confidence": 0.72 }
}
```

---

### GET /api/analysis/profiles

Get aggregated profiles for all people.

**Response (200 OK):**
```json
{
  "person-uuid-1": {
    "rank": 1,
    "percentile": 90,
    "trend": "strengthening",
    "summary": "..."
  }
}
```

---

### GET /api/analysis/graph/snapshot

Get graph-level aggregate metrics.

**Response (200 OK):**
```json
{
  "total_people": 25,
  "total_relationships": 45,
  "avg_reciprocity": 0.72,
  "bridge_count": 3,
  "strongest_connection": "PersonA ↔ PersonB"
}
```

---

## Relationship Endpoints

### GET /api/relationships/:id/health

Get relationship health data for a specific relationship (matched by source or target person ID).

**Response (200 OK):**
```json
{
  "source": "person-uuid-1",
  "target": "person-uuid-2",
  "quality": "nourishing",
  "strength": 0.85
}
```

---

## WebSocket

### /ws

Real-time bidirectional communication for chat and analysis updates.

**Connection:** Standard WebSocket upgrade at `/ws`

**Incoming messages** (client → server):
```json
{
  "type": "chat",
  "content": "Hello!"
}
```

**Outgoing messages** (server → client):
```json
{
  "id": "uuid",
  "sender": "BubbleTrack",
  "content": "Analysis complete.",
  "is_user": false,
  "created_at": "2025-01-15T10:30:00.000Z"
}
```

---

## Error Handling

All errors follow a consistent JSON format:

```json
{"error": "description of the error"}
```

| Status | When |
|--------|------|
| 400 | Invalid request body, missing required fields |
| 401 | Missing or invalid JWT token |
| 404 | Person or relationship not found |
| 409 | Email already registered (auth) |
| 500 | Internal processing error |
| 501 | Feature not configured (e.g., RAG, states) |

## Middleware Chain

1. `middleware.Recover()` — Panic recovery
2. `middleware.RequestID()` — Request ID generation
3. `middleware.RequestLoggerWithConfig()` — Request logging
4. `JWTMiddleware()` — JWT Bearer token validation (skips `/auth/*`, `/api/health`, `/ws`)
