package domain

type ThermodynamicProfile struct {
	SocialEntropy      float64 `json:"social_entropy"`
	OrderParameter     float64 `json:"order_parameter"`
	FreeEnergy         float64 `json:"free_energy"`
	InternalEnergy     float64 `json:"internal_energy"`
	Temperature        float64 `json:"temperature"` // agitation / conflict level
	Pressure           float64 `json:"pressure"`    // social obligations load
	Volume             float64 `json:"volume"`      // available social space
	PhaseState         string  `json:"phase_state"`
	PhaseTransitionRisk float64 `json:"phase_transition_risk"`
}

type ComplexSystemProfile struct {
	EmergenceLevel     float64 `json:"emergence_level"`
	SelfOrganization   float64 `json:"self_organization"`
	CriticalMass       float64 `json:"critical_mass"`
	TippingPoint       float64 `json:"tipping_point"`
	Resilience         float64 `json:"resilience"`
	Redundancy         float64 `json:"redundancy"`
	AdaptiveCapacity   float64 `json:"adaptive_capacity"`
	FeedbackStrength   float64 `json:"feedback_strength"`
	PositiveFeedback   float64 `json:"positive_feedback"`
	NegativeFeedback   float64 `json:"negative_feedback"`
	CascadeRisk        float64 `json:"cascade_risk"`
}

type DiffusionProfile struct {
	InnovationRate     float64 `json:"innovation_rate"`
	AdoptionSpeed      float64 `json:"adoption_speed"`
	Threshold          float64 `json:"threshold"`
	EarlyAdopters      float64 `json:"early_adopters"`
	MajorityAdoption   float64 `json:"majority_adoption"`
	Saturation         float64 `json:"saturation"`
}

type NetworkResilience struct {
	NodeConnectivity   float64 `json:"node_connectivity"`
	EdgeConnectivity   float64 `json:"edge_connectivity"`
	AttackTolerance    float64 `json:"attack_tolerance"`
	Robustness         float64 `json:"robustness"`
	Fragility          float64 `json:"fragility"`
	RecoverySpeed      float64 `json:"recovery_speed"`
	CriticalNodes      int     `json:"critical_nodes"`
	BackupPaths        float64 `json:"backup_paths"`
}

func ComputeSocialEntropy(roleDistribution map[SocialRole]int) float64 {
	total := 0
	for _, count := range roleDistribution {
		total += count
	}
	if total == 0 {
		return 0
	}

	entropy := 0.0
	for _, count := range roleDistribution {
		p := float64(count) / float64(total)
		if p > 0 {
			entropy -= p * log2f(p)
		}
	}
	return entropy
}

func log2f(x float64) float64 {
	if x <= 0 {
		return 0
	}
	ln2 := 0.6931471805599453
	return ln(x) / ln2
}

func ln(x float64) float64 {
	if x <= 0 {
		return 0
	}
	result := 0.0
	term := (x - 1) / (x + 1)
	termSq := term * term
	for n := 0; n < 100; n++ {
		result += term / float64(2*n+1)
		term *= termSq
		if absf(term/float64(2*n+1)) < 1e-15 {
			break
		}
	}
	return 2.0 * result
}

