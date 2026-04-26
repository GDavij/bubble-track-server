package domain

import (
	"math"
	"testing"
)

func TestComputePathDependency(t *testing.T) {
	tests := []struct {
		name           string
		pastDecisions  int
		repeatedPats   int
		totalDecisions int
		sunkCostEvid   int
	}{
		{"zero total decisions", 5, 3, 0, 2},
		{"high repetition", 50, 80, 100, 10},
		{"low repetition", 10, 5, 100, 2},
		{"all repeated", 50, 100, 100, 0},
		{"single decision", 1, 0, 1, 0},
		{"no past decisions", 0, 0, 50, 0},
		{"high sunk cost evidence", 50, 30, 100, 80},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ComputePathDependency(tc.pastDecisions, tc.repeatedPats, tc.totalDecisions, tc.sunkCostEvid)

			if tc.totalDecisions == 0 {
				if got != (HistoricalProfile{}) {
					t.Errorf("expected zero-value struct for zero totalDecisions, got %+v", got)
				}
				return
			}

			assertClampedHist(t, "PathDependency", got.PathDependency)
			assertClampedHist(t, "LegacyEffect", got.LegacyEffect)
			assertClampedHist(t, "LockInScore", got.LockInScore)
			assertClampedHist(t, "BranchingPointRisk", got.BranchingPointRisk)
			assertClampedHist(t, "GenerationalShift", got.GenerationalShift)
			assertClampedHist(t, "AccumulatedMomentum", got.AccumulatedMomentum)
		})
	}
}

func TestComputePathDependency_HighRepetition_LockIn(t *testing.T) {
	got := ComputePathDependency(50, 90, 100, 10)
	if got.LockInScore < 0.5 {
		t.Errorf("LockInScore = %v for high repetition, want >= 0.5", got.LockInScore)
	}
}

func TestComputePathDependency_BranchingInverseLockIn(t *testing.T) {
	got := ComputePathDependency(10, 5, 100, 2)
	expected := clamp(1-got.LockInScore, 0, 1)
	if math.Abs(got.BranchingPointRisk-expected) > 1e-9 {
		t.Errorf("BranchingPointRisk = %v, want %v (inverse of LockInScore)", got.BranchingPointRisk, expected)
	}
}

func TestComputePathDependency_CyclePhase(t *testing.T) {
	tests := []struct {
		name           string
		pastDecisions  int
		repeatedPats   int
		totalDecisions int
		sunkCostEvid   int
		wantPhase      string
	}{
		{"lock_in for high repetition", 80, 90, 100, 5, "lock_in"},
		{"transition for low repetition", 5, 2, 100, 2, "transition"},
		{"maintenance default", 30, 30, 100, 10, "maintenance"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ComputePathDependency(tc.pastDecisions, tc.repeatedPats, tc.totalDecisions, tc.sunkCostEvid)
			if got.CyclePhase != tc.wantPhase {
				t.Errorf("CyclePhase = %q, want %q", got.CyclePhase, tc.wantPhase)
			}
		})
	}
}

func TestDetectTemporalCycle(t *testing.T) {
	tests := []struct {
		name       string
		values     []float64
		windowSize int
	}{
		{"empty slice", []float64{}, 3},
		{"insufficient data", []float64{1, 2, 3}, 2},
		{"constant values", []float64{5, 5, 5, 5, 5, 5, 5, 5}, 3},
		{"increasing trend", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 3},
		{"decreasing trend", []float64{10, 9, 8, 7, 6, 5, 4, 3, 2, 1}, 3},
		{"oscillating", []float64{1, 3, 1, 3, 1, 3, 1, 3, 1, 3}, 3},
		{"zeros", []float64{0, 0, 0, 0, 0, 0, 0, 0}, 3},
		{"single value repeated", []float64{7, 7, 7, 7, 7, 7, 7}, 2},
		{"large values", []float64{1000, 2000, 3000, 4000, 5000, 6000, 7000, 8000}, 3},
		{"negative values", []float64{-5, -4, -3, -2, -1, 0, 1, 2, 3, 4}, 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := DetectTemporalCycle(tc.values, tc.windowSize)

			if len(tc.values) < tc.windowSize*2 {
				if got.CycleType != "insufficient_data" {
					t.Errorf("CycleType = %q, want insufficient_data", got.CycleType)
				}
				return
			}

			assertClampedHist(t, "Seasonality", got.Seasonality)
			assertClampedHist(t, "Predictability", got.Predictability)

			if got.CycleLength != float64(tc.windowSize) {
				t.Errorf("CycleLength = %v, want %v", got.CycleLength, float64(tc.windowSize))
			}
		})
	}
}

func TestDetectTemporalCycle_Constant_StableCycleType(t *testing.T) {
	vals := []float64{5, 5, 5, 5, 5, 5, 5, 5}
	got := DetectTemporalCycle(vals, 3)
	if got.CycleType != "stable" {
		t.Errorf("CycleType = %q for constant values, want stable", got.CycleType)
	}
}

