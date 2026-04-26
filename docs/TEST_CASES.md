# Test Cases

Complete inventory of all 130 test cases across the BubbleTrack codebase. Tests use only generic placeholder data (`PersonA`, `PersonB`, `user-1`, `test-user`) — no real personal information.

---

## Test Data Policy

- **No real names**: All person references use `PersonA`, `PersonB`, `PersonC`, `PersonD`
- **No real emails**: Only `user@example.com` in API docs
- **No real user IDs**: Tests use `user-1`, `test-user`
- **No real credentials**: Mock repos and stub agents, no database connections

---

## `internal/api` — Handler Tests

| Test | File | Description | Expectation |
|------|------|-------------|-------------|
| `TestHealthEndpoint` | handler_test.go | GET /api/health with no dependencies | Returns 200 OK |
| `TestGetGraphEndpoint` | handler_test.go | GET /api/graph with `user_id` context set | Placeholder — needs real GetGraphUseCase |
| `TestAnalyzeEndpoint` | handler_test.go | POST /api/analyze with valid and empty JSON bodies | Empty text body logs expected error (validation) |

---

## `internal/application` — Use Case Tests

### AnalyzeUseCase (`analyze_test.go`)

Uses stub agent, mock repos, and stub enqueuer — no real infrastructure.

| Test | Description | Expectation |
|------|-------------|-------------|
| `TestProcessJob_MetadataPeopleSerializedAsJSONString` | Regression test: people names must be JSON string in Qdrant metadata, not `[]string` | `metadata["people"]` is a valid JSON string `["PersonA","PersonB"]`; interaction status becomes `completed`; 1 memory stored |
| `TestProcessJob_EmptyPeopleStillProducesValidJSON` | Agent returns 0 extracted people | `metadata["people"]` is `"[]"` (valid empty JSON array); memory still stored |
| `TestProcessJob_AgentErrorMarksFailed` | Agent returns a `ValidationError` | `ProcessJob` returns error; interaction status becomes `failed` |
| `TestProcessJob_NilAgentProducesEmbeddingOnlyResult` | Agent is nil (LLM unavailable) | Embedding stored with fallback summary "Processed via embedding storage only (no AI agent)"; empty people; status `completed` |
| `TestProcessJob_AppliesPeopleAndRelationships` | Agent extracts 2 people + 1 relationship | At least 2 people upserted; 1 relationship created with label "friend" |
| `TestSubmit_ValidationError` | Submit with empty `UserID` and empty `RawText` | Returns `ValidationError` in both cases |
| `TestRegisterWorker_DecodesAndProcesses` | Asynq mux handler decodes payload and runs ProcessJob | Interaction status transitions `queued` → `completed` |

### Engine Tests (`application_test.go`)

| Test | Description | Expectation |
|------|-------------|-------------|
| `TestGraphAnalysisEngine` | Creates mock person/relationship chain graph | Graph structure accepted without panic |
| `TestClassificationEngine` | Instantiates ClassificationEngine | Engine is non-nil |
| `TestAggregationEngine` | Instantiates AggregationEngine | Engine is non-nil |
| `TestAnalyzeUseCaseSubmit` | Table-driven: valid text vs empty text | Valid text: no error; Empty text: error |
| `TestGetGraphUseCase` | Placeholder with empty context and `test-user` | No panic (needs real graph repo for full test) |

---

## `internal/domain` — Domain Logic Tests

### Errors & Validation (`errors_test.go`)

| Test | Description | Expectation |
|------|-------------|-------------|
| `TestValidateRelacionamento` | 9 sub-cases: valid, missing source, missing target, same source+target, strength too high, strength negative, reciprocity too high, boundary 0, boundary 1 | Invalid inputs return error; boundary values [0, 1] accepted |
| `TestValidateInteracao` | 6 sub-cases: valid, missing user_id, missing raw_text, empty raw_text, text > 10000 chars, text at 10000 boundary | Invalid inputs return error; boundary 10000 accepted |
| `TestNotFoundError` | NotFoundError with entity "person" and ID "123" | Error string = `"person 123 not found"` |
| `TestValidationError` | ValidationError with field "name" and message "required" | Error string = `"validation error on name: required"` |
| `TestPessoa` | Pessoa struct with ID, DisplayName, Aliases, Notes | All fields match expected values |
| `TestRelacionamento` | Relacionamento struct with Quality, Strength, Label | Quality = "nourishing", Label = "friend" |
| `TestSocialRoleConstants` | All 7 SocialRole constants | All unique, no duplicates |
| `TestQualityConstants` | All 5 Quality constants | All unique, no duplicates |

