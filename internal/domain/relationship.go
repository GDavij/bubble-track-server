package domain

import (
	"time"
)

// Relacionamento represents an edge between two people in the social graph.
// It captures the quality and direction of a relationship.
type Relacionamento struct {
	ID               string    `json:"id"`
	SourcePersonID   string    `json:"source_person_id"`
	TargetPersonID   string    `json:"target_person_id"`
	Quality          Quality   `json:"quality"`
	Strength         float64   `json:"strength"`
	SourceWeight     float64   `json:"source_weight"`
	TargetWeight     float64   `json:"target_weight"`
	Protocol         Protocol  `json:"protocol,omitempty"`
	Label            string    `json:"label,omitempty"`
	ReciprocityIndex float64   `json:"reciprocity_index"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Quality captures the emotional character of a relationship
// under the Humanist lens — agency, mutual respect, authenticity.
type Quality string

const (
	QualityNourishing Quality = "nourishing" // Mutually supportive, energizing
	QualityNeutral    Quality = "neutral"    // Neither positive nor negative
	QualityDraining   Quality = "draining"   // One-sided, energy-depleting
	QualityConflicted Quality = "conflicted" // Complex mix of positive and negative
	QualityUnknown    Quality = "unknown"    // Not yet assessed
)
