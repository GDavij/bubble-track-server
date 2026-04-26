package domain

import "time"

// Interacao represents a free-text report of a social interaction,
// submitted by the user for AI analysis.
type Interacao struct {
	ID       string    `json:"id"`
	UserID  string    `json:"user_id"`
	RawText string    `json:"raw_text"`
	Summary string   `json:"summary,omitempty"`
	People []string  `json:"people,omitempty"`
	JobID  string    `json:"job_id,omitempty"`
	Segment string   `json:"segment,omitempty"`
	Session string   `json:"session,omitempty"`
	Status JobStatus `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// JobStatus tracks the async processing lifecycle of an interaction.
type JobStatus string

const (
	StatusQueued    JobStatus = "queued"
	StatusProcessing JobStatus = "processing"
	StatusCompleted JobStatus = "completed"
	StatusFailed    JobStatus = "failed"
)

// AnalysisResult is the structured output from the AI agent
// after processing an interaction.
type AnalysisResult struct {
	InteractionID    string               `json:"interaction_id"`
	PeopleExtracted  []ExtractedPerson    `json:"people_extracted"`
	Relationships    []RelationshipUpdate `json:"relationships"`
	SocialRoles      map[string]SocialRole `json:"social_roles"`
	StatesExtracted  []PersonState        `json:"states_extracted"`
	Protocols        map[string]Protocol  `json:"protocols,omitempty"`
	Summary          string               `json:"summary"`
	ReciprocityNotes map[string]float64   `json:"reciprocity_notes"`
}

// ExtractedPerson is a person identified by the AI from free text.
type ExtractedPerson struct {
	Name        string   `json:"name"`
	ExistingID  string   `json:"existing_id,omitempty"` // If matched to existing person
	Aliases     []string `json:"aliases,omitempty"`
	Context     string   `json:"context"` // How they appeared in the text
}

// RelationshipUpdate captures changes the AI wants to make to the graph.
type RelationshipUpdate struct {
	SourcePersonID string  `json:"source_person_id"`
	TargetPersonID string  `json:"target_person_id"`
	Quality        Quality `json:"quality"`
	Strength       float64 `json:"strength"`
	Label          string  `json:"label,omitempty"`
	ReciprocityDelta float64 `json:"reciprocity_delta"`
}
