# Database & Data Storage

BubbleTrack uses PostgreSQL for relational data, Qdrant for vector search, and Redis for queuing and pub/sub.

## PostgreSQL — Tables & Schema

All tables are auto-created via migrations in `RunMigrations()` (postgres_graph.go) and `AddNodeMetricsMigrations()` (graph_engine.go).

### people

Stores every person extracted from interactions.

| Column | Type | Constraints | Description |
|--------|------|------------|-------------|
| id | UUID | PK, auto-generated | Unique person ID |
| display_name | TEXT | NOT NULL | Person's display name |
| aliases | TEXT[] | DEFAULT '{}' | Alternative names / nicknames |
| notes | TEXT | DEFAULT '' | Free-text notes |
| social_role | TEXT | NOT NULL DEFAULT 'unknown' | Bridge/Mentor/Anchor/Catalyst/Observer/Drain |
| current_mood | TEXT | NOT NULL DEFAULT 'neutral' | Happy/Anxious/Tired/Energized/Sad/Neutral/Angry/Hopeful/Lonely/Grateful |
| current_energy | FLOAT | NOT NULL DEFAULT 0.5 | Energy level [0, 1] |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | Creation timestamp |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | Last update timestamp |

**Self-reference filter:** Queries exclude display names `'user', 'eu', 'me', 'you', 'I', 'myself'` — these represent the user themselves, not tracked people.

---

### relationships

Stores edges between people in the social graph.

| Column | Type | Constraints | Description |
|--------|------|------------|-------------|
| id | UUID | PK, auto-generated | Unique relationship ID |
| source_person_id | UUID | FK → people(id) ON DELETE CASCADE | Source person |
| target_person_id | UUID | FK → people(id) ON DELETE CASCADE | Target person |
| quality | TEXT | NOT NULL DEFAULT 'unknown' | Nourishing/Neutral/Draining/Conflicted/Unknown |
| strength | FLOAT | NOT NULL DEFAULT 0.5, CHECK [0,1] | Relationship strength |
| label | TEXT | DEFAULT '' | Descriptive label (e.g., "friend", "colleague") |
| reciprocity_index | FLOAT | NOT NULL DEFAULT 0.5, CHECK [0,1] | Bidirectional balance |
| source_weight | FLOAT | NOT NULL DEFAULT 0.5 | Source's investment weight |
| target_weight | FLOAT | NOT NULL DEFAULT 0.5 | Target's investment weight |
| protocol | TEXT | NOT NULL DEFAULT '' | Deep/Casual/Professional/Digital/Mixed |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | Creation timestamp |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | Last update timestamp |

**Unique index:** `(source_person_id, target_person_id)` — one relationship per pair.

---

### interactions

Stores raw interaction text submitted for analysis.

| Column | Type | Constraints | Description |
|--------|------|------------|-------------|
| id | UUID | PK, auto-generated | Interaction ID |
| user_id | TEXT | NOT NULL | Owner's user ID |
| raw_text | TEXT | NOT NULL | Original free-text description |
| summary | TEXT | DEFAULT '' | AI-generated summary |
| people | UUID[] | DEFAULT '{}' | Referenced person IDs |
| job_id | TEXT | DEFAULT '' | Asynq background job ID |
| status | TEXT | NOT NULL DEFAULT 'queued' | queued/processing/completed/failed |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | Submission timestamp |

**Indexes:** `user_id`, `status`

---

### person_ownership

Maps which user "owns" which person (multi-tenancy).

| Column | Type | Constraints | Description |
|--------|------|------------|-------------|
| user_id | TEXT | PK | User identifier |
| person_id | UUID | PK, FK → people(id) ON DELETE CASCADE | Person identifier |

**Index:** `user_id`

---

### chat_messages

Stores chat conversation history between user and BubbleTrack.

| Column | Type | Constraints | Description |
|--------|------|------------|-------------|
| id | UUID | PK, auto-generated | Message ID |
| user_id | TEXT | NOT NULL | Owner's user ID |
| sender | TEXT | NOT NULL | "You" or "BubbleTrack" |
| content | TEXT | NOT NULL | Message text |
| is_user | BOOLEAN | NOT NULL DEFAULT FALSE | True = user sent, False = system response |
| session_id | TEXT | DEFAULT '' | Optional session grouping |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | Timestamp |

**Indexes:** `(user_id, created_at DESC)`, `(session_id, created_at DESC)`

---

### person_states

Emotional state snapshots per person, recorded after each analysis.

| Column | Type | Constraints | Description |
|--------|------|------------|-------------|
| id | UUID | PK, auto-generated | State ID |
| user_id | TEXT | NOT NULL | Owner's user ID |
| person_id | UUID | NULLABLE | Person ID (NULL = self-reflection) |
| mood | TEXT | NOT NULL DEFAULT 'neutral' | Happy/Anxious/Tired/Energized/Sad/Neutral/Angry/Hopeful/Lonely/Grateful |
| energy | FLOAT | NOT NULL DEFAULT 0.5, CHECK [0,1] | Energy level |
| valence | FLOAT | NOT NULL DEFAULT 0, CHECK [-1,1] | Emotional valence (negative to positive) |
| context | TEXT | DEFAULT '' | Setting (gathering, workplace, home, event, online) |
| trigger | TEXT | DEFAULT '' | What caused this state |
| interaction_id | UUID | NULLABLE | Link to originating interaction |
| notes | TEXT | DEFAULT '' | Additional notes |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | Timestamp |

