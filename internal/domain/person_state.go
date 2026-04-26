package domain

import "time"

type Mood string

const (
	MoodHappy     Mood = "happy"
	MoodAnxious   Mood = "anxious"
	MoodTired     Mood = "tired"
	MoodEnergized Mood = "energized"
	MoodSad       Mood = "sad"
	MoodNeutral   Mood = "neutral"
	MoodAngry     Mood = "angry"
	MoodHopeful   Mood = "hopeful"
	MoodLonely    Mood = "lonely"
	MoodGrateful  Mood = "grateful"
)

type Protocol string

const (
	ProtocolDeep         Protocol = "deep"
	ProtocolCasual       Protocol = "casual"
	ProtocolProfessional Protocol = "professional"
	ProtocolDigital      Protocol = "digital"
	ProtocolMixed        Protocol = "mixed"
)

type PersonState struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	PersonID      string    `json:"person_id,omitempty"`
	Mood          Mood      `json:"mood"`
	Energy        float64   `json:"energy"`
	Valence       float64   `json:"valence"`
	Context       string    `json:"context,omitempty"`
	Trigger       string    `json:"trigger,omitempty"`
	InteractionID string    `json:"interaction_id,omitempty"`
	Notes         string    `json:"notes,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