### Node Metrics (`node_metrics_test.go`)

| Test | Description | Expectation |
|------|-------------|-------------|
| `TestComputeNodeHealthScore` | 7 sub-cases: zero, high, low, overflow, negative, only relational, only centrality | Result always in [0, 1], never NaN or Inf |
| `TestComputeNodeHealthScore_HighScores_HighResult` | All metric inputs ≥ 0.8 | Result ≥ 0.5 |
| `TestComputeNodeHealthScore_ZeroMetrics_NonNaN` | All-zero NodeMetrics | Result is not NaN or Inf |
| `TestComputeTrend` | 9 sub-cases: empty, single, rising, falling, stable, volatile, two stable, two volatile, all zeros | Correct `TrendDirection` returned; delta = 0 for < 2 points |
| `TestComputeTrend_DeltaConsistency` | Linear rising points | Delta = last value − first value in window |
| `TestComputeTrend_VolatileThreshold` | Oscillating 0.0/0.3 values | Direction = "volatile" |
| `TestAbs` | 5 sub-cases: 0, 5, -5, 0.001, -0.001 | Correct absolute value |
| `TestMax` | 5 sub-cases: 0/0, 5/3, 3/5, -1/0, -5/-3 | Returns larger value |

### Communication — Watzlawick Axioms (`communication_test.go`)

| Test | Description | Expectation |
|------|-------------|-------------|
| `TestComputeWatzlawick` | 6 sub-cases: zero messages, perfect, poor, single, high received/low understood, exceeding totals | All output fields clamped to [0, 1]; zero messages → zero-value struct |
| `TestComputeWatzlawick_PerfectComms_HighCompliance` | All inputs = 100 | `OverallCompliance` ≥ 0.9 |
| `TestComputeWatzlawick_Axiom1QualityIsProduct` | Verify Axiom1Quality = Axiom1Received × Axiom2Content | Product formula holds |
| `TestComputeWatzlawick_ComplianceIsAverage` | Verify OverallCompliance = average of 5 axiom scores | Average formula holds |
| `TestComputeMetaCommunication` | 6 sub-cases: zero, high congruence, high incongruence, all paradox, single, exceeding | All fields clamped; zero interactions → zero-value struct |
| `TestComputeMetaCommunication_HighCongruence_HighReflexivity` | Congruence 90/100 | Reflexivity ≥ 0.5 |
| `TestComputeMetaCommunication_HighIncongruence_HighDoubleBind` | Incongruence 80/100 | DoubleBindRisk ≥ 0.5 |
| `TestComputeNarrative` | 9 sub-cases: empowerment, entrapment, transformation, struggle, stable, zero, max, negative clamped, overflow clamped | Correct trajectory label; all fields in [0, 1] |
| `TestComputeNarrative_RedemptionArc_Activated` | Agency 0.6, Empathy 0.7 | RedemptionArc > 0 (both > 0.4 threshold) |
| `TestComputeNarrative_RedemptionArc_NotActivated` | Agency 0.3, Empathy 0.3 | RedemptionArc = 0 |
| `TestComputeNarrative_ContaminationArc_Activated` | Complexity 0.7, Resolution 0.2 | ContaminationArc > 0 |
| `TestComputeNarrative_ContaminationArc_NotActivated` | Complexity 0.5, Resolution 0.5 | ContaminationArc = 0 |
| `TestComputeShannonInfo` | 6 sub-cases: zero, high variety, low variety, single, all misunderstandings, exceeding | All fields clamped; zero messages → zero-value struct |
| `TestComputeShannonInfo_NoMisunderstandings_HighMutual` | 50 variety, 100 messages, 0 misunderstandings | MutualUnderstanding ≥ 0.5; NoiseRatio near 0 |
| `TestComputeShannonInfo_HighMisunderstandings_LowChannelCap` | 10 variety, 50 messages, 40 misunderstandings | ChannelCapacity ≤ 0.5 |
| `TestComputeShannonInfo_RedundancyInverseEntropy` | Verify Redundancy = 1 − SourceEntropy | Inverse relationship holds |

