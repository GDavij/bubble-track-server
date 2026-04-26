package tui

import (
	"fmt"
	"sort"
	"strings"
)

type SortMode int

const (
	SortByName SortMode = iota
	SortByQuality
	SortByStrength
)

func RenderRelationshipTable(edges []Edge, selected int, width int) string {
	if len(edges) == 0 {
		return SimpleBox("Relationships", NormalStyle().Render("No relationships found"), T().TableBox)
	}

	if selected < 0 {
		selected = 0
	}
	if selected >= len(edges) {
		selected = len(edges) - 1
	}

	var content strings.Builder
	for i, edge := range edges {
		line := formatEdge(edge)
		if i == selected {
			content.WriteString(SelectedStyle().Width(width - 2).Render(line))
		} else {
			content.WriteString(NormalStyle().Width(width - 2).Render(line))
		}
		content.WriteString("\n")
	}

	if len(edges) > 10 {
		scrollbar := formatScrollbar(selected, len(edges))
		content.WriteString(MutedStyle().Render(scrollbar))
	}

	return SimpleBox("Relationships", content.String(), T().TableBox)
}

func formatEdge(edge Edge) string {
	dot := QualityStyle(edge.Quality).Render(IconDot)
	strengthBar := formatStrengthBar(edge.Strength, 10)
	strengthPct := int(edge.Strength * 100)

	reciprocity := getReciprocityIndex(edge, []Edge{})
	reciprocityText := fmt.Sprintf("r:%.2f", reciprocity)

	return fmt.Sprintf("%s %s %s %s %s [%s] %d%% %s",
		dot, edge.Source, IconArrow, edge.Target,
		QualityStyle(edge.Quality).Render(edge.Quality),
		strengthBar, strengthPct, reciprocityText,
	)
}

func formatStrengthBar(strength float64, barWidth int) string {
	if barWidth < 1 {
		barWidth = 10
	}
	filled := int(strength * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	if filled < 0 {
		filled = 0
	}
	return strings.Repeat("▓", filled) + strings.Repeat("░", barWidth-filled)
}

func getReciprocityIndex(edge Edge, edges []Edge) float64 {
	for _, e := range edges {
		if e.Source == edge.Target && e.Target == edge.Source {
			return (edge.Strength + e.Strength) / 2
		}
	}
	return 0.0
}

func SortEdges(edges []Edge, mode SortMode) []Edge {
	sorted := make([]Edge, len(edges))
	copy(sorted, edges)

	switch mode {
	case SortByName:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Source < sorted[j].Source
		})
	case SortByQuality:
		qualityOrder := map[string]int{
			"nourishing": 0,
			"conflicted": 1,
			"neutral":    2,
			"draining":   3,
		}
		sort.Slice(sorted, func(i, j int) bool {
			qi, qj := qualityOrder[sorted[i].Quality], qualityOrder[sorted[j].Quality]
			if qi != qj {
				return qi < qj
			}
			return sorted[i].Source < sorted[j].Source
		})
	case SortByStrength:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Strength > sorted[j].Strength
		})
	}

	return sorted
}

func formatScrollbar(selected, total int) string {
	visible := 10
	if total <= visible {
		return ""
	}

	pos := float64(selected) / float64(total-1)
	scrollPos := int(pos * float64(visible-2))

	var sb strings.Builder
	for i := 0; i < visible; i++ {
		if i == scrollPos {
			sb.WriteString("█")
		} else {
			sb.WriteString("░")
		}
	}
	return "\n" + sb.String()
}
