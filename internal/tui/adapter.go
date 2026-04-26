package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/bubbletrack/server/internal/domain"
	"github.com/bubbletrack/server/internal/infrastructure/repository"
	"github.com/google/uuid"
)

type GraphAdapter struct {
	repo *repository.PostgresGraphRepository
}

func NewGraphAdapter(repo *repository.PostgresGraphRepository) *GraphAdapter {
	return &GraphAdapter{repo: repo}
}

func (a *GraphAdapter) GetGraph(userID string) (*GraphData, error) {
	ctx := context.Background()
	graph, err := a.repo.GetGraph(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	return convertGraph(graph), nil
}

type ChatAdapter struct {
	repo *repository.ChatMessageRepository
}

func NewChatAdapter(repo *repository.ChatMessageRepository) *ChatAdapter {
	return &ChatAdapter{repo: repo}
}

func (a *ChatAdapter) SaveMessage(userID, sender, content string, isUser bool) error {
	ctx := context.Background()
	msg := &repository.ChatMessage{
		ID:        uuid.New(),
		UserID:    userID,
		Sender:    sender,
		Content:   content,
		IsUser:    isUser,
		CreatedAt: time.Now().UTC(),
	}
	return a.repo.Create(ctx, msg)
}

func (a *ChatAdapter) LoadMessages(userID string, limit int) ([]ChatMessage, error) {
	ctx := context.Background()
	dbMsgs, err := a.repo.GetByUserID(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to load messages: %w", err)
	}

	msgs := make([]ChatMessage, len(dbMsgs))
	for i, m := range dbMsgs {
		msgs[i] = ChatMessage{
			ID:        m.ID.String(),
			Sender:    m.Sender,
			Content:   m.Content,
			IsUser:    m.IsUser,
			Timestamp: m.CreatedAt.Format("15:04"),
		}
	}
	return msgs, nil
}

func convertGraph(graph *domain.Graph) *GraphData {
	data := &GraphData{
		Nodes: make([]Node, 0, len(graph.Nodes)),
		Edges: make([]Edge, 0, len(graph.Edges)),
		Stats: GraphStats{
			TotalPeople:         graph.Stats.TotalPeople,
			TotalRelationships:  graph.Stats.TotalRelationships,
			AvgReciprocity:      graph.Stats.AvgReciprocity,
			BridgeCount:         graph.Stats.BridgeCount,
			StrongestConnection: graph.Stats.StrongestConnection,
		},
	}

	for _, n := range graph.Nodes {
		data.Nodes = append(data.Nodes, Node{
			ID:           n.ID,
			Name:         n.DisplayName,
			Role:         string(n.SocialRole),
			Mood:         string(n.CurrentMood),
			Energy:       n.CurrentEnergy,
			InteractCount: n.InteractionCount,
		})
	}

	for _, e := range graph.Edges {
		data.Edges = append(data.Edges, Edge{
			Source:       e.Source,
			Target:       e.Target,
			Quality:     string(e.Quality),
			Strength:    e.Strength,
			Protocol:    string(e.Protocol),
			SourceW:     e.SourceWeight,
			TargetW:     e.TargetWeight,
			ReciprocityIndex: e.ReciprocityIndex,
		})
	}

	return data
}
