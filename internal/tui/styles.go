package tui

import "github.com/charmbracelet/lipgloss"

// Theme holds semantic color tokens for the entire TUI.
type Theme struct {
	MainFg       lipgloss.Color
	Title        lipgloss.Color
	Subtitle     lipgloss.Color
	MutedText    lipgloss.Color
	SelectedBg   lipgloss.Color
	SelectedFg   lipgloss.Color
	GraphBox     lipgloss.Color
	TableBox     lipgloss.Color
	Border       lipgloss.Color
	BorderActive lipgloss.Color
	Nourishing   lipgloss.Color
	Draining     lipgloss.Color
	Neutral      lipgloss.Color
	Conflicted   lipgloss.Color
	ErrorFg      lipgloss.Color
	SuccessFg    lipgloss.Color
	StatusBarBg  lipgloss.Color
	StatusBarFg  lipgloss.Color
}

// DefaultTheme returns the default color scheme.
func DefaultTheme() Theme {
	return Theme{
		MainFg:       "213", // light purple/pink
		Title:        "183", // light lavender
		Subtitle:     "147", // soft blue
		MutedText:    "245", // dim gray
		SelectedBg:   "61",  // deep blue
		SelectedFg:   "230", // warm white
		GraphBox:     "61",  // muted blue border
		TableBox:     "61",  // muted blue border
		Border:       "59",  // subtle gray-blue
		BorderActive: "111", // brighter blue
		Nourishing:   "86",  // green
		Draining:     "203", // red
		Neutral:      "245", // gray
		Conflicted:   "226", // yellow
		ErrorFg:      "203", // red
		SuccessFg:    "86",  // green
		StatusBarBg:  "236", // dark bg
		StatusBarFg:  "243", // dim fg
	}
}

// Active theme instance (can be swapped for dark/light mode later)
var activeTheme = DefaultTheme()

// T returns the active theme.
func T() Theme { return activeTheme }

// TitleStyle returns bold style with Title color and padding.
func TitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(T().Title).
		Bold(true).
		Padding(0, 1)
}

// SubtitleStyle returns style with Subtitle color and padding.
func SubtitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(T().Subtitle).
		Padding(0, 1)
}

// MutedStyle returns style with MutedText color.
func MutedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(T().MutedText)
}

// SelectedStyle returns bold style with SelectedBg/SelectedFg colors.
func SelectedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(T().SelectedFg).
		Background(T().SelectedBg).
		Bold(true).
		Padding(0, 1)
}

// NormalStyle returns style with MainFg color and padding.
func NormalStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(T().MainFg).
		Padding(0, 1)
}

// BorderStyle returns style with Border color.
func BorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(T().Border)
}

// ActiveBorderStyle returns style with BorderActive color.
func ActiveBorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(T().BorderActive)
}

// ErrorStyle returns bold style with ErrorFg color.
func ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(T().ErrorFg).
		Bold(true)
}

// SuccessStyle returns style with SuccessFg color.
func SuccessStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(T().SuccessFg)
}

// QualityColors maps quality strings to lipgloss colors.
var QualityColors = map[string]lipgloss.Color{
	"nourishing": "86",
	"draining":   "203",
	"neutral":    "245",
	"conflicted": "226",
}

// QualityStyle returns a lipgloss style for the given relationship quality.
func QualityStyle(quality string) lipgloss.Style {
	color, ok := QualityColors[quality]
	if !ok {
		color = QualityColors["neutral"]
	}
	return lipgloss.NewStyle().Foreground(color)
}

// RoundedBorderStyle returns a lipgloss style with rounded border.
func RoundedBorderStyle(borderColor lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)
}

// NormalBorderStyle returns a lipgloss style with normal (sharp) border.
func NormalBorderStyle(borderColor lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)
}
