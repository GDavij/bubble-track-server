package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TabItem represents a single tab in the tab bar.
type TabItem struct {
	Name   string
	Active bool
}

// RenderTabBar renders tabs side by side with active/inactive styling.
func RenderTabBar(tabs []TabItem, width int) string {
	if len(tabs) == 0 {
		return ""
	}

	var renderedTabs []string

	for _, tab := range tabs {
		base := lipgloss.NewStyle().
			Width(14).
			Align(lipgloss.Center).
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1)

		if tab.Active {
			base = base.
				Bold(true).
				Foreground(T().SelectedFg).
				Background(T().SelectedBg).
				BorderForeground(T().BorderActive)
		} else {
			base = base.
				Foreground(T().MainFg).
				BorderForeground(T().Border)
		}

		renderedTabs = append(renderedTabs, base.Render(tab.Name))
	}

	tabRow := lipgloss.JoinHorizontal(lipgloss.Left, renderedTabs...)
	innerWidth := width - 4
	if innerWidth < lipgloss.Width(tabRow) {
		innerWidth = lipgloss.Width(tabRow)
	}

	content := tabRow
	if innerWidth > lipgloss.Width(tabRow) {
		content += strings.Repeat(" ", innerWidth-lipgloss.Width(tabRow))
	}

	return lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(T().Border).
		Padding(0, 1).
		Render(content)
}