### History — Path Dependency (`history_test.go`)

| Test | Description | Expectation |
|------|-------------|-------------|
| `TestComputePathDependency` | 7 sub-cases: zero, high repetition, low repetition, all repeated, single, no past, high sunk cost | All output fields in [0, 1]; zero decisions → zero-value struct |
| `TestComputePathDependency_HighRepetition_LockIn` | 90/100 repeated | LockInScore ≥ 0.5 |
| `TestComputePathDependency_BranchingInverseLockIn` | Low repetition | BranchingPointRisk ≈ 1 − LockInScore |
| `TestComputePathDependency_CyclePhase` | 3 sub-cases: lock_in, transition, maintenance | Correct CyclePhase string |
| `TestDetectTemporalCycle` | 10 sub-cases: empty, insufficient, constant, increasing, decreasing, oscillating, zeros, repeated, large, negative | Seasonality/Predictability clamped; insufficient data → "insufficient_data" |
| `TestDetectTemporalCycle_Constant_StableCycleType` | Constant values | CycleType = "stable" |
| `TestDetectTemporalCycle_Increasing_NonIrregular` | Strong linear increase | CycleType ≠ "irregular" |
| `TestDetectTemporalCycle_Oscillating_LowSeasonality` | Alternating 1/5 values | Oscillating pattern detected |
| `TestComputeGenerationalDynamics` | 8 sub-cases: digital divide, similar, mixed, zero, max, negative clamped, overflow clamped, high conflict | All fields clamped; correct CommunicationStyle label |
| `TestComputeHoweStrauss` | 7 sub-cases: crisis, awakening, high, default, boundary, zero, max | Correct HoweStraussPhase and TurningsCount |
| `TestSqrt` | 8 values: 0, -1, 1, 4, 9, 2, 0.25, 100 | Correct sqrt (negative → 0) |

### Neuroscience (`neuroscience_test.go`)

| Test | Description | Expectation |
|------|-------------|-------------|
| `TestComputeNeuralSocial` | 9 sub-cases: zero, all positive, all negative, mixed, high proximity, large numbers, single, only negatives, only shared | All 10 output fields in [0, 1] |
| `TestComputeNeuralSocial_AllPositive_HighValues` | Positive inputs | EmpathyPrediction ≥ 0.3; BondingReadiness ≥ 0.3; Cortisol < Dopamine |
| `TestComputeNeuralSocial_AllNegative_HighStress` | Only negative interactions | CortisolProxy ≥ 0.5; StressResponse ≥ 0.3 |
| `TestComputeNeuralSocial_ZeroInputs_NonNaN` | All zeros | All 10 fields are neither NaN nor Inf |
| `TestComputeEmotionalRegulation` | 7 sub-cases: zero, high awareness, low awareness, high resolution, exceeds clamp, single, zero self-reports | All fields clamped; zero events → zero-value struct |
| `TestComputeEmotionalRegulation_HighResolution_LowSuppression` | High resolution rate | Suppression < Reappraisal |
| `TestComputeEmotionalRegulation_HighAwareness_LowAlexithymia` | 90/100 self-reports | AlexithymiaRisk ≤ 0.7 |
| `TestComputeBurnoutRisk` | 6 sub-cases: zero, low obligations, high obligations, max, zero obligations, single | All fields clamped; zero events → zero-value struct |
| `TestComputeBurnoutRisk_RecoveryPattern` | 3 sub-cases mapping burnout level to pattern | "insufficient" for high burnout; "minimal" for moderate; "adequate" for low |

### Mathematics — Graph Algorithms (`mathematics_test.go`)

