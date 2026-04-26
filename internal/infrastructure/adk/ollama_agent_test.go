package adk

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestOllamaAgent_ToolCallingScenario(t *testing.T) {
	var toolCallsSeen []string
	var mu sync.Mutex

	updateGraphCalled := false
	classifyRoleCalled := false

	agent := NewOllamaAgent("http://will-be-overridden", "test-model", slog.Default())

	agent.RegisterToolHandler("update_graph", func(args map[string]any) (map[string]any, error) {
		mu.Lock()
		toolCallsSeen = append(toolCallsSeen, "update_graph")
		updateGraphCalled = true
		mu.Unlock()
		return map[string]any{"status": "recorded"}, nil
	})

	agent.RegisterToolHandler("classify_role", func(args map[string]any) (map[string]any, error) {
		mu.Lock()
		toolCallsSeen = append(toolCallsSeen, "classify_role")
		classifyRoleCalled = true
		mu.Unlock()
		return map[string]any{"status": "classified"}, nil
	})

	agent.RegisterToolHandler("store_memory", func(args map[string]any) (map[string]any, error) {
		mu.Lock()
		toolCallsSeen = append(toolCallsSeen, "store_memory")
		mu.Unlock()
		return map[string]any{"status": "stored"}, nil
	})

	iteration := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ollamaChatRequest
		json.NewDecoder(r.Body).Decode(&req)

		iteration++

		if iteration == 1 {
			w.WriteHeader(200)
			json.NewEncoder(w).Encode(ollamaChatResponse{
				Message: ollamaChatMessage{
					Role:    "assistant",
					Content: "",
					ToolCalls: []ollamaToolCall{
						{Function: ollamaToolFunction{
							Name: "update_graph",
							Arguments: map[string]any{
								"source_person": "user",
								"target_person": "PersonA",
								"quality":       "nourishing",
								"strength":      0.7,
							},
						}},
						{Function: ollamaToolFunction{
							Name: "update_graph",
							Arguments: map[string]any{
								"source_person": "user",
								"target_person": "PersonB",
								"quality":       "neutral",
								"strength":      0.4,
							},
						}},
						{Function: ollamaToolFunction{
							Name: "classify_role",
							Arguments: map[string]any{
								"person":   "PersonB",
								"role":     "bridge",
								"evidence": "Unexpectedly joined meeting",
							},
						}},
					},
				},
				Done: true,
			})
			return
		}

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(ollamaChatResponse{
			Message: ollamaChatMessage{
				Role:    "assistant",
				Content: "Analysis: PersonA and user had a planned meeting. PersonB's unexpected arrival created a dynamic shift, acting as a bridge connecting planned interaction to a strained relationship.",
			},
			Done: true,
		})
	}))
	defer server.Close()

	agent.baseURL = server.URL
	agent.httpClient = &http.Client{Timeout: 10 * time.Second}

	result, err := agent.ProcessInteraction(context.Background(), "user-1", "Had meeting with PersonA today. PersonB joined us unexpectedly.")
	if err != nil {
		t.Fatalf("ProcessInteraction failed: %v", err)
	}

	if result.Summary == "" {
		t.Fatal("expected non-empty summary")
	}

	if !updateGraphCalled {
		t.Error("update_graph handler was never called — model did not invoke tools")
	}

	if !classifyRoleCalled {
		t.Error("classify_role handler was never called")
	}

	mu.Lock()
	count := len(toolCallsSeen)
	mu.Unlock()
	if count != 3 {
		t.Errorf("expected 3 tool calls (2 update_graph + 1 classify_role), got %d: %v", count, toolCallsSeen)
	}

	t.Logf("Summary: %s", result.Summary)
	t.Logf("Tool calls: %v", toolCallsSeen)
	t.Logf("Iterations: %d", iteration)
}

