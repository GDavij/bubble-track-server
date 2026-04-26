package domain

import (
	"math"
	"testing"
	"time"
)

func TestComputeNodeHealthScore(t *testing.T) {
	tests := []struct {
		name    string
		metrics NodeMetrics
	}{
		{"zero metrics", NodeMetrics{}},
		{"high scores across all", NodeMetrics{
			RelationalHealth: RelationshipHealth{OverallScore: 0.9},
			Centrality:       CentralityScores{Degree: 0.8, Betweenness: 0.7, Closeness: 0.9},
			SocialCapital:    SocialCapital{TotalCapital: 0.85},
			HumanistScore:    HumanistProfile{AgencyScore: 0.8, EmpathicCapacity: 0.7, RelationalEthics: 0.9},
			Community:        CommunityMetrics{Embeddedness: 0.7},
		}},
		{"low scores across all", NodeMetrics{
			RelationalHealth: RelationshipHealth{OverallScore: 0.1},
			Centrality:       CentralityScores{Degree: 0.05, Betweenness: 0.05, Closeness: 0.05},
			SocialCapital:    SocialCapital{TotalCapital: 0.1},
			HumanistScore:    HumanistProfile{AgencyScore: 0.1, EmpathicCapacity: 0.1, RelationalEthics: 0.1},
			Community:        CommunityMetrics{Embeddedness: 0.1},
		}},
		{"overflow scores clamped", NodeMetrics{
			RelationalHealth: RelationshipHealth{OverallScore: 5.0},
			Centrality:       CentralityScores{Degree: 3.0, Betweenness: 2.0, Closeness: 4.0},
			SocialCapital:    SocialCapital{TotalCapital: 10.0},
			HumanistScore:    HumanistProfile{AgencyScore: 5.0, EmpathicCapacity: 3.0, RelationalEthics: 4.0},
			Community:        CommunityMetrics{Embeddedness: 8.0},
		}},
		{"negative scores clamped", NodeMetrics{
			RelationalHealth: RelationshipHealth{OverallScore: -1},
			Centrality:       CentralityScores{Degree: -2, Betweenness: -1, Closeness: -3},
			SocialCapital:    SocialCapital{TotalCapital: -5},
			HumanistScore:    HumanistProfile{AgencyScore: -1, EmpathicCapacity: -2, RelationalEthics: -3},
			Community:        CommunityMetrics{Embeddedness: -1},
		}},
		{"only relational health nonzero", NodeMetrics{
			RelationalHealth: RelationshipHealth{OverallScore: 0.5},
		}},
		{"only centrality nonzero", NodeMetrics{
			Centrality: CentralityScores{Degree: 0.5, Betweenness: 0.5, Closeness: 0.5},
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ComputeNodeHealthScore(tc.metrics)

			if got < 0 || got > 1 {
				t.Errorf("ComputeNodeHealthScore = %v, want in [0,1]", got)
			}
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Errorf("ComputeNodeHealthScore = %v (NaN or Inf)", got)
			}
		})
	}
}

func TestComputeNodeHealthScore_HighScores_HighResult(t *testing.T) {
	m := NodeMetrics{
		RelationalHealth: RelationshipHealth{OverallScore: 0.9},
		Centrality:       CentralityScores{Degree: 0.8, Betweenness: 0.8, Closeness: 0.8},
		SocialCapital:    SocialCapital{TotalCapital: 0.9},
		HumanistScore:    HumanistProfile{AgencyScore: 0.8, EmpathicCapacity: 0.8, RelationalEthics: 0.8},
		Community:        CommunityMetrics{Embeddedness: 0.8},
	}
	got := ComputeNodeHealthScore(m)
	if got < 0.5 {
		t.Errorf("ComputeNodeHealthScore = %v for high inputs, want >= 0.5", got)
	}
}

func TestComputeNodeHealthScore_ZeroMetrics_NonNaN(t *testing.T) {
	got := ComputeNodeHealthScore(NodeMetrics{})
	if math.IsNaN(got) || math.IsInf(got, 0) {
		t.Errorf("ComputeNodeHealthScore = %v for zero metrics", got)
	}
}

