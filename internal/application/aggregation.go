package application

import (
	"math"
	"time"

	"github.com/bubbletrack/server/internal/domain"
	"github.com/bubbletrack/server/internal/infrastructure/repository"
)

type AggregationEngine struct {
	graphEngine *repository.GraphEngine
}

func NewAggregationEngine(engine *repository.GraphEngine) *AggregationEngine {
	return &AggregationEngine{graphEngine: engine}
}

type AggregatedProfile struct {
	PersonID            string                       `json:"person_id"`
	Period              string                       `json:"period"`
	From                time.Time                    `json:"from"`
	To                  time.Time                    `json:"to"`
	Metrics             *domain.NodeMetrics           `json:"metrics"`
	Classification      *InterdisciplinaryRoleScore  `json:"classification"`
	GraphPosition       *GraphPositionSummary        `json:"graph_position"`
	PatternAnalysis      *BehavioralPatternAnalysis   `json:"pattern_analysis"`
}

type GraphPositionSummary struct {
	RankByDegree       int     `json:"rank_by_degree"`
	RankByBetweenness  int     `json:"rank_by_betweenness"`
	RankByPageRank      int     `json:"rank_by_pagerank"`
	PercentileByDegree float64 `json:"percentile_degree"`
	PercentileByBridge  float64 `json:"percentile_bridge"`
}

type BehavioralPatternAnalysis struct {
	InteractionTrend    domain.TrendDirection `json:"interaction_trend"`
	ConsistencyScore    float64                `json:"consistency_score"`
	EvolutionScore      float64                `json:"evolution_score"`
	StabilityLevel      float64                `json:"stability_level"`
	AnomalyCount        int                    `json:"anomaly_count"`
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func (a *AggregationEngine) AggregateProfile(
	m *domain.NodeMetrics,
	allMetrics map[string]*domain.NodeMetrics,
) *AggregatedProfile {
	classEngine := NewClassificationEngine()
	classification := classEngine.ClassifyRole(m)

	position := computeGraphPosition(m, allMetrics)
	pattern := computeBehavioralPattern(m, allMetrics)

	return &AggregatedProfile{
		PersonID:       m.PersonID,
	Period:         m.TimeWindow,
		From:           m.ComputedAt,
	To:             m.ComputedAt,
		Metrics:        m,
		Classification: classification,
		GraphPosition:  position,
		PatternAnalysis: pattern,
	}
}

func (a *AggregationEngine) AggregateAll(metrics map[string]*domain.NodeMetrics) []*AggregatedProfile {
	profiles := make([]*AggregatedProfile, 0, len(metrics))
	for _, m := range metrics {
		profiles = append(profiles, a.AggregateProfile(m, metrics))
	}
	return profiles
}

func computeGraphPosition(m *domain.NodeMetrics, all map[string]*domain.NodeMetrics) *GraphPositionSummary {
	rankDegree := 1
	rankBetweenness := 1
	rankPageRank := 1
	degreeCount := 0
	betweenCount := 0
	pageRankCount := 0

	for _, other := range all {
		if other.Centrality.Degree > m.Centrality.Degree {
			rankDegree++
		}
		if other.Centrality.Betweenness > m.Centrality.Betweenness {
			rankBetweenness++
		}
		if other.Centrality.PageRank > m.Centrality.PageRank {
			rankPageRank++
		}
		degreeCount++
		betweenCount++
		pageRankCount++
	}

	percentileDegree := 0.0
	if degreeCount > 0 {
		percentileDegree = 1.0 - float64(rankDegree-1)/float64(degreeCount)
	}
	percentileBridge := 0.0
	if betweenCount > 0 {
		percentileBridge = 1.0 - float64(rankBetweenness-1)/float64(betweenCount)
	}

	return &GraphPositionSummary{
		RankByDegree:       rankDegree,
		RankByBetweenness:  rankBetweenness,
		RankByPageRank:      rankPageRank,
		PercentileByDegree:  percentileDegree,
		PercentileByBridge:  percentileBridge,
	}
}

func computeBehavioralPattern(m *domain.NodeMetrics, all map[string]*domain.NodeMetrics) *BehavioralPatternAnalysis {
	avgDegree := 0.0
	count := 0
	for _, other := range all {
		avgDegree += other.Centrality.Degree
		count++
	}
	if count > 0 {
		avgDegree /= float64(count)
	}

	stability := 1.0 - math.Abs(m.Centrality.Degree-avgDegree)/(avgDegree+0.01)
	stability = clamp(stability, 0, 1)

	evolution := (m.Community.BridgeScore + (1-m.Community.Embeddedness)) / 2.0
	consistency := m.Community.Embeddedness

	return &BehavioralPatternAnalysis{
		InteractionTrend: m.TrendDirection,
		ConsistencyScore: consistency,
		EvolutionScore:   clamp(evolution, 0, 1),
		StabilityLevel:   stability,
		AnomalyCount:     0,
	}
}
