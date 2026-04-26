package repository

import (
	"context"
	"time"

	"github.com/bubbletrack/server/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresWoundRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresWoundRepository(pool *pgxpool.Pool) *PostgresWoundRepository {
	return &PostgresWoundRepository{pool: pool}
}

func (r *PostgresWoundRepository) Create(ctx context.Context, w *domain.AttachmentWound) error {
	w.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO attachment_wounds (id, person_id, type, source, year, description, severity, processed, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		w.ID, w.PersonID, w.Type, w.Source, w.Year, w.Description, w.Severity, w.Processed, w.CreatedAt,
	)
	return err
}

func (r *PostgresWoundRepository) GetByPerson(ctx context.Context, personID string) ([]domain.AttachmentWound, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, person_id, type, source, year, description, severity, processed, created_at
		 FROM attachment_wounds WHERE person_id = $1 ORDER BY created_at DESC`, personID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wounds []domain.AttachmentWound
	for rows.Next() {
		var w domain.AttachmentWound
		if err := rows.Scan(&w.ID, &w.PersonID, &w.Type, &w.Source, &w.Year, &w.Description, &w.Severity, &w.Processed, &w.CreatedAt); err != nil {
			return nil, err
		}
		wounds = append(wounds, w)
	}
	return wounds, rows.Err()
}

func (r *PostgresWoundRepository) UpdateProcessed(ctx context.Context, woundID string, processed bool) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE attachment_wounds SET processed = $2 WHERE id = $1`, woundID, processed,
	)
	return err
}

type PostgresGhostingRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresGhostingRepository(pool *pgxpool.Pool) *PostgresGhostingRepository {
	return &PostgresGhostingRepository{pool: pool}
}

func (r *PostgresGhostingRepository) Create(ctx context.Context, g *domain.GhostingPattern) error {
	g.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO ghosting_patterns (id, person_id, target_id, frequency, last_instance, triggers, perceived_reason, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		g.ID, g.PersonID, g.TargetID, g.Frequency, g.LastInstance, g.Triggers, g.PerceivedReason, g.CreatedAt,
	)
	return err
}

func (r *PostgresGhostingRepository) GetByPerson(ctx context.Context, personID string) ([]domain.GhostingPattern, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, person_id, target_id, frequency, last_instance, triggers, perceived_reason, created_at
		 FROM ghosting_patterns WHERE person_id = $1`, personID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patterns []domain.GhostingPattern
	for rows.Next() {
		var g domain.GhostingPattern
		if err := rows.Scan(&g.ID, &g.PersonID, &g.TargetID, &g.Frequency, &g.LastInstance, &g.Triggers, &g.PerceivedReason, &g.CreatedAt); err != nil {
			return nil, err
		}
		patterns = append(patterns, g)
	}
	return patterns, rows.Err()
}

func (r *PostgresGhostingRepository) UpdateFrequency(ctx context.Context, personID, targetID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE ghosting_patterns SET frequency = frequency + 1, last_instance = NOW() 
		 WHERE person_id = $1 AND target_id = $2`, personID, targetID,
	)
	return err
}

type PostgresDecisionRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresDecisionRepository(pool *pgxpool.Pool) *PostgresDecisionRepository {
	return &PostgresDecisionRepository{pool: pool}
}

func (r *PostgresDecisionRepository) Create(ctx context.Context, d *domain.DecisionContext) error {
	d.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO decisions (id, person_id, session_id, situation, options, chosen, reasoning, confidence, fear_factor, regret, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		d.ID, d.PersonID, d.SessionID, d.Situation, d.Options, d.Chosen, d.Reasoning, d.Confidence, d.FearFactor, d.Regret, d.CreatedAt,
	)
	return err
}

func (r *PostgresDecisionRepository) GetBySession(ctx context.Context, sessionID string) ([]domain.DecisionContext, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, person_id, session_id, situation, options, chosen, reasoning, confidence, fear_factor, regret, created_at
		 FROM decisions WHERE session_id = $1 ORDER BY created_at DESC`, sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decisions []domain.DecisionContext
	for rows.Next() {
		var d domain.DecisionContext
		if err := rows.Scan(&d.ID, &d.PersonID, &d.SessionID, &d.Situation, &d.Options, &d.Chosen, &d.Reasoning, &d.Confidence, &d.FearFactor, &d.Regret, &d.CreatedAt); err != nil {
			return nil, err
		}
		decisions = append(decisions, d)
	}
	return decisions, rows.Err()
}

func (r *PostgresDecisionRepository) UpdateRegret(ctx context.Context, decisionID string, regret bool) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE decisions SET regret = $2 WHERE id = $1`, decisionID, regret,
	)
	return err
}

type PostgresDramaRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresDramaRepository(pool *pgxpool.Pool) *PostgresDramaRepository {
	return &PostgresDramaRepository{pool: pool}
}

func (r *PostgresDramaRepository) Create(ctx context.Context, d *domain.InterpersonalDrama) error {
	d.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO dramas (id, session_id, person_id, other_id, type, trigger, escalation, resolution, emotional_cost, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		d.ID, d.SessionID, d.PersonID, d.OtherID, d.Type, d.Trigger, d.Escalation, d.Resolution, d.EmotionalCost, d.CreatedAt,
	)
	return err
}

func (r *PostgresDramaRepository) GetBySession(ctx context.Context, sessionID string) ([]domain.InterpersonalDrama, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, session_id, person_id, other_id, type, trigger, escalation, resolution, emotional_cost, created_at
		 FROM dramas WHERE session_id = $1 ORDER BY created_at DESC`, sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dramas []domain.InterpersonalDrama
	for rows.Next() {
		var d domain.InterpersonalDrama
		if err := rows.Scan(&d.ID, &d.SessionID, &d.PersonID, &d.OtherID, &d.Type, &d.Trigger, &d.Escalation, &d.Resolution, &d.EmotionalCost, &d.CreatedAt); err != nil {
			return nil, err
		}
		dramas = append(dramas, d)
	}
	return dramas, rows.Err()
}

func (r *PostgresDramaRepository) UpdateResolution(ctx context.Context, dramaID, resolution string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE dramas SET resolution = $2 WHERE id = $1`, dramaID, resolution,
	)
	return err
}