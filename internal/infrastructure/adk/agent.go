package adk

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bubbletrack/server/internal/domain"
	"google.golang.org/genai"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
)

type BubbleTrackAgent struct {
	agent  agent.Agent
	runner *runner.Runner
	logger *slog.Logger
}

func NewBubbleTrackAgent(
	ctx context.Context,
	apiKey string,
	modelName string,
	tools []tool.Tool,
	logger *slog.Logger,
) (*BubbleTrackAgent, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY required for ADK agent")
	}

	model, err := gemini.NewModel(ctx, modelName, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("create gemini model: %w", err)
	}

	a, err := llmagent.New(llmagent.Config{
		Name:        "bubble_track",
		Description: "Analyzes social dynamics from free-text interaction reports using interdisciplinary frameworks.",
		Model:       model,
		Instruction: BubbleTrackSystemPrompt(),
		Tools:       tools,
		GenerateContentConfig: &genai.GenerateContentConfig{
			Temperature: genai.Ptr(float32(0.7)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create agent: %w", err)
	}

	sessionService := session.InMemoryService()

	r, err := runner.New(runner.Config{
		AppName:           "bubble_track",
		Agent:             a,
		SessionService:    sessionService,
		AutoCreateSession: true,
	})
	if err != nil {
		return nil, fmt.Errorf("create runner: %w", err)
	}

	return &BubbleTrackAgent{
		agent:  a,
		runner: r,
		logger: logger,
	}, nil
}

func (a *BubbleTrackAgent) ProcessInteraction(ctx context.Context, userID, text string) (*domain.AnalysisResult, error) {
	sessionID := fmt.Sprintf("session-%s", userID)

	msg := genai.NewContentFromText(text, genai.RoleUser)

	var summary string
	for event, err := range a.runner.Run(ctx, userID, sessionID, msg, agent.RunConfig{}) {
		if err != nil {
			a.logger.Warn("agent event error", "error", err)
			continue
		}
		if event != nil && event.Content != nil {
			for _, part := range event.Content.Parts {
				if part.Text != "" {
					summary += part.Text
				}
			}
		}
	}

	if summary == "" {
		summary = "Analysis complete."
	}

	return &domain.AnalysisResult{
		Summary:         summary,
		PeopleExtracted: []domain.ExtractedPerson{},
		Relationships:   []domain.RelationshipUpdate{},
		SocialRoles:     map[string]domain.SocialRole{},
	}, nil
}

func (a *BubbleTrackAgent) Name() string {
	return a.agent.Name()
}
