package adk

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/bubbletrack/server/internal/domain"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

type deps struct {
	memoryRepo       domain.MemoryRepository
	embedder         domain.Embedder
	personRepo       domain.PersonRepository
	relationshipRepo domain.RelationshipRepository
	logger           *slog.Logger
}

func newDeps(
	memoryRepo domain.MemoryRepository,
	embedder domain.Embedder,
	personRepo domain.PersonRepository,
	relationshipRepo domain.RelationshipRepository,
	logger *slog.Logger,
) *deps {
	return &deps{
		memoryRepo:       memoryRepo,
		embedder:         embedder,
		personRepo:       personRepo,
		relationshipRepo: relationshipRepo,
		logger:           logger,
	}
}

func userIDFromCtx(ctx tool.Context) string {
	if state := ctx.State(); state != nil {
		if uid, err := state.Get("user_id"); err == nil {
			if s, ok := uid.(string); ok {
				return s
			}
		}
	}
	return "unknown"
}

type SearchMemoryInput struct {
	Query string `json:"query" jsonschema:"description=The search query to find relevant past interactions or memories"`
	Limit int    `json:"limit,omitempty" jsonschema:"description=Maximum number of memories to return. Defaults to 5"`
}

type SearchMemoryOutput struct {
	Memories []memoryItem `json:"memories"`
	Count    int          `json:"count"`
}

type memoryItem struct {
	Content  string         `json:"content"`
	Score    float32        `json:"score"`
	Metadata map[string]any `json:"metadata"`
}

func (d *deps) searchMemoryHandler(ctx tool.Context, input SearchMemoryInput) (SearchMemoryOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 5
	}

	userID := userIDFromCtx(ctx)
	embedding, err := d.embedder.Embed(context.Background(), input.Query)
	if err != nil {
		return SearchMemoryOutput{}, fmt.Errorf("embed query: %w", err)
	}

	results, err := d.memoryRepo.Search(context.Background(), userID, embedding, limit)
	if err != nil {
		return SearchMemoryOutput{}, fmt.Errorf("search memory: %w", err)
	}

	items := make([]memoryItem, 0, len(results))
	for _, r := range results {
		rawText := ""
		if v, ok := r.Metadata["raw_text"].(string); ok {
			rawText = v
		}
		items = append(items, memoryItem{
			Content:  rawText,
			Score:    r.Score,
			Metadata: r.Metadata,
		})
	}

	return SearchMemoryOutput{Memories: items, Count: len(items)}, nil
}

type StoreMemoryInput struct {
	Content string   `json:"content" jsonschema:"description=The content of the memory to store"`
	People  []string `json:"people,omitempty" jsonschema:"description=Names of people mentioned in this memory"`
}

type StoreMemoryOutput struct {
	Status         string `json:"status"`
	InteractionID  string `json:"interaction_id"`
}

