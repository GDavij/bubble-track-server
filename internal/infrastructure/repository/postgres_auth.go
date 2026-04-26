package repository

import (
	"context"
	"time"

	"github.com/bubbletrack/server/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRefreshTokenRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRefreshTokenRepository(pool *pgxpool.Pool) *PostgresRefreshTokenRepository {
	return &PostgresRefreshTokenRepository{pool: pool}
}

func (r *PostgresRefreshTokenRepository) Store(ctx context.Context, t *domain.RefreshToken) error {
	t.ID = uuid.New().String()
	t.CreatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (id, user_id, token, expires_at, revoked, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		t.ID, t.UserID, t.Token, t.ExpiresAt, false, t.CreatedAt,
	)
	return err
}

func (r *PostgresRefreshTokenRepository) GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	var t domain.RefreshToken
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, token, expires_at, revoked, created_at
		 FROM refresh_tokens WHERE token = $1`, token,
	).Scan(&t.ID, &t.UserID, &t.Token, &t.ExpiresAt, &t.Revoked, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *PostgresRefreshTokenRepository) Revoke(ctx context.Context, tokenID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked = true WHERE id = $1`, tokenID,
	)
	return err
}

func (r *PostgresRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`, userID,
	)
	return err
}
