package domain

type PhilosophicalFramework string

const (
	FrameworkExistentialism  PhilosophicalFramework = "existentialism"
	FrameworkEthicsOfCare    PhilosophicalFramework = "ethics_of_care"
	FrameworkDialogism        PhilosophicalFramework = "dialogism"
	FrameworkVirtueEthics     PhilosophicalFramework = "virtue_ethics"
	FrameworkHermeneutics     PhilosophicalFramework = "hermeneutics"
	FrameworkPhenomenology    PhilosophicalFramework = "phenomenology"
	FrameworkAgentialRealism  PhilosophicalFramework = "agential_realism"
)

type HumanistProfile struct {
	AgencyScore          float64  `json:"agency_score"`
	AuthenticityScore    float64  `json:"authenticity_score"`
	ResponsibilityScore  float64  `json:"responsibility_score"`
	GrowthOrientation    float64  `json:"growth_orientation"`
	SelfAwareness        float64  `json:"self_awareness"`
	EmpathicCapacity     float64  `json:"empathic_capacity"`
	MeaningMaking        float64  `json:"meaning_making"`
	RelationalEthics     float64  `json:"relational_ethics"`
	NarrativeCoherence   float64  `json:"narrative_coherence"`
	FreedomActualization float64  `json:"freedom_actualization"`
}

type ExistentialAnalysis struct {
	RadicalFreedom   float64 `json:"radical_freedom"`
	BadFaithScore   float64 `json:"bad_faith_score"`
	AngstLevel      float64 `json:"angst_level"`
	Authenticity    float64 `json:"authenticity"`
	ProjectEngagement float64 `json:"project_engagement"`
	FinitudeAwareness float64 `json:"finitude_awareness"`
}

type EthicsOfCareAnalysis struct {
	CareReceptivity    float64 `json:"care_receptivity"`
	CareResponsiveness float64 `json:"care_responsiveness"`
	CareCompetence     float64 `json:"care_competence"`
	MutualRecognition  float64 `json:"mutual_recognition"`
	DependencyAcceptance float64 `json:"dependency_acceptance"`
	VoiceAmplification float64 `json:"voice_amplification"`
}

type DialogicProfile struct {
	Polyphony       float64 `json:"polyphony"`
	Heteroglossia   float64 `json:"heteroglossia"`
	Chronotope      string  `json:"chronotope"`
	Answerability   float64 `json:"answerability"`
	Unfinalizability float64 `json:"unfinalizability"`
	ExcessOfVision  float64 `json:"excess_of_vision"`
}

type VirtueProfile struct {
	Phronesis       float64 `json:"phronesis"`        // practical wisdom
	Eudaimonia      float64 `json:"eudaimonia"`       // flourishing
	MeanCourage     float64 `json:"mean_courage"`     // between deficiency and excess
	MeanTemperance  float64 `json:"mean_temperance"`
	MeanJustice     float64 `json:"mean_justice"`
	FriendshipLevel float64 `json:"friendship_level"`  // philia
}

type HermeneuticProfile struct {
	PrejudiceAwareness float64 `json:"prejudice_awareness"`
	FusionOfHorizons   float64 `json:"fusion_of_horizons"`
	HistoricalConsciousness float64 `json:"historical_consciousness"`
	InterpretiveOpenness float64 `json:"interpretive_openness"`
	TraditionEngagement float64 `json:"tradition_engagement"`
}

func ComputeExistentialAnalysis(agency, conformity, avoidance, projectCount int, totalPossibleProjects int) ExistentialAnalysis {
	freedom := clamp(float64(agency)/(float64(agency+conformity+avoidance)+0.01), 0, 1)
	badFaith := clamp((float64(conformity)*1.5+float64(avoidance))/float64(agency+conformity+avoidance+1), 0, 1)
	angst := clamp(freedom*0.7+badFaith*0.3, 0, 1)
	authenticity := clamp(freedom-badFaith, 0, 1)
	engagement := clamp(float64(projectCount)/(float64(totalPossibleProjects)+0.01), 0, 1)
	finitude := clamp(1.0-float64(avoidance)/float64(agency+conformity+avoidance+1), 0, 1)

	return ExistentialAnalysis{
		RadicalFreedom:    freedom,
		BadFaithScore:    badFaith,
		AngstLevel:       angst,
		Authenticity:     authenticity,
		ProjectEngagement: engagement,
		FinitudeAwareness: finitude,
	}
}

