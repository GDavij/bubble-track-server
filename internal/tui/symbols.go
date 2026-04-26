package tui

// BoxChars holds box-drawing characters for borders.
type BoxChars struct {
	HLine     string
	VLine     string
	LeftUp    string
	RightUp   string
	LeftDown  string
	RightDown string
	TopTee    string
	BottomTee string
	LeftTee   string
	RightTee  string
	Cross     string
}

// Rounded and Sharp border variants
var (
	Rounded = BoxChars{
		HLine:     "─",
		VLine:     "│",
		LeftUp:    "╭",
		RightUp:   "╮",
		LeftDown:  "╰",
		RightDown: "╯",
		TopTee:    "┬",
		BottomTee: "┴",
		LeftTee:   "├",
		RightTee:  "┤",
		Cross:     "┼",
	}
	Sharp = BoxChars{
		HLine:     "─",
		VLine:     "│",
		LeftUp:    "┌",
		RightUp:   "┐",
		LeftDown:  "└",
		RightDown: "┘",
		TopTee:    "┬",
		BottomTee: "┴",
		LeftTee:   "├",
		RightTee:  "┤",
		Cross:     "┼",
	}
)

// EdgeChars for graph rendering
var EdgeChars = struct {
	Horizontal string
	Vertical   string
}{
	Horizontal: "─",
	Vertical:   "│",
}

// SpinnerFrames for loading animation
var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// UI icons - using ASCII/Unicode box-drawing characters only, no emojis
const (
	IconDot     = "+"
	IconEmpty   = "o"
	IconArrow   = "->"
	IconDiamond = "<>"
	IconCheck   = "[x]"
	IconCross   = "[ ]"
	IconLock    = "[!]"
)
