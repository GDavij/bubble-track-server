package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/bubbletrack/server/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRelationshipRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRelationshipRepository(pool *pgxpool.Pool) *PostgresRelationshipRepository {
	return &PostgresRelationshipRepository{pool: pool}
}

func (r *PostgresRelationshipRepository) Create(ctx context.Context, rel *domain.Relacionamento) error {
	now := time.Now().UTC()
	rel.CreatedAt = now
	rel.UpdatedAt = now

	_, err := r.pool.Exec(ctx,
		`INSERT INTO relationships (id, source_person_id, target_person_id, quality, strength, source_weight, target_weight, protocol, label, reciprocity_index, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		rel.ID, rel.SourcePersonID, rel.TargetPersonID, rel.Quality,
		rel.Strength, rel.SourceWeight, rel.TargetWeight, rel.Protocol,
		rel.Label, rel.ReciprocityIndex, rel.CreatedAt, rel.UpdatedAt,
	)
	return err
}

func (r *PostgresRelationshipRepository) GetByID(ctx context.Context, id string) (*domain.Relacionamento, error) {
	var rel domain.Relacionamento
	err := r.pool.QueryRow(ctx,
		`SELECT id, source_person_id, target_person_id, quality, strength, source_weight, target_weight, protocol, label, reciprocity_index, created_at, updated_at
		 FROM relationships WHERE id = $1`, id,
	).Scan(&rel.ID, &rel.SourcePersonID, &rel.TargetPersonID, &rel.Quality,
		&rel.Strength, &rel.SourceWeight, &rel.TargetWeight, &rel.Protocol,
		&rel.Label, &rel.ReciprocityIndex, &rel.CreatedAt, &rel.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, &domain.NotFoundError{Entity: "relationship", ID: id}
	}
	return &rel, err
}

func (r *PostgresRelationshipRepository) GetByPeople(ctx context.Context, sourceID, targetID string) (*domain.Relacionamento, error) {
	var rel domain.Relacionamento
	err := r.pool.QueryRow(ctx,
		`SELECT id, source_person_id, target_person_id, quality, strength, source_weight, target_weight, protocol, label, reciprocity_index, created_at, updated_at
		 FROM relationships
		 WHERE (source_person_id = $1 AND target_person_id = $2)
		    OR (source_person_id = $2 AND target_person_id = $1)`,
		sourceID, targetID,
	).Scan(&rel.ID, &rel.SourcePersonID, &rel.TargetPersonID, &rel.Quality,
		&rel.Strength, &rel.SourceWeight, &rel.TargetWeight, &rel.Protocol,
		&rel.Label, &rel.ReciprocityIndex, &rel.CreatedAt, &rel.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, &domain.NotFoundError{Entity: "relationship", ID: fmt.Sprintf("%s-%s", sourceID, targetID)}
	}
	return &rel, err
}

func (r *PostgresRelationshipRepository) GetByUser(ctx context.Context, userID string) ([]domain.Relacionamento, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT r.id, r.source_person_id, r.target_person_id, r.quality, r.strength, r.source_weight, r.target_weight, r.protocol, r.label, r.reciprocity_index, r.created_at, r.updated_at
		 FROM relationships r
		 JOIN person_ownership po1 ON r.source_person_id = po1.person_id
		 JOIN person_ownership po2 ON r.target_person_id = po2.person_id
		 WHERE po1.user_id = $1 OR po2.user_id = $1`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rels []domain.Relacionamento
	for rows.Next() {
		var rel domain.Relacionamento
		if err := rows.Scan(&rel.ID, &rel.SourcePersonID, &rel.TargetPersonID, &rel.Quality,
			&rel.Strength, &rel.SourceWeight, &rel.TargetWeight, &rel.Protocol,
			&rel.Label, &rel.ReciprocityIndex, &rel.CreatedAt, &rel.UpdatedAt); err != nil {
			return nil, err
		}
		rels = append(rels, rel)
	}
	return rels, rows.Err()
}

func (r *PostgresRelationshipRepository) Update(ctx context.Context, rel *domain.Relacionamento) error {
	rel.UpdatedAt = time.Now().UTC()

	tag, err := r.pool.Exec(ctx,
		`UPDATE relationships SET quality = $2, strength = $3, source_weight = $4, target_weight = $5, protocol = $6, label = $7, reciprocity_index = $8, updated_at = $9
		 WHERE id = $1`,
		rel.ID, rel.Quality, rel.Strength, rel.SourceWeight, rel.TargetWeight, rel.Protocol, rel.Label, rel.ReciprocityIndex, rel.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return &domain.NotFoundError{Entity: "relationship", ID: rel.ID}
	}
	return nil
}

func (r *PostgresRelationshipRepository) Upsert(ctx context.Context, rel *domain.Relacionamento) error {
	now := time.Now().UTC()
	rel.UpdatedAt = now
	if rel.CreatedAt.IsZero() {
		rel.CreatedAt = now
	}

	_, err := r.pool.Exec(ctx,
		`INSERT INTO relationships (id, source_person_id, target_person_id, quality, strength, source_weight, target_weight, protocol, label, reciprocity_index, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 ON CONFLICT (source_person_id, target_person_id) DO UPDATE SET
		   quality = EXCLUDED.quality,
		   strength = EXCLUDED.strength,
		   source_weight = EXCLUDED.source_weight,
		   target_weight = EXCLUDED.target_weight,
		   protocol = EXCLUDED.protocol,
		   label = EXCLUDED.label,
		   reciprocity_index = EXCLUDED.reciprocity_index,
		   updated_at = EXCLUDED.updated_at`,
		rel.ID, rel.SourcePersonID, rel.TargetPersonID, rel.Quality,
		rel.Strength, rel.SourceWeight, rel.TargetWeight, rel.Protocol,
		rel.Label, rel.ReciprocityIndex, rel.CreatedAt, rel.UpdatedAt,
	)
	return err
}
