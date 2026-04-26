package domain

import "time"

type UserSession struct {
	ID          string    `json:"id"`
	UserID     string    `json:"user_id"`
	Name       string    `json:"name"`
	Topic      string    `json:"topic"`
	Status     SessionStatus `json:"status"`
	StartedAt  time.Time   `json:"started_at"`
	EndedAt    *time.Time `json:"ended_at,omitempty"`
	EmotionArc []EmotionState `json:"emotion_arc"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type SessionStatus string

const (
	SessionActive   SessionStatus = "active"
	SessionPaused  SessionStatus = "paused"
	SessionComplete SessionStatus = "completed"
	SessionAbandoned SessionStatus = "abandoned"
)

type SessionInsight struct {
	ID          string    `json:"id"`
	SessionID  string    `json:"session_id"`
	Type       InsightType `json:"type"`
	Content    string     `json:"content"`
	TriggeredBy string    `json:"triggered_by"`
	CreatedAt  time.Time `json:"created_at"`
}

type InsightType string

const (
	InsightPattern    InsightType = "pattern"
	InsightTrigger   InsightType = "trigger"
	InsightDecision  InsightType = "decision"
	InsightConflict  InsightType = "conflict"
	InsightGrowth    InsightType = "growth"
	InsightWarning    InsightType = "warning"
	InsightBreakthrough InsightType = "breakthrough"
)

func NewUserSession(userID, name, topic string) *UserSession {
	now := time.Now().UTC()
	return &UserSession{
		ID:         generateID(),
		UserID:    userID,
		Name:      name,
		Topic:     topic,
		Status:    SessionActive,
		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func generateID() string {
	return "sess_" + time.Now().Format("20060102150405")
}