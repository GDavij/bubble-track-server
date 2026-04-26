package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Layout holds calculated box dimensions for all TUI regions.
type Layout struct {
	Width   int
	Height  int
	TabBar  TabBarLayout
	Sidebar SidebarLayout
	Graph   GraphLayout
	Status  StatusLayout
}

// TabBarLayout defines the tab bar region dimensions.
type TabBarLayout struct {
	Width  int
	Height int
	Y      int
}

// SidebarLayout defines the sidebar region dimensions.
type SidebarLayout struct {
	X      int
	Y      int
	Width  int
	Height int
}

// GraphLayout defines the graph region dimensions.
type GraphLayout struct {
	X      int
	Y      int
	Width  int
	Height int
}

// StatusLayout defines the status bar region dimensions.
type StatusLayout struct {
	Y      int
	Width  int
	Height int
}

// CalcLayout calculates percentage-based dimensions for all TUI regions.
func CalcLayout(termWidth, termHeight int) Layout {
	// Ensure minimum dimensions to prevent negative values
	if termWidth < 10 {
		termWidth = 10
	}
	if termHeight < 5 {
		termHeight = 5
	}

	// Tab bar: full width, 1 row, at top
	tabBar := TabBarLayout{
		Width:  termWidth,
		Height: 1,
		Y:      0,
	}

	// Status bar: full width, 1 row, at bottom
	status := StatusLayout{
		Y:      termHeight - 1,
		Width:  termWidth,
		Height: 1,
	}

	// Sidebar: 30% of width (min 25, max 40)
	sidebarWidth := termWidth * 30 / 100
	if sidebarWidth < 25 {
		sidebarWidth = 25
	}
	if sidebarWidth > 40 {
		sidebarWidth = 40
	}
	if sidebarWidth >= termWidth-5 {
		sidebarWidth = termWidth / 2
	}

	// Calculate content area height (between tab bar and status bar)
	contentHeight := termHeight - 2
	if contentHeight < 1 {
		contentHeight = 1
	}

	sidebar := SidebarLayout{
		X:      0,
		Y:      1,
		Width:  sidebarWidth,
		Height: contentHeight,
	}

	// Graph: remaining width (70%)
	graphWidth := termWidth - sidebarWidth
	if graphWidth < 5 {
		graphWidth = 5
	}

	graph := GraphLayout{
		X:      sidebarWidth,
		Y:      1,
		Width:  graphWidth,
		Height: contentHeight,
	}

	return Layout{
		Width:   termWidth,
		Height:  termHeight,
		TabBar:  tabBar,
		Sidebar: sidebar,
		Graph:   graph,
		Status:  status,
	}
}

// JoinHorizontal joins two multi-line strings side by side with a separator.
func JoinHorizontal(left, right string, totalWidth int) string {
	separator := lipgloss.NewStyle().
		Foreground(T().Border).
		Render(Rounded.VLine)

	leftStyle := lipgloss.NewStyle().
		Width(totalWidth/2 - 1).
		MaxWidth(totalWidth/2 - 1)

	rightStyle := lipgloss.NewStyle().
		Width(totalWidth/2 - 1).
		MaxWidth(totalWidth/2 - 1)

	leftRendered := leftStyle.Render(left)
	rightRendered := rightStyle.Render(right)

	return lipgloss.JoinHorizontal(lipgloss.Left, leftRendered, separator, rightRendered)
}
