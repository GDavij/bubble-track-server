package domain

type SocialCapital struct {
	BondingScore float64 `json:"bonding_score"`
	BridgingScore float64 `json:"bridging_score"`
	LinkingScore float64 `json:"linking_score"`
	TotalCapital float64 `json:"total_capital"`
	Diversity    float64 `json:"diversity"`
	Redundancy   float64 `json:"redundancy"`
}

type WeakTieAnalysis struct {
	BridgeEdges      int     `json:"bridge_edges"`
	BridgeRatio      float64 `json:"bridge_ratio"`
	InformationFlow  float64 `json:"information_flow"`
	NoveltyAccess    float64 `json:"novelty_access"`
	GranovetterScore float64 `json:"granovetter_score"`
}

type HomophilyProfile struct {
	AttributeSimilarity float64 `json:"attribute_similarity"`
	BehavioralHomophily float64 `json:"behavioral_homophily"`
	EchoChamberRisk    float64 `json:"echo_chamber_risk"`
	DiversityGain      float64 `json:"diversity_gain"`
}

type StructuralHoleProfile struct {
	EntrepreneurialIndex float64 `json:"entrepreneurial_index"`
	BrokerageScore       float64 `json:"brokerage_score"`
	ConstraintScore      float64 `json:"constraint_score"`
	NetworkEfficiency    float64 `json:"network_efficiency"`
}

type SocialIdentityProfile struct {
	InGroupStrength     float64 `json:"in_group_strength"`
	OutGroupOpenness    float64 `json:"out_group_openness"`
	IdentityComplexity  float64 `json:"identity_complexity"`
	CategorizationRigor float64 `json:"categorization_rigor"`
}

type GroupDynamicsProfile struct {
	Stage           string  `json:"stage"` // forming, storming, norming, performing, adjourning
	Cohesion        float64 `json:"cohesion"`
	Polarization    float64 `json:"polarization"`
	RoleClarity     float64 `json:"role_clarity"`
	DecisionQuality float64 `json:"decision_quality"`
	GroupthinkRisk  float64 `json:"groupthink_risk"`
}

type DunbarAnalysis struct {
	IntimateCount    int `json:"intimate_count"`    // ~5
	CloseCount       int `json:"close_count"`       // ~15
	AcquaintanceCount int `json:"acquaintance_count"` // ~50
	RecognizableCount int `json:"recognizable_count"` // ~150
	NetworkPressure  float64 `json:"network_pressure"`
	CapacityUtilization float64 `json:"capacity_utilization"`
}

func ComputeSocialCapital(bondingInteractions, bridgingInteractions, linkingInteractions int, uniqueGroups int) SocialCapital {
	total := bondingInteractions + bridgingInteractions + linkingInteractions
	if total == 0 {
		return SocialCapital{}
	}

	bonding := float64(bondingInteractions) / float64(total)
	bridging := float64(bridgingInteractions) / float64(total)
	linking := float64(linkingInteractions) / float64(total)

	diversity := clamp(float64(uniqueGroups)/10.0, 0, 1)
	redundancy := clamp(bonding*1.5-0.5, 0, 1)

	totalCapital := bonding*0.3 + bridging*0.5 + linking*0.2

	return SocialCapital{
		BondingScore: bonding,
		BridgingScore: bridging,
		LinkingScore: linking,
		TotalCapital: totalCapital,
		Diversity:    diversity,
		Redundancy:   redundancy,
	}
}

func AnalyzeWeakTies(totalEdges, bridgeEdges int, clusterCount int) WeakTieAnalysis {
	if totalEdges == 0 {
		return WeakTieAnalysis{}
	}

	bridgeRatio := float64(bridgeEdges) / float64(totalEdges)
	informationFlow := clamp(bridgeRatio*1.5, 0, 1)
	noveltyAccess := clamp(float64(clusterCount)*bridgeRatio*0.5, 0, 1)
	granovetter := clamp(bridgeRatio*0.6+noveltyAccess*0.4, 0, 1)

	return WeakTieAnalysis{
		BridgeEdges:      bridgeEdges,
		BridgeRatio:      bridgeRatio,
		InformationFlow:  informationFlow,
		NoveltyAccess:    noveltyAccess,
		GranovetterScore: granovetter,
	}
}

func AnalyzeHomophily(inGroupEdgeRatio, outGroupEdgeRatio float64, attributeOverlap float64) HomophilyProfile {
	behavioral := clamp(inGroupEdgeRatio, 0, 1)
	echoRisk := clamp(behavioral*1.5-0.5, 0, 1)
	diversityGain := clamp(outGroupEdgeRatio*2.0, 0, 1)

	return HomophilyProfile{
		AttributeSimilarity: attributeOverlap,
		BehavioralHomophily: behavioral,
		EchoChamberRisk:    echoRisk,
		DiversityGain:      diversityGain,
	}
}

func AnalyzeStructuralHoles(brokerageOpportunities, totalConnections, networkSize int) StructuralHoleProfile {
	if totalConnections == 0 || networkSize == 0 {
		return StructuralHoleProfile{}
	}

	brokerage := float64(brokerageOpportunities) / float64(totalConnections)
	constraint := 1.0 - brokerage
	efficiency := clamp(float64(totalConnections)/float64(networkSize)*brokerage, 0, 1)
	entrepreneurial := clamp(brokerage*0.6+efficiency*0.4, 0, 1)

	return StructuralHoleProfile{
		EntrepreneurialIndex: entrepreneurial,
		BrokerageScore:       brokerage,
		ConstraintScore:      constraint,
		NetworkEfficiency:    efficiency,
	}
}

func AnalyzeDunbarLayer(intimate, close, acquaintances int) DunbarAnalysis {
	recognizable := intimate + close + acquaintances
	pressure := 0.0
	if close > 20 {
		pressure += float64(close-20) * 0.05
	}
	if acquaintances > 60 {
		pressure += float64(acquaintances-60) * 0.02
	}
	pressure = clamp(pressure, 0, 1)

	capacity := clamp(float64(recognizable)/150.0, 0, 1)

	return DunbarAnalysis{
		IntimateCount:     intimate,
		CloseCount:        close,
		AcquaintanceCount: acquaintances,
		RecognizableCount: recognizable,
		NetworkPressure:   pressure,
		CapacityUtilization: capacity,
	}
}
