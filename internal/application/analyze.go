package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/bubbletrack/server/internal/domain"
	"github.com/bubbletrack/server/internal/infrastructure/queue"
	"github.com/bubbletrack/server/internal/infrastructure/repository"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type Agent interface {
	ProcessInteraction(ctx context.Context, userID, text string) (*domain.AnalysisResult, error)
}

type ChatMessageRepository interface {
	CreateChatMessage(ctx context.Context, userID, sender, content string, isUser bool) error
}

type AnalyzeUseCase struct {
	agent            Agent
	interactionRepo  domain.InteractionRepository
	personRepo       domain.PersonRepository
	relationshipRepo domain.RelationshipRepository
	memoryRepo       domain.MemoryRepository
	embedder         domain.Embedder
	notifier         domain.AnalysisNotifier
	chatRepo         ChatMessageRepository
	chatNotifier     domain.ChatMessageNotifier
	stateRepo        domain.PersonStateRepository
	enqueuer         *queue.Enqueuer
	logger           *slog.Logger
}

func NewAnalyzeUseCase(
	agent Agent,
	interactionRepo domain.InteractionRepository,
	personRepo domain.PersonRepository,
	relationshipRepo domain.RelationshipRepository,
	memoryRepo domain.MemoryRepository,
	embedder domain.Embedder,
	notifier domain.AnalysisNotifier,
	enqueuer *queue.Enqueuer,
	chatRepo ChatMessageRepository,
	chatNotifier domain.ChatMessageNotifier,
	stateRepo domain.PersonStateRepository,
	logger *slog.Logger,
) *AnalyzeUseCase {
	return &AnalyzeUseCase{
		agent:            agent,
		interactionRepo:  interactionRepo,
		personRepo:       personRepo,
		relationshipRepo: relationshipRepo,
		memoryRepo:       memoryRepo,
		embedder:         embedder,
		notifier:         notifier,
		chatRepo:         chatRepo,
		chatNotifier:     chatNotifier,
		stateRepo:        stateRepo,
		enqueuer:         enqueuer,
		logger:           logger,
	}
}

type AnalyzeRequest struct {
	UserID  string
	RawText string
}

type AnalyzeResponse struct {
	InteractionID string `json:"interaction_id"`
	JobID         string `json:"job_id"`
	Status        string `json:"status"`
}

func (uc *AnalyzeUseCase) Submit(ctx context.Context, req AnalyzeRequest) (*AnalyzeResponse, error) {
	if err := domain.ValidateInteracao(&domain.Interacao{UserID: req.UserID, RawText: req.RawText}); err != nil {
		return nil, err
	}

	interaction := &domain.Interacao{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		RawText:   req.RawText,
		Status:    domain.StatusQueued,
		CreatedAt: time.Now().UTC(),
	}

	if err := uc.interactionRepo.Create(ctx, interaction); err != nil {
		return nil, fmt.Errorf("save interaction: %w", err)
	}

	jobID, err := uc.enqueuer.EnqueueAnalyze(ctx, queue.AnalyzePayload{
		InteractionID: interaction.ID,
		UserID:        req.UserID,
		RawText:       req.RawText,
	})
	if err != nil {
		return nil, fmt.Errorf("enqueue: %w", err)
	}

	interaction.JobID = jobID
	if err := uc.interactionRepo.UpdateStatus(ctx, interaction.ID, domain.StatusQueued); err != nil {
		uc.logger.Warn("failed to update job_id", "error", err)
	}

	uc.logger.Info("interaction queued", "interaction_id", interaction.ID, "job_id", jobID)

	return &AnalyzeResponse{
		InteractionID: interaction.ID,
		JobID:         jobID,
		Status:        "queued",
	}, nil
}

