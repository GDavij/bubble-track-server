package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/bubbletrack/server/internal/config"
	"github.com/bubbletrack/server/internal/domain"
	"github.com/redis/go-redis/v9"
)

const channelPrefix = "bubble:analysis:"

type RedisNotifier struct {
	client *redis.Client
	logger *slog.Logger
}

func NewRedisNotifier(cfg config.RedisConfig, logger *slog.Logger) *RedisNotifier {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return &RedisNotifier{client: client, logger: logger}
}

func (n *RedisNotifier) Notify(ctx context.Context, userID string, result *domain.AnalysisResult) error {
	channel := channelPrefix + userID
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}

	if err := n.client.Publish(ctx, channel, data).Err(); err != nil {
		return fmt.Errorf("publish to redis: %w", err)
	}

	n.logger.Info("notification published", "channel", channel, "interaction_id", result.InteractionID)
	return nil
}

func (n *RedisNotifier) Subscribe(ctx context.Context, userID string) <-chan *domain.AnalysisResult {
	ch := make(chan *domain.AnalysisResult, 10)
	channel := channelPrefix + userID

	sub := n.client.Subscribe(ctx, channel)
	_, err := sub.Receive(ctx)
	if err != nil {
		n.logger.Error("subscribe failed", "channel", channel, "error", err)
		close(ch)
		return ch
	}

	go func() {
		defer close(ch)
		defer sub.Close()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-sub.Channel():
				if !ok {
					return
				}
				var result domain.AnalysisResult
				if err := json.Unmarshal([]byte(msg.Payload), &result); err != nil {
					n.logger.Error("unmarshal notification", "error", err)
					continue
				}
				ch <- &result
			}
		}
	}()

	return ch
}

func (n *RedisNotifier) Close() error {
	return n.client.Close()
}
