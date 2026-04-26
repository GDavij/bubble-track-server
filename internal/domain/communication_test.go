package domain

import (
	"math"
	"testing"
)

func TestComputeWatzlawick(t *testing.T) {
	tests := []struct {
		name           string
		received       int
		understood     int
		relevant       int
		timely         int
		mediumRichness int
		initFrequency  int
		totalMessages  int
	}{
		{"zero total messages", 0, 0, 0, 0, 0, 0, 0},
		{"perfect communication", 100, 100, 100, 100, 100, 100, 100},
		{"poor communication", 10, 5, 3, 20, 10, 2, 100},
		{"single message", 1, 1, 1, 1, 1, 1, 1},
		{"high received low understood", 90, 10, 50, 50, 50, 50, 100},
		{"exceeding totals", 200, 200, 200, 200, 200, 200, 100},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ComputeWatzlawick(tc.received, tc.understood, tc.relevant, tc.timely, tc.mediumRichness, tc.initFrequency, tc.totalMessages)

			if tc.totalMessages == 0 {
				if got != (WatzlawickAxioms{}) {
					t.Errorf("expected zero-value struct for zero totalMessages, got %+v", got)
				}
				return
			}

			assertClampedComm(t, "Axiom1Received", got.Axiom1Received)
			assertClampedComm(t, "Axiom1Quality", got.Axiom1Quality)
			assertClampedComm(t, "Axiom2Content", got.Axiom2Content)
			assertClampedComm(t, "Axiom2Relation", got.Axiom2Relation)
			assertClampedComm(t, "Axiom3Punctuation", got.Axiom3Punctuation)
			assertClampedComm(t, "Axiom4Digital", got.Axiom4Digital)
			assertClampedComm(t, "Axiom5Symmetric", got.Axiom5Symmetric)
			assertClampedComm(t, "OverallCompliance", got.OverallCompliance)
		})
	}
}

func TestComputeWatzlawick_PerfectComms_HighCompliance(t *testing.T) {
	got := ComputeWatzlawick(100, 100, 100, 100, 100, 100, 100)
	if got.OverallCompliance < 0.9 {
		t.Errorf("OverallCompliance = %v for perfect communication, want >= 0.9", got.OverallCompliance)
	}
}

func TestComputeWatzlawick_Axiom1QualityIsProduct(t *testing.T) {
	got := ComputeWatzlawick(50, 80, 50, 50, 50, 50, 100)
	expected := clamp(got.Axiom1Received*got.Axiom2Content, 0, 1)
	if math.Abs(got.Axiom1Quality-expected) > 1e-9 {
		t.Errorf("Axiom1Quality = %v, want %v", got.Axiom1Quality, expected)
	}
}

func TestComputeWatzlawick_ComplianceIsAverage(t *testing.T) {
	got := ComputeWatzlawick(50, 60, 70, 40, 80, 50, 100)
	avg := (got.Axiom1Received + got.Axiom2Content + got.Axiom3Punctuation + got.Axiom4Digital + got.Axiom5Symmetric) / 5.0
	if math.Abs(got.OverallCompliance-avg) > 1e-9 {
		t.Errorf("OverallCompliance = %v, want average %v", got.OverallCompliance, avg)
	}
}

func TestComputeMetaCommunication(t *testing.T) {
	tests := []struct {
		name               string
		incongruenceCount  int
		paradoxCount       int
		congruenceCount    int
		totalInteractions  int
	}{
		{"zero interactions", 0, 0, 0, 0},
		{"high congruence", 5, 2, 80, 100},
		{"high incongruence", 80, 30, 5, 100},
		{"all paradox", 0, 50, 0, 50},
		{"single interaction", 0, 0, 1, 1},
		{"exceeding interactions", 200, 200, 200, 100},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ComputeMetaCommunication(tc.incongruenceCount, tc.paradoxCount, tc.congruenceCount, tc.totalInteractions)

			if tc.totalInteractions == 0 {
				if got != (MetaCommunicationProfile{}) {
					t.Errorf("expected zero-value struct, got %+v", got)
				}
				return
			}

			assertClampedComm(t, "DoubleBindRisk", got.DoubleBindRisk)
			assertClampedComm(t, "ParadoxLevel", got.ParadoxLevel)
			assertClampedComm(t, "Congruence", got.Congruence)
			assertClampedComm(t, "MetaMessageAccuracy", got.MetaMessageAccuracy)
			assertClampedComm(t, "FrameControl", got.FrameControl)
			assertClampedComm(t, "Reflexivity", got.Reflexivity)
		})
	}
}

func TestComputeMetaCommunication_HighCongruence_HighReflexivity(t *testing.T) {
	got := ComputeMetaCommunication(1, 0, 90, 100)
	if got.Reflexivity < 0.5 {
		t.Errorf("Reflexivity = %v for high congruence, want >= 0.5", got.Reflexivity)
	}
}

func TestComputeMetaCommunication_HighIncongruence_HighDoubleBind(t *testing.T) {
	got := ComputeMetaCommunication(80, 10, 5, 100)
	if got.DoubleBindRisk < 0.5 {
		t.Errorf("DoubleBindRisk = %v for high incongruence, want >= 0.5", got.DoubleBindRisk)
	}
}

