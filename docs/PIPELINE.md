# Request Pipeline

Complete data flow for every request type in BubbleTrack.

## Authentication Pipeline

All API requests (except public auth endpoints) go through JWT authentication.

### POST /auth/register — Account Creation

```
┌──────────────────────────────────────────────────────────────────────────┐
│  1. HTTP Request                                                        │
│     POST /auth/register {"email": "...", "password": "...", "display_name": "..."}│
└──────────────────────────────┬───────────────────────────────────────────┘
                               │
┌──────────────────────────────▼───────────────────────────────────────────┐
│  2. AuthHandler.Register (auth_handler.go)                               │
│     a. Bind & validate request (email format, password ≥ 8 chars)       │
│     b. Check email uniqueness (UserRepository.GetByEmail)               │
│     c. Hash password with bcrypt                                         │
│     d. Create user record (UserRepository.Create)                       │
│     e. Generate JWT access token (HS256, 15min)                         │
│     f. Generate opaque refresh token (crypto/rand, 7 days)              │
│     g. Store refresh token (RefreshTokenRepository.Store)               │
│     h. Return 201 with tokens + user profile                            │
└──────────────────────────────────────────────────────────────────────────┘
```

### POST /auth/login — Authentication

```
┌──────────────────────────────────────────────────────────────────────────┐
│  1. HTTP Request                                                        │
│     POST /auth/login {"email": "...", "password": "..."}               │
└──────────────────────────────┬───────────────────────────────────────────┘
                               │
┌──────────────────────────────▼───────────────────────────────────────────┐
│  2. AuthHandler.Login (auth_handler.go)                                  │
│     a. Bind & validate request                                          │
│     b. Lookup user by email (UserRepository.GetByEmail)                 │
│     c. Compare password with bcrypt hash                                │
│     d. Generate access token + refresh token                            │
│     e. Store refresh token (RefreshTokenRepository.Store)               │
│     f. Return 200 with tokens + user profile                            │
└──────────────────────────────────────────────────────────────────────────┘
```

### POST /auth/refresh — Token Refresh

```
AuthHandler.Refresh
  └─→ AuthUseCase.RefreshToken(ctx, refreshToken)
        ├─→ Lookup token (RefreshTokenRepository.GetByToken)
        │     Check: not expired, not revoked
        ├─→ Lookup user (UserRepository.GetByID)
        ├─→ Generate new access token
        ├─→ Generate new refresh token
        ├─→ Revoke old token (RefreshTokenRepository.Revoke)
        ├─→ Store new refresh token
        └─→ Return new token pair
```

### POST /auth/logout — Session Termination

```
AuthHandler.Logout
  └─→ AuthUseCase.Logout(ctx, userID)
        └─→ RefreshTokenRepository.RevokeAllForUser(ctx, userID)
              UPDATE refresh_tokens SET revoked_at = NOW()
              WHERE user_id = $1 AND revoked_at IS NULL
```

### JWT Middleware — Request Authentication

```
Every API request passes through JWTMiddleware (except /auth/*, /api/health, /ws):

1. Extract Authorization header
2. Parse "Bearer <token>" format
3. ValidateAccessToken(token)
   ├─→ Parse JWT with HS256 + secret key
   ├─→ Verify not expired
   └─→ Extract sub (user ID) claim
4. Set c.Set("user_id", userID) for downstream handlers
5. If invalid → return 401 Unauthorized
```

---

## POST /api/analyze — Interaction Analysis Pipeline

The core pipeline. Submits free-text interaction descriptions for multi-disciplinary social analysis.