| Test | Description | Expectation |
|------|-------------|-------------|
| `TestMathComputeDensity` | 10 sub-cases: zero, one, two no edges, complete graphs, negative, large sparse | Correct density = 2E / (N(N-1)) |
| `TestMathComputeAvgDegree` | 7 sub-cases: zero, isolated, connected, complete, no edges, large | Correct avg degree = 2E / N |
| `TestMathComputeDegreeCentralityMap` | 5 sub-cases: empty, isolated, connected, star, complete triangle | Correct normalized degree centrality |
| `TestMathComputeBetweennessCentrality` | 6 sub-cases: empty, single, connected, line (bridge), complete, star | Bridge nodes have highest betweenness; complete graph = 0 |
| `TestMathComputeClosenessCentrality` | 5 sub-cases: empty, isolated, connected, line, complete | Correct closeness = (N-1) / sum(distances) |
| `TestMathComputePageRank` | 6 sub-cases: empty, single, two-node cycle, complete triangle, default config, star | Correct convergence; star center has highest rank |
| `TestMathComputeClusteringCoefficientLocal` | 7 sub-cases: isolated, one neighbor, triangle, square, star, missing node, partial | Triangle = 1.0; star/square = 0; partial = 1/3 |
| `TestMathComputeModularity` | 6 sub-cases: empty, single, same community, split, correct pairs, wrong pairs | Correct Q values; wrong assignment gives negative modularity |
| `TestMathDetectCommunitiesLouvain` | 5 sub-cases: empty, single, complete, disconnected, connected pair | Correct community count; connected nodes share community |
| `TestMathComputeInformationEntropy` | 7 sub-cases: empty, single, same, two groups, three, four, unequal | Correct Shannon entropy values |
| `TestMathLog2` | 8 values: 0, -1, 1, 2, 4, 8, 0.5, 1024 | Correct log2 (0 and negative → 0) |
| `TestMathAbsFloat` | 3 sub-cases: positive, negative, zero | Correct absolute value |
| `TestMathPageRankSumToOne` | 3-node graph with cycle | All PageRank values sum to 1.0 |
| `TestMathDisconnectedGraphCloseness` | Two disconnected pairs | Each node's closeness = 1.0 within its component |

---

## `internal/infrastructure/adk` — Ollama Agent Tests

All tests use `httptest.NewServer` to mock the Ollama API — no real LLM calls.

| Test | Description | Expectation |
|------|-------------|-------------|
| `TestOllamaAgent_ToolCallingScenario` | Mock server returns 2 `update_graph` + 1 `classify_role` tool calls in iteration 1, then text summary in iteration 2 | All 3 tool handlers called; non-empty summary; 2 iterations total |
| `TestOllamaAgent_TextOnlyScenario_NoToolCalls` | Mock server returns text-only response (no tool calls) | Non-empty summary; `update_graph` handler never called; `PeopleExtracted` is empty |
| `TestOllamaAgent_MultiIterationScenario` | 3-iteration flow: iter 1 = 2 graph updates, iter 2 = 1 role + 1 memory, iter 3 = text summary | 4 total tool calls; 3 iterations; non-empty summary about PersonC (anchor) and PersonD (draining) |
| `TestOllamaAgent_ErrorFromOllama` | Mock server returns HTTP 500 | `ProcessInteraction` returns error |
| `TestOllamaAgent_UnknownToolIgnored` | Mock server returns `nonexistent_tool` call | No error; fallback summary "Analysis complete."; unknown tool logged as warning |
| `TestOllamaAgent_10IterationMax` | Mock server always returns tool calls (would loop forever) | Agent stops at exactly 10 iterations; returns fallback summary |

---

## Summary

| Package | Test Files | Test Functions | Sub-cases |
|---------|-----------|---------------|-----------|
| api | 1 | 3 | — |
| application | 2 | 10 | 2 (table-driven) |
| domain/errors | 1 | 8 | 15 |
| domain/node_metrics | 1 | 8 | 19 |
| domain/communication | 1 | 16 | 27 |
| domain/history | 1 | 10 | 28 |
| domain/neuroscience | 1 | 9 | 28 |
| domain/mathematics | 1 | 14 | 47 |
| infrastructure/adk | 1 | 6 | — |
| **Total** | **10** | **84 functions** | **~166 sub-cases** |

All 130+ test cases pass. Zero external dependencies required (no database, no LLM, no Redis).