func TestOllamaAgent_TextOnlyScenario_NoToolCalls(t *testing.T) {
	graphHandlerCalled := false

	agent := NewOllamaAgent("http://will-be-overridden", "test-model", slog.Default())

	agent.RegisterToolHandler("update_graph", func(args map[string]any) (map[string]any, error) {
		graphHandlerCalled = true
		return map[string]any{"status": "recorded"}, nil
	})

	agent.RegisterToolHandler("classify_role", func(args map[string]any) (map[string]any, error) {
		return map[string]any{"status": "classified"}, nil
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(ollamaChatResponse{
			Message: ollamaChatMessage{
				Role: "assistant",
				Content: `PHASE 1: Observation & Extraction
* Nodes Identified: You, PersonA, PersonB.
* Observed Connections: You ↔ PersonA (coffee), You ↔ PersonB (strained), PersonA ↔ PersonB (co-presence).

PHASE 2: Insight Delivery
The core meeting was set up for a closed system. PersonB's arrival disrupted the anticipated structure.`,
			},
			Done: true,
		})
	}))
	defer server.Close()

	agent.baseURL = server.URL
	agent.httpClient = &http.Client{Timeout: 10 * time.Second}

	result, err := agent.ProcessInteraction(context.Background(), "user-1", "Had meeting with PersonA today. PersonB joined us unexpectedly.")
	if err != nil {
		t.Fatalf("ProcessInteraction failed: %v", err)
	}

	if result.Summary == "" {
		t.Fatal("expected non-empty summary even without tool calls")
	}

	if graphHandlerCalled {
		t.Error("did not expect update_graph to be called in text-only scenario")
	}

	if len(result.PeopleExtracted) != 0 {
		t.Log("PeopleExtracted is empty because model didn't call tools — expected for small models")
	}

	t.Logf("Summary length: %d chars", len(result.Summary))
	t.Logf("Tool calls: 0 (text-only response)")
}

func TestOllamaAgent_MultiIterationScenario(t *testing.T) {
	var toolCallsSeen []string
	var mu sync.Mutex

	agent := NewOllamaAgent("http://will-be-overridden", "test-model", slog.Default())

	agent.RegisterToolHandler("update_graph", func(args map[string]any) (map[string]any, error) {
		mu.Lock()
		toolCallsSeen = append(toolCallsSeen, "update_graph")
		mu.Unlock()
		return map[string]any{"status": "recorded"}, nil
	})

	agent.RegisterToolHandler("classify_role", func(args map[string]any) (map[string]any, error) {
		mu.Lock()
		toolCallsSeen = append(toolCallsSeen, "classify_role")
		mu.Unlock()
		return map[string]any{"status": "classified"}, nil
	})

	agent.RegisterToolHandler("store_memory", func(args map[string]any) (map[string]any, error) {
		mu.Lock()
		toolCallsSeen = append(toolCallsSeen, "store_memory")
		mu.Unlock()
		return map[string]any{"status": "stored"}, nil
	})

	iteration := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		iteration++

		switch iteration {
		case 1:
			json.NewEncoder(w).Encode(ollamaChatResponse{
				Message: ollamaChatMessage{
					ToolCalls: []ollamaToolCall{
						{Function: ollamaToolFunction{Name: "update_graph", Arguments: map[string]any{
							"source_person": "user", "target_person": "PersonC", "quality": "nourishing", "strength": 0.8,
						}}},
						{Function: ollamaToolFunction{Name: "update_graph", Arguments: map[string]any{
							"source_person": "user", "target_person": "PersonD", "quality": "draining", "strength": 0.3,
						}}},
					},
				},
				Done: true,
			})
		case 2:
			json.NewEncoder(w).Encode(ollamaChatResponse{
				Message: ollamaChatMessage{
					ToolCalls: []ollamaToolCall{
						{Function: ollamaToolFunction{Name: "classify_role", Arguments: map[string]any{
							"person": "PersonC", "role": "anchor", "evidence": "Consistent supportive presence",
						}}},
						{Function: ollamaToolFunction{Name: "store_memory", Arguments: map[string]any{
							"content": "Event with PersonC and PersonD",
							"people":  []any{"PersonC", "PersonD"},
						}}},
					},
				},
				Done: true,
			})
		default:
			json.NewEncoder(w).Encode(ollamaChatResponse{
				Message: ollamaChatMessage{
					Role:    "assistant",
					Content: "PersonC acts as your anchor — consistently supportive. PersonD appears draining — interactions with them deplete energy. Consider setting boundaries with PersonD while nurturing connection with PersonC.",
				},
				Done: true,
			})
		}
	}))
	defer server.Close()

	agent.baseURL = server.URL
	agent.httpClient = &http.Client{Timeout: 10 * time.Second}

	result, err := agent.ProcessInteraction(context.Background(), "user-1", "Met PersonC for event. PersonD crashed it and complained the whole time.")
	if err != nil {
		t.Fatalf("ProcessInteraction failed: %v", err)
	}

	if result.Summary == "" {
		t.Fatal("expected non-empty summary")
	}

	mu.Lock()
	count := len(toolCallsSeen)
	mu.Unlock()
	if count != 4 {
		t.Errorf("expected 4 tool calls (2 graph + 1 role + 1 memory), got %d: %v", count, toolCallsSeen)
	}

	if iteration != 3 {
		t.Errorf("expected 3 iterations (tools → tools → text), got %d", iteration)
	}

	t.Logf("Iterations: %d, Tool calls: %d", iteration, count)
	t.Logf("Summary: %s", result.Summary)
}

