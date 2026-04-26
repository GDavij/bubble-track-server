package domain

import "time"

type NodeMetrics struct {
	PersonID              string             `json:"person_id"`
	UserID                string             `json:"user_id"`
	ComputedAt            time.Time          `json:"computed_at"`
	TimeWindow            string             `json:"time_window"`

	Centrality            CentralityScores   `json:"centrality"`
	Community             CommunityMetrics   `json:"community"`
	RelationalHealth      RelationshipHealth `json:"relational_health"`
	SocialCapital         SocialCapital      `json:"social_capital"`
	Attachment            AttachmentProfile  `json:"attachment"`
	HumanistScore         HumanistProfile    `json:"humanist_score"`
	SocialExchange        SocialExchangeProfile `json:"social_exchange"`

	Degree                int                `json:"degree"`
	InteractionFrequency  float64            `json:"interaction_frequency"`
	LastInteractionAt     *time.Time         `json:"last_interaction_at"`
	EmotionalValence      float64            `json:"emotional_valence"`
	TrendDirection        TrendDirection     `json:"trend_direction"`
	TrendStrength         float64            `json:"trend_strength"`
}

type TrendDirection string

const (
	TrendRising      TrendDirection = "rising"
	TrendFalling     TrendDirection = "falling"
	TrendStable      TrendDirection = "stable"
	TrendVolatile    TrendDirection = "volatile"
)

type CentralityScores struct {
	Degree         float64 `json:"degree"`
	Betweenness    float64 `json:"betweenness"`
	Closeness      float64 `json:"closeness"`
	Eigenvector    float64 `json:"eigenvector"`
	PageRank       float64 `json:"page_rank"`
	ClusteringCoef float64 `json:"clustering_coefficient"`
}

type CommunityMetrics struct {
	CommunityID        string  `json:"community_id"`
	CommunityRole      string  `json:"community_role"`
	InternalEdges      int     `json:"internal_edges"`
	ExternalEdges      int     `json:"external_edges"`
	Embeddedness       float64 `json:"embeddedness"`
	BridgeScore        float64 `json:"bridge_score"`
	GatewayScore       float64 `json:"gateway_score"`
}

type NodeSnapshot struct {
	PersonID    string    `json:"person_id"`
	MetricName  string    `json:"metric_name"`
	Value       float64   `json:"value"`
	RecordedAt  time.Time `json:"recorded_at"`
}

type MetricTimeline struct {
	PersonID  string         `json:"person_id"`
	MetricName string         `json:"metric_name"`
	Points     []MetricPoint  `json:"points"`
}

type MetricPoint struct {
	Value      float64   `json:"value"`
	RecordedAt time.Time `json:"recorded_at"`
}

type GraphSnapshot struct {
	UserID      string        `json:"user_id"`
	TakenAt     time.Time     `json:"taken_at"`
	NodeCount   int           `json:"node_count"`
	EdgeCount   int           `json:"edge_count"`
	Density     float64       `json:"density"`
	AvgDegree   float64       `json:"avg_degree"`
	ComponentCount int        `json:"component_count"`
	LargestComponent int      `json:"largest_component"`
	GlobalClustering float64   `json:"global_clustering"`
	Assortativity   float64    `json:"assortativity"`
	IsSmallWorld   bool        `json:"is_small_world"`
	IsScaleFree    bool        `json:"is_scale_free"`
}

func ComputeNodeHealthScore(m NodeMetrics) float64 {
	centrality := (m.Centrality.Degree + m.Centrality.Betweenness + m.Centrality.Closeness) / 3.0
	capital := m.SocialCapital.TotalCapital
	humanist := (m.HumanistScore.AgencyScore + m.HumanistScore.EmpathicCapacity + m.HumanistScore.RelationalEthics) / 3.0
	health := (m.RelationalHealth.OverallScore + centrality*0.15 + capital*0.2 + humanist*0.15 + m.Community.Embeddedness*0.1)
	return clamp(health, 0, 1)
}

func ComputeTrend(points []MetricPoint) (TrendDirection, float64) {
	if len(points) < 2 {
		return TrendStable, 0
	}

	recent := points[len(points)-1].Value
	older := points[max(0, len(points)-4)].Value
	delta := recent - older

	volatility := 0.0
	for i := 1; i < len(points); i++ {
		volatility += abs(points[i].Value - points[i-1].Value)
	}
	volatility /= float64(len(points)-1)

	direction := TrendStable
	switch {
	case delta > 0.05 && volatility < 0.3:
		direction = TrendRising
	case delta < -0.05 && volatility < 0.3:
		direction = TrendFalling
	case volatility >= 0.3:
		direction = TrendVolatile
	}

	return direction, delta
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
