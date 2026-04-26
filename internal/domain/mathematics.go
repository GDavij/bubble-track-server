package domain

import "math"

type GraphMetrics struct {
	Density            float64 `json:"density"`
	AvgDegree          float64 `json:"avg_degree"`
	GlobalClustering   float64 `json:"global_clustering"`
	Assortativity      float64 `json:"assortativity"`
	IsSmallWorld       bool    `json:"is_small_world"`
	IsScaleFree        bool    `json:"is_scale_free"`
	ConnectedComponents int    `json:"connected_components"`
	Diameter           int     `json:"diameter"`
	AvgPathLength      float64 `json:"avg_path_length"`
	Modularity         float64 `json:"modularity"`
}

type PageRankConfig struct {
	DampingFactor float64
	MaxIterations int
	Tolerance     float64
}

func ComputeDensity(nodeCount, edgeCount int) float64 {
	if nodeCount <= 1 {
		return 0
	}
	maxEdges := float64(nodeCount) * float64(nodeCount-1) / 2.0
	if maxEdges == 0 {
		return 0
	}
	return float64(edgeCount) / maxEdges
}

func ComputeAvgDegree(nodeCount, edgeCount int) float64 {
	if nodeCount == 0 {
		return 0
	}
	return 2.0 * float64(edgeCount) / float64(nodeCount)
}

func ComputeDegreeCentralityMap(adjList map[string][]string) map[string]float64 {
	n := len(adjList)
	if n <= 1 {
		result := make(map[string]float64)
		for k := range adjList {
			result[k] = 0
		}
		return result
	}

	result := make(map[string]float64)
	for node, neighbors := range adjList {
		degree := float64(len(neighbors))
		result[node] = degree / float64(n-1)
	}
	return result
}

func ComputeBetweennessCentrality(adjList map[string][]string) map[string]float64 {
	nodes := make([]string, 0, len(adjList))
	for n := range adjList {
		nodes = append(nodes, n)
	}

	bc := make(map[string]float64)
	for _, s := range nodes {
		shortest := bfsShortestPaths(adjList, s)
		dependencies := make(map[string]float64)

		sorted := topoSort(shortest, nodes)
		for _, w := range sorted {
			for _, v := range shortest[w].Predecessors {
				c := float64(shortest[v].Count) / float64(shortest[w].Count)
				dependencies[v] += c * (1 + dependencies[w])
			}
			if w != s {
				bc[w] += dependencies[w]
			}
		}
	}

	total := float64(len(nodes) - 1)
	if total <= 0 {
		return bc
	}
	for k := range bc {
		bc[k] /= total
	}
	return bc
}

func ComputeClosenessCentrality(adjList map[string][]string) map[string]float64 {
	cc := make(map[string]float64)
	for node := range adjList {
		shortest := bfsShortestPaths(adjList, node)
		totalDist := 0
		reachable := 0
		for _, dists := range shortest {
			if dists.Distance > 0 {
				totalDist += dists.Distance
				reachable++
			}
		}
		if reachable > 0 && totalDist > 0 {
			cc[node] = float64(reachable) / float64(totalDist)
		}
	}
	return cc
}

func ComputePageRank(adjList map[string][]string, cfg PageRankConfig) map[string]float64 {
	if cfg.DampingFactor == 0 {
		cfg.DampingFactor = 0.85
	}
	if cfg.MaxIterations == 0 {
		cfg.MaxIterations = 100
	}
	if cfg.Tolerance == 0 {
		cfg.Tolerance = 1e-6
	}

	nodes := make([]string, 0, len(adjList))
	for n := range adjList {
		nodes = append(nodes, n)
	}
	n := len(nodes)
	if n == 0 {
		return map[string]float64{}
	}

	pr := make(map[string]float64)
	initRank := 1.0 / float64(n)
	for _, node := range nodes {
		pr[node] = initRank
	}

	outDegree := make(map[string]float64)
	for node, neighbors := range adjList {
		outDegree[node] = float64(len(neighbors))
	}

	inLinks := make(map[string][]string)
	for node, neighbors := range adjList {
		for _, neighbor := range neighbors {
			inLinks[neighbor] = append(inLinks[neighbor], node)
		}
	}

	for iter := 0; iter < cfg.MaxIterations; iter++ {
		newPR := make(map[string]float64)
		danglingSum := 0.0
		for _, node := range nodes {
			if outDegree[node] == 0 {
				danglingSum += pr[node]
			}
		}

		for _, node := range nodes {
			rank := (1-cfg.DampingFactor)/float64(n)
			rank += cfg.DampingFactor * danglingSum / float64(n)
			for _, src := range inLinks[node] {
				if outDegree[src] > 0 {
					rank += cfg.DampingFactor * pr[src] / outDegree[src]
				}
			}
			newPR[node] = rank
		}

		diff := 0.0
		for _, node := range nodes {
			diff += absFloat(newPR[node] - pr[node])
		}
		pr = newPR
		if diff < cfg.Tolerance {
			break
		}
	}
	return pr
}

