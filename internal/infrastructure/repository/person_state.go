package repository

import (
	"context"

	"github.com/bubbletrack/server/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PersonStateRepository struct {
	pool *pgxpool.Pool
}

func NewPersonStateRepository(pool *pgxpool.Pool) *PersonStateRepository {
	return &PersonStateRepository{pool: pool}
}

func (r *PersonStateRepository) Create(ctx context.Context, state *domain.PersonState) error {
	state.ID = uuid.New().String()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO person_states (id, user_id, person_id, mood, energy, valence, context, trigger, interaction_id, notes, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())`,
		state.ID, state.UserID, nullableUUID(state.PersonID), string(state.Mood),
		state.Energy, state.Valence, state.Context, state.Trigger,
		nullableUUID(state.InteractionID), state.Notes,
	)
	return err
}

func (r *PersonStateRepository) GetByPerson(ctx context.Context, userID, personID string, limit int) ([]domain.PersonState, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, COALESCE(person_id::text, ''), mood, energy, valence, context, trigger, COALESCE(interaction_id::text, ''), notes, created_at
		 FROM person_states
		 WHERE user_id = $1 AND person_id = $2::uuid
		 ORDER BY created_at DESC
		 LIMIT $3`, userID, personID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanStates(rows)
}

func (r *PersonStateRepository) GetByUser(ctx context.Context, userID string, limit int) ([]domain.PersonState, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, COALESCE(person_id::text, ''), mood, energy, valence, context, trigger, COALESCE(interaction_id::text, ''), notes, created_at
		 FROM person_states
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2`, userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanStates(rows)
}

func (r *PersonStateRepository) GetTimeline(ctx context.Context, userID string, limit int) ([]domain.PersonState, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, COALESCE(person_id::text, ''), mood, energy, valence, context, trigger, COALESCE(interaction_id::text, ''), notes, created_at
		 FROM person_states
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2`, userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanStates(rows)
}

func (r *PersonStateRepository) GetSelfStates(ctx context.Context, userID string, limit int) ([]domain.PersonState, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, COALESCE(person_id::text, ''), mood, energy, valence, context, trigger, COALESCE(interaction_id::text, ''), notes, created_at
		 FROM person_states
		 WHERE user_id = $1 AND person_id IS NULL
		 ORDER BY created_at DESC
		 LIMIT $2`, userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanStates(rows)
}

func scanStates(rows pgx.Rows) ([]domain.PersonState, error) {
	var states []domain.PersonState
	for rows.Next() {
		var s domain.PersonState
		var personID, interactionID string
		if err := rows.Scan(&s.ID, &s.UserID, &personID, &s.Mood, &s.Energy, &s.Valence,
			&s.Context, &s.Trigger, &interactionID, &s.Notes, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.PersonID = personID
		s.InteractionID = interactionID
		states = append(states, s)
	}
	return states, rows.Err()
}

func nullableUUID(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
