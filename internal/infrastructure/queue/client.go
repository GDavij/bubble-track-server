package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/bubbletrack/server/internal/config"
	"github.com/hibiken/asynq"
)

func NewClient(cfg config.RedisConfig) (*asynq.Client, error) {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Address(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return client, nil
}

func NewServer(cfg config.RedisConfig, logger *slog.Logger) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.Address(),
			Password: cfg.Password,
			DB:       cfg.DB,
		},
		asynq.Config{
			Concurrency: 5,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				retried, _ := asynq.GetRetryCount(ctx)
				maxRetry, _ := asynq.GetMaxRetry(ctx)
				logger.Error("task failed",
					"type", task.Type(),
					"error", err,
					"retry", fmt.Sprintf("%d/%d", retried, maxRetry),
				)
			}),
			Logger:   nil,
			LogLevel: asynq.InfoLevel,
			ShutdownTimeout: 30 * time.Second,
		},
	)
}

type Enqueuer struct {
	client *asynq.Client
}

func NewEnqueuer(client *asynq.Client) *Enqueuer {
	return &Enqueuer{client: client}
}

func (e *Enqueuer) EnqueueAnalyze(ctx context.Context, payload AnalyzePayload) (string, error) {
	task, err := NewAnalyzeTask(payload)
	if err != nil {
		return "", fmt.Errorf("create task: %w", err)
	}

	info, err := e.client.Enqueue(
		task,
		asynq.Queue("default"),
		asynq.MaxRetry(3),
		asynq.Timeout(5*time.Minute),
		asynq.Retention(24*time.Hour),
	)
	if err != nil {
		return "", fmt.Errorf("enqueue task: %w", err)
	}

	return info.ID, nil
}

func DecodeAnalyzePayload(t *asynq.Task) (*AnalyzePayload, error) {
	var payload AnalyzePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return nil, fmt.Errorf("unmarshal payload: %w", err)
	}
	return &payload, nil
}