func TestOllamaAgent_ErrorFromOllama(t *testing.T) {
	agent := NewOllamaAgent("http://will-be-overridden", "test-model", slog.Default())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("model not found"))
	}))
	defer server.Close()

	agent.baseURL = server.URL
	agent.httpClient = &http.Client{Timeout: 5 * time.Second}

	_, err := agent.ProcessInteraction(context.Background(), "user-1", "test")
	if err == nil {
		t.Fatal("expected error from ollama 500")
	}
}

func TestOllamaAgent_UnknownToolIgnored(t *testing.T) {
	agent := NewOllamaAgent("http://will-be-overridden", "test-model", slog.Default())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ollamaChatResponse{
			Message: ollamaChatMessage{
				ToolCalls: []ollamaToolCall{
					{Function: ollamaToolFunction{Name: "nonexistent_tool", Arguments: map[string]any{}}},
				},
			},
			Done: true,
		})
	}))
	defer server.Close()

	agent.baseURL = server.URL
	agent.httpClient = &http.Client{Timeout: 5 * time.Second}

	result, err := agent.ProcessInteraction(context.Background(), "user-1", "test")
	if err != nil {
		t.Fatalf("should not error on unknown tool: %v", err)
	}

	if result.Summary != "Analysis complete." {
		t.Errorf("expected fallback summary, got: %s", result.Summary)
	}
}

func TestOllamaAgent_10IterationMax(t *testing.T) {
	agent := NewOllamaAgent("http://will-be-overridden", "test-model", slog.Default())

	agent.RegisterToolHandler("update_graph", func(args map[string]any) (map[string]any, error) {
		return map[string]any{"status": "recorded"}, nil
	})

	iteration := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		iteration++
		if iteration <= 15 {
			json.NewEncoder(w).Encode(ollamaChatResponse{
				Message: ollamaChatMessage{
					ToolCalls: []ollamaToolCall{
						{Function: ollamaToolFunction{Name: "update_graph", Arguments: map[string]any{
							"source_person": "a", "target_person": "b", "quality": "neutral", "strength": 0.5,
						}}},
					},
				},
				Done: true,
			})
		} else {
			json.NewEncoder(w).Encode(ollamaChatResponse{
				Message: ollamaChatMessage{Role: "assistant", Content: "done"},
				Done:    true,
			})
		}
	}))
	defer server.Close()

	agent.baseURL = server.URL
	agent.httpClient = &http.Client{Timeout: 5 * time.Second}

	result, err := agent.ProcessInteraction(context.Background(), "user-1", "test")
	if err != nil {
		t.Fatalf("ProcessInteraction failed: %v", err)
	}

	if iteration > 10 {
		t.Errorf("agent should stop at 10 iterations, got %d", iteration)
	}

	if result.Summary == "" {
		t.Error("expected fallback summary when max iterations reached")
	}
}