**Indexes:** `(user_id, created_at DESC)`, `(user_id, person_id, created_at DESC)`

---

### node_metrics

Computed dynamic metrics per person node from 10+ academic disciplines.

| Column | Type | Description |
|--------|------|-------------|
| person_id | TEXT | Person identifier |
| user_id | TEXT | Owner's user ID |
| time_window | TEXT | "all", "30d", etc. |
| computed_at | TIMESTAMPTZ | When metrics were computed |
| degree | INT | Direct connections |
| interaction_frequency | FLOAT | How often this person appears |
| emotional_valence | FLOAT | [-1, 1] average valence |
| trend_direction | TEXT | improving/stable/declining |
| trend_strength | FLOAT | Magnitude of trend |
| centrality_degree | FLOAT | Degree centrality [0, 1] |
| centrality_betweenness | FLOAT | Betweenness centrality [0, 1] |
| centrality_closeness | FLOAT | Closeness centrality [0, 1] |
| centrality_eigenvector | FLOAT | Eigenvector centrality [0, 1] |
| centrality_pagerank | FLOAT | PageRank score |
| centrality_clustering | FLOAT | Local clustering coefficient |
| community_id | TEXT | Louvain community assignment |
| community_role | TEXT | Role within community |
| community_embeddedness | FLOAT | Internal edges ratio |
| community_bridge | FLOAT | External edges ratio |
| relational_health_overall | FLOAT | Composite health score |
| relational_health_reciprocity | FLOAT | Reciprocity component |
| social_capital_total | FLOAT | Bourdieu capital index |
| humanist_agency | FLOAT | Sartre agency score |
| humanist_empathic | FLOAT | Gilligan care score |
| exchange_satisfaction | FLOAT | Thibaut & Kelley satisfaction |
| attachment_style | TEXT | Bowlby classification |

**PK:** `(person_id, user_id, time_window)`
**Indexes:** `(user_id, computed_at DESC)`, `(person_id, computed_at DESC)`

---

### users

Stores registered user accounts for authentication.

| Column | Type | Constraints | Description |
|--------|------|------------|-------------|
| id | UUID | PK, auto-generated | Unique user ID |
| email | TEXT | NOT NULL, UNIQUE | Login email |
| password_hash | TEXT | NOT NULL | bcrypt hashed password |
| display_name | TEXT | NOT NULL | User's display name |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | Account creation timestamp |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | Last update timestamp |

**Index:** `UNIQUE(email)`

---

### refresh_tokens

Stores JWT refresh tokens for session management.

| Column | Type | Constraints | Description |
|--------|------|------------|-------------|
| id | UUID | PK, auto-generated | Token ID |
| user_id | UUID | FK → users(id) ON DELETE CASCADE | Owning user |
| token | TEXT | NOT NULL, UNIQUE | Opaque refresh token string |
| expires_at | TIMESTAMPTZ | NOT NULL | Token expiration time |
| revoked_at | TIMESTAMPTZ | NULLABLE | When token was revoked (NULL = active) |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | Token creation time |

**Indexes:** `UNIQUE(token)`, `(user_id)`, `(expires_at)`

---

## Qdrant — Vector Store

Used for semantic memory search (RAG) on past interactions.

**Collection:** `bubble_memories` (configurable via `QDRANT_COLLECTION`)
**Vector size:** 768 (Gemini embedding dimensions)
**Distance metric:** Cosine similarity

### Point Structure

| Field | Type | Description |
|-------|------|-------------|
| id | string | Interaction ID |
| vector | float32[768] | Gemini embedding of interaction text |
| payload.user_id | string | Owner's user ID |
| payload.interaction_id | string | Source interaction reference |
| payload.people | string[] | People mentioned |
| payload.content | string | Original text content |
| payload.timestamp | string | ISO 8601 timestamp |

### Operations

| Operation | Method | Description |
|-----------|--------|-------------|
| Store | `Store(ctx, userID, interactionID, embedding, metadata)` | Upsert vector point |
| Search | `Search(ctx, userID, queryEmbedding, limit)` | Cosine similarity search |
| Filtered Search | `SearchFiltered(ctx, userID, queryEmbedding, limit, filter)` | Search with person/session/segment filters |

---

## Redis — Queue & Pub/Sub

### Queue (Asynq)

Background job processing for analysis pipeline.

**Task Type:** `analyze_interaction`

**Payload:**
```json
{
  "interaction_id": "uuid",
  "user_id": "test-user",
  "raw_text": "Met PersonA for lunch..."
}
```

### Pub/Sub

Real-time notifications when analysis completes.

| Channel | Event | Payload |
|---------|-------|---------|
| `analysis:{userID}` | Analysis complete | `AnalysisResult` JSON |

---

## Data Flow Summary

```
User submits text → interactions table (status: queued)
                     ↓
              Redis Queue (Asynq)
                     ↓
              Worker picks up job
                     ↓
         ┌──────────────────────────┐
         │     AI Agent Loop        │
         │  (Gemini + Tool Calls)   │
         └──────────┬───────────────┘
                    ↓
    ┌───────────────┼───────────────┐
    ↓               ↓               ↓
people table   relationships    person_states
(upsert)       (upsert)         (create)
    ↓
Qdrant vectors (store embedding)
    ↓
Redis pub/sub (notify complete)
    ↓
interactions table (status: completed)
```