```
┌──────────────────────────────────────────────────────────────────────────┐
│  1. HTTP Request                                                        │
│     POST /api/analyze {"text": "Met PersonA for lunch..."}             │
└──────────────────────────────┬───────────────────────────────────────────┘
                               │
┌──────────────────────────────▼───────────────────────────────────────────┐
│  2. Middleware                                                           │
│     JWTMiddleware → validates Bearer token → sets user_id context       │
└──────────────────────────────┬───────────────────────────────────────────┘
                               │
┌──────────────────────────────▼───────────────────────────────────────────┐
│  3. Handler.Analyze (handler.go)                                        │
│     Validate request → bind JSON → check text is non-empty              │
└──────────────────────────────┬───────────────────────────────────────────┘
                               │
┌──────────────────────────────▼───────────────────────────────────────────┐
│  4. AnalyzeUseCase.Submit (analyze.go)                                  │
│     a. Create Interacao record (status: pending)                        │
│     b. Enqueue Asynq task → Redis                                       │
│     c. Return 202 Accepted with job_id                                  │
└──────────────────────────────┬───────────────────────────────────────────┘
                               │ async (Redis queue)
┌──────────────────────────────▼───────────────────────────────────────────┐
│  5. Worker: AnalyzeUseCase.ProcessJob (analyze.go)                      │
│     a. Embed raw text → vector via Embedder (Gemini/Ollama)             │
│     b. Search Qdrant for related past memories (RAG context)            │
│     c. Build DeepAnalysis prompt with memories + 10-discipline system   │
│     d. Run Agent loop (max 10 iterations with tool calling)              │
└──────────────────────────────┬───────────────────────────────────────────┘
                               │
┌──────────────────────────────▼───────────────────────────────────────────┐
│  6. Agent Loop (adk/agent.go)                                           │
│     Loop up to 10 times:                                                │
│     a. GenerateContent with system prompt + tools                       │
│     b. If response has FunctionCalls → execute tool → feed back         │
│     c. If no FunctionCalls → return final AnalysisResult                │
│                                                                          │
│     Available tools:                                                     │
│     - search_memory: Qdrant vector search                               │
│     - store_memory: Qdrant vector store                                 │
│     - update_graph: PostgreSQL person/relationship upsert               │
│     - classify_role: Social role classification                         │
│     - record_emotional_state: Person mood/energy/valence tracking       │
└──────────────────────────────┬───────────────────────────────────────────┘
                               │
┌──────────────────────────────▼───────────────────────────────────────────┐
│  7. applyResults (analyze.go)                                           │
│     a. Upsert people (PersonRepository)                                 │
│     b. Upsert relationships (RelationshipRepository)                    │
│     c. Store memory in Qdrant (MemoryRepository)                        │
│     d. Create PersonState records (PersonStateRepository)               │
│     e. Create ChatMessage with analysis summary                         │
│     f. Notify via Redis pub/sub → WebSocket/SSE                         │
│     g. Update Interacao status → completed                              │
└──────────────────────────────────────────────────────────────────────────┘
```

## GET /api/graph — Social Graph Retrieval

```
Handler.GetGraph
  └─→ GetGraphUseCase.Execute(ctx, userID)
        └─→ PostgresGraphRepository.GetGraph(ctx, userID)
              ├─→ getNodes(): SELECT from people + person_ownership
              │     WHERE display_name NOT IN ('user','eu','me','you','I','myself')
              │     LIMIT 20
              ├─→ getEdges(): SELECT from relationships
              │     JOIN people for source/target names
              │     LIMIT 100
              └─→ computeStats(): aggregates from nodes + edges
                    TotalPeople, TotalRelationships, AvgReciprocity,
                    BridgeCount, StrongestConnection
```

## GET /api/graph/full — Complete Graph Data

Single-request aggregation of all graph data. Returns nodes, edges, metrics, roles, profiles, person states, and timeline in one response.

```
Handler.GetFullGraph
  ├─→ GetGraphUseCase.Execute(ctx, userID)
  │     └─→ Returns nodes, edges, stats (same as GET /api/graph)
  ├─→ GraphAnalysisEngine.ComputeAllMetrics(ctx, userID)
  │     └─→ Returns all NodeMetrics per person
  ├─→ PersonStateRepository.GetTimeline(ctx, userID, 50)
  │     └─→ Returns emotional state timeline (last 50 entries)
  ├─→ ClassificationEngine.ClassifyRole(metrics) × each person
  │     └─→ Returns role classification with confidence scores
  ├─→ AggregationEngine.AggregateProfile(metrics, allMetrics) × each person
  │     └─→ Returns rank, percentile, trend, stability, summary
  └─→ PersonStateRepository.GetByPerson(ctx, userID, personID, 10) × each node
        └─→ Returns per-person emotional state history (last 10 each)
```

## GET /api/people/:id/metrics — Node Metrics Pipeline

