package domain

import (
	"time"
)

// Pessoa represents a person in the social graph.
// Each person is a node tracked across interactions.
type Pessoa struct {
	ID            string    `json:"id"`
	DisplayName   string    `json:"display_name"`
	Aliases       []string  `json:"aliases,omitempty"`
	SocialRole    SocialRole `json:"social_role,omitempty"`
	CurrentMood   Mood      `json:"current_mood,omitempty"`
	CurrentEnergy float64   `json:"current_energy"`
	Notes         string    `json:"notes,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// SocialRole represents the qualitative role a person plays in the social graph.
// This is a Humanist classification — it focuses on agency, bridge-building,
// and emotional quality rather than deterministic categories.
type SocialRole string

const (
	RoleBridge    SocialRole = "bridge"    // Connects disparate groups
	RoleMentor    SocialRole = "mentor"    // Actively invests in others' growth
	RoleAnchor    SocialRole = "anchor"    // Provides stability and continuity
	RoleCatalyst  SocialRole = "catalyst"  // Sparks new connections or ideas
	RoleObserver  SocialRole = "observer"  // Present but not actively engaging
	RoleDrain     SocialRole = "drain"     // Consistently consumes more than gives
	RoleUnknown   SocialRole = "unknown"   // Not yet classified
)
