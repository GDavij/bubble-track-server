package tui

import "sort"

type LayoutEdge struct {
	Source, Target, Quality string
	Strength                float64
	Reversed                bool
}

type NodePosition struct {
	ID    string
	X, Y  int
	Layer int
}

func LayoutGraph(nodes []Node, edges []Edge, width, height int) []NodePosition {
	if len(nodes) == 0 {
		return []NodePosition{}
	}
	le := make([]LayoutEdge, len(edges))
	for i, e := range edges {
		le[i] = LayoutEdge{Source: e.Source, Target: e.Target, Quality: e.Quality, Strength: e.Strength}
	}
	le = breakCycles(nodes, le)
	layer := assignLayers(nodes, le)
	reduceCrossings(le, layer)
	return calcPositions(nodes, layer, width, height)
}

func breakCycles(nodes []Node, edges []LayoutEdge) []LayoutEdge {
	visited, recStack := make(map[string]bool), make(map[string]bool)
	adj := make(map[string][]string)
	for _, e := range edges {
		adj[e.Source] = append(adj[e.Source], e.Target)
	}
	var dfs func(string) bool
	dfs = func(id string) bool {
		visited[id], recStack[id] = true, true
		for _, n := range adj[id] {
			if !visited[n] && dfs(n) || recStack[n] {
				return true
			}
		}
		recStack[id] = false
		return false
	}
	for _, n := range nodes {
		if !visited[n.ID] {
			dfs(n.ID)
		}
	}
	visited, recStack = make(map[string]bool), make(map[string]bool)
	var mark func(string)
	mark = func(id string) {
		visited[id], recStack[id] = true, true
		for i := range edges {
			if edges[i].Source == id && !edges[i].Reversed {
				if !visited[edges[i].Target] {
					mark(edges[i].Target)
				} else if recStack[edges[i].Target] {
					edges[i].Reversed = true
				}
			}
		}
		recStack[id] = false
	}
	for _, n := range nodes {
		if !visited[n.ID] {
			mark(n.ID)
		}
	}
	return edges
}

func assignLayers(nodes []Node, edges []LayoutEdge) map[string]int {
	if len(nodes) == 0 {
		return nil
	}
	deg := make(map[string]int)
	for _, e := range edges {
		deg[e.Source]++
		deg[e.Target]++
	}
	start, maxDeg := nodes[0].ID, 0
	for id, d := range deg {
		if d > maxDeg {
			maxDeg, start = d, id
		}
	}
	layer := make(map[string]int)
	layer[start] = 0
	queue := []string{start}
	visited := make(map[string]bool)
	visited[start] = true
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		cl := layer[curr]
		for _, e := range edges {
			if !e.Reversed && e.Source == curr && !visited[e.Target] {
				layer[e.Target] = cl + 1
				visited[e.Target] = true
				queue = append(queue, e.Target)
			} else if e.Reversed && e.Target == curr && !visited[e.Source] {
				layer[e.Source] = cl + 1
				visited[e.Source] = true
				queue = append(queue, e.Source)
			}
		}
	}
	for _, n := range nodes {
		if _, ok := layer[n.ID]; !ok {
			layer[n.ID] = 0
		}
	}
	return layer
}

func reduceCrossings(edges []LayoutEdge, layerMap map[string]int) {
	l2n := make(map[int][]string)
	maxL := 0
	for id, l := range layerMap {
		l2n[l] = append(l2n[l], id)
		if l > maxL {
			maxL = l
		}
	}
	for _ = range [3]int{} {
		for l := 0; l <= maxL; l++ {
			nodes := l2n[l]
			if len(nodes) < 2 {
				continue
			}
			type assign struct {
				id   string
				bary float64
			}
			assigns := make([]assign, len(nodes))
			for i, id := range nodes {
				assigns[i] = assign{id: id, bary: calcBary(id, l, edges, l2n)}
			}
			sort.Slice(assigns, func(i, j int) bool { return assigns[i].bary < assigns[j].bary })
			for i, a := range assigns {
				l2n[l][i] = a.id
			}
		}
	}
	pos := make(map[string]int)
	for _, ids := range l2n {
		for i, id := range ids {
			pos[id] = i
		}
	}
	for l := 0; l <= maxL; l++ {
		sort.Slice(l2n[l], func(i, j int) bool { return pos[l2n[l][i]] < pos[l2n[l][j]] })
	}
}

func calcBary(nodeID string, layer int, edges []LayoutEdge, l2n map[int][]string) float64 {
	var sum float64
	cnt := 0
	for _, e := range edges {
		var nID string
		var nL int
		if e.Source == nodeID {
			nID, nL = e.Target, layer+1
		} else if e.Target == nodeID {
			nID, nL = e.Source, layer-1
		} else {
			continue
		}
		for i, nid := range l2n[nL] {
			if nid == nID {
				sum += float64(i)
				cnt++
				break
			}
		}
	}
	if cnt == 0 {
		return 0
	}
	return sum / float64(cnt)
}

func calcPositions(nodes []Node, layerMap map[string]int, w, h int) []NodePosition {
	if len(nodes) == 0 {
		return nil
	}
	l2n := make(map[int][]string)
	maxL := 0
	for id, l := range layerMap {
		l2n[l] = append(l2n[l], id)
		if l > maxL {
			maxL = l
		}
	}
	numLayers := maxL + 1
	if numLayers == 0 {
		numLayers = 1
	}
	colSpacing := w / (numLayers + 1)
	if colSpacing < 10 {
		colSpacing = 10
	}
	pos := make([]NodePosition, len(nodes))
	for i, n := range nodes {
		l := layerMap[n.ID]
		layerNodes := l2n[l]
		row := 0
		for j, id := range layerNodes {
			if id == n.ID {
				row = j
				break
			}
		}
		rowSpacing := h / (len(layerNodes) + 1)
		if rowSpacing < 4 {
			rowSpacing = 4
		}
		x := (l + 1) * colSpacing
		y := (row + 1) * rowSpacing
		if x < 5 {
			x = 5
		}
		if y < 2 {
			y = 2
		}
		pos[i] = NodePosition{ID: n.ID, X: x, Y: y, Layer: l}
	}
	return pos
}
