package domain

import (
	"math"
	"testing"
)

func TestComputeNeuralSocial(t *testing.T) {
	tests := []struct {
		name                 string
		proximity            int
		positiveInteractions int
		negativeInteractions int
		sharedActivities     int
	}{
		{"zero values", 0, 0, 0, 0},
		{"all positive", 10, 20, 0, 5},
		{"all negative", 5, 0, 20, 0},
		{"mixed interactions", 8, 15, 5, 3},
		{"high proximity high shared", 50, 10, 2, 30},
		{"large numbers", 100, 200, 50, 80},
		{"single interaction", 1, 1, 0, 0},
		{"only negatives", 0, 0, 100, 0},
		{"only shared activities", 0, 0, 0, 100},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ComputeNeuralSocial(tc.proximity, tc.positiveInteractions, tc.negativeInteractions, tc.sharedActivities)

			assertClampedNeuro(t, "MirrorActivity", got.MirrorActivity)
			assertClampedNeuro(t, "OxytocinProxy", got.OxytocinProxy)
			assertClampedNeuro(t, "DopamineProxy", got.DopamineProxy)
			assertClampedNeuro(t, "CortisolProxy", got.CortisolProxy)
			assertClampedNeuro(t, "SerotoninProxy", got.SerotoninProxy)
			assertClampedNeuro(t, "EmpathyPrediction", got.EmpathyPrediction)
			assertClampedNeuro(t, "BondingReadiness", got.BondingReadiness)
			assertClampedNeuro(t, "StressResponse", got.StressResponse)
			assertClampedNeuro(t, "RewardPrediction", got.RewardPrediction)
			assertClampedNeuro(t, "NeuralSynchrony", got.NeuralSynchrony)
		})
	}
}

func TestComputeNeuralSocial_AllPositive_HighValues(t *testing.T) {
	got := ComputeNeuralSocial(20, 50, 0, 10)

	if got.EmpathyPrediction < 0.3 {
		t.Errorf("EmpathyPrediction too low for all-positive input: %v", got.EmpathyPrediction)
	}
	if got.BondingReadiness < 0.3 {
		t.Errorf("BondingReadiness too low for all-positive input: %v", got.BondingReadiness)
	}
	if got.CortisolProxy > got.DopamineProxy {
		t.Errorf("for all-positive input, cortisol (%v) should not exceed dopamine (%v)", got.CortisolProxy, got.DopamineProxy)
	}
}

func TestComputeNeuralSocial_AllNegative_HighStress(t *testing.T) {
	got := ComputeNeuralSocial(0, 0, 50, 0)

	if got.CortisolProxy < 0.5 {
		t.Errorf("CortisolProxy expected high for all-negative input: %v", got.CortisolProxy)
	}
	if got.StressResponse < 0.3 {
		t.Errorf("StressResponse expected elevated for all-negative input: %v", got.StressResponse)
	}
}

func TestComputeNeuralSocial_ZeroInputs_NonNaN(t *testing.T) {
	got := ComputeNeuralSocial(0, 0, 0, 0)

	fields := []struct {
		name  string
		value float64
	}{
		{"MirrorActivity", got.MirrorActivity},
		{"OxytocinProxy", got.OxytocinProxy},
		{"DopamineProxy", got.DopamineProxy},
		{"CortisolProxy", got.CortisolProxy},
		{"SerotoninProxy", got.SerotoninProxy},
		{"EmpathyPrediction", got.EmpathyPrediction},
		{"BondingReadiness", got.BondingReadiness},
		{"StressResponse", got.StressResponse},
		{"RewardPrediction", got.RewardPrediction},
		{"NeuralSynchrony", got.NeuralSynchrony},
	}
	for _, f := range fields {
		if math.IsNaN(f.value) {
			t.Errorf("%s is NaN", f.name)
		}
		if math.IsInf(f.value, 0) {
			t.Errorf("%s is Inf", f.name)
		}
	}
}