func ComputeEthicsOfCare(careReceived, careGiven, careReciprocated, voicesAmplified, voicesSilenced int) EthicsOfCareAnalysis {
	total := careGiven + careReceived + 1
	receptivity := clamp(float64(careReceived)/float64(total), 0, 1)
	responsiveness := clamp(float64(careGiven)/float64(total), 0, 1)
	competence := clamp(float64(careReciprocated)/(float64(careGiven)+1), 0, 1)
	mutualRecognition := clamp(competence*0.5+float64(careReciprocated)/float64(total)*0.5, 0, 1)
	dependency := clamp(1.0-float64(careGiven)/float64(total)*2.0, 0, 1)
	totalVoices := voicesAmplified + voicesSilenced + 1
	voiceAmp := clamp(float64(voicesAmplified)/float64(totalVoices), 0, 1)

	return EthicsOfCareAnalysis{
		CareReceptivity:    receptivity,
		CareResponsiveness: responsiveness,
		CareCompetence:     competence,
		MutualRecognition:  mutualRecognition,
		DependencyAcceptance: dependency,
		VoiceAmplification: voiceAmp,
	}
}

func ComputeVirtueProfile(wisdomActs, courageousActs, temperateActs, justActs int, totalActs int) VirtueProfile {
	if totalActs == 0 {
		return VirtueProfile{}
	}
	phronesis := clamp(float64(wisdomActs)/float64(totalActs)*1.5, 0, 1)
	eudaimonia := clamp((float64(wisdomActs)+float64(courageousActs)+float64(temperateActs)+float64(justActs))/float64(totalActs*4)*4, 0, 1)
	courage := clamp(float64(courageousActs)/float64(totalActs)*2.0, 0, 1)
	temperance := clamp(float64(temperateActs)/float64(totalActs)*2.0, 0, 1)
	justice := clamp(float64(justActs)/float64(totalActs)*2.0, 0, 1)
	friendship := clamp((float64(wisdomActs)+float64(justActs))/float64(totalActs*2)*2, 0, 1)

	return VirtueProfile{
		Phronesis:       phronesis,
		Eudaimonia:      eudaimonia,
		MeanCourage:     courage,
		MeanTemperance:  temperance,
		MeanJustice:     justice,
		FriendshipLevel: friendship,
	}
}

func ComputeHumanistProfile(existential ExistentialAnalysis, care EthicsOfCareAnalysis, virtues VirtueProfile) HumanistProfile {
	return HumanistProfile{
		AgencyScore:          existential.RadicalFreedom,
		AuthenticityScore:    existential.Authenticity,
		ResponsibilityScore:  clamp(care.CareResponsiveness+virtues.MeanJustice, 0, 1),
		GrowthOrientation:    clamp(existential.ProjectEngagement+virtues.Eudaimonia, 0, 1),
		SelfAwareness:        clamp(existential.FinitudeAwareness+virtues.Phronesis, 0, 1),
		EmpathicCapacity:     clamp(care.CareReceptivity+care.MutualRecognition, 0, 1),
		MeaningMaking:        clamp(existential.Authenticity+existential.ProjectEngagement, 0, 1),
		RelationalEthics:     clamp(care.CareCompetence+care.VoiceAmplification, 0, 1),
		NarrativeCoherence:   clamp(virtues.Phronesis*0.4+existential.Authenticity*0.3+care.MutualRecognition*0.3, 0, 1),
		FreedomActualization: clamp(existential.RadicalFreedom*0.4+virtues.MeanCourage*0.3+care.DependencyAcceptance*0.3, 0, 1),
	}
}