func (uc *AnalyzeUseCase) ProcessJob(ctx context.Context, payload queue.AnalyzePayload) error {
	uc.logger.Info("process_job_start", "interaction_id", payload.InteractionID, "user_id", payload.UserID)

	if err := uc.interactionRepo.UpdateStatus(ctx, payload.InteractionID, domain.StatusProcessing); err != nil {
		return fmt.Errorf("update status to processing: %w", err)
	}

	var result *domain.AnalysisResult
	var err error

	if uc.embedder != nil && uc.memoryRepo != nil {
		ragCtx, cancelRAG := context.WithTimeout(ctx, 8*time.Second)
		defer cancelRAG()

		uc.logger.Info("process_job_rag_start", "interaction_id", payload.InteractionID)
		embedding, embedErr := uc.embedder.Embed(ragCtx, payload.RawText)
		if embedErr == nil {
			memories, searchErr := uc.memoryRepo.Search(ragCtx, payload.UserID, embedding, 5)
			if searchErr == nil && len(memories) > 0 {
				ragContext := "## Contexto Relevante de Interações Anteriores:\n"
				for i, mem := range memories {
					if raw, ok := mem.Metadata["raw_text"].(string); ok {
						ragContext += fmt.Sprintf("\n%d. %s", i+1, raw)
					}
				}
				payload.RawText = ragContext + "\n\n" + payload.RawText
			}
			if searchErr != nil {
				uc.logger.Warn("process_job_rag_search_failed", "interaction_id", payload.InteractionID, "error", searchErr)
			}
		} else {
			uc.logger.Warn("process_job_rag_embed_failed", "interaction_id", payload.InteractionID, "error", embedErr)
		}
		uc.logger.Info("process_job_rag_done", "interaction_id", payload.InteractionID)
	}

	if uc.agent != nil {
		uc.logger.Info("process_job_agent_start", "interaction_id", payload.InteractionID)
		result, err = uc.agent.ProcessInteraction(ctx, payload.UserID, payload.RawText)
		if err != nil {
			uc.logger.Error("process_job_agent_failed", "interaction_id", payload.InteractionID, "error", err)
			uc.interactionRepo.UpdateStatus(ctx, payload.InteractionID, domain.StatusFailed)
			return fmt.Errorf("agent processing: %w", err)
		}
		uc.logger.Info("process_job_agent_done", "interaction_id", payload.InteractionID)
	} else {
		result = &domain.AnalysisResult{
			Summary:         "Processed via embedding storage only (no AI agent)",
			PeopleExtracted: []domain.ExtractedPerson{},
			Relationships:   []domain.RelationshipUpdate{},
			SocialRoles:     map[string]domain.SocialRole{},
		}
	}

	result.InteractionID = payload.InteractionID

	if err := uc.applyResults(ctx, payload.UserID, result); err != nil {
		uc.logger.Error("failed to apply results", "error", err)
	}

	if uc.embedder != nil {
		embedding, err := uc.embedder.Embed(ctx, payload.RawText)
		if err != nil {
			uc.logger.Warn("failed to embed interaction", "error", err)
		} else {
			peopleNames := make([]string, 0, len(result.PeopleExtracted))
			for _, p := range result.PeopleExtracted {
				peopleNames = append(peopleNames, p.Name)
			}
			peopleJSON, _ := json.Marshal(peopleNames)
			metadata := map[string]any{
				"raw_text":  payload.RawText,
				"summary":   result.Summary,
				"people":    string(peopleJSON),
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			}
			if err := uc.memoryRepo.Store(ctx, payload.UserID, payload.InteractionID, embedding, metadata); err != nil {
				uc.logger.Warn("failed to store memory", "error", err)
			}
		}
	}

	uc.interactionRepo.UpdateStatus(ctx, payload.InteractionID, domain.StatusCompleted)

	if uc.notifier != nil {
		if err := uc.notifier.Notify(ctx, payload.UserID, result); err != nil {
			uc.logger.Warn("failed to notify", "error", err)
		}
	}

	if result.Summary != "" && uc.chatRepo != nil {
		if err := uc.chatRepo.CreateChatMessage(ctx, payload.UserID, "BubbleTrack", result.Summary, false); err != nil {
			uc.logger.Warn("failed to save chat response", "error", err)
		}
		if uc.chatNotifier != nil {
			if err := uc.chatNotifier.NotifyChatMessage(payload.UserID, "BubbleTrack", result.Summary, false); err != nil {
				uc.logger.Warn("failed to broadcast chat response", "error", err)
			}
		}
	}

	if len(result.StatesExtracted) > 0 && uc.stateRepo != nil && uc.personRepo != nil {
		for _, state := range result.StatesExtracted {
			state.UserID = payload.UserID
			state.InteractionID = payload.InteractionID
			personID := state.PersonID

			if personID == "self" || personID == "" {
				personID = payload.UserID
			} else {
				person, err := uc.personRepo.GetOrCreateByName(ctx, personID)
				if err != nil {
					uc.logger.Warn("failed to find person", "name", personID, "error", err)
					continue
				}
				personID = person.ID
			}

			state.PersonID = personID
			if err := uc.stateRepo.Create(ctx, &state); err != nil {
				uc.logger.Warn("failed to save emotional state", "person", state.PersonID, "error", err)
			}
		}
		uc.logger.Info("emotional states saved", "count", len(result.StatesExtracted), "interaction_id", payload.InteractionID)
	}

	uc.logger.Info("interaction processed", "interaction_id", payload.InteractionID)
	uc.logger.Info("process_job_done", "interaction_id", payload.InteractionID)
	return nil
}

