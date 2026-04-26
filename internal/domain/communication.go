package domain

type WatzlawickAxioms struct {
	Axiom1Received   float64 `json:"axiom1_received"`
	Axiom1Quality    float64 `json:"axiom1_quality"`
	Axiom2Content    float64 `json:"axiom2_content"`
	Axiom2Relation   float64 `json:"axiom2_relation"`
	Axiom3Punctuation float64 `json:"axiom3_punctuation"`
	Axiom4Digital     float64 `json:"axiom4_digital"`
	Axiom5Symmetric   float64 `json:"axiom5_symmetric"`
	OverallCompliance float64 `json:"overall_compliance"`
}

type MetaCommunicationProfile struct {
	DoubleBindRisk      float64 `json:"double_bind_risk"`
	ParadoxLevel        float64 `json:"paradox_level"`
	Congruence          float64 `json:"congruence"`
	MetaMessageAccuracy float64 `json:"meta_message_accuracy"`
	FrameControl        float64 `json:"frame_control"`
	Reflexivity         float64 `json:"reflexivity"`
}

type NarrativeProfile struct {
	Coherence         float64 `json:"coherence"`
	Agency            float64 `json:"agency"`
	Complexity        float64 `json:"complexity"`
	Empathy           float64 `json:"empathy"`
	Resolution        float64 `json:"resolution"`
	Trajectory        string  `json:"trajectory"`
	RedemptionArc     float64 `json:"redemption_arc"`
	ContaminationArc   float64 `json:"contamination_arc"`
}

type ShannonInfo struct {
	SourceEntropy     float64 `json:"source_entropy"`
	ChannelCapacity   float64 `json:"channel_capacity"`
	Redundancy        float64 `json:"redundancy"`
	Efficiency        float64 `json:"efficiency"`
	NoiseRatio        float64 `json:"noise_ratio"`
	MutualUnderstanding float64 `json:"mutual_understanding"`
}

func ComputeWatzlawick(received, understood, relevant, timely, mediumRichness, initFrequency int, totalMessages int) WatzlawickAxioms {
	if totalMessages == 0 {
		return WatzlawickAxioms{}
	}
	n := float64(totalMessages)
	a1 := clamp(float64(received)/n, 0, 1)
	a2 := clamp(float64(understood)/n, 0, 1)
	a3 := clamp(float64(relevant)/n, 0, 1)
	a4 := clamp(float64(mediumRichness)/n, 0, 1)
	a5 := clamp(float64(initFrequency)/n, 0, 1)
	compliance := (a1 + a2 + a3 + a4 + a5) / 5.0

	return WatzlawickAxioms{
		Axiom1Received:   a1,
		Axiom1Quality:    clamp(a1*a2, 0, 1),
		Axiom2Content:    a2,
		Axiom2Relation:   a3,
		Axiom3Punctuation: a3,
		Axiom4Digital:     a4,
		Axiom5Symmetric:   a5,
		OverallCompliance: compliance,
	}
}

func ComputeMetaCommunication(incongruenceCount, paradoxCount, congruenceCount, totalInteractions int) MetaCommunicationProfile {
	if totalInteractions == 0 {
		return MetaCommunicationProfile{}
	}
	n := float64(totalInteractions)
	dbRisk := clamp(float64(incongruenceCount)/n*2, 0, 1)
	paradox := clamp(float64(paradoxCount)/n*3, 0, 1)
	congruence := clamp(float64(congruenceCount)/n*1.5, 0, 1)
	metaAcc := clamp(congruence*0.6+(1-dbRisk)*0.4, 0, 1)
	frameCtrl := clamp(dbRisk*0.5+paradox*0.3+(1-congruence)*0.2, 0, 1)
	reflexivity := clamp(metaAcc*0.4+congruence*0.3+(1-paradox)*0.3, 0, 1)

	return MetaCommunicationProfile{
		DoubleBindRisk:      dbRisk,
		ParadoxLevel:        paradox,
		Congruence:          congruence,
		MetaMessageAccuracy: metaAcc,
		FrameControl:        frameCtrl,
		Reflexivity:         reflexivity,
	}
}

func ComputeNarrative(coherence, agency, complexity, empathy, resolution float64) NarrativeProfile {
	trajectory := "stable"
	switch {
	case agency > 0.7 && resolution > 0.7:
		trajectory = "empowerment"
	case agency < 0.3 && complexity > 0.7:
		trajectory = "entrapment"
	case complexity > 0.6 && resolution > 0.5:
		trajectory = "transformation"
	case agency > 0.5 && resolution < 0.3:
		trajectory = "struggle"
	}

	redemption := 0.0
	if agency > 0.4 && empathy > 0.4 {
		redemption = clamp(agency*0.3+empathy*0.3+resolution*0.4, 0, 1)
	}

	contamination := 0.0
	if complexity > 0.6 && resolution < 0.3 {
		contamination = clamp(complexity*0.3+(1-resolution)*0.4+(1-empathy)*0.3, 0, 1)
	}

	return NarrativeProfile{
		Coherence:       clamp(coherence, 0, 1),
		Agency:          clamp(agency, 0, 1),
		Complexity:      clamp(complexity, 0, 1),
		Empathy:         clamp(empathy, 0, 1),
		Resolution:      clamp(resolution, 0, 1),
		Trajectory:      trajectory,
		RedemptionArc:   redemption,
		ContaminationArc: contamination,
	}
}

func ComputeShannonInfo(messageVariety, totalMessages, misunderstandings int) ShannonInfo {
	if totalMessages == 0 {
		return ShannonInfo{}
	}
	sourceEntropy := clamp(float64(messageVariety)/float64(totalMessages)*10, 0, 1)
	channelCap := clamp(1-float64(misunderstandings)/float64(totalMessages)*3, 0, 1)
	redundancy := clamp(1-sourceEntropy, 0, 1)
	efficiency := clamp(channelCap*(1-redundancy*0.5), 0, 1)
	noise := clamp(float64(misunderstandings)/float64(totalMessages)*2, 0, 1)
	mutual := clamp(efficiency*0.5+(1-noise)*0.5, 0, 1)

	return ShannonInfo{
		SourceEntropy:     sourceEntropy,
		ChannelCapacity:   channelCap,
		Redundancy:        redundancy,
		Efficiency:        efficiency,
		NoiseRatio:        noise,
		MutualUnderstanding: mutual,
	}
}
