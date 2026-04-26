package application

import (
	"math"

	"github.com/bubbletrack/server/internal/domain"
)

type ClassificationEngine struct{}

func NewClassificationEngine() *ClassificationEngine {
	return &ClassificationEngine{}
}

type InterdisciplinaryRoleScore struct {
	PersonID        string  `json:"person_id"`
	PrimaryRole     domain.SocialRole `json:"primary_role"`
	RoleScores      map[domain.SocialRole]float64 `json:"role_scores"`
	Confidence      float64 `json:"confidence"`
	EvidenceSummary string  `json:"evidence_summary"`
	Disciplines     DisciplinaryEvidence `json:"disciplines"`
}

type DisciplinaryEvidence struct {
	Sociological    float64 `json:"sociological"`
	Philosophical    float64 `json:"philosophical"`
	Psychological   float64 `json:"psychological"`
	Mathematical    float64 `json:"mathematical"`
	Anthropological float64 `json:"anthropological"`
	Economic        float64 `json:"economic"`
	Neuroscientific  float64 `json:"neuroscientific"`
	Communicative    float64 `json:"communicative"`
	Historical      float64 `json:"historical"`
}

func (c *ClassificationEngine) ClassifyRole(m *domain.NodeMetrics) *InterdisciplinaryRoleScore {
	scores := map[domain.SocialRole]float64{
		domain.RoleBridge:   0,
		domain.RoleMentor:   0,
		domain.RoleAnchor:   0,
		domain.RoleCatalyst: 0,
		domain.RoleObserver: 0,
		domain.RoleDrain:     0,
	}

	centrality := (m.Centrality.Betweenness + m.Centrality.Closeness) / 2.0
	bridge := centrality * 0.4 + m.Community.BridgeScore*0.6
	scores[domain.RoleBridge] = bridge

	mentor := m.Community.Embeddedness*0.4 + m.RelationalHealth.Trust*0.3 + (1-m.Community.BridgeScore)*0.3
	scores[domain.RoleMentor] = mentor

	anchor := m.RelationalHealth.OverallScore*0.4 + m.Centrality.Degree*0.3 + m.Community.Embeddedness*0.3
	scores[domain.RoleAnchor] = anchor

	catalyst := m.Community.BridgeScore*0.4 + (1-m.Community.Embeddedness)*0.3 + m.Centrality.Eigenvector*0.3
	scores[domain.RoleCatalyst] = catalyst

	observer := (1-m.Centrality.Degree)*0.4 + (1-m.Community.Embeddedness)*0.3 + m.Centrality.PageRank*0.3
	scores[domain.RoleObserver] = observer

	drain := m.RelationalHealth.ToxicityRisk*0.4 + (1-m.RelationalHealth.Reciprocity)*0.3 + m.RelationalHealth.Communication*0.3
	scores[domain.RoleDrain] = drain

	bestRole := domain.RoleUnknown
	bestScore := -1.0
	totalScore := 0.0
	for role, score := range scores {
		totalScore += score
		if score > bestScore {
			bestScore = score
			bestRole = role
		}
	}

	confidence := 0.0
	if totalScore > 0 {
		confidence = bestScore / totalScore
	}
	confidence = math.Min(confidence*2, 1.0)

	secondBest := 0.0
	for role, score := range scores {
		if role != bestRole && score > secondBest {
			secondBest = score
		}
	}
	if bestScore+secondBest > 0 {
		confidence = math.Min(confidence, 1-bestScore/(bestScore+secondBest+0.01))
	}

	disciplines := DisciplinaryEvidence{
		Sociological:    bridge*0.5 + scores[domain.RoleMentor]*0.3 + scores[domain.RoleAnchor]*0.2,
		Philosophical:    m.HumanistScore.AgencyScore*0.3 + m.HumanistScore.AuthenticityScore*0.4 + m.HumanistScore.MeaningMaking*0.3,
		Psychological:   m.Attachment.Confidence*0.3 + m.RelationalHealth.Trust*0.3 + (1-m.RelationalHealth.ToxicityRisk)*0.4,
		Mathematical:    (m.Centrality.Betweenness+m.Centrality.Eigenvector)/2.0*0.5 + m.Community.BridgeScore*0.3 + m.Community.Embeddedness*0.2,
	Anthropological: m.SocialCapital.BridgingScore*0.4 + m.SocialCapital.BondingScore*0.3 + m.SocialCapital.Diversity*0.3,
		Economic:        m.SocialExchange.Satisfaction*0.4 + m.SocialExchange.Investment*0.3 + (1-m.SocialExchange.CostScore)*0.3,
		Neuroscientific:  m.HumanistScore.EmpathicCapacity*0.3 + m.Community.Embeddedness*0.3 + m.RelationalHealth.Trust*0.2,
		Communicative:    m.HumanistScore.NarrativeCoherence*0.3 + m.RelationalHealth.Communication*0.4 + m.Community.Embeddedness*0.3,
		Historical:      0.5,
	}

	evidence := "primary: "
	switch bestRole {
	case domain.RoleBridge:
		evidence += "high betweenness centrality and community bridging"
	case domain.RoleMentor:
		evidence += "strong community embeddedness with trust and support"
	case domain.RoleAnchor:
		evidence += "stable high-quality connections with consistent presence"
	case domain.RoleCatalyst:
		evidence += "connects disparate groups, high eigenvector centrality"
	case domain.RoleObserver:
		evidence += "low engagement, peripheral network position"
	case domain.RoleDrain:
		evidence += "high toxicity risk and low reciprocity"
	default:
		evidence += "insufficient data"
	}

	return &InterdisciplinaryRoleScore{
		PersonID:        m.PersonID,
		PrimaryRole:     bestRole,
		RoleScores:      scores,
		Confidence:      confidence,
		EvidenceSummary: evidence,
		Disciplines:     disciplines,
	}
}

func (c *ClassificationEngine) ClassifyAllRoles(metrics map[string]*domain.NodeMetrics) map[string]*InterdisciplinaryRoleScore {
	results := make(map[string]*InterdisciplinaryRoleScore, len(metrics))
	for id, m := range metrics {
		results[id] = c.ClassifyRole(m)
	}
	return results
}
