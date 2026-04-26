package domain

import "time"

type AttachmentStyle string

const (
	AttachmentSecure   AttachmentStyle = "secure"
	AttachmentAnxious  AttachmentStyle = "anxious_preoccupied"
	AttachmentAvoidant AttachmentStyle = "dismissive_avoidant"
	AttachmentFearful  AttachmentStyle = "fearful_avoidant"
	AttachmentUnknown  AttachmentStyle = "unknown"
)

type AttachmentProfile struct {
	PersonID         string         `json:"person_id"`
	Style            AttachmentStyle `json:"style"`
	Confidence       float64        `json:"confidence"`
	EvidenceSnippets []string       `json:"evidence_snippets"`
	AssessedAt       time.Time      `json:"assessed_at"`
}

type AttachmentSignals struct {
	ProximitySeeking    float64 `json:"proximity_seeking"`
	SeparationAnxiety   float64 `json:"separation_anxiety"`
	EmotionalAvailability float64 `json:"emotional_availability"`
	SelfDisclosure      float64 `json:"self_disclosure"`
	TrustLevel         float64 `json:"trust_level"`
	ConflictResponse   float64 `json:"conflict_response"`
	Consistency        float64 `json:"consistency"`
}

func ClassifyAttachment(s AttachmentSignals) AttachmentProfile {
	score := map[AttachmentStyle]float64{
		AttachmentSecure:   0,
		AttachmentAnxious:  0,
		AttachmentAvoidant: 0,
		AttachmentFearful:  0,
	}

	secureScore := (s.EmotionalAvailability*0.3 + s.TrustLevel*0.25 + s.Consistency*0.2 +
		s.SelfDisclosure*0.15 + (1-s.SeparationAnxiety)*0.1)
	score[AttachmentSecure] = secureScore

	anxiousScore := (s.SeparationAnxiety*0.3 + s.ProximitySeeking*0.25 + (1-s.TrustLevel)*0.2 +
		s.SelfDisclosure*0.15 + (1-s.Consistency)*0.1)
	score[AttachmentAnxious] = anxiousScore

	avoidantScore := ((1-s.EmotionalAvailability)*0.3 + (1-s.ProximitySeeking)*0.25 +
		(1-s.SelfDisclosure)*0.2 + (1-s.TrustLevel)*0.15 + (1-s.Consistency)*0.1)
	score[AttachmentAvoidant] = avoidantScore

	fearfulScore := (s.SeparationAnxiety*0.2 + (1-s.EmotionalAvailability)*0.2 +
		(1-s.TrustLevel)*0.2 + s.ProximitySeeking*0.2 + (1-s.Consistency)*0.2)
	score[AttachmentFearful] = fearfulScore

	bestStyle := AttachmentSecure
	bestScore := -1.0
	for style, s := range score {
		if s > bestScore {
			bestScore = s
			bestStyle = style
		}
	}

	return AttachmentProfile{
		Style:      bestStyle,
		Confidence: bestScore,
	}
}