```
Handler.GetPersonMetrics
  └─→ GraphAnalysisEngine.ComputeAllMetrics(ctx, userID)
        ├─→ Load all people and relationships from PostgreSQL
        ├─→ Build adjacency graph in memory
        ├─→ Compute per-node metrics:
        │     ├─ Degree Centrality (mathematics.go)
        │     ├─ Betweenness Centrality (mathematics.go)
        │     ├─ Closeness Centrality (mathematics.go)
        │     ├─ PageRank (mathematics.go)
        │     ├─ Clustering Coefficient (mathematics.go)
        │     ├─ Community Detection — Louvain (mathematics.go)
        │     ├─ Relational Health Score (6 dimensions)
        │     ├─ Social Capital — Bourdieu (sociology.go)
        │     ├─ Attachment Profile — Bowlby (attachment.go)
        │     ├─ Social Exchange — Thibaut & Kelley (social_exchange.go)
        │     ├─ Humanist Profile — Sartre/Aristotle/Gilligan (philosophy.go)
        │     ├─ Social Entropy (physics.go)
        │     ├─ Gravity Model proximity (geography.go)
        │     ├─ Gift Economy balance (anthropology.go)
        │     ├─ Game Theory cooperation (economics.go)
        │     ├─ Neurochemical proxies (neuroscience.go)
        │     ├─ Communication coherence (communication.go)
        │     └─ Path dependency index (history.go)
        └─→ Persist to node_metrics table (GraphEngine.UpsertNodeMetrics)
```

## GET /api/analysis/roles — Role Classification Pipeline

```
Handler.GetAllRoles
  └─→ GraphAnalysisEngine.ComputeAllMetrics(ctx, userID)
        └─→ ClassificationEngine.ClassifyAllRoles(metrics)
              └─→ For each person:
                    ClassificationEngine.ClassifyRole(metrics)
                      ├─→ Compute InterdisciplinaryRoleScore:
                      │     Bridge, Mentor, Anchor, Catalyst, Observer, Drain
                      │     Based on centrality, community position, capital,
                      │     attachment, exchange, energy patterns
                      └─→ Return confidence-weighted role assignment
```

## POST /api/chat — Chat Message Pipeline

```
Handler.SendChatMessage
  └─→ ChatMessageRepository.Create(ctx, msg)
        └─→ INSERT INTO chat_messages (id, user_id, sender, content, is_user, session_id)
              └─→ If WebSocket hub connected:
                    └─→ Hub.BroadcastMessage → all connected clients
              └─→ If no WS:
                    └─→ Poll-based: client polls GET /api/chat
```

## GET /api/states — Emotional Timeline Pipeline

```
Handler.GetEmotionalTimeline
  └─→ PersonStateRepository.GetTimeline(ctx, userID, limit)
        └─→ SELECT from person_states
              JOIN people ON person_id = people.id
              ORDER BY created_at DESC
              LIMIT N
        └─→ Returns: [{person_name, mood, energy, valence, context, trigger, timestamp}]
```

## WebSocket /ws — Real-time Updates

```
Client connects via WebSocket upgrade
  └─→ Hub.Register(client)
        └─→ Client.ReadPump() → reads messages from client
        └─→ Client.WritePump() → sends messages to client
              └─→ On new analysis result:
                    Hub.BroadcastAnalysis(result)
                    → all connected clients receive JSON
              └─→ On new chat message:
                    Hub.NotifyChatMessage(userID, sender, content, isUser)
                    → client receives chat message in real-time
```

## Background Job Processing

```
┌─────────────────┐     ┌──────────┐     ┌──────────────┐
│  HTTP Handler    │────▶│  Redis   │────▶│  Asynq Worker│
│  (Enqueue)       │     │  Queue   │     │  (Process)    │
└─────────────────┘     └──────────┘     └──────┬───────┘
                                                 │
                                                 ▼
                                          ┌──────────────┐
                                          │  Agent Loop   │
                                          │  (Gemini AI)  │
                                          └──────┬───────┘
                                                 │
                                                 ▼
                                          ┌──────────────┐
                                          │  Apply Results│
                                          │  (Persist)    │
                                          └──────┬───────┘
                                                 │
                                    ┌────────────┼────────────┐
                                    ▼            ▼            ▼
                              ┌──────────┐ ┌──────────┐ ┌──────────┐
                              │ PostgreSQL│ │  Qdrant   │ │  Redis   │
                              │ People,   │ │  Memory   │ │  Pub/Sub │
                              │ Rels,     │ │  Vectors  │ │  Notify  │
                              │ States    │ │           │ │          │
                              └──────────┘ └──────────┘ └──────────┘
```
