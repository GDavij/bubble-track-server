package domain

// Graph represents the full social graph for a user.
// Used for returning the graph to the frontend for visualization.
type Graph struct {
	UserID         string           `json:"user_id"`
	Nodes          []GraphNode      `json:"nodes"`
	Edges          []GraphEdge      `json:"edges"`
	Stats          GraphStats       `json:"stats"`
}

// GraphNode is a person rendered as a graph node.
type GraphNode struct {
	ID               string     `json:"id"`
	DisplayName      string     `json:"display_name"`
	SocialRole       SocialRole `json:"social_role"`
	CurrentMood      Mood       `json:"current_mood,omitempty"`
	CurrentEnergy    float64    `json:"current_energy"`
	InteractionCount int        `json:"interaction_count"`
}

type GraphEdge struct {
	Source           string    `json:"source"`
	Target           string    `json:"target"`
	Quality          Quality   `json:"quality"`
	Strength         float64   `json:"strength"`
	SourceWeight     float64   `json:"source_weight"`
	TargetWeight     float64   `json:"target_weight"`
	Protocol         Protocol  `json:"protocol,omitempty"`
	Label            string    `json:"label,omitempty"`
	ReciprocityIndex float64   `json:"reciprocity_index"`
}

type GraphStats struct {
	TotalPeople         int     `json:"total_people"`
	TotalRelationships  int     `json:"total_relationships"`
	AvgReciprocity      float64 `json:"avg_reciprocity"`
	BridgeCount         int     `json:"bridge_count"`
	StrongestConnection string  `json:"strongest_connection,omitempty"`
}