func TestComputeEmotionalRegulation(t *testing.T) {
	tests := []struct {
		name                   string
		selfReports            int
		totalEmotionalEvents   int
		conflictResolutionRate float64
	}{
		{"zero total events returns zero", 5, 0, 0.5},
		{"high awareness high resolution", 80, 100, 0.9},
		{"low awareness low resolution", 5, 100, 0.1},
		{"high conflict resolution rate", 50, 100, 1.5},
		{"exceeds clamp range", 200, 100, 2.0},
		{"single event", 1, 1, 1.0},
		{"zero self reports", 0, 100, 0.5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ComputeEmotionalRegulation(tc.selfReports, tc.totalEmotionalEvents, tc.conflictResolutionRate)

			if tc.totalEmotionalEvents == 0 {
				if got != (EmotionalRegulationProfile{}) {
					t.Errorf("expected zero-value struct for zero totalEmotionalEvents, got %+v", got)
				}
				return
			}

			assertClampedNeuro(t, "Awareness", got.Awareness)
			assertClampedNeuro(t, "Reappraisal", got.Reappraisal)
			assertClampedNeuro(t, "Suppression", got.Suppression)
			assertClampedNeuro(t, "Acceptance", got.Acceptance)
			assertClampedNeuro(t, "EQScore", got.EQScore)
			assertClampedNeuro(t, "AlexithymiaRisk", got.AlexithymiaRisk)
			assertClampedNeuro(t, "EmotionalLability", got.EmotionalLability)
		})
	}
}

func TestComputeEmotionalRegulation_HighResolution_LowSuppression(t *testing.T) {
	got := ComputeEmotionalRegulation(80, 100, 0.95)
	if got.Suppression > got.Reappraisal {
		t.Errorf("Suppression (%v) should be lower than Reappraisal (%v) for high resolution rate", got.Suppression, got.Reappraisal)
	}
}

func TestComputeEmotionalRegulation_HighAwareness_LowAlexithymia(t *testing.T) {
	got := ComputeEmotionalRegulation(90, 100, 0.8)
	if got.AlexithymiaRisk > 0.7 {
		t.Errorf("AlexithymiaRisk should be low for high awareness: %v", got.AlexithymiaRisk)
	}
}

func TestComputeBurnoutRisk(t *testing.T) {
	tests := []struct {
		name              string
		socialObligations int
		recoveryTime      int
		conflictEvents    int
		totalEvents       int
	}{
		{"zero total events returns zero", 5, 5, 2, 0},
		{"low obligations high recovery", 2, 50, 1, 100},
		{"high obligations low recovery", 20, 1, 30, 50},
		{"max obligations", 100, 0, 0, 100},
		{"zero obligations", 0, 100, 0, 100},
		{"single event", 1, 0, 0, 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ComputeBurnoutRisk(tc.socialObligations, tc.recoveryTime, tc.conflictEvents, tc.totalEvents)

			if tc.totalEvents == 0 {
				if got != (CircadianSocialProfile{}) {
					t.Errorf("expected zero-value struct for zero totalEvents, got %+v", got)
				}
				return
			}

			assertClampedNeuro(t, "CircadianAlign", got.CircadianAlign)
			assertClampedNeuro(t, "BurnoutRisk", got.BurnoutRisk)
			assertClampedNeuro(t, "SocialBattery", got.SocialBattery)
		})
	}
}

func TestComputeBurnoutRisk_RecoveryPattern(t *testing.T) {
	tests := []struct {
		name              string
		socialObligations int
		recoveryTime      int
		conflictEvents    int
		totalEvents       int
		wantPattern       string
	}{
		{"high burnout gives insufficient", 19, 0, 40, 50, "insufficient"},
		{"moderate burnout gives minimal", 8, 5, 10, 50, "minimal"},
		{"low burnout gives adequate", 2, 50, 1, 100, "adequate"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ComputeBurnoutRisk(tc.socialObligations, tc.recoveryTime, tc.conflictEvents, tc.totalEvents)
			if got.RecoveryPattern != tc.wantPattern {
				t.Errorf("RecoveryPattern = %q, want %q (burnout=%.4f)", got.RecoveryPattern, tc.wantPattern, got.BurnoutRisk)
			}
		})
	}
}

func assertClampedNeuro(t *testing.T, field string, v float64) {
	t.Helper()
	if math.IsNaN(v) || math.IsInf(v, 0) {
		t.Errorf("%s = %v (NaN or Inf)", field, v)
		return
	}
	if v < 0 || v > 1 {
		t.Errorf("%s = %v, want in [0,1]", field, v)
	}
}
