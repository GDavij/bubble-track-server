package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Box defines a bordered panel.
type Box struct {
	Width       int
	Height      int
	Title       string
	BorderColor lipgloss.Color
	Rounded     bool
	Content     string
}

// RenderBox creates a bordered panel with title and content using lipgloss.
// The box auto-sizes to content if Width/Height are 0.
func RenderBox(box Box) string {
	t := T()
	if box.BorderColor == "" {
		box.BorderColor = t.Border
	}

	var style lipgloss.Style
	if box.Rounded {
		style = RoundedBorderStyle(box.BorderColor)
	} else {
		style = NormalBorderStyle(box.BorderColor)
	}

	if box.Title != "" {
		style = style.BorderForeground(box.BorderColor).BorderTop(true)
		style = style.BorderLeft(true).BorderRight(true).BorderBottom(true)
	}

	if box.Width > 0 {
		style = style.Width(box.Width)
	}
	if box.Height > 0 {
		style = style.Height(box.Height)
	}

	if box.Title != "" {
		return style.Render(fmt.Sprintf(" %s ", box.Title) + "\n" + box.Content)
	}
	return style.Render(box.Content)
}

// SimpleBox is a convenience function for a quick bordered box.
func SimpleBox(title, content string, borderColor lipgloss.Color) string {
	return RenderBox(Box{
		Title:       title,
		Content:     content,
		BorderColor: borderColor,
		Rounded:     true,
	})
}
