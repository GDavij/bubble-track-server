package mock

import (
	"context"
	"time"

	"github.com/bubbletrack/server/internal/domain"
)

type MockPersonRepository struct {
	People   map[string]*domain.Pessoa
	CreateFn func(ctx context.Context, p *domain.Pessoa) error
	GetFn    func(ctx context.Context, id string) (*domain.Pessoa, error)
}

func NewMockPersonRepository() *MockPersonRepository {
	return &MockPersonRepository{People: make(map[string]*domain.Pessoa)}
}

func (m *MockPersonRepository) Create(ctx context.Context, p *domain.Pessoa) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, p)
	}
	m.People[p.ID] = p
	return nil
}

func (m *MockPersonRepository) GetByID(ctx context.Context, id string) (*domain.Pessoa, error) {
	if m.GetFn != nil {
		return m.GetFn(ctx, id)
	}
	p, ok := m.People[id]
	if !ok {
		return nil, &domain.NotFoundError{Entity: "person", ID: id}
	}
	return p, nil
}

func (m *MockPersonRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Pessoa, error) {
	var result []domain.Pessoa
	for _, p := range m.People {
		result = append(result, *p)
	}
	return result, nil
}

func (m *MockPersonRepository) Update(ctx context.Context, p *domain.Pessoa) error {
	m.People[p.ID] = p
	return nil
}

func (m *MockPersonRepository) Upsert(ctx context.Context, p *domain.Pessoa) error {
	m.People[p.ID] = p
	return nil
}

func (m *MockPersonRepository) GetOrCreateByName(ctx context.Context, name string) (*domain.Pessoa, error) {
	for _, p := range m.People {
		if p.DisplayName == name {
			return p, nil
		}
	}
	p := &domain.Pessoa{
		ID:          "p-" + name,
		DisplayName: name,
	}
	m.People[p.ID] = p
	return p, nil
}

func (m *MockPersonRepository) UpdateSocialRole(ctx context.Context, personID string, role domain.SocialRole) error {
	if p, ok := m.People[personID]; ok {
		p.SocialRole = role
	}
	return nil
}

type MockRelationshipRepository struct {
	Relations map[string]*domain.Relacionamento
}

func NewMockRelationshipRepository() *MockRelationshipRepository {
	return &MockRelationshipRepository{Relations: make(map[string]*domain.Relacionamento)}
}

func (m *MockRelationshipRepository) Create(ctx context.Context, r *domain.Relacionamento) error {
	m.Relations[r.ID] = r
	return nil
}

func (m *MockRelationshipRepository) GetByID(ctx context.Context, id string) (*domain.Relacionamento, error) {
	r, ok := m.Relations[id]
	if !ok {
		return nil, &domain.NotFoundError{Entity: "relationship", ID: id}
	}
	return r, nil
}

func (m *MockRelationshipRepository) GetByPeople(ctx context.Context, sourceID, targetID string) (*domain.Relacionamento, error) {
	for _, r := range m.Relations {
		if r.SourcePersonID == sourceID && r.TargetPersonID == targetID {
			return r, nil
		}
	}
	return nil, &domain.NotFoundError{Entity: "relationship", ID: sourceID + "-" + targetID}
}

func (m *MockRelationshipRepository) GetByUser(ctx context.Context, userID string) ([]domain.Relacionamento, error) {
	var result []domain.Relacionamento
	for _, r := range m.Relations {
		result = append(result, *r)
	}
	return result, nil
}

func (m *MockRelationshipRepository) Update(ctx context.Context, r *domain.Relacionamento) error {
	m.Relations[r.ID] = r
	return nil
}

func (m *MockRelationshipRepository) Upsert(ctx context.Context, r *domain.Relacionamento) error {
	m.Relations[r.ID] = r
	return nil
}

type MockInteractionRepository struct {
	Interactions map[string]*domain.Interacao
}

func NewMockInteractionRepository() *MockInteractionRepository {
	return &MockInteractionRepository{Interactions: make(map[string]*domain.Interacao)}
}

func (m *MockInteractionRepository) Create(ctx context.Context, i *domain.Interacao) error {
	m.Interactions[i.ID] = i
	return nil
}

func (m *MockInteractionRepository) GetByID(ctx context.Context, id string) (*domain.Interacao, error) {
	i, ok := m.Interactions[id]
	if !ok {
		return nil, &domain.NotFoundError{Entity: "interaction", ID: id}
	}
	return i, nil
}

func (m *MockInteractionRepository) UpdateStatus(ctx context.Context, id string, status domain.JobStatus) error {
	if i, ok := m.Interactions[id]; ok {
		i.Status = status
	}
	return nil
}

type MockMemoryRepository struct {
	Memories map[string][]domain.MemoryResult
}

func NewMockMemoryRepository() *MockMemoryRepository {
	return &MockMemoryRepository{Memories: make(map[string][]domain.MemoryResult)}
}

func (m *MockMemoryRepository) Store(ctx context.Context, userID string, interactionID string, embedding []float32, metadata map[string]any) error {
	m.Memories[userID] = append(m.Memories[userID], domain.MemoryResult{
		InteractionID: interactionID,
		Metadata:       metadata,
	})
	return nil
}

func (m *MockMemoryRepository) Search(ctx context.Context, userID string, queryEmbedding []float32, limit int) ([]domain.MemoryResult, error) {
	results := m.Memories[userID]
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

func (m *MockMemoryRepository) SearchFiltered(ctx context.Context, userID string, queryEmbedding []float32, limit int, filter domain.MemoryFilter) ([]domain.MemoryResult, error) {
	return m.Search(ctx, userID, queryEmbedding, limit)
}

type MockEmbedder struct{}

func NewMockEmbedder() *MockEmbedder {
	return &MockEmbedder{}
}

func (m *MockEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	emb := make([]float32, 768)
	h := 0
	for _, c := range text {
		h = h*31 + int(c)
	}
	for i := range emb {
		emb[i] = float32((h+i)%1000) / 1000.0
	}
	return emb, nil
}

type MockNotifier struct {
	Notifications []domain.AnalysisResult
}

func NewMockNotifier() *MockNotifier {
	return &MockNotifier{Notifications: make([]domain.AnalysisResult, 0)}
}

func (m *MockNotifier) Notify(ctx context.Context, userID string, result *domain.AnalysisResult) error {
	m.Notifications = append(m.Notifications, *result)
	return nil
}

func NewMockPessoa(id, name string) *domain.Pessoa {
	return &domain.Pessoa{ID: id, DisplayName: name, CreatedAt: time.Now()}
}

func NewMockRelacionamento(id, source, target string, strength float64) *domain.Relacionamento {
	return &domain.Relacionamento{
		ID:              id,
		SourcePersonID:  source,
		TargetPersonID:  target,
		Strength:        strength,
		Quality:         domain.QualityNourishing,
		ReciprocityIndex: 0.5,
	}
}