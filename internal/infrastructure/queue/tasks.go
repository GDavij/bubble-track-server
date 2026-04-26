package queue

import (
	"encoding/json"

	"github.com/bubbletrack/server/internal/domain"
	"github.com/hibiken/asynq"
)

const (
	TypeAnalyzeInteraction = "interaction:analyze"
)

type AnalyzePayload struct {
	InteractionID string `json:"interaction_id"`
	UserID        string `json:"user_id"`
	RawText       string `json:"raw_text"`
}

func NewAnalyzeTask(payload AnalyzePayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeAnalyzeInteraction, data), nil
}

type TaskResult struct {
	InteractionID string                `json:"interaction_id"`
	Status        domain.JobStatus      `json:"status"`
	Error         string                `json:"error,omitempty"`
	Analysis      *domain.AnalysisResult `json:"analysis,omitempty"`
}

func EncodeTaskResult(result TaskResult) ([]byte, error) {
	return json.Marshal(result)
}

func DecodeTaskResult(data []byte) (*TaskResult, error) {
	var result TaskResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
