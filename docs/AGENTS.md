# Agent System

Bubble Track uses a custom AI agent loop (not the Python Google ADK) that orchestrates Gemini's function-calling capabilities to analyze social interactions.

## Agent Loop Flow

```
┌──────────────────────────────────────────────────────────────┐
│                   INPUT: Interaction Text                      │
│  "Met PersonA for lunch today. PersonB joined unexpectedly..."  │
└──────────────────────────┬──────────────────────────────────────┘
                           │
        ┌──────────────────▼──────────────────┐
        │  GenerateContent (System + Tools)    │
        │  System: 10-discipline prompt        │
        │  Tools: search, store, update, classify│
        └──────────────┬───────────────────────┘
                     │
        ┌─────────────▼─────────────┐
        │  Response from Gemini    │
        │  May contain FunctionCalls│
        └──────────────┬───────────┘
                     │
          ┌─────────┴─────────┐
          │               │
      [No FC]         [Has FC]
          │               │
          ▼               ▼
    ┌─────────┐    ┌──────▼──────────┐
    │ Return │    │ Execute Tool   │
    │Result │    │ (executor.go) │
    └───────┘    └──────┬─────────┘
                        │
              ┌────────┴────────┐
              │ Send FunctionResponse │
              │ back to Gemini     │
              └─────────┬─────────┘
                        │
              ┌────────▼────────┐
              │ Next Iteration   │
              │ (loop up to 10) │
              └────────────────┘
```

## Configuration

| Parameter | Value |
|-----------|-------|
| Max Iterations | 10 |
| Tool Timeout | 30s per tool |
| Model | Gemini (via google.golang.org/genai) |

## Tools Provided

### 1. search_memory
Searches past interaction memories in Qdrant vector store.

```json
{
  "query": "interactions with PersonA",
  "limit": 5
}
```

### 2. store_memory
Stores new interaction memory in Qdrant.

```json
{
  "content": "Met PersonA and PersonB for a gathering today...",
  "people": ["PersonA", "PersonB"]
}
```

### 3. update_graph
Updates PostgreSQL graph with relationship changes.

```json
{
  "source_person": "user",
  "target_person": "PersonA",
  "quality": "positive",
  "strength": 0.8,
  "label": "friend",
  "reciprocity_delta": 0.1
}
```

### 4. classify_role
Classifies a person's social role.

```json
{
  "person": "PersonB",
  "role": "bridge",
  "evidence": "connects different social groups"
}
```

## System Prompts

### BubbleTrackSystemPrompt
Multidisciplinary analysis framework covering 10 lenses:
1. **Psychological**: Bowlby attachment, Social exchange
2. **Sociological**: Bourdieu capital, Granovetter ties, Dunbar
3. **Philosophical**: Sartre, Aristotle, Gilligan
4. **Mathematical**: Graph centrality, PageRank
5. **Physical**: Thermodynamics, Complex systems
6. **Anthropological**: Mauss gift, Fictive kin
7. **Economic**: Game theory, Nash equilibrium
8. **Neuroscientific**: Mirror neurons, Hormones
9. **Communication**: Watzlawick axioms
10. **Historical**: Path dependency

### EnhancedToolDefinitions
Four function declarations for the model.

### DeepAnalysisSystemPrompt
System prompt + user ID context for specific analysis.

## Tool Executor

The `ToolExecutor` interface defines 4 methods:
- `ExecuteSearchMemory(ctx, query, limit) -> string`
- `ExecuteStoreMemory(ctx, content, people) -> string`
- `ExecuteUpdateGraph(ctx, source, target, quality, strength, label, reciprocity) -> string`
- `ExecuteClassifyRole(ctx, person, role, evidence) -> string`

Implementation in `executor.go` bridges to:
- **Qdrant**: Store and search memories
- **PostgreSQL**: Graph persistence
- **Repositories**: Person and relationship CRUD

## Error Handling

- Tool execution errors → Return JSON error message in FunctionResponse
- Timeout → 30s context timeout
- Unknown tool → Error message, continue loop
- Model without function calls → Return final response

## Output

Returns `AnalysisResult` containing:
- `PeopleExtracted`: []ExtractedPerson
- `Relationships`: []RelationshipUpdate
- `SocialRoles`: map[PersonID]SocialRole
- `Summary`: string (final analysis)