package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type GraphStats struct {
	TotalPeople         int
	TotalRelationships  int
	AvgReciprocity      float64
	BridgeCount         int
	StrongestConnection string
}

func RenderStatsPanel(stats GraphStats, width int) string {
	t := T()

	lines := []string{
		fmt.Sprintf("  People:        %s", NormalStyle().Render(fmt.Sprintf("%d", stats.TotalPeople))),
		fmt.Sprintf("  Relationships: %s", NormalStyle().Render(fmt.Sprintf("%d", stats.TotalRelationships))),
		fmt.Sprintf("  Reciprocity:    %s", NormalStyle().Render(fmt.Sprintf("%.2f", stats.AvgReciprocity))),
		fmt.Sprintf("  Bridges:       %s", NormalStyle().Render(fmt.Sprintf("%d", stats.BridgeCount))),
	}

	if stats.StrongestConnection != "" {
		lines = append(lines, fmt.Sprintf("  Strongest:     %s", SuccessStyle().Render(stats.StrongestConnection)))
	}

	content := strings.Join(lines, "\n")

	return RenderBox(Box{
		Title:       "Stats",
		Content:     content,
		BorderColor: t.Border,
		Rounded:     true,
		Width:       width - 2,
	})
}

func RenderDetailPanel(edge Edge, allEdges []Edge, width int) string {
	t := T()

	reciprocity := 0.0
	for _, e := range allEdges {
		if e.Source == edge.Target && e.Target == edge.Source {
			reciprocity = (edge.Strength + e.Strength) / 2
			break
		}
	}

	qualityStyled := QualityStyle(edge.Quality).Render(strings.Title(edge.Quality))
	strengthPct := int(edge.Strength * 100)
	bar := formatStrengthBar(edge.Strength, 20)

	lines := []string{
		fmt.Sprintf("  %s", TitleStyle().Render(fmt.Sprintf("%s  %s  %s", edge.Source, IconArrow, edge.Target))),
		"",
		fmt.Sprintf("  Quality:     %s", qualityStyled),
		fmt.Sprintf("  Strength:    %s %d%%", bar, strengthPct),
		fmt.Sprintf("  Reciprocity: %.2f", reciprocity),
	}

	content := strings.Join(lines, "\n")
	boxWidth := min(50, width-4)

	box := RenderBox(Box{
		Title:       "Relationship Detail",
		Content:     content,
		BorderColor: t.BorderActive,
		Rounded:     true,
		Width:       boxWidth,
	})

	closeHint := MutedStyle().Render("Press Esc to close")

	return lipgloss.JoinVertical(lipgloss.Center, box, closeHint)
}
