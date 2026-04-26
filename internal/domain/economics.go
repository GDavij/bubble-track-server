package domain

type GameTheoryProfile struct {
	CooperationRate    float64 `json:"cooperation_rate"`
	DefectionRate      float64 `json:"defection_rate"`
	PayoffSum          float64 `json:"payoff_sum"`
	AveragePayoff      float64 `json:"average_payoff"`
	NashEquilibrium    float64 `json:"nash_equilibrium"`
	TragedyRisk        float64 `json:"tragedy_risk"`
	ParetoOptimality   float64 `json:"pareto_optimality"`
	Strategy           string  `json:"strategy"`
	StabilityIndex     float64 `json:"stability_index"`
}

type PrisonersDilemmaOutcome struct {
	PlayerACooperated bool    `json:"player_a_cooperated"`
	PlayerBCooperated bool    `json:"player_b_cooperated"`
	PlayerAPayoff     float64 `json:"player_a_payoff"`
	PlayerBPayoff     float64 `json:"player_b_payoff"`
	JointPayoff       float64 `json:"joint_payoff"`
	OutcomeType       string  `json:"outcome_type"`
}

type StagHuntProfile struct {
	CoordinationLevel float64 `json:"coordination_level"`
	TrustThreshold    float64 `json:"trust_threshold"`
	RiskAssessment    float64 `json:"risk_assessment"`
	CollectivePayoff  float64 `json:"collective_payoff"`
}

type RepeatedGameHistory struct {
	Rounds            int                    `json:"rounds"`
	CooperationCount  int                    `json:"cooperation_count"`
	DefectionCount    int                    `json:"defection_count"`
	TitForTatScore    float64                `json:"tit_for_tat_score"`
	GrimTriggerRisk   float64                `json:"grim_trigger_risk"`
	PavlovScore        float64                `json:"pavlov_score"`
	WinStayLoseShift   float64                `json:"win_stay_lose_shift"`
	EvolutionaryFit    float64                `json:"evolutionary_fit"`
}

type PublicGoodsProfile struct {
	ContributionRate    float64 `json:"contribution_rate"`
	FreeRiderRate      float64 `json:"free_rider_rate"`
	PunishmentRate      float64 `json:"punishment_rate"`
	GroupBenefit        float64 `json:"group_benefit"`
	IndividualReturn   float64 `json:"individual_return"`
	Sustainability     float64 `json:"sustainability"`
}

func ComputePrisonersDilemma(aCooperates, bCooperates bool) PrisonersDilemmaOutcome {
	var aPayoff, bPayoff float64
	var outcome string

	if aCooperates && bCooperates {
		aPayoff, bPayoff = 3, 3
		outcome = "mutual_cooperation"
	} else if aCooperates && !bCooperates {
		aPayoff, bPayoff = 0, 5
		outcome = "exploited"
	} else if !aCooperates && bCooperates {
		aPayoff, bPayoff = 5, 0
		outcome = "exploiter"
	} else {
		aPayoff, bPayoff = 1, 1
		outcome = "mutual_defection"
	}

	return PrisonersDilemmaOutcome{
		PlayerACooperated: aCooperates,
		PlayerBCooperated: bCooperates,
		PlayerAPayoff:     aPayoff,
		PlayerBPayoff:     bPayoff,
		JointPayoff:       aPayoff + bPayoff,
		OutcomeType:       outcome,
	}
}

