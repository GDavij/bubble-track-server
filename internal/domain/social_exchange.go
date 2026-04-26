package domain

type SocialExchangeProfile struct {
	PersonID        string  `json:"person_id"`
	BenefitScore    float64 `json:"benefit_score"`
	CostScore       float64 `json:"cost_score"`
	OutcomeValue    float64 `json:"outcome_value"`
	ComparisonLevel float64 `json:"comparison_level"`
	Satisfaction    float64 `json:"satisfaction"`
	Stability       float64 `json:"stability"`
	Investment      float64 `json:"investment"`
}

type ExchangeFactors struct {
	EmotionalSupport float64 `json:"emotional_support"`
	PracticalHelp    float64 `json:"practical_help"`
	IntellectualStim float64 `json:"intellectual_stimulation"`
	JoyFun          float64 `json:"joy_fun"`
	TrustSafety     float64 `json:"trust_safety"`
	EmotionalDrain  float64 `json:"emotional_drain"`
	TimeCost        float64 `json:"time_cost"`
	ConflictFreq    float64 `json:"conflict_frequency"`
	Unreciprocated  float64 `json:"unreciprocated_effort"`
	PowerImbalance  float64 `json:"power_imbalance"`
}

func ComputeSocialExchange(f ExchangeFactors) SocialExchangeProfile {
	benefits := (f.EmotionalSupport*0.3 + f.PracticalHelp*0.15 + f.IntellectualStim*0.15 +
		f.JoyFun*0.2 + f.TrustSafety*0.2)

	costs := (f.EmotionalDrain*0.3 + f.TimeCost*0.15 + f.ConflictFreq*0.2 +
		f.Unreciprocated*0.2 + f.PowerImbalance*0.15)

	outcome := benefits - costs
	cl := 0.5
	satisfaction := clamp(outcome-cl, -1, 1)
	stability := clamp(1.0-f.ConflictFreq-f.Unreciprocated, 0, 1)
	investment := clamp(benefits/(costs+0.01)-1, -1, 1)

	return SocialExchangeProfile{
		BenefitScore:    benefits,
		CostScore:       costs,
		OutcomeValue:    outcome,
		ComparisonLevel: cl,
		Satisfaction:    satisfaction,
		Stability:       stability,
		Investment:      investment,
	}
}

type RelationshipHealth struct {
	OverallScore    float64 `json:"overall_score"`
	Reciprocity     float64 `json:"reciprocity"`
	Communication   float64 `json:"communication"`
	Trust           float64 `json:"trust"`
	ConflictMgmt    float64 `json:"conflict_management"`
	GrowthPotential float64 `json:"growth_potential"`
	ToxicityRisk    float64 `json:"toxicity_risk"`
	Recommendation  string  `json:"recommendation"`
}

func AssessRelationshipHealth(exchange SocialExchangeProfile, reciprocity float64) RelationshipHealth {
	reciprocityScore := clamp(reciprocity, 0, 1)
	communication := clamp(exchange.OutcomeValue*0.5+exchange.Satisfaction*0.5, 0, 1)
	trust := clamp(exchange.Stability, 0, 1)
	conflictMgmt := clamp(1-exchange.CostScore*0.5, 0, 1)
	growth := clamp(exchange.Investment*0.3+exchange.BenefitScore*0.3+exchange.Satisfaction*0.4, 0, 1)
	toxicity := clamp(exchange.CostScore*0.4+exchange.CostScore*0.3+exchange.CostScore*0.3, 0, 1)

	overall := (reciprocityScore*0.25 + communication*0.2 + trust*0.2 + conflictMgmt*0.15 + growth*0.1 + (1-toxicity)*0.1)

	rec := "maintain"
	switch {
	case overall >= 0.8:
		rec = "nurture — this is a high-quality connection"
	case overall >= 0.6:
		rec = "maintain — solid foundation, minor improvements possible"
	case overall >= 0.4:
		rec = "evaluate — check if effort matches return"
	case overall >= 0.2:
		rec = "consider boundaries — signs of imbalance"
	default:
		rec = "protect yourself — this connection may be harmful"
	}

	return RelationshipHealth{
		OverallScore:    overall,
		Reciprocity:     reciprocityScore,
		Communication:   communication,
		Trust:           trust,
		ConflictMgmt:    conflictMgmt,
		GrowthPotential: growth,
		ToxicityRisk:    toxicity,
		Recommendation:  rec,
	}
}