func TestDetectTemporalCycle_Increasing_NonIrregular(t *testing.T) {
	vals := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	got := DetectTemporalCycle(vals, 3)
	if got.CycleType == "irregular" {
		t.Errorf("CycleType = %q for strongly increasing values, want non-irregular", got.CycleType)
	}
}

func TestDetectTemporalCycle_Oscillating_LowSeasonality(t *testing.T) {
	vals := []float64{1, 5, 1, 5, 1, 5, 1, 5, 1, 5}
	got := DetectTemporalCycle(vals, 3)
	if got.CycleType != "irregular" && got.Seasonality <= 0 {
		t.Errorf("Seasonality = %v for alternating values, expected negative autocorrelation clamped", got.Seasonality)
	}
}

func TestComputeGenerationalDynamics(t *testing.T) {
	tests := []struct {
		name           string
		valueOverlap   float64
		techComfortA   float64
		techComfortB   float64
		conflictEvents int
		knowledgeShare int
		wantStyle      string
	}{
		{"digital divide", 0.5, 0.9, 0.1, 5, 10, "digital_divide"},
		{"similar style", 0.7, 0.5, 0.55, 2, 20, "similar"},
		{"mixed style", 0.5, 0.6, 0.2, 5, 10, "mixed"},
		{"zero overlap", 0, 0.5, 0.5, 0, 0, "similar"},
		{"max overlap", 1, 0.8, 0.8, 0, 30, "similar"},
		{"negative overlap clamped", -0.5, 0.5, 0.5, 0, 0, "similar"},
		{"overflow overlap clamped", 2, 0.5, 0.5, 0, 0, "similar"},
		{"high conflict events", 0.5, 0.5, 0.5, 50, 10, "similar"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ComputeGenerationalDynamics(tc.valueOverlap, tc.techComfortA, tc.techComfortB, tc.conflictEvents, tc.knowledgeShare)

			assertClampedHist(t, "ValueAlignment", got.ValueAlignment)
			assertClampedHist(t, "ConflictPotential", got.ConflictPotential)
			assertClampedHist(t, "KnowledgeTransfer", got.KnowledgeTransfer)

			if got.CommunicationStyle != tc.wantStyle {
				t.Errorf("CommunicationStyle = %q, want %q", got.CommunicationStyle, tc.wantStyle)
			}
		})
	}
}

func TestComputeHoweStrauss(t *testing.T) {
	tests := []struct {
		name               string
		institutionalTrust float64
		individualism      float64
		crisisLevel        float64
		wantPhase          string
		wantTurnings       int
	}{
		{"crisis phase", 0.2, 0.5, 0.8, "crisis", 4},
		{"awakening phase", 0.3, 0.8, 0.4, "awakening", 1},
		{"high phase", 0.8, 0.3, 0.1, "high", 3},
		{"default high when no match", 0.6, 0.4, 0.4, "high", 3},
		{"boundary crisis", 0.25, 0.7, 0.75, "crisis", 4},
		{"zero values", 0, 0, 0, "high", 3},
		{"max values stays high", 1, 1, 1, "high", 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ComputeHoweStrauss(tc.institutionalTrust, tc.individualism, tc.crisisLevel)

			if got.HoweStraussPhase != tc.wantPhase {
				t.Errorf("HoweStraussPhase = %q, want %q", got.HoweStraussPhase, tc.wantPhase)
			}
			if got.TurningsCount != tc.wantTurnings {
				t.Errorf("TurningsCount = %d, want %d", got.TurningsCount, tc.wantTurnings)
			}
			if got.InstitutionalTrust != tc.institutionalTrust {
				t.Errorf("InstitutionalTrust = %v, want %v", got.InstitutionalTrust, tc.institutionalTrust)
			}
			if got.Individualism != tc.individualism {
				t.Errorf("Individualism = %v, want %v", got.Individualism, tc.individualism)
			}
			if got.CrisisSeverity != tc.crisisLevel {
				t.Errorf("CrisisSeverity = %v, want %v", got.CrisisSeverity, tc.crisisLevel)
			}
		})
	}
}

func TestSqrt(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{0, 0},
		{-1, 0},
		{1, 1},
		{4, 2},
		{9, 3},
		{2, math.Sqrt(2)},
		{0.25, 0.5},
		{100, 10},
	}

	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			got := sqrt(tc.input)
			if math.Abs(got-tc.want) > 1e-6 {
				t.Errorf("sqrt(%v) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func assertClampedHist(t *testing.T, field string, v float64) {
	t.Helper()
	if math.IsNaN(v) || math.IsInf(v, 0) {
		t.Errorf("%s = %v (NaN or Inf)", field, v)
		return
	}
	if v < 0 || v > 1 {
		t.Errorf("%s = %v, want in [0,1]", field, v)
	}
}
