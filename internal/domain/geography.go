package domain

type SpatialProfile struct {
	DistanceDecayBeta  float64  `json:"distance_decay_beta"`
	GravityScore       float64  `json:"gravity_score"`
	ProximityIndex     float64  `json:"proximity_index"`
	TerritorialOverlap float64  `json:"territorial_overlap"`
	MobilityRange      float64  `json:"mobility_range"`
	CentralityPlace    float64  `json:"centrality_place"`
}

type InteractionFrequencyModel struct {
	ObservedFrequency float64 `json:"observed_frequency"`
	PredictedFrequency float64 `json:"predicted_frequency"`
	DistanceDecay      float64 `json:"distance_decay"`
	GravityPrediction float64 `json:"gravity_prediction"`
	Deviation         float64 `json:"deviation"`
	AnomalyScore      float64 `json:"anomaly_score"`
}

type TemporalProximity struct {
	InteractionRecency  float64 `json:"interaction_recency"`
	FrequencyRegularity float64 `json:"frequency_regularity"`
	TimingOverlap       float64 `json:"timing_overlap"`
	RoutineStrength     float64 `json:"routine_strength"`
	SeasonalityScore    float64 `json:"seasonality_score"`
}

func ComputeGravityModel(mass1, mass2 float64, distance float64, beta float64) float64 {
	if distance <= 0 {
		return mass1 * mass2
	}
	if beta <= 0 {
		beta = 2.0
	}
	denominator := powf(distance, beta)
	if denominator == 0 {
		return 0
	}
	return (mass1 * mass2) / denominator
}

func ComputeDistanceDecay(interactionCount int, distance float64, beta float64) float64 {
	if beta <= 0 {
		beta = 1.0
	}
	predicted := float64(interactionCount) / (1.0 + powf(distance, beta))
	return predicted
}

func ComputeProximityIndex(interactionFrequency float64, physicalDistance float64, socialDistance float64) float64 {
	invPhysical := 1.0 / (1.0 + physicalDistance)
	invSocial := 1.0 / (1.0 + socialDistance)
	return clamp((invPhysical*0.4+invSocial*0.6)*interactionFrequency, 0, 1)
}

func ComputeMobilityRange(uniqueLocations int, timeSpanDays int) float64 {
	if timeSpanDays == 0 {
		return 0
	}
	return clamp(float64(uniqueLocations)/float64(timeSpanDays)*10, 0, 1)
}

func powf(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}

func ComputeTemporalProximity(lastInteractionDaysAgo int, avgDaysBetween int, interactions []string) TemporalProximity {
	recency := clamp(1.0/float64(lastInteractionDaysAgo+1), 0, 1)

	regularity := 0.0
	if avgDaysBetween > 0 && len(interactions) > 2 {
		regularity = 1.0 / (1.0 + absf(float64(avgDaysBetween)-7.0)/7.0)
	}

	routine := clamp(regularity*0.6+recency*0.4, 0, 1)

	seasonality := 0.0
	if len(interactions) > 5 {
		seasonality = clamp(float64(len(interactions))/20.0, 0, 1)
	}

	return TemporalProximity{
		InteractionRecency:  recency,
		FrequencyRegularity: regularity,
		TimingOverlap:       0,
		RoutineStrength:     routine,
		SeasonalityScore:    seasonality,
	}
}