func ComputeGameTheoryProfile(history []PrisonersDilemmaOutcome) GameTheoryProfile {
	if len(history) == 0 {
		return GameTheoryProfile{}
	}

	coopA := 0
	coopB := 0
	totalPayoffA := 0.0
	totalPayoffB := 0.0

	for _, round := range history {
		if round.PlayerACooperated {
			coopA++
		}
		if round.PlayerBCooperated {
			coopB++
		}
		totalPayoffA += round.PlayerAPayoff
		totalPayoffB += round.PlayerBPayoff
	}

	n := float64(len(history))
	coopRate := float64(coopA+coopB) / (2.0 * n)
	defectRate := 1 - coopRate
	avgPayoff := (totalPayoffA + totalPayoffB) / (2.0 * n)

	nashEq := clamp(1.0-coopRate, 0, 1)
	tragedyRisk := clamp(defectRate*1.5-0.5, 0, 1)
	pareto := clamp(float64(coopA+coopB)/(2*n), 0, 1)

	strategy := "mixed"
	if coopRate > 0.8 {
		strategy = "cooperative"
	} else if defectRate > 0.8 {
		strategy = "defective"
	} else if defectRate > coopRate {
		strategy = "cautious"
	}

	stability := 0.0
	if len(history) > 2 {
		recent := history[len(history)-1].OutcomeType
		prev := history[len(history)-2].OutcomeType
		if recent == prev {
			stability = 0.8
		}
		stability += pareto * 0.2
	}

	return GameTheoryProfile{
		CooperationRate:  coopRate,
		DefectionRate:    defectRate,
		PayoffSum:        totalPayoffA + totalPayoffB,
		AveragePayoff:    avgPayoff,
		NashEquilibrium:  nashEq,
		TragedyRisk:      tragedyRisk,
		ParetoOptimality: pareto,
		Strategy:         strategy,
		StabilityIndex:   clamp(stability, 0, 1),
	}
}

func ComputeRepeatedGame(history []PrisonersDilemmaOutcome) RepeatedGameHistory {
	if len(history) == 0 {
		return RepeatedGameHistory{}
	}

	cooperations := 0
	defections := 0
	titForTatHits := 0
	grimTriggerRisk := 0.0

	for i, round := range history {
		if round.PlayerACooperated {
			cooperations++
		} else {
			defections++
		}

		if i > 0 {
			prev := history[i-1]
			if round.PlayerACooperated == prev.PlayerBCooperated {
				titForTatHits++
			}
			if !prev.PlayerBCooperated && !round.PlayerACooperated {
				grimTriggerRisk += 0.2
			}
		}
	}

	n := len(history)
	titForTatScore := clamp(float64(titForTatHits)/float64(n-1), 0, 1)
	pavlov := 0.0
	winStayLoseShift := 0.0

	if len(history) > 1 {
		pavlovCount := 0
		shiftCount := 0
		for i := 1; i < n; i++ {
			prev := history[i-1]
			curr := history[i]
			if prev.PlayerAPayoff >= 3 && curr.PlayerACooperated {
				pavlovCount++
			}
			if prev.PlayerACooperated != curr.PlayerACooperated {
				shiftCount++
			}
		}
		pavlov = clamp(float64(pavlovCount)/float64(n-1), 0, 1)
		winStayLoseShift = clamp(float64(shiftCount)/float64(n-1), 0, 1)
	}

	evolutionaryFit := clamp((titForTatScore*0.4+pavlov*0.3+(1-grimTriggerRisk)*0.3), 0, 1)

	return RepeatedGameHistory{
		Rounds:          n,
		CooperationCount: cooperations,
		DefectionCount:  defections,
		TitForTatScore:  titForTatScore,
		GrimTriggerRisk: clamp(grimTriggerRisk, 0, 1),
		PavlovScore:     pavlov,
		WinStayLoseShift: winStayLoseShift,
		EvolutionaryFit: evolutionaryFit,
	}
}

func ComputePublicGoods(contributions []float64, costPerUnit float64, groupBenefitMultiplier float64) PublicGoodsProfile {
	if len(contributions) == 0 {
		return PublicGoodsProfile{}
	}

	n := float64(len(contributions))
	totalContrib := 0.0
	contributors := 0
	for _, c := range contributions {
		totalContrib += c
		if c > costPerUnit*0.5 {
			contributors++
		}
	}

	contribRate := float64(contributors) / n
	freeRiderRate := 1 - contribRate
	groupBenefit := totalContrib * groupBenefitMultiplier
	individualReturn := groupBenefit / n
	sustainability := clamp(groupBenefit/(totalContrib*costPerUnit+0.01), 0, 1)

	return PublicGoodsProfile{
		ContributionRate:  contribRate,
		FreeRiderRate:    freeRiderRate,
		PunishmentRate:    0,
		GroupBenefit:      groupBenefit,
		IndividualReturn: individualReturn,
		Sustainability:    sustainability,
	}
}
