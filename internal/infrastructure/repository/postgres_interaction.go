package repository

import (
	"context"

	"github.com/bubbletrack/server/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresInteractionRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresInteractionRepository(pool *pgxpool.Pool) *PostgresInteractionRepository {
	return &PostgresInteractionRepository{pool: pool}
}

func (r *PostgresInteractionRepository) Create(ctx context.Context, i *domain.Interacao) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO interactions (id, user_id, raw_text, summary, people, job_id, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		i.ID, i.UserID, i.RawText, i.Summary, i.People, i.JobID, i.Status, i.CreatedAt,
	)
	return err
}

func (r *PostgresInteractionRepository) GetByID(ctx context.Context, id string) (*domain.Interacao, error) {
	var i domain.Interacao
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, raw_text, COALESCE(summary, ''), COALESCE(people, '{}'), COALESCE(job_id, ''), status, created_at
		 FROM interactions WHERE id = $1`, id,
	).Scan(&i.ID, &i.UserID, &i.RawText, &i.Summary, &i.People, &i.JobID, &i.Status, &i.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, &domain.NotFoundError{Entity: "interaction", ID: id}
	}
	return &i, err
}

func (r *PostgresInteractionRepository) UpdateStatus(ctx context.Context, id string, status domain.JobStatus) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE interactions SET status = $2 WHERE id = $1`, id, status,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return &domain.NotFoundError{Entity: "interaction", ID: id}
	}
	return nil
}
