package domain

type NeuralSocialProfile struct {
	MirrorActivity     float64 `json:"mirror_activity"`
	OxytocinProxy     float64 `json:"oxytocin_proxy"`
	DopamineProxy     float64 `json:"dopamine_proxy"`
	CortisolProxy     float64 `json:"cortisol_proxy"`
	SerotoninProxy    float64 `json:"serotonin_proxy"`
	EmpathyPrediction float64 `json:"empathy_prediction"`
	BondingReadiness  float64 `json:"bonding_readiness"`
	StressResponse    float64 `json:"stress_response"`
	RewardPrediction  float64 `json:"reward_prediction"`
	NeuralSynchrony   float64 `json:"neural_synchrony"`
}

type EmotionalRegulationProfile struct {
	Awareness         float64 `json:"awareness"`
	Reappraisal        float64 `json:"reappraisal"`
	Suppression       float64 `json:"suppression"`
	Acceptance        float64 `json:"acceptance"`
	EQScore           float64 `json:"eq_score"`
	AlexithymiaRisk   float64 `json:"alexithymia_risk"`
	EmotionalLability float64 `json:"emotional_lability"`
}

type CircadianSocialProfile struct {
	SocialPeakHour    string  `json:"social_peak_hour"`
	InteractionRhythm string  `json:"interaction_rhythm"`
	CircadianAlign    float64 `json:"circadian_align"`
	BurnoutRisk       float64 `json:"burnout_risk"`
	SocialBattery     float64 `json:"social_battery"`
	RecoveryPattern   string  `json:"recovery_pattern"`
}

func ComputeNeuralSocial(proximity, positiveInteractions, negativeInteractions, sharedActivities int) NeuralSocialProfile {
	total := positiveInteractions + negativeInteractions + 1
	mirror := clamp(float64(proximity+sharedActivities)/float64(total+10)*5, 0, 1)
	oxytocin := clamp(float64(proximity+sharedActivities)/20.0, 0, 1)
	dopamine := clamp(float64(positiveInteractions)/float64(total)*2.0, 0, 1)
	cortisol := clamp(float64(negativeInteractions)/float64(total)*2, 0, 1)
	serotonin := clamp(1-cortisol+dopamine*0.3, 0, 1)
	empathy := clamp(mirror*0.4+oxytocin*0.3+serotonin*0.3, 0, 1)
	bonding := clamp(oxytocin*0.5+mirror*0.3+dopamine*0.2, 0, 1)
	stress := clamp(cortisol*0.6+(1-serotonin)*0.2+(1-bonding)*0.2, 0, 1)
	reward := clamp(dopamine*0.5+oxytocin*0.3+(1-stress)*0.2, 0, 1)
	synchrony := clamp(mirror*0.4+float64(sharedActivities)*0.3+oxytocin*0.3, 0, 1)

	return NeuralSocialProfile{
		MirrorActivity:     mirror,
		OxytocinProxy:     oxytocin,
		DopamineProxy:     dopamine,
		CortisolProxy:     cortisol,
		SerotoninProxy:    serotonin,
		EmpathyPrediction: empathy,
		BondingReadiness:  bonding,
		StressResponse:    stress,
		RewardPrediction:  reward,
		NeuralSynchrony:   synchrony,
	}
}

func ComputeEmotionalRegulation(selfReports int, totalEmotionalEvents int, conflictResolutionRate float64) EmotionalRegulationProfile {
	if totalEmotionalEvents == 0 {
		return EmotionalRegulationProfile{}
	}

	awareness := clamp(float64(selfReports)/float64(totalEmotionalEvents)*2, 0, 1)
	reappraisal := clamp(conflictResolutionRate*1.2, 0, 1)
	suppression := clamp(1-reappraisal, 0, 1)
	acceptance := clamp(awareness*0.4+reappraisal*0.4+(1-suppression)*0.2, 0, 1)
	eq := clamp((awareness+reappraisal+acceptance)/3.0, 0, 1)
	alexithymia := clamp(1-awareness*0.5-(1-suppression)*0.5, 0, 1)
	lability := clamp(suppression*0.5+alexithymia*0.3+(1-eq)*0.2, 0, 1)

	return EmotionalRegulationProfile{
		Awareness:         awareness,
		Reappraisal:        reappraisal,
		Suppression:       suppression,
		Acceptance:        acceptance,
		EQScore:           eq,
		AlexithymiaRisk:   alexithymia,
		EmotionalLability: lability,
	}
}

func ComputeBurnoutRisk(socialObligations, recoveryTime, conflictEvents, totalEvents int) CircadianSocialProfile {
	if totalEvents == 0 {
		return CircadianSocialProfile{}
	}

	obligationLoad := clamp(float64(socialObligations)/20.0, 0, 1)
	recoveryRatio := clamp(1.0-float64(recoveryTime)/float64(totalEvents+1)*10, 0, 1)
	conflictRatio := clamp(float64(conflictEvents)/float64(totalEvents), 0, 1)
	battery := clamp(recoveryRatio*0.5+(1-obligationLoad)*0.3+(1-conflictRatio)*0.2, 0, 1)
	burnout := clamp(obligationLoad*0.4+conflictRatio*0.3+(1-battery)*0.3, 0, 1)

	recovery := "adequate"
	if burnout > 0.7 {
		recovery = "insufficient"
	} else if burnout > 0.4 {
		recovery = "minimal"
	}

	return CircadianSocialProfile{
		SocialPeakHour:    "",
		InteractionRhythm: "",
		CircadianAlign:    clamp(battery*0.5+recoveryRatio*0.5, 0, 1),
		BurnoutRisk:       burnout,
		SocialBattery:     battery,
		RecoveryPattern:   recovery,
	}
}
