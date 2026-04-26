package repository

import (
	"context"
	"time"

	"github.com/bubbletrack/server/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) Create(ctx context.Context, a *domain.Account) error {
	a.ID = uuid.New().String()
	a.CreatedAt = time.Now().UTC()
	a.UpdatedAt = a.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, display_name, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		a.ID, a.Email, a.PasswordHash, a.DisplayName, a.CreatedAt, a.UpdatedAt,
	)
	return err
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	var a domain.Account
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, display_name, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&a.ID, &a.Email, &a.PasswordHash, &a.DisplayName, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.Account, error) {
	var a domain.Account
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, display_name, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(&a.ID, &a.Email, &a.PasswordHash, &a.DisplayName, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *PostgresUserRepository) UpdatePassword(ctx context.Context, id string, hash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`,
		hash, time.Now().UTC(), id,
	)
	return err
}
