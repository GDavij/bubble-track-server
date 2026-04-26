package adk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/bubbletrack/server/internal/domain"
)

type ollamaChatMessage struct {
	Role      string         `json:"role"`
	Content   string         `json:"content"`
	ToolCalls []ollamaToolCall `json:"tool_calls,omitempty"`
}

type ollamaToolCall struct {
	Function ollamaToolFunction `json:"function"`
}

type ollamaToolFunction struct {
	Name      string `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type ollamaTool struct {
	Type     string             `json:"type"`
	Function ollamaToolFunctionDef `json:"function"`
}

type ollamaToolFunctionDef struct {
	Name        string                      `json:"name"`
	Description string                      `json:"description,omitempty"`
	Parameters  ollamaToolFunctionParameters `json:"parameters"`
}

type ollamaToolFunctionParameters struct {
	Type       string                       `json:"type"`
	Required   []string                     `json:"required,omitempty"`
	Properties map[string]ollamaToolProperty `json:"properties"`
}

type ollamaToolProperty struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

type ollamaChatRequest struct {
	Model    string              `json:"model"`
	Messages []ollamaChatMessage `json:"messages"`
	Tools    []ollamaTool        `json:"tools,omitempty"`
	Stream   bool                `json:"stream"`
}

type ollamaChatResponse struct {
	Message ollamaChatMessage `json:"message"`
	Done    bool              `json:"done"`
}

type OllamaAgent struct {
	baseURL    string
	model      string
	tools      []ollamaTool
	handler    map[string]func(map[string]any) (map[string]any, error)
	logger     *slog.Logger
	httpClient *http.Client
}

func NewOllamaAgent(baseURL, model string, logger *slog.Logger) *OllamaAgent {
	agent := &OllamaAgent{
		baseURL: baseURL,
		model:   model,
		tools:   ollamaToolDefinitions(),
		handler: make(map[string]func(map[string]any) (map[string]any, error)),
		logger:  logger,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
	return agent
}

func (a *OllamaAgent) RegisterToolHandler(name string, handler func(map[string]any) (map[string]any, error)) {
	a.handler[name] = handler
}

func (a *OllamaAgent) ProcessInteraction(ctx context.Context, userID, text string) (*domain.AnalysisResult, error) {
	a.logger.Info("ollama_process_start", "user_id", userID, "model", a.model, "text_len", len(text))

	messages := []ollamaChatMessage{
		{Role: "system", Content: BubbleTrackSystemPrompt()},
		{Role: "user", Content: text},
	}

	result := &domain.AnalysisResult{
		PeopleExtracted:  []domain.ExtractedPerson{},
		Relationships:    []domain.RelationshipUpdate{},
		SocialRoles:      map[string]domain.SocialRole{},
		StatesExtracted:  []domain.PersonState{},
		Protocols:        map[string]domain.Protocol{},
		ReciprocityNotes: map[string]float64{},
	}

	for i := 0; i < 10; i++ {
		a.logger.Info("ollama_iteration_start", "iteration", i+1, "model", a.model)
		resp, err := a.chat(ctx, messages)
		if err != nil {
			a.logger.Error("ollama_iteration_failed", "iteration", i+1, "error", err)
			return nil, fmt.Errorf("ollama chat iteration %d: %w", i+1, err)
		}

		if len(resp.Message.ToolCalls) == 0 {
			result.Summary = resp.Message.Content
			a.logger.Info("ollama agent completed", "iterations", i+1)
			break
		}

		messages = append(messages, resp.Message)

		for _, tc := range resp.Message.ToolCalls {
			a.logger.Info("tool call", "name", tc.Function.Name, "args", tc.Function.Arguments)
			handler, ok := a.handler[tc.Function.Name]
			if !ok {
				a.logger.Warn("unknown tool", "name", tc.Function.Name)
				continue
			}
			toolResult, err := handler(tc.Function.Arguments)
			if err != nil {
				a.logger.Warn("tool error", "name", tc.Function.Name, "error", err)
				continue
			}

			// Extract data from tool calls and populate AnalysisResult
			if tc.Function.Name == "update_graph" {
				source := ""
				target := ""
				quality := domain.QualityNeutral
				strength := 0.5

				if v, ok := tc.Function.Arguments["source_person"]; ok {
					source, _ = v.(string)
				}
				if v, ok := tc.Function.Arguments["target_person"]; ok {
					target, _ = v.(string)
				}
				if v, ok := tc.Function.Arguments["quality"]; ok {
					q, _ := v.(string)
					switch q {
					case "nourishing":
						quality = domain.QualityNourishing
					case "draining":
						quality = domain.QualityDraining
					case "conflicted":
						quality = domain.QualityConflicted
					}
				}
				if v, ok := tc.Function.Arguments["strength"]; ok {
					strength, _ = v.(float64)
				}

				// Add to relationships
				if source != "" && target != "" {
					result.Relationships = append(result.Relationships, domain.RelationshipUpdate{
						SourcePersonID:   source,
						TargetPersonID:   target,
						Quality:          quality,
						Strength:        strength,
						ReciprocityDelta: 0,
					})
				}
				// Add people
				for _, name := range []string{source, target} {
					if name != "" {
						found := false
						for _, p := range result.PeopleExtracted {
							if p.Name == name {
								found = true
								break
							}
						}
						if !found {
							result.PeopleExtracted = append(result.PeopleExtracted, domain.ExtractedPerson{
								Name:    name,
								Context: "extracted from tool call",
							})
						}
					}
				}
			}

			if tc.Function.Name == "classify_role" {
				person := ""
				role := domain.RoleUnknown

				if v, ok := tc.Function.Arguments["person"]; ok {
					person, _ = v.(string)
				}
				if v, ok := tc.Function.Arguments["role"]; ok {
					r, _ := v.(string)
					switch r {
					case "bridge":
						role = domain.RoleBridge
					case "mentor":
						role = domain.RoleMentor
					case "anchor":
						role = domain.RoleAnchor
					case "catalyst":
						role = domain.RoleCatalyst
					case "observer":
						role = domain.RoleObserver
					case "drain":
						role = domain.RoleDrain
					}
				}

				if person != "" {
					result.SocialRoles[person] = role
				}
			}

			if tc.Function.Name == "record_emotional_state" {
				personName := "self"
				if v, ok := tc.Function.Arguments["person_name"]; ok {
					personName, _ = v.(string)
				}
				mood := domain.MoodNeutral
				if v, ok := tc.Function.Arguments["mood"]; ok {
					mood = domain.Mood(v.(string))
				}
				energy := 0.5
				if v, ok := tc.Function.Arguments["energy"]; ok {
					energy, _ = v.(float64)
				}
				valence := 0.0
				if v, ok := tc.Function.Arguments["valence"]; ok {
					valence, _ = v.(float64)
				}
				contextStr := ""
				if v, ok := tc.Function.Arguments["context"]; ok {
					contextStr, _ = v.(string)
				}
				trigger := ""
				if v, ok := tc.Function.Arguments["trigger"]; ok {
					trigger, _ = v.(string)
				}
				notes := ""
				if v, ok := tc.Function.Arguments["notes"]; ok {
					notes, _ = v.(string)
				}

				result.StatesExtracted = append(result.StatesExtracted, domain.PersonState{
					UserID:  userID,
					PersonID: personName,
					Mood:    mood,
					Energy:  energy,
					Valence: valence,
					Context: contextStr,
					Trigger: trigger,
					Notes:   notes,
				})
			}

			resultBytes, _ := json.Marshal(toolResult)
			messages = append(messages, ollamaChatMessage{
				Role:    "tool",
				Content: string(resultBytes),
			})
		}
	}

	if result.Summary == "" {
		result.Summary = "Analysis complete."
	}

	a.logger.Info("ollama_process_done", "user_id", userID, "model", a.model)

	return result, nil
}

func (a *OllamaAgent) chat(ctx context.Context, messages []ollamaChatMessage) (*ollamaChatResponse, error) {
	started := time.Now()
	a.logger.Info("ollama_http_start", "url", a.baseURL+"/api/chat", "model", a.model, "messages", len(messages))

	req := ollamaChatRequest{
		Model:    a.model,
		Messages: messages,
		Tools:    a.tools,
		Stream:   false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		a.logger.Error("ollama_http_failed", "error", err, "elapsed", time.Since(started))
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		a.logger.Error("ollama_http_non_200", "status", resp.StatusCode, "elapsed", time.Since(started))
		return nil, fmt.Errorf("ollama: %s", string(respBody))
	}

	var chatResp ollamaChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	a.logger.Info("ollama_http_done", "status", resp.StatusCode, "elapsed", time.Since(started))

	return &chatResp, nil
}

func ollamaToolDefinitions() []ollamaTool {
	return []ollamaTool{
		{
			Type: "function",
			Function: ollamaToolFunctionDef{
				Name:        "search_memory",
				Description: "Searches past interaction memories using semantic similarity.",
				Parameters: ollamaToolFunctionParameters{
					Type:     "object",
					Required: []string{"query"},
					Properties: map[string]ollamaToolProperty{
						"query": {Type: "string", Description: "The search query"},
						"limit": {Type: "integer", Description: "Max results. Defaults to 5"},
					},
				},
			},
		},
		{
			Type: "function",
			Function: ollamaToolFunctionDef{
				Name:        "store_memory",
				Description: "Stores a new interaction memory for future retrieval.",
				Parameters: ollamaToolFunctionParameters{
					Type:     "object",
					Required: []string{"content"},
					Properties: map[string]ollamaToolProperty{
						"content": {Type: "string", Description: "The content to store"},
						"people":  {Type: "array", Description: "Names of people mentioned"},
					},
				},
			},
		},
		{
			Type: "function",
			Function: ollamaToolFunctionDef{
				Name:        "update_graph",
				Description: "Updates the social graph with a relationship between two people.",
				Parameters: ollamaToolFunctionParameters{
					Type:     "object",
					Required: []string{"source_person", "target_person", "quality", "strength"},
					Properties: map[string]ollamaToolProperty{
						"source_person":     {Type: "string", Description: "Source person name"},
						"target_person":     {Type: "string", Description: "Target person name"},
						"quality":           {Type: "string", Description: "Relationship quality", Enum: []string{"nourishing", "neutral", "draining", "conflicted"}},
						"strength":          {Type: "number", Description: "Strength 0.0 to 1.0"},
						"label":             {Type: "string", Description: "Label (e.g. colleague, friend)"},
						"reciprocity_delta": {Type: "number", Description: "Reciprocity change -1.0 to 1.0"},
					},
				},
			},
		},
		{
			Type: "function",
			Function: ollamaToolFunctionDef{
				Name:        "classify_role",
				Description: "Classifies a person's social role with evidence.",
				Parameters: ollamaToolFunctionParameters{
					Type:     "object",
					Required: []string{"person", "role", "evidence"},
					Properties: map[string]ollamaToolProperty{
						"person":   {Type: "string", Description: "Person name"},
						"role":     {Type: "string", Description: "Social role", Enum: []string{"bridge", "mentor", "anchor", "catalyst", "observer", "drain"}},
						"evidence": {Type: "string", Description: "Evidence from interaction"},
					},
				},
			},
		},
		{
			Type: "function",
			Function: ollamaToolFunctionDef{
				Name:        "set_user_preferences",
				Description: "Stores user preferences for analysis delivery.",
				Parameters: ollamaToolFunctionParameters{
					Type:     "object",
					Required: []string{"philosophical_lens"},
					Properties: map[string]ollamaToolProperty{
						"philosophical_lens": {Type: "string", Description: "User outlook", Enum: []string{"humanist", "pragmatic", "spiritual", "stoic", "romantic", "existential", "utilitarian", "conservative", "scientific"}},
						"preferences":        {Type: "string", Description: "Delivery preferences"},
					},
				},
			},
		},
		{
			Type: "function",
			Function: ollamaToolFunctionDef{
				Name:        "record_emotional_state",
				Description: "Records the emotional state of the user or a person during an interaction. Use this for every person mentioned to track how they felt.",
				Parameters: ollamaToolFunctionParameters{
					Type:     "object",
					Required: []string{"person_name", "mood"},
					Properties: map[string]ollamaToolProperty{
						"person_name": {Type: "string", Description: "Name of the person (use 'self' for the user's own state)"},
						"mood":        {Type: "string", Description: "Emotional state", Enum: []string{"happy", "anxious", "tired", "energized", "sad", "neutral", "angry", "hopeful", "lonely", "grateful"}},
						"energy":      {Type: "number", Description: "Energy level 0.0 to 1.0"},
						"valence":     {Type: "number", Description: "Emotional valence -1.0 (negative) to 1.0 (positive)"},
						"context":     {Type: "string", Description: "Where/what setting: gathering, workplace, home, event, online"},
						"trigger":     {Type: "string", Description: "What caused this state: interaction, reflection, event"},
						"notes":       {Type: "string", Description: "Additional observations"},
					},
				},
			},
		},
		{
			Type: "function",
			Function: ollamaToolFunctionDef{
				Name:        "update_relationship_protocol",
				Description: "Records the communication protocol and investment levels of a relationship. Captures how deep the connection is and how mutual.",
				Parameters: ollamaToolFunctionParameters{
					Type:     "object",
					Required: []string{"source_person", "target_person", "protocol"},
					Properties: map[string]ollamaToolProperty{
						"source_person":       {Type: "string", Description: "Source person name"},
						"target_person":       {Type: "string", Description: "Target person name"},
						"protocol":            {Type: "string", Description: "Communication type", Enum: []string{"deep", "casual", "professional", "digital", "mixed"}},
						"source_investment":   {Type: "number", Description: "How much source invests 0.0 to 1.0"},
						"target_investment":   {Type: "number", Description: "How much target invests 0.0 to 1.0"},
					},
				},
			},
		},
	}
}
