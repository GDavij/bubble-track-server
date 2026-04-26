package domain

import "time"

type AttachmentWound struct {
	ID           string        `json:"id"`
	PersonID    string        `json:"person_id"`
	Type        WoundType    `json:"type"`
	Source      string        `json:"source"`
	Year        int           `json:"year,omitempty"`
	Description string      `json:"description"`
	Severity    WoundSeverity `json:"severity"`
	Processed  bool          `json:"processed"`
	CreatedAt  time.Time   `json:"created_at"`
}

type WoundType string

const (
	WoundAbandonment  WoundType = "abandonment"
	WoundBetrayal    WoundType = "betrayal"
	WoundRejection   WoundType = "rejection"
	WoundNeglect    WoundType = "neglect"
	WoundManipulation WoundType = "manipulation"
	WoundViolence   WoundType = "violence"
)

type WoundSeverity string

const (
	SeverityLow     WoundSeverity = "low"
	SeverityMedium WoundSeverity = "medium"
	SeverityHigh  WoundSeverity = "high"
	SeverityCricket WoundSeverity = "critical"
)

type GhostingPattern struct {
	ID             string    `json:"id"`
	PersonID      string    `json:"person_id"`
	TargetID      string    `json:"target_id"`
	Frequency     int       `json:"frequency"`
	LastInstance time.Time  `json:"last_instance"`
	Triggers     []string  `json:"triggers"`
	PerceivedReason string `json:"perceived_reason"`
	CreatedAt    time.Time `json:"created_at"`
}

type DecisionContext struct {
	ID             string        `json:"id"`
	PersonID      string        `json:"person_id"`
	SessionID    string        `json:"session_id"`
	Situation    string        `json:"situation"`
	Options      []string      `json:"options"`
	Chosen       string        `json:"chosen"`
	Reasoning    string        `json:"reasoning"`
	EmotionalState EmotionState `json:"emotional_state"`
	Confidence   float64      `json:"confidence"`
	FearFactor   float64      `json:"fear_factor"`
	Regret       bool        `json:"regret"`
	CreatedAt    time.Time    `json:"created_at"`
}

type EmotionState struct {
	Primary   string  `json:"primary"`
	Intensity float64 `json:"intensity"`
	Valence  float64 `json:"valence"`
	Arousal  float64 `json:"arousal"`
}

type BetrayalDetection struct {
	ID            string    `json:"id"`
	PersonID     string    `json:"person_id"`
	TargetID    string    `json:"target_id"`
	Type        string    `json:"type"`
	Evidence    string    `json:"evidence"`
	Impact      float64   `json:"impact"`
	Forgiven   bool      `json:"forgiven"`
	LastBreachAt time.Time `json:"last_breach_at"`
	BreachCount int       `json:"breach_count"`
	CreatedAt  time.Time `json:"created_at"`
}

type InterpersonalDrama struct {
	ID            string    `json:"id"`
	SessionID   string    `json:"session_id"`
	PersonID   string    `json:"person_id"`
	OtherID    string    `json:"other_id"`
	Type       DramaType `json:"type"`
	Trigger    string    `json:"trigger"`
	Escalation string    `json:"escalation"`
	Resolution string   `json:"resolution"`
	EmotionalCost float64 `json:"emotional_cost"`
	CreatedAt  time.Time `json:"created_at"`
}

type DramaType string

const (
	DramaArgument      DramaType = "argument"
	DramaColdShoulder DramaType = "cold_shoulder"
	DramaPower        DramaType = "power_struggle"
	DramaBoundary     DramaType = "boundary_violation"
	DramaExpectation  DramaType = "expectation_mismatch"
	DramaSilence     DramaType = "silence_treatment"
)