func ComputeClusteringCoefficientLocal(adjList map[string][]string, node string) float64 {
	neighbors := adjList[node]
	k := len(neighbors)
	if k < 2 {
		return 0
	}

	neighborSet := make(map[string]bool)
	for _, n := range neighbors {
		neighborSet[n] = true
	}

	triangles := 0
	for _, n1 := range neighbors {
		for _, n2 := range neighbors {
			if n1 >= n2 {
				continue
			}
			if neighborSet[n2] {
				for _, nn := range adjList[n1] {
					if nn == n2 {
						triangles++
						break
					}
				}
			}
		}
	}

	maxTriangles := k * (k - 1) / 2
	if maxTriangles == 0 {
		return 0
	}
	return float64(triangles) / float64(maxTriangles)
}

func ComputeModularity(adjList map[string][]string, communities map[string]int) float64 {
	m := 0
	for _, neighbors := range adjList {
		m += len(neighbors)
	}
	m /= 2
	if m == 0 {
		return 0
	}

	ki := make(map[string]float64)
	for node, neighbors := range adjList {
		ki[node] = float64(len(neighbors))
	}

	Q := 0.0
	for node, neighbors := range adjList {
		ci := communities[node]
		for _, neighbor := range neighbors {
			cj := communities[neighbor]
			delta := 1.0
			if ci != cj {
				delta = 0.0
			}
			Q += delta - (ki[node]*ki[neighbor])/(2.0*float64(m))
		}
	}
	return Q / (2.0 * float64(m))
}

func DetectCommunitiesLouvain(adjList map[string][]string) map[string]int {
	communities := make(map[string]int)
	i := 0
	for node := range adjList {
		communities[node] = i
		i++
	}

	improved := true
	for iteration := 0; iteration < 50 && improved; iteration++ {
		improved = false
		for node := range adjList {
			bestCommunity := communities[node]
			bestGain := 0.0

			neighborCommunities := make(map[int]float64)
			for _, neighbor := range adjList[node] {
				neighborCommunities[communities[neighbor]]++
			}

			currentComm := communities[node]
			for comm, weight := range neighborCommunities {
				if comm == currentComm {
					continue
				}
				gain := weight
				if gain > bestGain {
					bestGain = gain
					bestCommunity = comm
					improved = true
				}
			}

			communities[node] = bestCommunity
		}
	}

	renamed := make(map[int]int)
	newID := 0
	for node := range adjList {
		commVal := communities[node]
		if _, ok := renamed[commVal]; !ok {
			renamed[commVal] = newID
			newID++
		}
		communities[node] = renamed[commVal]
	}
	return communities
}

type bfsResult struct {
	Distance     int
	Predecessors []string
	Count        int
}

func bfsShortestPaths(adjList map[string][]string, source string) map[string]*bfsResult {
	dist := make(map[string]*bfsResult)
	dist[source] = &bfsResult{Distance: 0, Count: 1}

	queue := []string{source}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, neighbor := range adjList[current] {
			if _, ok := dist[neighbor]; !ok {
				dist[neighbor] = &bfsResult{
					Distance:     dist[current].Distance + 1,
					Predecessors: []string{current},
					Count:        dist[current].Count,
				}
				queue = append(queue, neighbor)
			} else if dist[neighbor].Distance == dist[current].Distance+1 {
				dist[neighbor].Count += dist[current].Count
				dist[neighbor].Predecessors = append(dist[neighbor].Predecessors, current)
			}
		}
	}
	return dist
}

func topoSort(dist map[string]*bfsResult, nodes []string) []string {
	type nodeDist struct {
		node string
		dist int
	}
	sorted := make([]nodeDist, 0, len(nodes))
	for _, n := range nodes {
		if d, ok := dist[n]; ok {
			sorted = append(sorted, nodeDist{n, d.Distance})
		}
	}
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].dist < sorted[j].dist {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	result := make([]string, len(sorted))
	for i, s := range sorted {
		result[i] = s.node
	}
	return result
}

func ComputeInformationEntropy(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	freq := make(map[float64]int)
	for _, v := range values {
		bucket := float64(int(v*10)) / 10.0
		freq[bucket]++
	}

	n := float64(len(values))
	entropy := 0.0
	for _, count := range freq {
		p := float64(count) / n
		if p > 0 {
			entropy -= p * log2(p)
		}
	}
	return entropy
}

func log2(x float64) float64 {
	if x <= 0 {
		return 0
	}
	return math.Log2(x)
}

func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