func absf(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func ComputeThermodynamicProfile(entropy, cohesion, conflict, obligations, connections int, maxConnections int) ThermodynamicProfile {
	socialEntropy := clamp(float64(entropy), 0, 1)
	order := clamp(1-socialEntropy, 0, 1)
	temperature := clamp(float64(conflict)/10.0, 0, 1)
	pressure := clamp(float64(obligations)/20.0, 0, 1)
	volume := clamp(float64(maxConnections-connections)/float64(maxConnections+1), 0, 1)

	internalEnergy := temperature*0.4 + pressure*0.3 + (1-volume)*0.3
	freeEnergy := internalEnergy - temperature*socialEntropy*0.5

	phaseState := "ordered"
	transitionRisk := 0.0
	if temperature > 0.7 && socialEntropy > 0.6 {
		phaseState = "transitional"
		transitionRisk = clamp((temperature-0.7)*2+(socialEntropy-0.6)*2, 0, 1)
	}
	if temperature > 0.85 {
		phaseState = "chaotic"
	}

	return ThermodynamicProfile{
		SocialEntropy:      socialEntropy,
		OrderParameter:     order,
		FreeEnergy:         freeEnergy,
		InternalEnergy:     internalEnergy,
		Temperature:        temperature,
		Pressure:           pressure,
		Volume:             volume,
		PhaseState:         phaseState,
		PhaseTransitionRisk: transitionRisk,
	}
}

func ComputeComplexSystem(thermo ThermodynamicProfile, graph GraphMetrics) ComplexSystemProfile {
	emergence := clamp((1-graph.GlobalClustering)*graph.Density+graph.Modularity*0.5, 0, 1)
	selfOrg := clamp(1-graph.Assortativity, 0, 1)
	criticalMass := clamp(graph.AvgDegree*0.1+graph.Density*0.5, 0, 1)
	tipping := clamp(thermo.PhaseTransitionRisk*0.6+criticalMass*0.4, 0, 1)
	connectedBool := 0.0
	if graph.ConnectedComponents > 1 {
		connectedBool = -0.2
	}
	resilience := clamp(1-graph.Density*0.3+graph.GlobalClustering*0.3+connectedBool, 0, 1)
	redundancy := clamp(1-graph.Density, 0, 1)
	adaptive := clamp(resilience*0.5+redundancy*0.3+emergence*0.2, 0, 1)
	feedbackStr := clamp(thermo.Temperature*thermo.SocialEntropy, 0, 1)
	posFeedback := clamp(feedbackStr*0.6, 0, 1)
	negFeedback := clamp(feedbackStr*0.4+thermo.OrderParameter*0.3, 0, 1)
	cascade := clamp(tipping*0.4+posFeedback*0.3+(1-resilience)*0.3, 0, 1)

	return ComplexSystemProfile{
		EmergenceLevel:   emergence,
		SelfOrganization: selfOrg,
		CriticalMass:     criticalMass,
		TippingPoint:     tipping,
		Resilience:       resilience,
		Redundancy:       redundancy,
		AdaptiveCapacity: adaptive,
		FeedbackStrength: feedbackStr,
		PositiveFeedback: posFeedback,
		NegativeFeedback: negFeedback,
		CascadeRisk:      cascade,
	}
}

func ComputeNetworkResilience(nodeCount, edgeCount, articulationPoints int, avgDegree float64) NetworkResilience {
	if nodeCount == 0 {
		return NetworkResilience{}
	}

	nodeConn := clamp(float64(nodeCount-articulationPoints)/float64(nodeCount), 0, 1)
	edgeConn := clamp(1-1.0/(avgDegree+0.01), 0, 1)
	attackTol := clamp(1-float64(articulationPoints)/float64(nodeCount+1), 0, 1)
	robustness := clamp(nodeConn*0.4+edgeConn*0.3+attackTol*0.3, 0, 1)
	fragility := 1 - robustness
	recovery := clamp(robustness*0.5+float64(edgeCount)/float64(nodeCount*2+1)*0.5, 0, 1)
	backupPaths := clamp(avgDegree/float64(nodeCount)*2, 0, 1)

	return NetworkResilience{
		NodeConnectivity: nodeConn,
		EdgeConnectivity: edgeConn,
		AttackTolerance:  attackTol,
		Robustness:       robustness,
		Fragility:        fragility,
		RecoverySpeed:    recovery,
		CriticalNodes:    articulationPoints,
		BackupPaths:      backupPaths,
	}
}

func ComputeDiffusion(adoptedCount, totalPopulation int, timeSteps int) DiffusionProfile {
	if totalPopulation == 0 || timeSteps == 0 {
		return DiffusionProfile{}
	}

	adoptionRate := float64(adoptedCount) / float64(totalPopulation)
	earlyAdopterThreshold := 0.16
	earlyAdopters := clamp(adoptionRate/earlyAdopterThreshold, 0, 1)
	majorityAdoption := clamp((adoptionRate-earlyAdopterThreshold)/(0.5-earlyAdopterThreshold), 0, 1)
	saturation := clamp(adoptionRate, 0, 1)
	innovationRate := clamp(adoptionRate/float64(timeSteps)*10, 0, 1)
	threshold := 0.25

	return DiffusionProfile{
		InnovationRate:   innovationRate,
		AdoptionSpeed:    clamp(adoptionRate/float64(timeSteps)*5, 0, 1),
		Threshold:        threshold,
		EarlyAdopters:    earlyAdopters,
		MajorityAdoption: majorityAdoption,
		Saturation:       saturation,
	}
}