func TestComputeNarrative(t *testing.T) {
	tests := []struct {
		name       string
		coherence  float64
		agency     float64
		complexity float64
		empathy    float64
		resolution float64
		wantTraj   string
	}{
		{"empowerment trajectory", 0.8, 0.8, 0.5, 0.6, 0.8, "empowerment"},
		{"entrapment trajectory", 0.5, 0.2, 0.8, 0.3, 0.2, "entrapment"},
		{"transformation trajectory", 0.6, 0.5, 0.7, 0.6, 0.6, "transformation"},
		{"struggle trajectory", 0.5, 0.6, 0.4, 0.5, 0.2, "struggle"},
		{"stable default", 0.5, 0.5, 0.4, 0.5, 0.5, "stable"},
		{"zero values", 0, 0, 0, 0, 0, "stable"},
		{"max values", 1, 1, 1, 1, 1, "empowerment"},
		{"negative values clamped", -1, -1, -1, -1, -1, "stable"},
		{"overflow values clamped", 5, 5, 5, 5, 5, "empowerment"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ComputeNarrative(tc.coherence, tc.agency, tc.complexity, tc.empathy, tc.resolution)

			assertClampedComm(t, "Coherence", got.Coherence)
			assertClampedComm(t, "Agency", got.Agency)
			assertClampedComm(t, "Complexity", got.Complexity)
			assertClampedComm(t, "Empathy", got.Empathy)
			assertClampedComm(t, "Resolution", got.Resolution)
			assertClampedComm(t, "RedemptionArc", got.RedemptionArc)
			assertClampedComm(t, "ContaminationArc", got.ContaminationArc)

			if got.Trajectory != tc.wantTraj {
				t.Errorf("Trajectory = %q, want %q", got.Trajectory, tc.wantTraj)
			}
		})
	}
}

func TestComputeNarrative_RedemptionArc_Activated(t *testing.T) {
	got := ComputeNarrative(0.7, 0.6, 0.5, 0.7, 0.6)
	if got.RedemptionArc <= 0 {
		t.Errorf("RedemptionArc should be > 0 when agency>0.4 and empathy>0.4, got %v", got.RedemptionArc)
	}
}

func TestComputeNarrative_RedemptionArc_NotActivated(t *testing.T) {
	got := ComputeNarrative(0.5, 0.3, 0.5, 0.3, 0.5)
	if got.RedemptionArc != 0 {
		t.Errorf("RedemptionArc should be 0 when agency<=0.4 or empathy<=0.4, got %v", got.RedemptionArc)
	}
}

func TestComputeNarrative_ContaminationArc_Activated(t *testing.T) {
	got := ComputeNarrative(0.5, 0.5, 0.7, 0.3, 0.2)
	if got.ContaminationArc <= 0 {
		t.Errorf("ContaminationArc should be > 0 when complexity>0.6 and resolution<0.3, got %v", got.ContaminationArc)
	}
}

func TestComputeNarrative_ContaminationArc_NotActivated(t *testing.T) {
	got := ComputeNarrative(0.5, 0.5, 0.5, 0.5, 0.5)
	if got.ContaminationArc != 0 {
		t.Errorf("ContaminationArc should be 0 when complexity<=0.6 or resolution>=0.3, got %v", got.ContaminationArc)
	}
}

func TestComputeShannonInfo(t *testing.T) {
	tests := []struct {
		name              string
		messageVariety    int
		totalMessages     int
		misunderstandings int
	}{
		{"zero messages", 0, 0, 0},
		{"high variety no misunderstandings", 50, 100, 0},
		{"low variety high misunderstandings", 5, 100, 30},
		{"single message", 1, 1, 0},
		{"all misunderstandings", 10, 50, 50},
		{"exceeding variety", 200, 100, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ComputeShannonInfo(tc.messageVariety, tc.totalMessages, tc.misunderstandings)

			if tc.totalMessages == 0 {
				if got != (ShannonInfo{}) {
					t.Errorf("expected zero-value struct for zero totalMessages, got %+v", got)
				}
				return
			}

			assertClampedComm(t, "SourceEntropy", got.SourceEntropy)
			assertClampedComm(t, "ChannelCapacity", got.ChannelCapacity)
			assertClampedComm(t, "Redundancy", got.Redundancy)
			assertClampedComm(t, "Efficiency", got.Efficiency)
			assertClampedComm(t, "NoiseRatio", got.NoiseRatio)
			assertClampedComm(t, "MutualUnderstanding", got.MutualUnderstanding)
		})
	}
}

func TestComputeShannonInfo_NoMisunderstandings_HighMutual(t *testing.T) {
	got := ComputeShannonInfo(50, 100, 0)
	if got.MutualUnderstanding < 0.5 {
		t.Errorf("MutualUnderstanding = %v for zero misunderstandings, want >= 0.5", got.MutualUnderstanding)
	}
	if got.NoiseRatio > 0.1 {
		t.Errorf("NoiseRatio = %v for zero misunderstandings, want near 0", got.NoiseRatio)
	}
}

func TestComputeShannonInfo_HighMisunderstandings_LowChannelCap(t *testing.T) {
	got := ComputeShannonInfo(10, 50, 40)
	if got.ChannelCapacity > 0.5 {
		t.Errorf("ChannelCapacity = %v for many misunderstandings, want low", got.ChannelCapacity)
	}
}

func TestComputeShannonInfo_RedundancyInverseEntropy(t *testing.T) {
	got := ComputeShannonInfo(50, 100, 0)
	expected := clamp(1-got.SourceEntropy, 0, 1)
	if math.Abs(got.Redundancy-expected) > 1e-9 {
		t.Errorf("Redundancy = %v, want %v", got.Redundancy, expected)
	}
}

func assertClampedComm(t *testing.T, field string, v float64) {
	t.Helper()
	if math.IsNaN(v) || math.IsInf(v, 0) {
		t.Errorf("%s = %v (NaN or Inf)", field, v)
		return
	}
	if v < 0 || v > 1 {
		t.Errorf("%s = %v, want in [0,1]", field, v)
	}
}
