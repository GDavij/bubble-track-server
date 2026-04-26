package domain

type HistoricalProfile struct {
	PathDependency     float64 `json:"path_dependency"`
	LegacyEffect       float64 `json:"legacy_effect"`
	LockInScore        float64 `json:"lock_in_score"`
	BranchingPointRisk float64 `json:"branching_point_risk"`
	GenerationalShift  float64 `json:"generational_shift"`
	CyclePhase         string  `json:"cycle_phase"`
	AccumulatedMomentum float64 `json:"accumulated_momentum"`
}

type TemporalPattern struct {
	CycleType          string  `json:"cycle_type"`
	CycleLength        float64 `json:"cycle_length"`
	Amplitude          float64 `json:"amplitude"`
	Phase              float64 `json:"phase"`
	Seasonality        float64 `json:"seasonality"`
	TrendSlope         float64 `json:"trend_slope"`
	Predictability     float64 `json:"predictability"`
}

type GenerationalDynamics struct {
	Generation         string  `json:"generation"`
	ValueAlignment     float64 `json:"value_alignment"`
	CommunicationStyle string  `json:"communication_style"`
	ConflictPotential  float64 `json:"conflict_potential"`
	KnowledgeTransfer  float64 `json:"knowledge_transfer"`
}

type CyclicTheoryProfile struct {
	HoweStraussPhase   string  `json:"howe_strauss_phase"`
	TurningsCount      int     `json:"turnings_count"`
	InstitutionalTrust float64 `json:"institutional_trust"`
	Individualism      float64 `json:"individualism"`
	CrisisSeverity    float64 `json:"crisis_severity"`
}

func ComputePathDependency(pastDecisions int, repeatedPatterns int, totalDecisions int, sunkCostEvidence int) HistoricalProfile {
	if totalDecisions == 0 {
		return HistoricalProfile{}
	}
	pathDep := clamp(float64(repeatedPatterns)/float64(totalDecisions)*2, 0, 1)
	legacy := clamp(float64(pastDecisions)/float64(totalDecisions+1)*0.5+pathDep*0.5, 0, 1)
	lockIn := clamp(pathDep*0.6+legacy*0.4, 0, 1)
	branching := clamp(1-lockIn, 0, 1)
	genShift := clamp(float64(sunkCostEvidence)/float64(totalDecisions+1)*3, 0, 1)

	phase := "maintenance"
	if branching > 0.7 {
		phase = "transition"
	} else if lockIn > 0.7 {
		phase = "lock_in"
	}

	momentum := clamp(legacy*0.4+lockIn*0.3+pathDep*0.3, 0, 1)

	return HistoricalProfile{
		PathDependency:      pathDep,
		LegacyEffect:        legacy,
		LockInScore:         lockIn,
		BranchingPointRisk:  branching,
		GenerationalShift:   genShift,
		CyclePhase:          phase,
		AccumulatedMomentum: momentum,
	}
}

func DetectTemporalCycle(values []float64, windowSize int) TemporalPattern {
	if len(values) < windowSize*2 {
		return TemporalPattern{CycleType: "insufficient_data"}
	}

	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))

	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(values))
	std := 0.0
	if variance > 0 {
		std = sqrt(variance)
	}
	amplitude := std * 2

	autocorr := 0.0
	if len(values) > 1 && std > 0 {
		var num, den float64
		for i := 1; i < len(values); i++ {
			devI := values[i] - mean
			devIm1 := values[i-1] - mean
			num += devI * devIm1
			den += devI * devI
		}
		if den > 0 {
			autocorr = num / den
		}
	}

	seasonality := clamp(autocorr, 0, 1)
	predictability := clamp(1-variance/(mean*mean+0.01), 0, 1)

	trendSlope := 0.0
	if len(values) >= 2 {
	var num, den float64
	for i := 0; i < len(values); i++ {
		num += float64(i) * (values[i] - mean)
		den += float64(i) * float64(i)
	}
	if den > 0 {
		trendSlope = num / den
	}
		if den > 0 {
			trendSlope = num / den
		}
	}

	cycleType := "irregular"
	if seasonality > 0.6 {
		cycleType = "seasonal"
	} else if predictability > 0.7 {
		cycleType = "stable"
	} else if absf(trendSlope) > std*0.5 {
		if trendSlope > 0 {
			cycleType = "growth"
		} else {
			cycleType = "decline"
		}
	}

	return TemporalPattern{
		CycleType:      cycleType,
		CycleLength:    float64(windowSize),
		Amplitude:      amplitude,
		Phase:          0,
		Seasonality:    seasonality,
		TrendSlope:     trendSlope,
		Predictability: predictability,
	}
}

func ComputeGenerationalDynamics(valueOverlap float64, techComfortA, techComfortB float64, conflictEvents int, knowledgeSharing int) GenerationalDynamics {
	alignment := clamp(valueOverlap, 0, 1)
	style := "mixed"
	if techComfortA > 0.7 && techComfortB < 0.3 {
		style = "digital_divide"
	} else if absf(techComfortA-techComfortB) < 0.2 {
		style = "similar"
	}
	conflict := clamp(float64(conflictEvents)/20.0, 0, 1)
	transfer := clamp(float64(knowledgeSharing)/30.0, 0, 1)

	return GenerationalDynamics{
		ValueAlignment:    alignment,
		CommunicationStyle: style,
		ConflictPotential: conflict,
		KnowledgeTransfer: transfer,
	}
}

func ComputeHoweStrauss(institutionalTrust float64, individualism float64, crisisLevel float64) CyclicTheoryProfile {
	phase := "high"
	if crisisLevel > 0.7 && institutionalTrust < 0.3 {
		phase = "crisis"
	} else if individualism > 0.6 && institutionalTrust < 0.5 {
		phase = "awakening"
	} else if institutionalTrust > 0.5 && crisisLevel < 0.3 {
		phase = "high"
	}

	turnings := 0
	switch phase {
	case "crisis":
		turnings = 4
	case "awakening":
		turnings = 1
	case "high":
		turnings = 3
	}

	return CyclicTheoryProfile{
		HoweStraussPhase:   phase,
		TurningsCount:      turnings,
		InstitutionalTrust: institutionalTrust,
		Individualism:      individualism,
		CrisisSeverity:    crisisLevel,
	}
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 20; i++ {
		z = (z + x/z) / 2.0
	}
	return z
}
