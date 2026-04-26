package domain

import (
	"context"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, a *Account) error
	GetByID(ctx context.Context, id string) (*Account, error)
	GetByEmail(ctx context.Context, email string) (*Account, error)
	UpdatePassword(ctx context.Context, id string, hash string) error
}

type RefreshTokenRepository interface {
	Store(ctx context.Context, token *RefreshToken) error
	GetByToken(ctx context.Context, token string) (*RefreshToken, error)
	Revoke(ctx context.Context, tokenID string) error
	RevokeAllForUser(ctx context.Context, userID string) error
}

type PersonRepository interface {
	Create(ctx context.Context, p *Pessoa) error
	GetByID(ctx context.Context, id string) (*Pessoa, error)
	GetByUserID(ctx context.Context, userID string) ([]Pessoa, error)
	GetOrCreateByName(ctx context.Context, name string) (*Pessoa, error)
	Update(ctx context.Context, p *Pessoa) error
	Upsert(ctx context.Context, p *Pessoa) error
}

type RelationshipRepository interface {
	Create(ctx context.Context, r *Relacionamento) error
	GetByID(ctx context.Context, id string) (*Relacionamento, error)
	GetByPeople(ctx context.Context, sourceID, targetID string) (*Relacionamento, error)
	GetByUser(ctx context.Context, userID string) ([]Relacionamento, error)
	Update(ctx context.Context, r *Relacionamento) error
	Upsert(ctx context.Context, r *Relacionamento) error
}

type InteractionRepository interface {
	Create(ctx context.Context, i *Interacao) error
	GetByID(ctx context.Context, id string) (*Interacao, error)
	UpdateStatus(ctx context.Context, id string, status JobStatus) error
}

type MemoryRepository interface {
	Store(ctx context.Context, userID string, interactionID string, embedding []float32, metadata map[string]any) error
	Search(ctx context.Context, userID string, queryEmbedding []float32, limit int) ([]MemoryResult, error)
	SearchFiltered(ctx context.Context, userID string, queryEmbedding []float32, limit int, filter MemoryFilter) ([]MemoryResult, error)
}

type MemoryFilter struct {
	Segment string
	Session string
	Person  string
	Since   *time.Time
	Until   *time.Time
}

type MemoryResult struct {
	InteractionID string
	Score         float32
	Metadata      map[string]any
}

type UserProvider interface {
	GetCurrentUserID(ctx context.Context) string
	GetUserDisplayName(ctx context.Context) string
}

type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

type AnalysisNotifier interface {
	Notify(ctx context.Context, userID string, result *AnalysisResult) error
}

type ChatMessageNotifier interface {
	NotifyChatMessage(userID, sender, content string, isUser bool) error
}

type PersonStateRepository interface {
	Create(ctx context.Context, state *PersonState) error
	GetByPerson(ctx context.Context, userID, personID string, limit int) ([]PersonState, error)
	GetByUser(ctx context.Context, userID string, limit int) ([]PersonState, error)
	GetTimeline(ctx context.Context, userID string, limit int) ([]PersonState, error)
	GetSelfStates(ctx context.Context, userID string, limit int) ([]PersonState, error)
}

type WoundRepository interface {
	Create(ctx context.Context, w *AttachmentWound) error
	GetByPerson(ctx context.Context, personID string) ([]AttachmentWound, error)
	UpdateProcessed(ctx context.Context, woundID string, processed bool) error
}

type GhostingRepository interface {
	Create(ctx context.Context, g *GhostingPattern) error
	GetByPerson(ctx context.Context, personID string) ([]GhostingPattern, error)
	UpdateFrequency(ctx context.Context, personID, targetID string) error
}

type DecisionRepository interface {
	Create(ctx context.Context, d *DecisionContext) error
	GetBySession(ctx context.Context, sessionID string) ([]DecisionContext, error)
	UpdateRegret(ctx context.Context, decisionID string, regret bool) error
}

type DramaRepository interface {
	Create(ctx context.Context, d *InterpersonalDrama) error
	GetBySession(ctx context.Context, sessionID string) ([]InterpersonalDrama, error)
	UpdateResolution(ctx context.Context, dramaID, resolution string) error
}

type SessionRepository interface {
	Create(ctx context.Context, s *UserSession) error
	GetByID(ctx context.Context, id string) (*UserSession, error)
	GetActive(ctx context.Context, userID string) (*UserSession, error)
	UpdateStatus(ctx context.Context, id string, status SessionStatus) error
	AddInsight(ctx context.Context, s *SessionInsight) error
	GetInsights(ctx context.Context, sessionID string) ([]SessionInsight, error)
}
