package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/bubbletrack/server/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GraphEngine struct {
	pool *pgxpool.Pool
}

func NewGraphEngine(pool *pgxpool.Pool) *GraphEngine {
	return &GraphEngine{pool: pool}
}

func (g *GraphEngine) ComputeDegreeCentrality(ctx context.Context, userID string) (map[string]float64, error) {
	rows, err := g.pool.Query(ctx, `
		SELECT p.id,
		       (SELECT count(*)::float FROM relationships r
		        WHERE r.source_person_id = p.id OR r.target_person_id = p.id
		        AND (EXISTS (
		            SELECT 1 FROM person_ownership po WHERE po.person_id = r.source_person_id AND po.user_id = $1
		        ) OR EXISTS (
		            SELECT 1 FROM person_ownership po WHERE po.person_id = r.target_person_id AND po.user_id = $1
		        ))) / NULLIF(
		            (SELECT count(*) FROM people pp
		             JOIN person_ownership po ON pp.id = po.person_id WHERE po.user_id = $1) - 1, 0
		        ) as degree
		FROM people p
		JOIN person_ownership po ON p.id = po.person_id
		WHERE po.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	centrality := make(map[string]float64)
	for rows.Next() {
		var id string
		var degree float64
		if err := rows.Scan(&id, &degree); err != nil {
			return nil, err
		}
		centrality[id] = degree
	}
	return centrality, rows.Err()
}

func (g *GraphEngine) GetNeighbors(ctx context.Context, personID string) ([]string, error) {
	rows, err := g.pool.Query(ctx, `
		SELECT target_person_id FROM relationships WHERE source_person_id = $1
		UNION
		SELECT source_person_id FROM relationships WHERE target_person_id = $1
	`, personID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var neighbors []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		neighbors = append(neighbors, id)
	}
	return neighbors, rows.Err()
}

func (g *GraphEngine) GetAdjacentPairs(ctx context.Context, personID string) ([][]string, error) {
	rows, err := g.pool.Query(ctx, `
		SELECT n.id, m.id
		FROM people n
		JOIN person_ownership po ON n.id = po.person_id
		JOIN relationships r ON (r.source_person_id = n.id OR r.target_person_id = n.id)
		JOIN people m ON (m.id = CASE WHEN r.source_person_id = n.id THEN r.target_person_id ELSE r.source_person_id END)
		WHERE po.user_id = $1 AND m.id != $2
	`, personID, personID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pairs [][]string
	for rows.Next() {
		var a, b string
		if err := rows.Scan(&a, &b); err != nil {
			return nil, err
		}
		pairs = append(pairs, []string{a, b})
	}
	return pairs, rows.Err()
}

func (g *GraphEngine) GetEdgeCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := g.pool.QueryRow(ctx, `
		SELECT count(DISTINCT r.id)
		FROM relationships r
		JOIN person_ownership po1 ON r.source_person_id = po1.person_id
		WHERE po1.user_id = $1
	`, userID).Scan(&count)
	return count, err
}

func (g *GraphEngine) GetNodeCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := g.pool.QueryRow(ctx, `
		SELECT count(*) FROM person_ownership WHERE user_id = $1
	`, userID).Scan(&count)
	return count, err
}

func (g *GraphEngine) ComputeClusteringCoefficient(ctx context.Context, personID string) (float64, error) {
	var coef float64
	err := g.pool.QueryRow(ctx, `
		WITH neighbors AS (
		    SELECT target_person_id AS nid FROM relationships WHERE source_person_id = $1
		    UNION
		    SELECT source_person_id FROM relationships WHERE target_person_id = $1
		),
		neighbor_pairs AS (
		    SELECT n1.nid AS a, n2.nid AS b
		    FROM neighbors n1, neighbors n2
		    WHERE n1.nid < n2.nid
		),
		triangles AS (
		    SELECT count(*)::float FROM neighbor_pairs np
		    WHERE EXISTS (
		        SELECT 1 FROM relationships r
		        WHERE (r.source_person_id = np.a AND r.target_person_id = np.b)
		        OR (r.source_person_id = np.b AND r.target_person_id = np.a)
		    )
		)
		SELECT
		    CASE WHEN (SELECT count(*) FROM neighbors) < 2 THEN 0
		    ELSE (SELECT count FROM triangles) / (SELECT count(*)::float FROM neighbor_pairs)
		    END
	`, personID).Scan(&coef)
	if err == pgx.ErrNoRows {
		return 0, nil
	}
	return coef, err
}

func (g *GraphEngine) UpsertNodeMetrics(ctx context.Context, m *domain.NodeMetrics) error {
	_, err := g.pool.Exec(ctx, `
		INSERT INTO node_metrics (
			person_id, user_id, computed_at, time_window,
			degree, interaction_frequency, emotional_valence,
			trend_direction, trend_strength,
			centrality_degree, centrality_betweenness, centrality_closeness,
			centrality_eigenvector, centrality_pagerank, centrality_clustering,
			community_id, community_role, community_embeddedness, community_bridge,
			relational_health_overall, relational_health_reciprocity,
			social_capital_total, humanist_agency, humanist_empathic,
			exchange_satisfaction, attachment_style
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26)
		ON CONFLICT (person_id, user_id, time_window) DO UPDATE SET
			computed_at = EXCLUDED.computed_at,
			degree = EXCLUDED.degree,
			interaction_frequency = EXCLUDED.interaction_frequency,
			emotional_valence = EXCLUDED.emotional_valence,
			trend_direction = EXCLUDED.trend_direction,
			trend_strength = EXCLUDED.trend_strength,
			centrality_degree = EXCLUDED.centrality_degree,
			centrality_betweenness = EXCLUDED.centrality_betweenness,
			centrality_closeness = EXCLUDED.centrality_closeness,
			centrality_eigenvector = EXCLUDED.centrality_eigenvector,
			centrality_pagerank = EXCLUDED.centrality_pagerank,
			centrality_clustering = EXCLUDED.centrality_clustering,
			community_id = EXCLUDED.community_id,
			community_role = EXCLUDED.community_role,
			community_embeddedness = EXCLUDED.community_embeddedness,
			community_bridge = EXCLUDED.community_bridge,
			relational_health_overall = EXCLUDED.relational_health_overall,
			relational_health_reciprocity = EXCLUDED.relational_health_reciprocity,
			social_capital_total = EXCLUDED.social_capital_total,
			humanist_agency = EXCLUDED.humanist_agency,
			humanist_empathic = EXCLUDED.humanist_empathic,
			exchange_satisfaction = EXCLUDED.exchange_satisfaction,
			attachment_style = EXCLUDED.attachment_style
	`, m.PersonID, m.UserID, m.ComputedAt, m.TimeWindow,
		m.Degree, m.InteractionFrequency, m.EmotionalValence,
		m.TrendDirection, m.TrendStrength,
		m.Centrality.Degree, m.Centrality.Betweenness, m.Centrality.Closeness,
		m.Centrality.Eigenvector, m.Centrality.PageRank, m.Centrality.ClusteringCoef,
		m.Community.CommunityID, m.Community.CommunityRole,
		m.Community.Embeddedness, m.Community.BridgeScore,
		m.RelationalHealth.OverallScore, m.RelationalHealth.Reciprocity,
		m.SocialCapital.TotalCapital,
		m.HumanistScore.AgencyScore, m.HumanistScore.EmpathicCapacity,
		m.SocialExchange.Satisfaction, string(m.Attachment.Style))
	return err
}

func (g *GraphEngine) GetMetricHistory(ctx context.Context, personID, metricName string, limit int) ([]domain.MetricPoint, error) {
	if limit <= 0 {
		limit = 30
	}

	colMap := map[string]string{
		"centrality_degree":         "centrality_degree",
		"centrality_betweenness":    "centrality_betweenness",
		"centrality_closeness":      "centrality_closeness",
		"centrality_pagerank":       "centrality_pagerank",
		"relational_health_overall": "relational_health_overall",
		"social_capital_total":      "social_capital_total",
		"emotional_valence":         "emotional_valence",
		"interaction_frequency":     "interaction_frequency",
	}

	col, ok := colMap[metricName]
	if !ok {
		return nil, fmt.Errorf("unknown metric: %s", metricName)
	}

	query := fmt.Sprintf(
		`SELECT %s, computed_at FROM node_metrics
		 WHERE person_id = $1
		 ORDER BY computed_at DESC
		 LIMIT $2`, col)

	rows, err := g.pool.Query(ctx, query, personID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []domain.MetricPoint
	for rows.Next() {
		var p domain.MetricPoint
		if err := rows.Scan(&p.Value, &p.RecordedAt); err != nil {
			return nil, err
		}
		points = append(points, p)
	}

	for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
		points[i], points[j] = points[j], points[i]
	}
	return points, rows.Err()
}

func AddNodeMetricsMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS node_metrics (
			person_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			time_window TEXT NOT NULL DEFAULT 'all',
			computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			degree INT DEFAULT 0,
			interaction_frequency FLOAT DEFAULT 0,
			emotional_valence FLOAT DEFAULT 0 CHECK (emotional_valence >= -1 AND emotional_valence <= 1),
			trend_direction TEXT DEFAULT 'stable',
			trend_strength FLOAT DEFAULT 0,
			centrality_degree FLOAT DEFAULT 0,
			centrality_betweenness FLOAT DEFAULT 0,
			centrality_closeness FLOAT DEFAULT 0,
			centrality_eigenvector FLOAT DEFAULT 0,
			centrality_pagerank FLOAT DEFAULT 0,
			centrality_clustering FLOAT DEFAULT 0,
			community_id TEXT DEFAULT '',
			community_role TEXT DEFAULT '',
			community_embeddedness FLOAT DEFAULT 0,
			community_bridge FLOAT DEFAULT 0,
			relational_health_overall FLOAT DEFAULT 0,
			relational_health_reciprocity FLOAT DEFAULT 0,
			social_capital_total FLOAT DEFAULT 0,
			humanist_agency FLOAT DEFAULT 0,
			humanist_empathic FLOAT DEFAULT 0,
			exchange_satisfaction FLOAT DEFAULT 0,
			attachment_style TEXT DEFAULT 'unknown',
			PRIMARY KEY (person_id, user_id, time_window)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_node_metrics_user_time ON node_metrics (user_id, computed_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_node_metrics_person ON node_metrics (person_id, computed_at DESC)`,
	}

	for _, m := range migrations {
		if _, err := pool.Exec(ctx, m); err != nil {
			return fmt.Errorf("node_metrics migration failed: %w", err)
		}
	}
	_ = time.Now()
	return nil
}
