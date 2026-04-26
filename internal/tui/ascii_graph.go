package tui

import (
	"strings"
)

type renderEdge struct {
	Source   string
	Target   string
	Quality  string
	Reversed bool
}

type nodeRender struct {
	ID     string
	Name   string
	Mood   string
	Energy float64
	Role   string
	X      int
	Y      int
	Width  int
	Height int
}

func RenderASCIIGraph(data *GraphData, gridWidth, gridHeight int, highlightNodes map[string]bool) string {
	if len(data.Nodes) == 0 {
		return "\nNo nodes to display"
	}

	if gridWidth < 20 || gridHeight < 5 {
		gridWidth = 20
		gridHeight = 5
	}

	positions := LayoutGraph(data.Nodes, data.Edges, gridWidth, gridHeight)

	nodeMap := make(map[string]*nodeRender)
	nodeWidth := 12
	nodeHeight := 3

	for _, pos := range positions {
		name := ""
		mood := ""
		energy := 0.0
		role := ""
		for _, n := range data.Nodes {
			if n.ID == pos.ID {
				name = n.Name
				mood = n.Mood
				energy = n.Energy
				role = n.Role
				break
			}
		}
		nodeMap[pos.ID] = &nodeRender{
			ID:     pos.ID,
			Name:   name,
			Mood:   mood,
			Energy: energy,
			Role:   role,
			X:      pos.X,
			Y:      pos.Y,
			Width:  nodeWidth,
			Height: nodeHeight,
		}
	}

	edges := make([]renderEdge, len(data.Edges))
	for i, e := range data.Edges {
		edges[i] = renderEdge{Source: e.Source, Target: e.Target, Quality: e.Quality}
	}

	grid := make([][]rune, gridHeight)
	for i := range grid {
		grid[i] = make([]rune, gridWidth)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	drawEdges(grid, edges, nodeMap, gridWidth, gridHeight)
	drawNodes(grid, nodeMap, highlightNodes)

	var b strings.Builder
	for _, row := range grid {
		b.WriteString(string(row))
		b.WriteString("\n")
	}

	return TitleStyle().Render("Graph View") + "\n\n" + b.String()
}

func drawNodes(grid [][]rune, nodes map[string]*nodeRender, highlightNodes map[string]bool) {
	chars := Rounded

	for _, node := range nodes {
		if node.X < 0 || node.Y < 0 || node.X+node.Width > len(grid[0]) || node.Y+node.Height > len(grid) {
			continue
		}

		name := node.Name
		if len(name) > 10 {
			name = name[:9] + "…"
		}
		padding := strings.Repeat(" ", 10-len(name))

		top := chars.LeftUp + strings.Repeat(chars.HLine, node.Width-2) + chars.RightUp
		mid := chars.VLine + " " + name + padding + " " + chars.VLine
		bot := chars.LeftDown + strings.Repeat(chars.HLine, node.Width-2) + chars.RightDown

		if node.Role != "" || node.Mood != "" {
			roleChar := getRoleLetter(node.Role)
			moodChar := getMoodChar(node.Mood)
			energyBar := getEnergyBar(node.Energy)

			statusLine := roleChar + " " + moodChar + " " + energyBar
			statusPadding := strings.Repeat(" ", 10-len(statusLine))
			mid = chars.VLine + " " + statusLine + statusPadding + " " + chars.VLine
		}

		isHighlighted := highlightNodes != nil && highlightNodes[node.ID]

		lines := []string{top, mid, bot}

		for r, line := range lines {
			row := node.Y + r
			for c, ch := range line {
				col := node.X + c
				if row >= 0 && row < len(grid) && col >= 0 && col < len(grid[0]) {
					if isHighlighted && r == 1 && c > 1 && c < len(line)-2 {
						grid[row][col] = ch
					} else if !isHighlighted {
						grid[row][col] = ch
					}
				}
			}
		}

		if isHighlighted {
			for r, line := range lines {
				row := node.Y + r
				for c, ch := range line {
					col := node.X + c
					if row >= 0 && row < len(grid) && col >= 0 && col < len(grid[0]) {
						if r == 0 || r == len(lines)-1 {
							grid[row][col] = ch
						} else if c == 0 || c == len(line)-1 {
							grid[row][col] = ch
						}
					}
				}
			}
		}
	}
}

func getRoleLetter(role string) string {
	switch role {
	case "bridge", "Bridge":
		return "B"
	case "mentor", "Mentor":
		return "M"
	case "anchor", "Anchor":
		return "A"
	case "catalyst", "Catalyst":
		return "C"
	case "observer", "Observer":
		return "O"
	case "drain", "Drain":
		return "D"
	default:
		return "?"
	}
}

func getMoodChar(mood string) string {
	switch mood {
	case "happy", "Happy":
		return ":)"
	case "anxious", "Anxious":
		return "!"
	case "tired", "Tired":
		return "z"
	case "energized", "Energized":
		return "*"
	case "sad", "Sad":
		return ":("
	case "neutral", "Neutral":
		return "-"
	case "angry", "Angry":
		return "@"
	case "hopeful", "Hopeful":
		return "^"
	case "lonely", "Lonely":
		return "."
	case "grateful", "Grateful":
		return "+"
	default:
		return "?"
	}
}

func getEnergyBar(energy float64) string {
	if energy < 0 {
		energy = 0
	} else if energy > 1 {
		energy = 1
	}

	filled := int(energy * 3)
	empty := 3 - filled

	return strings.Repeat("#", filled) + strings.Repeat(".", empty)
}

func drawEdges(grid [][]rune, edges []renderEdge, nodes map[string]*nodeRender, gridWidth, gridHeight int) {
	chars := Rounded

	for _, edge := range edges {
		src, srcOk := nodes[edge.Source]
		tgt, tgtOk := nodes[edge.Target]

		if !srcOk || !tgtOk || src.ID == tgt.ID {
			continue
		}

		srcX, srcY := src.X+src.Width/2, src.Y+src.Height/2
		tgtX, tgtY := tgt.X+tgt.Width/2, tgt.Y+tgt.Height/2

		if srcX > tgtX {
			srcX, tgtX = tgtX, srcX
		}

		hChar := chars.HLine
		vChar := chars.VLine

		if edge.Reversed {
			hChar = "┄"
			vChar = "╎"
		}

		if edge.Quality == "draining" {
			hChar = "┄"
		}

		midX := (srcX + tgtX) / 2

		for x := srcX; x <= midX && x < gridWidth; x++ {
			if srcY >= 0 && srcY < gridHeight && x >= 0 && x < gridWidth {
				if grid[srcY][x] == ' ' {
					grid[srcY][x] = []rune(hChar)[0]
				}
			}
		}

		step := 1
		if tgtY < srcY {
			step = -1
		}

		for y := srcY; y != tgtY; y += step {
			if y >= 0 && y < gridHeight && midX >= 0 && midX < gridWidth {
				if grid[y][midX] == ' ' {
					grid[y][midX] = []rune(vChar)[0]
				}
			}
		}

		for x := midX; x <= tgtX && x < gridWidth; x++ {
			if tgtY >= 0 && tgtY < gridHeight && x >= 0 && x < gridWidth {
				if grid[tgtY][x] == ' ' {
					grid[tgtY][x] = []rune(hChar)[0]
				}
			}
		}

		if srcY >= 0 && srcY < gridHeight && midX-1 >= 0 && midX-1 < gridWidth {
			if grid[srcY][midX-1] == ' ' {
				grid[srcY][midX-1] = []rune(chars.LeftTee)[0]
			}
		}

		if tgtY >= 0 && tgtY < gridHeight && midX >= 0 && midX < gridWidth {
			if grid[tgtY][midX] == ' ' {
				grid[tgtY][midX] = []rune(chars.RightTee)[0]
			}
		}

		if midX >= 0 && midX < gridWidth && srcY-step >= 0 && srcY-step < gridHeight {
			if grid[srcY-step][midX] == ' ' {
				grid[srcY-step][midX] = []rune(chars.BottomTee)[0]
			}
		}

		if midX >= 0 && midX < gridWidth && tgtY+step >= 0 && tgtY+step < gridHeight {
			if grid[tgtY+step][midX] == ' ' {
				grid[tgtY+step][midX] = []rune(chars.TopTee)[0]
			}
		}
	}
}