func (uc *AnalyzeUseCase) applyResults(ctx context.Context, userID string, result *domain.AnalysisResult) error {
	for _, pe := range result.PeopleExtracted {
		person := &domain.Pessoa{
			ID:          uuid.New().String(),
			DisplayName: pe.Name,
			Aliases:     pe.Aliases,
		}
		if err := uc.personRepo.Upsert(ctx, person); err != nil {
			uc.logger.Warn("failed to upsert person", "name", pe.Name, "error", err)
			continue
		}
	}

	for _, ru := range result.Relationships {
		// Get UUIDs from display names
		sourcePerson, err := uc.personRepo.GetOrCreateByName(ctx, ru.SourcePersonID)
		if err != nil {
			uc.logger.Warn("applyResults: failed to get source person", "name", ru.SourcePersonID, "error", err)
			continue
		}
		targetPerson, err := uc.personRepo.GetOrCreateByName(ctx, ru.TargetPersonID)
		if err != nil {
			uc.logger.Warn("applyResults: failed to get target person", "name", ru.TargetPersonID, "error", err)
			continue
		}

		rel, err := uc.relationshipRepo.GetByPeople(ctx, sourcePerson.ID, targetPerson.ID)
		if err != nil {
			if _, ok := err.(*domain.NotFoundError); ok {
				rel = &domain.Relacionamento{
					ID:               uuid.New().String(),
					SourcePersonID:   sourcePerson.ID,
					TargetPersonID:   targetPerson.ID,
					Quality:          ru.Quality,
					Strength:         ru.Strength,
					Label:            ru.Label,
					ReciprocityIndex: 0.5,
				}
				if err := uc.relationshipRepo.Create(ctx, rel); err != nil {
					uc.logger.Warn("failed to create relationship", "error", err)
					continue
				}
			} else {
				uc.logger.Warn("failed to check relationship", "error", err)
				continue
			}
		} else {
			rel.Strength = clamp(rel.Strength+ru.Strength*0.2, 0, 1)
			rel.Quality = ru.Quality
			rel.ReciprocityIndex = clamp(rel.ReciprocityIndex+ru.ReciprocityDelta*0.1, 0, 1)
			if ru.Label != "" {
				rel.Label = ru.Label
			}
			if err := uc.relationshipRepo.Update(ctx, rel); err != nil {
				uc.logger.Warn("failed to update relationship", "error", err)
			}
		}
	}

	return nil
}

type GetGraphUseCase struct {
	graphRepo repository.PostgresGraphRepository
}

func NewGetGraphUseCase(graphRepo repository.PostgresGraphRepository) *GetGraphUseCase {
	return &GetGraphUseCase{graphRepo: graphRepo}
}

func (uc *GetGraphUseCase) Execute(ctx context.Context, userID string) (*domain.Graph, error) {
	return uc.graphRepo.GetGraph(ctx, userID)
}

func RegisterWorker(mux *asynq.ServeMux, uc *AnalyzeUseCase, logger *slog.Logger) {
	mux.HandleFunc(queue.TypeAnalyzeInteraction, func(ctx context.Context, t *asynq.Task) error {
		payload, err := queue.DecodeAnalyzePayload(t)
		if err != nil {
			logger.Error("failed to decode task", "error", err)
			return fmt.Errorf("decode: %w", err)
		}
		logger.Info("processing interaction", "interaction_id", payload.InteractionID)
		return uc.ProcessJob(ctx, *payload)
	})
}

type chatMessageRepoAdapter struct {
	inner *repository.ChatMessageRepository
}

func NewChatMessageRepoAdapter(inner *repository.ChatMessageRepository) ChatMessageRepository {
	return &chatMessageRepoAdapter{inner: inner}
}

func (a *chatMessageRepoAdapter) CreateChatMessage(ctx context.Context, userID, sender, content string, isUser bool) error {
	msg := &repository.ChatMessage{
		UserID:  userID,
		Sender:  sender,
		Content: content,
		IsUser:  isUser,
	}
	return a.inner.Create(ctx, msg)
}
