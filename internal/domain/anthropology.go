package domain

type GiftEconomyProfile struct {
	GiftGiven         int     `json:"gift_given"`
	GiftReceived      int     `json:"gift_received"`
	Balance           float64 `json:"balance"`
	GenerosityIndex   float64 `json:"generosity_index"`
	ReciprocityRate   float64 `json:"reciprocity_rate"`
	ObligationLoad    float64 `json:"obligation_load"`
	CreditAccumulated float64 `json:"credit_accumulated"`
	DebtAccumulated  float64 `json:"debt_accumulated"`
	Status            string  `json:"status"`
}

type KinshipProfile struct {
	FictiveKinCount    int     `json:"fictive_kin_count"`
	ChosenFamilyCount  int     `json:"chosen_family_count"`
	BloodRelationCount int     `json:"blood_relation_count"`
	KinshipDensity     float64 `json:"kinship_density"`
	FamilyRole         string  `json:"family_role"`
	LoyaltyScore       float64 `json:"loyalty_score"`
	ObligationStrength float64 `json:"obligation_strength"`
}

type RitualProfile struct {
	RitualName         string  `json:"ritual_name"`
	Frequency          string  `json:"frequency"`
	Participants       int     `json:"participants"`
	CohesionBoost      float64 `json:"cohesion_boost"`
	Meaningfulness     float64 `json:"meaningfulness"`
	Stability          float64 `json:"stability"`
	InclusionScore     float64 `json:"inclusion_score"`
}

type SymbolicProfile struct {
	SharedSymbols      int     `json:"shared_symbols"`
	InsideJokes        int     `json:"inside_jokes"`
	CulturalCapital    float64 `json:"cultural_capital"`
	MutualUnderstanding float64 `json:"mutual_understanding"`
	IngroupMarkers     int     `json:"ingroup_markers"`
}

func ComputeGiftEconomy(given, received int) GiftEconomyProfile {
	total := given + received
	if total == 0 {
		return GiftEconomyProfile{Status: "no_exchange"}
	}

	generosity := clamp(float64(given)/float64(total), 0, 1)
	reciprocity := 0.0
	if given > 0 {
		reciprocity = clamp(float64(received)/float64(given), 0, 2) / 2.0
	}

	balance := float64(received-given) / float64(total)
	obligation := 0.0
	if given > received {
		obligation = clamp(float64(given-received)/float64(given), 0, 1)
	}

	credit := clamp(float64(max(0, received-given))/10.0, 0, 1)
	debt := clamp(float64(max(0, given-received))/10.0, 0, 1)

	status := "balanced"
	switch {
	case obligation > 0.7:
		status = "overextended"
	case debt > 0.5:
		status = "indebted"
	case credit > 0.5:
		status = "beneficiary"
	case reciprocity > 0.8:
		status = "harmonious"
	}

	return GiftEconomyProfile{
		GiftGiven:         given,
		GiftReceived:      received,
		Balance:           balance,
		GenerosityIndex:   generosity,
		ReciprocityRate:   reciprocity,
		ObligationLoad:    obligation,
		CreditAccumulated: credit,
		DebtAccumulated:  debt,
		Status:            status,
	}
}

func ComputeKinshipProfile(fictive, chosen, blood int, loyaltyInteractions int, obligationInteractions int) KinshipProfile {
	total := fictive + chosen + blood
	if total == 0 {
		return KinshipProfile{}
	}

	density := clamp(float64(total)/20.0, 0, 1)
	chosenRatio := float64(chosen) / float64(total)
	fictiveRatio := float64(fictive) / float64(total)

	role := "nuclear"
	if chosenRatio > 0.6 {
		role = "chosen_family_oriented"
	} else if fictiveRatio > 0.4 {
		role = "community_integrated"
	} else if float64(blood)/float64(total) > 0.8 {
		role = "traditional"
	}

	totalInteractions := loyaltyInteractions + obligationInteractions + 1
	loyalty := clamp(float64(loyaltyInteractions)/float64(totalInteractions), 0, 1)
	obligation := clamp(float64(obligationInteractions)/float64(totalInteractions), 0, 1)

	return KinshipProfile{
		FictiveKinCount:    fictive,
		ChosenFamilyCount:  chosen,
		BloodRelationCount: blood,
		KinshipDensity:     density,
		FamilyRole:         role,
		LoyaltyScore:       loyalty,
		ObligationStrength: obligation,
	}
}

func ComputeSymbolicProfile(sharedSymbols, insideJokes, markers int, interactions int) SymbolicProfile {
	if interactions == 0 {
		return SymbolicProfile{}
	}

	capital := clamp(float64(sharedSymbols+insideJokes+markers)/float64(interactions)*5, 0, 1)
	understanding := clamp(float64(insideJokes)/float64(sharedSymbols+1)*2, 0, 1)

	return SymbolicProfile{
		SharedSymbols:       sharedSymbols,
		InsideJokes:         insideJokes,
		CulturalCapital:     capital,
		MutualUnderstanding: understanding,
		IngroupMarkers:      markers,
	}
}
