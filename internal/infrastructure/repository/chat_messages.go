package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatMessageRepository struct {
	pool *pgxpool.Pool
}

type ChatMessage struct {
	ID        uuid.UUID `json:"id"`
	UserID    string    `json:"user_id"`
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	IsUser    bool      `json:"is_user"`
	CreatedAt time.Time `json:"created_at"`
	SessionID string    `json:"session_id"`
}

func NewChatMessageRepository(pool *pgxpool.Pool) *ChatMessageRepository {
	return &ChatMessageRepository{pool: pool}
}

func (r *ChatMessageRepository) Create(ctx context.Context, msg *ChatMessage) error {
	msg.ID = uuid.New()
	msg.CreatedAt = time.Now().UTC()

	_, err := r.pool.Exec(ctx,
		`INSERT INTO chat_messages (id, user_id, sender, content, is_user, created_at, session_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		msg.ID, msg.UserID, msg.Sender, msg.Content, msg.IsUser, msg.CreatedAt, msg.SessionID,
	)
	return err
}

func (r *ChatMessageRepository) GetByUserID(ctx context.Context, userID string, limit int) ([]ChatMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, sender, content, is_user, created_at, session_id
		 FROM chat_messages
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2`,
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var msg ChatMessage
		if err := rows.Scan(&msg.ID, &msg.UserID, &msg.Sender, &msg.Content, &msg.IsUser, &msg.CreatedAt, &msg.SessionID); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

func (r *ChatMessageRepository) GetBySessionID(ctx context.Context, sessionID string, limit int) ([]ChatMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, sender, content, is_user, created_at, session_id
		 FROM chat_messages
		 WHERE session_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2`,
		sessionID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var msg ChatMessage
		if err := rows.Scan(&msg.ID, &msg.UserID, &msg.Sender, &msg.Content, &msg.IsUser, &msg.CreatedAt, &msg.SessionID); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}
