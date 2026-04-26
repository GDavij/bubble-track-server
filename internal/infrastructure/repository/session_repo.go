package repository

import (
	"context"
	"time"

	"github.com/bubbletrack/server/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresSessionRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresSessionRepository(pool *pgxpool.Pool) *PostgresSessionRepository {
	return &PostgresSessionRepository{pool: pool}
}

func (r *PostgresSessionRepository) Create(ctx context.Context, s *domain.UserSession) error {
	s.CreatedAt = time.Now()
	s.UpdatedAt = s.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_sessions (id, user_id, name, topic, status, started_at, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		s.ID, s.UserID, s.Name, s.Topic, s.Status, s.StartedAt, s.CreatedAt, s.UpdatedAt,
	)
	return err
}

func (r *PostgresSessionRepository) GetByID(ctx context.Context, id string) (*domain.UserSession, error) {
	var s domain.UserSession
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, name, topic, status, started_at, ended_at, created_at, updated_at
		 FROM user_sessions WHERE id = $1`, id,
	).Scan(&s.ID, &s.UserID, &s.Name, &s.Topic, &s.Status, &s.StartedAt, &s.EndedAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *PostgresSessionRepository) GetActive(ctx context.Context, userID string) (*domain.UserSession, error) {
	var s domain.UserSession
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, name, topic, status, started_at, ended_at, created_at, updated_at
		 FROM user_sessions WHERE user_id = $1 AND status = 'active' ORDER BY started_at DESC LIMIT 1`, userID,
	).Scan(&s.ID, &s.UserID, &s.Name, &s.Topic, &s.Status, &s.StartedAt, &s.EndedAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *PostgresSessionRepository) UpdateStatus(ctx context.Context, id string, status domain.SessionStatus) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE user_sessions SET status = $2, updated_at = NOW(), ended_at = CASE WHEN $2 = 'completed' THEN NOW() ELSE ended_at END
		 WHERE id = $1`, id, status,
	)
	return err
}

func (r *PostgresSessionRepository) AddInsight(ctx context.Context, s *domain.SessionInsight) error {
	s.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO session_insights (id, session_id, type, content, triggered_by, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		s.ID, s.SessionID, s.Type, s.Content, s.TriggeredBy, s.CreatedAt,
	)
	return err
}

func (r *PostgresSessionRepository) GetInsights(ctx context.Context, sessionID string) ([]domain.SessionInsight, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, session_id, type, content, triggered_by, created_at
		 FROM session_insights WHERE session_id = $1 ORDER BY created_at DESC`, sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var insights []domain.SessionInsight
	for rows.Next() {
		var s domain.SessionInsight
		if err := rows.Scan(&s.ID, &s.SessionID, &s.Type, &s.Content, &s.TriggeredBy, &s.CreatedAt); err != nil {
			return nil, err
		}
		insights = append(insights, s)
	}
	return insights, rows.Err()
}