func (d *deps) storeMemoryHandler(ctx tool.Context, input StoreMemoryInput) (StoreMemoryOutput, error) {
	userID := userIDFromCtx(ctx)
	embedding, err := d.embedder.Embed(context.Background(), input.Content)
	if err != nil {
		return StoreMemoryOutput{}, fmt.Errorf("embed content: %w", err)
	}

	interactionID := fmt.Sprintf("tool-%d", time.Now().UnixNano())
	metadata := map[string]any{
		"raw_text":  input.Content,
		"people":    input.People,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if err := d.memoryRepo.Store(context.Background(), userID, interactionID, embedding, metadata); err != nil {
		return StoreMemoryOutput{}, fmt.Errorf("store memory: %w", err)
	}

	return StoreMemoryOutput{Status: "stored", InteractionID: interactionID}, nil
}

type UpdateGraphInput struct {
	SourcePerson    string  `json:"source_person" jsonschema:"description=Name of the source person in the relationship"`
	TargetPerson    string  `json:"target_person" jsonschema:"description=Name of the target person in the relationship"`
	Quality         string  `json:"quality" jsonschema:"description=Emotional quality of the relationship,enum=nourishing,enum=neutral,enum=draining,enum=conflicted"`
	Strength        float64 `json:"strength" jsonschema:"description=Relationship strength from 0.0 to 1.0"`
	Label           string  `json:"label,omitempty" jsonschema:"description=Label describing the relationship (e.g. colleague, friend, family)"`
	ReciprocityDelta float64 `json:"reciprocity_delta,omitempty" jsonschema:"description=Change in reciprocity from -1.0 to 1.0"`
}

type UpdateGraphOutput struct {
	Status   string  `json:"status"`
	Source   string  `json:"source"`
	Target   string  `json:"target"`
	Quality  string  `json:"quality"`
	Strength float64 `json:"strength"`
}

func (d *deps) updateGraphHandler(ctx tool.Context, input UpdateGraphInput) (UpdateGraphOutput, error) {
	c := context.Background()
	sourcePerson, err := d.findOrCreatePerson(c, input.SourcePerson)
	if err != nil {
		return UpdateGraphOutput{}, fmt.Errorf("source person: %w", err)
	}
	targetPerson, err := d.findOrCreatePerson(c, input.TargetPerson)
	if err != nil {
		return UpdateGraphOutput{}, fmt.Errorf("target person: %w", err)
	}

	rel, err := d.relationshipRepo.GetByPeople(c, sourcePerson.ID, targetPerson.ID)
	if err != nil {
		if _, ok := err.(*domain.NotFoundError); ok {
			rel = &domain.Relacionamento{
				ID:               fmt.Sprintf("rel-%s-%s", input.SourcePerson, input.TargetPerson),
				SourcePersonID:   sourcePerson.ID,
				TargetPersonID:   targetPerson.ID,
				Quality:          domain.Quality(input.Quality),
				Strength:         input.Strength,
				Label:            input.Label,
				ReciprocityIndex: 0.5,
			}
			if verr := domain.ValidateRelacionamento(rel); verr != nil {
				return UpdateGraphOutput{}, verr
			}
			if err := d.relationshipRepo.Create(c, rel); err != nil {
				return UpdateGraphOutput{}, fmt.Errorf("create relationship: %w", err)
			}
		} else {
			return UpdateGraphOutput{}, fmt.Errorf("check relationship: %w", err)
		}
	} else {
		rel.Strength = input.Strength
		rel.Quality = domain.Quality(input.Quality)
		if input.Label != "" {
			rel.Label = input.Label
		}
		rel.ReciprocityIndex += input.ReciprocityDelta * 0.1
		if rel.ReciprocityIndex < 0 {
			rel.ReciprocityIndex = 0
		}
		if rel.ReciprocityIndex > 1 {
			rel.ReciprocityIndex = 1
		}
		if err := d.relationshipRepo.Update(c, rel); err != nil {
			return UpdateGraphOutput{}, fmt.Errorf("update relationship: %w", err)
		}
	}

	return UpdateGraphOutput{
		Status: "updated", Source: input.SourcePerson, Target: input.TargetPerson,
		Quality: input.Quality, Strength: input.Strength,
	}, nil
}

type ClassifyRoleInput struct {
	Person   string `json:"person" jsonschema:"description=Name of the person to classify"`
	Role     string `json:"role" jsonschema:"description=The social role,enum=bridge,enum=mentor,enum=anchor,enum=catalyst,enum=observer,enum=drain"`
	Evidence string `json:"evidence" jsonschema:"description=Textual evidence supporting this classification"`
}

type ClassifyRoleOutput struct {
	Status   string `json:"status"`
	Person   string `json:"person"`
	Role     string `json:"role"`
	Evidence string `json:"evidence"`
}

func (d *deps) classifyRoleHandler(_ tool.Context, input ClassifyRoleInput) (ClassifyRoleOutput, error) {
	d.logger.Info("role classified", "person", input.Person, "role", input.Role)
	return ClassifyRoleOutput{Status: "classified", Person: input.Person, Role: input.Role, Evidence: input.Evidence}, nil
}

type SetUserPreferencesInput struct {
	PhilosophicalLens string `json:"philosophical_lens" jsonschema:"description=User outlook: humanist, pragmatic, spiritual, stoic, romantic, existential, utilitarian, conservative, scientific"`
	Preferences       string `json:"preferences,omitempty" jsonschema:"description=Specific preferences on how to deliver feedback"`
}

type SetUserPreferencesOutput struct {
	Status            string `json:"status"`
	PhilosophicalLens string `json:"philosophical_lens"`
}

func (d *deps) setUserPreferencesHandler(ctx tool.Context, input SetUserPreferencesInput) (SetUserPreferencesOutput, error) {
	userID := userIDFromCtx(ctx)
	d.logger.Info("user preferences set", "lens", input.PhilosophicalLens)

	metadata := map[string]any{
		"philosophical_lens": input.PhilosophicalLens,
		"preferences":        input.Preferences,
		"timestamp":          time.Now().UTC().Format(time.RFC3339),
	}

	embedding, err := d.embedder.Embed(context.Background(), fmt.Sprintf("philosophical lens: %s preference: %s", input.PhilosophicalLens, input.Preferences))
	if err != nil {
		return SetUserPreferencesOutput{}, fmt.Errorf("embed preferences: %w", err)
	}

	if err := d.memoryRepo.Store(context.Background(), userID, "user-preferences", embedding, metadata); err != nil {
		return SetUserPreferencesOutput{}, fmt.Errorf("store preferences: %w", err)
	}

	return SetUserPreferencesOutput{Status: "saved", PhilosophicalLens: input.PhilosophicalLens}, nil
}

func (d *deps) findOrCreatePerson(ctx context.Context, name string) (*domain.Pessoa, error) {
	people, err := d.personRepo.GetByUserID(ctx, "unknown")
	if err != nil {
		return nil, err
	}

	for _, p := range people {
		if p.DisplayName == name {
			return &p, nil
		}
		for _, alias := range p.Aliases {
			if alias == name {
				return &p, nil
			}
		}
	}

	person := &domain.Pessoa{
		ID:          fmt.Sprintf("person-%s-%d", name, time.Now().UnixNano()),
		DisplayName: name,
	}
	if err := d.personRepo.Create(ctx, person); err != nil {
		return nil, fmt.Errorf("create person: %w", err)
	}
	return person, nil
}

func CreateBubbleTrackTools(
	memoryRepo domain.MemoryRepository,
	embedder domain.Embedder,
	personRepo domain.PersonRepository,
	relationshipRepo domain.RelationshipRepository,
	logger *slog.Logger,
) ([]tool.Tool, error) {
	d := newDeps(memoryRepo, embedder, personRepo, relationshipRepo, logger)

	searchMemTool, err := functiontool.New(functiontool.Config{
		Name:        "search_memory",
		Description: "Searches past interaction memories using semantic similarity. Use this to recall details about people, interactions, or patterns from previous conversations.",
	}, d.searchMemoryHandler)
	if err != nil {
		return nil, fmt.Errorf("create search_memory tool: %w", err)
	}

	storeMemTool, err := functiontool.New(functiontool.Config{
		Name:        "store_memory",
		Description: "Stores a new interaction memory for future retrieval. Use this to persist important observations about people and interactions.",
	}, d.storeMemoryHandler)
	if err != nil {
		return nil, fmt.Errorf("create store_memory tool: %w", err)
	}

	updateGraphTool, err := functiontool.New(functiontool.Config{
		Name:        "update_graph",
		Description: "Updates the social graph with a relationship between two people. Use this when you identify or infer a relationship from the interaction text.",
	}, d.updateGraphHandler)
	if err != nil {
		return nil, fmt.Errorf("create update_graph tool: %w", err)
	}

	classifyRoleTool, err := functiontool.New(functiontool.Config{
		Name:        "classify_role",
		Description: "Classifies a person's social role with evidence from interactions. Roles are based on Humanist principles of agency, bridge-building, and mutual growth.",
	}, d.classifyRoleHandler)
	if err != nil {
		return nil, fmt.Errorf("create classify_role tool: %w", err)
	}

	setPrefsTool, err := functiontool.New(functiontool.Config{
		Name:        "set_user_preferences",
		Description: "Stores user preferences for how they want to receive analysis.",
	}, d.setUserPreferencesHandler)
	if err != nil {
		return nil, fmt.Errorf("create set_user_preferences tool: %w", err)
	}

	return []tool.Tool{searchMemTool, storeMemTool, updateGraphTool, classifyRoleTool, setPrefsTool}, nil
}

func toJSON(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}
