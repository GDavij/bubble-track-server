package repository

import (
	"context"
	"fmt"

	"github.com/bubbletrack/server/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresGraphRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresGraphRepository(pool *pgxpool.Pool) *PostgresGraphRepository {
	return &PostgresGraphRepository{pool: pool}
}

func (r *PostgresGraphRepository) GetGraph(ctx context.Context, userID string) (*domain.Graph, error) {
	people, err := r.getNodes(ctx, userID)
	if err != nil {
		return nil, err
	}

	edges, err := r.getEdges(ctx, userID)
	if err != nil {
		return nil, err
	}

	stats := r.computeStats(people, edges)

	return &domain.Graph{
		UserID: userID,
		Nodes:  people,
		Edges:  edges,
		Stats:  stats,
	}, nil
}

func (r *PostgresGraphRepository) getNodes(ctx context.Context, userID string) ([]domain.GraphNode, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT ON (LOWER(p.display_name)) 
			p.id, p.display_name, p.social_role, p.current_mood, p.current_energy,
			(SELECT COUNT(*) FROM interactions i 
			 JOIN people ep ON i.raw_text ILIKE '%' || ep.display_name || '%'
			 WHERE ep.display_name = p.display_name) as ic
		 FROM people p
		 WHERE p.display_name NOT IN ('user', 'eu', 'me', 'you', 'I', 'myself')
		 ORDER BY LOWER(p.display_name), p.created_at DESC
		 LIMIT 20`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []domain.GraphNode
	for rows.Next() {
		var n domain.GraphNode
		if err := rows.Scan(&n.ID, &n.DisplayName, &n.SocialRole, &n.CurrentMood, &n.CurrentEnergy, &n.InteractionCount); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	return nodes, rows.Err()
}

func (r *PostgresGraphRepository) getEdges(ctx context.Context, userID string) ([]domain.GraphEdge, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT 
			COALESCE(src.display_name, '?') as src,
			COALESCE(tgt.display_name, '?') as tgt,
			r.quality, r.strength, r.source_weight, r.target_weight, r.protocol, COALESCE(r.label, ''), r.reciprocity_index
		 FROM relationships r
		 LEFT JOIN people src ON r.source_person_id = src.id
		 LEFT JOIN people tgt ON r.target_person_id = tgt.id
		 ORDER BY r.strength DESC
		 LIMIT 100`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []domain.GraphEdge
	for rows.Next() {
		var e domain.GraphEdge
		if err := rows.Scan(&e.Source, &e.Target, &e.Quality, &e.Strength, &e.SourceWeight, &e.TargetWeight, &e.Protocol, &e.Label, &e.ReciprocityIndex); err != nil {
			return nil, err
		}
		edges = append(edges, e)
	}
	return edges, rows.Err()
}

func (r *PostgresGraphRepository) computeStats(nodes []domain.GraphNode, edges []domain.GraphEdge) domain.GraphStats {
	stats := domain.GraphStats{
		TotalPeople:        len(nodes),
		TotalRelationships: len(edges),
	}

	if len(edges) > 0 {
		totalReciprocity := 0.0
		maxStrength := 0.0
		var strongestSource, strongestTarget string

		for _, e := range edges {
			totalReciprocity += e.ReciprocityIndex
			if e.Strength > maxStrength {
				maxStrength = e.Strength
				strongestSource = e.Source
				strongestTarget = e.Target
			}
		}

		stats.AvgReciprocity = totalReciprocity / float64(len(edges))
		if maxStrength > 0 {
			stats.StrongestConnection = fmt.Sprintf("%s → %s", strongestSource, strongestTarget)
		}
	}

	bridgeCount := 0
	for _, n := range nodes {
		if n.SocialRole == domain.RoleBridge {
			bridgeCount++
		}
	}
	stats.BridgeCount = bridgeCount

	return stats
}

func (r *PostgresGraphRepository) GetInteractionCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT count(*) FROM interactions WHERE user_id = $1`, userID,
	).Scan(&count)
	return count, err
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS people (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			display_name TEXT NOT NULL,
			aliases TEXT[] DEFAULT '{}',
			notes TEXT DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS relationships (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			source_person_id UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
			target_person_id UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
			quality TEXT NOT NULL DEFAULT 'unknown',
			strength FLOAT NOT NULL DEFAULT 0.5 CHECK (strength >= 0 AND strength <= 1),
			label TEXT DEFAULT '',
			reciprocity_index FLOAT NOT NULL DEFAULT 0.5 CHECK (reciprocity_index >= 0 AND reciprocity_index <= 1),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS relationships_uniq ON relationships (source_person_id, target_person_id)`,
		`CREATE TABLE IF NOT EXISTS interactions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id TEXT NOT NULL,
			raw_text TEXT NOT NULL,
			summary TEXT DEFAULT '',
			people UUID[] DEFAULT '{}',
			job_id TEXT DEFAULT '',
			status TEXT NOT NULL DEFAULT 'queued',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS person_ownership (
			user_id TEXT NOT NULL,
			person_id UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
			PRIMARY KEY (user_id, person_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_interactions_user_id ON interactions (user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_interactions_status ON interactions (status)`,
		`CREATE INDEX IF NOT EXISTS idx_person_ownership_user ON person_ownership (user_id)`,
		`CREATE TABLE IF NOT EXISTS chat_messages (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id TEXT NOT NULL,
			sender TEXT NOT NULL,
			content TEXT NOT NULL,
			is_user BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			session_id TEXT DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_messages_user_id ON chat_messages (user_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_messages_session_id ON chat_messages (session_id, created_at DESC)`,
		`ALTER TABLE people ADD COLUMN IF NOT EXISTS social_role TEXT NOT NULL DEFAULT 'unknown'`,
		`ALTER TABLE people ADD COLUMN IF NOT EXISTS current_mood TEXT NOT NULL DEFAULT 'neutral'`,
		`ALTER TABLE people ADD COLUMN IF NOT EXISTS current_energy FLOAT NOT NULL DEFAULT 0.5`,
		`ALTER TABLE relationships ADD COLUMN IF NOT EXISTS source_weight FLOAT NOT NULL DEFAULT 0.5`,
		`ALTER TABLE relationships ADD COLUMN IF NOT EXISTS target_weight FLOAT NOT NULL DEFAULT 0.5`,
		`ALTER TABLE relationships ADD COLUMN IF NOT EXISTS protocol TEXT NOT NULL DEFAULT ''`,
		`CREATE TABLE IF NOT EXISTS person_states (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id TEXT NOT NULL,
			person_id UUID DEFAULT NULL,
			mood TEXT NOT NULL DEFAULT 'neutral',
			energy FLOAT NOT NULL DEFAULT 0.5 CHECK (energy >= 0 AND energy <= 1),
			valence FLOAT NOT NULL DEFAULT 0 CHECK (valence >= -1 AND valence <= 1),
			context TEXT DEFAULT '',
			trigger TEXT DEFAULT '',
			interaction_id UUID DEFAULT NULL,
			notes TEXT DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_person_states_user ON person_states (user_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_person_states_person ON person_states (user_id, person_id, created_at DESC)`,
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			display_name TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email)`,
		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token TEXT NOT NULL UNIQUE,
			expires_at TIMESTAMPTZ NOT NULL,
			revoked BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens (user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens (token)`,
	}

	for _, m := range migrations {
		if _, err := pool.Exec(ctx, m); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}
