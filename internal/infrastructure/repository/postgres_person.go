package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bubbletrack/server/internal/config"
	"github.com/bubbletrack/server/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresPersonRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresPersonRepository(pool *pgxpool.Pool) *PostgresPersonRepository {
	return &PostgresPersonRepository{pool: pool}
}

func (r *PostgresPersonRepository) Create(ctx context.Context, p *domain.Pessoa) error {
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now

	_, err := r.pool.Exec(ctx,
		`INSERT INTO people (id, display_name, aliases, social_role, current_mood, current_energy, notes, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		p.ID, p.DisplayName, p.Aliases, p.SocialRole, p.CurrentMood, p.CurrentEnergy, p.Notes, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *PostgresPersonRepository) GetByID(ctx context.Context, id string) (*domain.Pessoa, error) {
	var p domain.Pessoa
	err := r.pool.QueryRow(ctx,
		`SELECT id, display_name, aliases, social_role, current_mood, current_energy, notes, created_at, updated_at
		 FROM people WHERE id = $1`, id,
	).Scan(&p.ID, &p.DisplayName, &p.Aliases, &p.SocialRole, &p.CurrentMood, &p.CurrentEnergy, &p.Notes, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, &domain.NotFoundError{Entity: "person", ID: id}
	}
	return &p, err
}

func (r *PostgresPersonRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Pessoa, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT p.id, p.display_name, p.aliases, p.social_role, p.current_mood, p.current_energy, p.notes, p.created_at, p.updated_at
		 FROM people p
		 JOIN person_ownership po ON p.id = po.person_id
		 WHERE po.user_id = $1`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var people []domain.Pessoa
	for rows.Next() {
		var p domain.Pessoa
		if err := rows.Scan(&p.ID, &p.DisplayName, &p.Aliases, &p.SocialRole, &p.CurrentMood, &p.CurrentEnergy, &p.Notes, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		people = append(people, p)
	}
	return people, rows.Err()
}

func (r *PostgresPersonRepository) Update(ctx context.Context, p *domain.Pessoa) error {
	p.UpdatedAt = time.Now().UTC()

	tag, err := r.pool.Exec(ctx,
		`UPDATE people SET display_name = $2, aliases = $3, social_role = $4, current_mood = $5, current_energy = $6, notes = $7, updated_at = $8
		 WHERE id = $1`,
		p.ID, p.DisplayName, p.Aliases, p.SocialRole, p.CurrentMood, p.CurrentEnergy, p.Notes, p.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return &domain.NotFoundError{Entity: "person", ID: p.ID}
	}
	return nil
}

func (r *PostgresPersonRepository) Upsert(ctx context.Context, p *domain.Pessoa) error {
	now := time.Now().UTC()
	p.UpdatedAt = now
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}

	_, err := r.pool.Exec(ctx,
		`INSERT INTO people (id, display_name, aliases, social_role, current_mood, current_energy, notes, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (id) DO UPDATE SET
		   display_name = EXCLUDED.display_name,
		   aliases = EXCLUDED.aliases,
		   social_role = EXCLUDED.social_role,
		   current_mood = EXCLUDED.current_mood,
		   current_energy = EXCLUDED.current_energy,
		   notes = EXCLUDED.notes,
		   updated_at = EXCLUDED.updated_at`,
		p.ID, p.DisplayName, p.Aliases, p.SocialRole, p.CurrentMood, p.CurrentEnergy, p.Notes, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *PostgresPersonRepository) GetOrCreateByName(ctx context.Context, name string) (*domain.Pessoa, error) {
	normalizedName := normalizeName(name)

	var p domain.Pessoa
	err := r.pool.QueryRow(ctx,
		`SELECT id, display_name, aliases, social_role, current_mood, current_energy, notes, created_at, updated_at
		 FROM people 
		 WHERE LOWER(display_name) = LOWER($1) 
		    OR $1 = ANY(aliases)
		 LIMIT 1`, name,
	).Scan(&p.ID, &p.DisplayName, &p.Aliases, &p.SocialRole, &p.CurrentMood, &p.CurrentEnergy, &p.Notes, &p.CreatedAt, &p.UpdatedAt)
	if err == nil {
		return &p, nil
	}
	if err != pgx.ErrNoRows {
		p, fuzzyErr := r.fuzzyFind(ctx, normalizedName)
		if fuzzyErr == nil && p != nil {
			r.addAlias(ctx, p.ID, name)
			return p, nil
		}
		return nil, err
	}

	now := time.Now().UTC()
	aliases := []string{}
	if normalizedName == "eu" || normalizedName == "me" || normalizedName == "i" || normalizedName == "myself" {
		aliases = []string{"eu", "me", "I", "myself", "user", "you"}
	}

	p = domain.Pessoa{
		DisplayName: name,
		Aliases:     aliases,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	err = r.pool.QueryRow(ctx,
		`INSERT INTO people (display_name, aliases, notes, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		p.DisplayName, p.Aliases, p.Notes, p.CreatedAt, p.UpdatedAt,
	).Scan(&p.ID)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func normalizeName(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	// Common normalizations
	switch lower {
	case "eu", "me", "myself", "i":
		return "eu"
	case "you", "você", "tu":
		return "you"
	case "ele", "he", "him":
		return "he"
	case "ela", "she", "her":
		return "she"
	case "nós", "we", "us", "nos":
		return "we"
	}
	return lower
}

func (r *PostgresPersonRepository) fuzzyFind(ctx context.Context, normalized string) (*domain.Pessoa, error) {
	var p domain.Pessoa
	err := r.pool.QueryRow(ctx,
		`SELECT id, display_name, aliases, social_role, current_mood, current_energy, notes, created_at, updated_at
		 FROM people 
		 WHERE LOWER(display_name) LIKE $1
		 LIMIT 1`,
		"%"+normalized+"%",
	).Scan(&p.ID, &p.DisplayName, &p.Aliases, &p.SocialRole, &p.CurrentMood, &p.CurrentEnergy, &p.Notes, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func (r *PostgresPersonRepository) addAlias(ctx context.Context, personID, alias string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE people SET aliases = array_append(aliases, $2), updated_at = NOW() WHERE id = $1`,
		personID, alias,
	)
	return err
}

func (r *PostgresPersonRepository) UpdateSocialRole(ctx context.Context, personID string, role domain.SocialRole) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE people SET social_role = $2, updated_at = NOW() WHERE id = $1`,
		personID, role,
	)
	return err
}

func NewPostgresPool(ctx context.Context, cfg config.PostgresConfig) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}
	poolCfg.MaxConns = 10
	poolCfg.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}