func TestComputeTrend(t *testing.T) {
	baseTime := time.Now()

	tests := []struct {
		name    string
		points  []MetricPoint
		wantDir TrendDirection
	}{
		{"empty points", []MetricPoint{}, TrendStable},
		{"single point", []MetricPoint{{Value: 0.5, RecordedAt: baseTime}}, TrendStable},
		{"rising trend", []MetricPoint{
			{Value: 0.1, RecordedAt: baseTime},
			{Value: 0.2, RecordedAt: baseTime.Add(1)},
			{Value: 0.3, RecordedAt: baseTime.Add(2)},
			{Value: 0.5, RecordedAt: baseTime.Add(3)},
		}, TrendRising},
		{"falling trend", []MetricPoint{
			{Value: 0.9, RecordedAt: baseTime},
			{Value: 0.7, RecordedAt: baseTime.Add(1)},
			{Value: 0.5, RecordedAt: baseTime.Add(2)},
			{Value: 0.3, RecordedAt: baseTime.Add(3)},
		}, TrendFalling},
		{"stable flat", []MetricPoint{
			{Value: 0.5, RecordedAt: baseTime},
			{Value: 0.51, RecordedAt: baseTime.Add(1)},
			{Value: 0.5, RecordedAt: baseTime.Add(2)},
			{Value: 0.5, RecordedAt: baseTime.Add(3)},
		}, TrendStable},
		{"volatile values", []MetricPoint{
			{Value: 0.1, RecordedAt: baseTime},
			{Value: 0.9, RecordedAt: baseTime.Add(1)},
			{Value: 0.1, RecordedAt: baseTime.Add(2)},
			{Value: 0.9, RecordedAt: baseTime.Add(3)},
			{Value: 0.1, RecordedAt: baseTime.Add(4)},
		}, TrendVolatile},
		{"two points stable", []MetricPoint{
			{Value: 0.5, RecordedAt: baseTime},
			{Value: 0.52, RecordedAt: baseTime.Add(1)},
		}, TrendStable},
		{"two points volatile due to large jump", []MetricPoint{
			{Value: 0.1, RecordedAt: baseTime},
			{Value: 0.7, RecordedAt: baseTime.Add(1)},
		}, TrendVolatile},
		{"all zeros", []MetricPoint{
			{Value: 0, RecordedAt: baseTime},
			{Value: 0, RecordedAt: baseTime.Add(1)},
			{Value: 0, RecordedAt: baseTime.Add(2)},
			{Value: 0, RecordedAt: baseTime.Add(3)},
		}, TrendStable},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir, delta := ComputeTrend(tc.points)

			if dir != tc.wantDir {
				t.Errorf("direction = %q, want %q (delta=%v)", dir, tc.wantDir, delta)
			}

			if len(tc.points) < 2 {
				if delta != 0 {
					t.Errorf("delta = %v for < 2 points, want 0", delta)
				}
			}
		})
	}
}

func TestComputeTrend_DeltaConsistency(t *testing.T) {
	baseTime := time.Now()

	points := []MetricPoint{
		{Value: 0.1, RecordedAt: baseTime},
		{Value: 0.3, RecordedAt: baseTime.Add(1)},
		{Value: 0.5, RecordedAt: baseTime.Add(2)},
		{Value: 0.7, RecordedAt: baseTime.Add(3)},
	}

	_, delta := ComputeTrend(points)
	recent := points[len(points)-1].Value
	older := points[max(0, len(points)-4)].Value
	expected := recent - older

	if math.Abs(delta-expected) > 1e-9 {
		t.Errorf("delta = %v, want %v", delta, expected)
	}
}

func TestComputeTrend_VolatileThreshold(t *testing.T) {
	baseTime := time.Now()

	points := []MetricPoint{
		{Value: 0.0, RecordedAt: baseTime},
		{Value: 0.3, RecordedAt: baseTime.Add(1)},
		{Value: 0.0, RecordedAt: baseTime.Add(2)},
		{Value: 0.3, RecordedAt: baseTime.Add(3)},
	}

	dir, _ := ComputeTrend(points)
	if dir != TrendVolatile {
		t.Errorf("direction = %q for volatile data, want volatile", dir)
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{0, 0},
		{5, 5},
		{-5, 5},
		{0.001, 0.001},
		{-0.001, 0.001},
	}

	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			got := abs(tc.input)
			if got != tc.want {
				t.Errorf("abs(%v) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		a, b int
		want int
	}{
		{0, 0, 0},
		{5, 3, 5},
		{3, 5, 5},
		{-1, 0, 0},
		{-5, -3, -3},
	}

	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			got := max(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("max(%d,%d) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}
