package domain

import (
	"math"
	"testing"
)

const floatTolerance = 1e-9

func floatEqual(a, b float64) bool {
	return math.Abs(a-b) < floatTolerance
}

// ---------------------------------------------------------------------------
// ComputeDensity
// ---------------------------------------------------------------------------

func TestMathComputeDensity(t *testing.T) {
	tests := []struct {
		name      string
		nodeCount int
		edgeCount int
		want      float64
	}{
		{name: "zero nodes", nodeCount: 0, edgeCount: 0, want: 0},
		{name: "one node", nodeCount: 1, edgeCount: 0, want: 0},
		{name: "two nodes no edges", nodeCount: 2, edgeCount: 0, want: 0},
		{name: "two nodes one edge (complete)", nodeCount: 2, edgeCount: 1, want: 1.0},
		{name: "three nodes complete graph", nodeCount: 3, edgeCount: 3, want: 1.0},
		{name: "three nodes two edges", nodeCount: 3, edgeCount: 2, want: 2.0 / 3.0},
		{name: "four nodes complete graph", nodeCount: 4, edgeCount: 6, want: 1.0},
		{name: "four nodes three edges", nodeCount: 4, edgeCount: 3, want: 0.5},
		{name: "negative node count", nodeCount: -1, edgeCount: 5, want: 0},
		{name: "large graph sparse", nodeCount: 100, edgeCount: 10, want: 10.0 / (100.0 * 99.0 / 2.0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeDensity(tt.nodeCount, tt.edgeCount)
			if !floatEqual(got, tt.want) {
				t.Errorf("ComputeDensity(%d, %d) = %v, want %v", tt.nodeCount, tt.edgeCount, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ComputeAvgDegree
// ---------------------------------------------------------------------------

func TestMathComputeAvgDegree(t *testing.T) {
	tests := []struct {
		name      string
		nodeCount int
		edgeCount int
		want      float64
	}{
		{name: "zero nodes", nodeCount: 0, edgeCount: 0, want: 0},
		{name: "one node no edges", nodeCount: 1, edgeCount: 0, want: 0},
		{name: "two nodes one edge", nodeCount: 2, edgeCount: 1, want: 1.0},
		{name: "three nodes three edges", nodeCount: 3, edgeCount: 3, want: 2.0},
		{name: "four nodes six edges complete", nodeCount: 4, edgeCount: 6, want: 3.0},
		{name: "five nodes zero edges", nodeCount: 5, edgeCount: 0, want: 0},
		{name: "large graph", nodeCount: 1000, edgeCount: 5000, want: 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeAvgDegree(tt.nodeCount, tt.edgeCount)
			if !floatEqual(got, tt.want) {
				t.Errorf("ComputeAvgDegree(%d, %d) = %v, want %v", tt.nodeCount, tt.edgeCount, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ComputeDegreeCentralityMap
// ---------------------------------------------------------------------------

func TestMathComputeDegreeCentralityMap(t *testing.T) {
	tests := []struct {
		name    string
		adjList map[string][]string
		want    map[string]float64
	}{
		{
			name:    "empty graph",
			adjList: map[string][]string{},
			want:    map[string]float64{},
		},
		{
			name:    "single isolated node",
			adjList: map[string][]string{"A": {}},
			want:    map[string]float64{"A": 0},
		},
		{
			name:    "two connected nodes",
			adjList: map[string][]string{"A": {"B"}, "B": {"A"}},
			want:    map[string]float64{"A": 1.0, "B": 1.0},
		},
		{
			name:    "star graph center has degree 1",
			adjList: map[string][]string{"A": {"B", "C"}, "B": {"A"}, "C": {"A"}},
			want:    map[string]float64{"A": 1.0, "B": 0.5, "C": 0.5},
		},
		{
			name:    "complete triangle",
			adjList: map[string][]string{"A": {"B", "C"}, "B": {"A", "C"}, "C": {"A", "B"}},
			want:    map[string]float64{"A": 1.0, "B": 1.0, "C": 1.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeDegreeCentralityMap(tt.adjList)
			if len(got) != len(tt.want) {
				t.Fatalf("ComputeDegreeCentralityMap() returned %d entries, want %d", len(got), len(tt.want))
			}
			for k, wantV := range tt.want {
				gotV, ok := got[k]
				if !ok {
					t.Errorf("missing key %q in result", k)
					continue
				}
				if !floatEqual(gotV, wantV) {
					t.Errorf("ComputeDegreeCentralityMap()[%q] = %v, want %v", k, gotV, wantV)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ComputeBetweennessCentrality
// ---------------------------------------------------------------------------

func TestMathComputeBetweennessCentrality(t *testing.T) {
	tests := []struct {
		name    string
		adjList map[string][]string
		want    map[string]float64
		tol     float64
	}{
		{
			name:    "empty graph",
			adjList: map[string][]string{},
			want:    map[string]float64{},
			tol:     floatTolerance,
		},
		{
			name:    "single node",
			adjList: map[string][]string{"A": {}},
			want:    map[string]float64{},
			tol:     floatTolerance,
		},
		{
			name:    "two connected nodes",
			adjList: map[string][]string{"A": {"B"}, "B": {"A"}},
			want:    map[string]float64{"A": 0, "B": 0},
			tol:     floatTolerance,
		},
		{
			name:    "line graph A-B-C, B is bridge",
			adjList: map[string][]string{"A": {"B"}, "B": {"A", "C"}, "C": {"B"}},
			want:    map[string]float64{"A": 0, "B": 1.0, "C": 0},
			tol:     floatTolerance,
		},
		{
			name:    "complete triangle, no betweenness",
			adjList: map[string][]string{"A": {"B", "C"}, "B": {"A", "C"}, "C": {"A", "B"}},
			want:    map[string]float64{"A": 0, "B": 0, "C": 0},
			tol:     floatTolerance,
		},
		{
			name:    "star graph, center should be highest betweenness",
			adjList: map[string][]string{"A": {"B", "C", "D"}, "B": {"A"}, "C": {"A"}, "D": {"A"}},
			want:    map[string]float64{"A": 2, "B": 0, "C": 0, "D": 0}, // unnormalized, center is highest
			tol:     floatTolerance,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeBetweennessCentrality(tt.adjList)
			if len(got) != len(tt.want) {
				t.Fatalf("ComputeBetweennessCentrality() returned %d entries, want %d", len(got), len(tt.want))
			}
			for k, wantV := range tt.want {
				gotV, ok := got[k]
				if !ok {
					t.Errorf("missing key %q in result", k)
					continue
				}
				if math.Abs(gotV-wantV) > tt.tol {
					t.Errorf("ComputeBetweennessCentrality()[%q] = %v, want %v", k, gotV, wantV)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ComputeClosenessCentrality
// ---------------------------------------------------------------------------

func TestMathComputeClosenessCentrality(t *testing.T) {
	tests := []struct {
		name    string
		adjList map[string][]string
		want    map[string]float64
		tol     float64
	}{
		{
			name:    "empty graph",
			adjList: map[string][]string{},
			want:    map[string]float64{},
			tol:     floatTolerance,
		},
		{
			name:    "single isolated node",
			adjList: map[string][]string{"A": {}},
			want:    map[string]float64{},
			tol:     floatTolerance,
		},
		{
			name:    "two connected nodes",
			adjList: map[string][]string{"A": {"B"}, "B": {"A"}},
			want:    map[string]float64{"A": 1.0, "B": 1.0},
			tol:     floatTolerance,
		},
		{
			name:    "line graph A-B-C",
			adjList: map[string][]string{"A": {"B"}, "B": {"A", "C"}, "C": {"B"}},
			want:    map[string]float64{"A": 2.0 / 3.0, "B": 1.0, "C": 2.0 / 3.0},
			tol:     floatTolerance,
		},
		{
			name:    "complete triangle",
			adjList: map[string][]string{"A": {"B", "C"}, "B": {"A", "C"}, "C": {"A", "B"}},
			want:    map[string]float64{"A": 1.0, "B": 1.0, "C": 1.0},
			tol:     floatTolerance,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeClosenessCentrality(tt.adjList)
			if len(got) != len(tt.want) {
				t.Fatalf("ComputeClosenessCentrality() returned %d entries, want %d (got=%v)", len(got), len(tt.want), got)
			}
			for k, wantV := range tt.want {
				gotV, ok := got[k]
				if !ok {
					t.Errorf("missing key %q in result", k)
					continue
				}
				if math.Abs(gotV-wantV) > tt.tol {
					t.Errorf("ComputeClosenessCentrality()[%q] = %v, want %v", k, gotV, wantV)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ComputePageRank
// ---------------------------------------------------------------------------

func TestMathComputePageRank(t *testing.T) {
	tests := []struct {
		name    string
		adjList map[string][]string
		cfg     PageRankConfig
		want    map[string]float64
		tol     float64
	}{
		{
			name:    "empty graph",
			adjList: map[string][]string{},
			cfg:     PageRankConfig{},
			want:    map[string]float64{},
			tol:     floatTolerance,
		},
		{
			name:    "single node converges to 1",
			adjList: map[string][]string{"A": {}},
			cfg:     PageRankConfig{DampingFactor: 0.85, MaxIterations: 100, Tolerance: 1e-6},
			want:    map[string]float64{"A": 1.0},
			tol:     1e-4,
		},
		{
			name:    "two node cycle equal rank",
			adjList: map[string][]string{"A": {"B"}, "B": {"A"}},
			cfg:     PageRankConfig{DampingFactor: 0.85, MaxIterations: 100, Tolerance: 1e-6},
			want:    map[string]float64{"A": 0.5, "B": 0.5},
			tol:     1e-4,
		},
		{
			name:    "complete triangle equal rank",
			adjList: map[string][]string{"A": {"B", "C"}, "B": {"A", "C"}, "C": {"A", "B"}},
			cfg:     PageRankConfig{DampingFactor: 0.85, MaxIterations: 100, Tolerance: 1e-6},
			want:    map[string]float64{"A": 1.0 / 3.0, "B": 1.0 / 3.0, "C": 1.0 / 3.0},
			tol:     1e-4,
		},
		{
			name:    "default config applied",
			adjList: map[string][]string{"A": {"B"}, "B": {"A"}},
			cfg:     PageRankConfig{}, // zero values should use defaults
			want:    map[string]float64{"A": 0.5, "B": 0.5},
			tol:     1e-4,
		},
		{
			name: "star graph center receives links, should have higher rank",
			adjList: map[string][]string{
				"A": {"B", "C"},
				"B": {"A"},
				"C": {"A"},
			},
			cfg:  PageRankConfig{DampingFactor: 0.85, MaxIterations: 200, Tolerance: 1e-8},
			want: nil,
			tol:  1e-4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputePageRank(tt.adjList, tt.cfg)

			if tt.want == nil {
				// Custom verification: star graph center (receives links from B and C) should have HIGHER rank
				sum := 0.0
				for _, v := range got {
					sum += v
				}
				if math.Abs(sum-1.0) > tt.tol {
					t.Errorf("PageRank sum = %v, want 1.0", sum)
				}
				if got["A"] <= got["B"] {
					t.Errorf("star graph: expected A rank > B rank, got A=%v B=%v", got["A"], got["B"])
				}
				return
			}

			if len(got) != len(tt.want) {
				t.Fatalf("ComputePageRank() returned %d entries, want %d", len(got), len(tt.want))
			}
			for k, wantV := range tt.want {
				gotV, ok := got[k]
				if !ok {
					t.Errorf("missing key %q in result", k)
					continue
				}
				if math.Abs(gotV-wantV) > tt.tol {
					t.Errorf("ComputePageRank()[%q] = %v, want %v", k, gotV, wantV)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ComputeClusteringCoefficientLocal
// ---------------------------------------------------------------------------

func TestMathComputeClusteringCoefficientLocal(t *testing.T) {
	tests := []struct {
		name    string
		adjList map[string][]string
		node    string
		want    float64
	}{
		{
			name:    "isolated node",
			adjList: map[string][]string{"A": {}},
			node:    "A",
			want:    0,
		},
		{
			name:    "one neighbor",
			adjList: map[string][]string{"A": {"B"}, "B": {"A"}},
			node:    "A",
			want:    0,
		},
		{
			name:    "triangle fully clustered",
			adjList: map[string][]string{"A": {"B", "C"}, "B": {"A", "C"}, "C": {"A", "B"}},
			node:    "A",
			want:    1.0,
		},
		{
			name:    "square no triangles",
			adjList: map[string][]string{"A": {"B", "D"}, "B": {"A", "C"}, "C": {"B", "D"}, "D": {"C", "A"}},
			node:    "A",
			want:    0,
		},
		{
			name: "star center no triangles",
			adjList: map[string][]string{
				"A": {"B", "C", "D"},
				"B": {"A"},
				"C": {"A"},
				"D": {"A"},
			},
			node: "A",
			want: 0,
		},
		{
			name:    "node not in graph",
			adjList: map[string][]string{"A": {"B"}, "B": {"A"}},
			node:    "Z",
			want:    0,
		},
		{
			name: "partial clustering two of three triangles",
			adjList: map[string][]string{
				"A": {"B", "C", "D"},
				"B": {"A", "C"},
				"C": {"A", "B"},
				"D": {"A"},
			},
			node: "A",
			// A has 3 neighbors (B,C,D). Possible triangles = 3*2/2 = 3.
			// Actual: (B,C) connected → yes. (B,D) not connected. (C,D) not connected.
			// Triangles = 1. Coeff = 1/3.
			want: 1.0 / 3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeClusteringCoefficientLocal(tt.adjList, tt.node)
			if !floatEqual(got, tt.want) {
				t.Errorf("ComputeClusteringCoefficientLocal(%q) = %v, want %v", tt.node, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ComputeModularity
// ---------------------------------------------------------------------------

func TestMathComputeModularity(t *testing.T) {
	tests := []struct {
		name        string
		adjList     map[string][]string
		communities map[string]int
		want        float64
		tol         float64
	}{
		{
			name:        "empty graph",
			adjList:     map[string][]string{},
			communities: map[string]int{},
			want:        0,
			tol:         floatTolerance,
		},
		{
			name:        "single node no edges",
			adjList:     map[string][]string{"A": {}},
			communities: map[string]int{"A": 0},
			want:        0,
			tol:         floatTolerance,
		},
		{
			name:        "triangle same community",
			adjList:     map[string][]string{"A": {"B", "C"}, "B": {"A", "C"}, "C": {"A", "B"}},
			communities: map[string]int{"A": 0, "B": 0, "C": 0},
			want:        1.0 / 3.0,
			tol:         floatTolerance,
		},
		{
			name:        "triangle split communities lower modularity",
			adjList:     map[string][]string{"A": {"B", "C"}, "B": {"A", "C"}, "C": {"A", "B"}},
			communities: map[string]int{"A": 0, "B": 1, "C": 1},
			want:        -1.0 / 3.0,
			tol:         floatTolerance,
		},
		{
			name: "two disconnected pairs same community",
			adjList: map[string][]string{
				"A": {"B"}, "B": {"A"},
				"C": {"D"}, "D": {"C"},
			},
			communities: map[string]int{"A": 0, "B": 0, "C": 1, "D": 1},
			// m = 4/2 = 2, ki each = 1
			// Q: A-B: delta=1, 1-(1*1)/(2*2)=1-0.25=0.75
			//    B-A: same 0.75
			//    C-D: same 0.75
			//    D-C: same 0.75
			// Q_total = 3.0, result = 3.0/(2*2) = 0.75
			want: 0.75,
			tol:  floatTolerance,
		},
		{
			name: "two disconnected pairs wrong community assignment",
			adjList: map[string][]string{
				"A": {"B"}, "B": {"A"},
				"C": {"D"}, "D": {"C"},
			},
			communities: map[string]int{"A": 0, "B": 1, "C": 0, "D": 1},
			// All edges cross communities, delta=0 for each edge
			// Q: A-B: delta=0, 0-0.25=-0.25; B-A: -0.25; C-D: -0.25; D-C: -0.25
			// Q_total = -1.0, result = -1.0/4 = -0.25
			want: -0.25,
			tol:  floatTolerance,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeModularity(tt.adjList, tt.communities)
			if math.Abs(got-tt.want) > tt.tol {
				t.Errorf("ComputeModularity() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DetectCommunitiesLouvain
// ---------------------------------------------------------------------------

func TestMathDetectCommunitiesLouvain(t *testing.T) {
	tests := []struct {
		name          string
		adjList       map[string][]string
		wantNumComms  int // expected number of distinct communities
		wantSameGroup [][]string // nodes that should share a community
	}{
		{
			name:         "empty graph",
			adjList:      map[string][]string{},
			wantNumComms: 0,
		},
		{
			name:         "single node",
			adjList:      map[string][]string{"A": {}},
			wantNumComms: 1,
		},
		{
			name:         "complete triangle one community",
			adjList:      map[string][]string{"A": {"B", "C"}, "B": {"A", "C"}, "C": {"A", "B"}},
			wantNumComms: 1,
			wantSameGroup: [][]string{{"A", "B", "C"}},
		},
		{
			name: "two disconnected components",
			adjList: map[string][]string{
				"A": {"B"}, "B": {"A"},
				"C": {"D"}, "D": {"C"},
			},
			wantNumComms: 2,
			wantSameGroup: [][]string{{"A", "B"}, {"C", "D"}},
		},
		{
			name:         "two connected nodes",
			adjList:      map[string][]string{"A": {"B"}, "B": {"A"}},
			wantNumComms: 1,
			wantSameGroup: [][]string{{"A", "B"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectCommunitiesLouvain(tt.adjList)

			// Count distinct communities
			commSet := make(map[int]bool)
			for _, c := range got {
				commSet[c] = true
			}
			if len(commSet) != tt.wantNumComms {
				t.Errorf("DetectCommunitiesLouvain() produced %d communities, want %d (map=%v)", len(commSet), tt.wantNumComms, got)
			}

			// Check that specified nodes share communities
			for _, group := range tt.wantSameGroup {
				if len(group) < 2 {
					continue
				}
				expected := got[group[0]]
				for _, node := range group[1:] {
					if got[node] != expected {
						t.Errorf("expected %q and %q in same community, got %d and %d", group[0], node, expected, got[node])
					}
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ComputeInformationEntropy
// ---------------------------------------------------------------------------

func TestMathComputeInformationEntropy(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
		tol    float64
	}{
		{
			name:   "empty slice",
			values: []float64{},
			want:   0,
			tol:    floatTolerance,
		},
		{
			name:   "single value",
			values: []float64{0.5},
			want:   0, // p=1, log2(1)=0
			tol:    floatTolerance,
		},
		{
			name:   "all same values",
			values: []float64{0.3, 0.3, 0.3, 0.3},
			want:   0,
			tol:    floatTolerance,
		},
		{
			name:   "two equal groups",
			values: []float64{0.1, 0.1, 0.2, 0.2},
			want:   1.0, // 2 groups of 2, entropy = -(0.5*log2(0.5) + 0.5*log2(0.5)) = 1.0
			tol:    floatTolerance,
		},
		{
			name:   "three unique buckets",
			values: []float64{0.1, 0.2, 0.3},
			want:   math.Log2(3), // uniform over 3 buckets
			tol:    1e-9,
		},
		{
			name:   "four unique buckets uniform",
			values: []float64{0.1, 0.2, 0.3, 0.4},
			want:   math.Log2(4), // = 2.0
			tol:    1e-9,
		},
		{
			name:   "unequal distribution",
			values: []float64{0.1, 0.1, 0.1, 0.3},
			want:   -(0.75*math.Log2(0.75) + 0.25*math.Log2(0.25)),
			tol:    1e-9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeInformationEntropy(tt.values)
			if math.Abs(got-tt.want) > tt.tol {
				t.Errorf("ComputeInformationEntropy() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// log2 (unexported, same package access)
// ---------------------------------------------------------------------------

func TestMathLog2(t *testing.T) {
	tests := []struct {
		name string
		x    float64
		want float64
	}{
		{name: "zero returns zero (guard)", x: 0, want: 0},
		{name: "negative returns zero (guard)", x: -1, want: 0},
		{name: "log2(1) = 0", x: 1, want: 0},
		{name: "log2(2) = 1", x: 2, want: 1},
		{name: "log2(4) = 2", x: 4, want: 2},
		{name: "log2(8) = 3", x: 8, want: 3},
		{name: "log2(0.5) = -1", x: 0.5, want: -1},
		{name: "log2(1024) = 10", x: 1024, want: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := log2(tt.x)
			if !floatEqual(got, tt.want) {
				t.Errorf("log2(%v) = %v, want %v", tt.x, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// absFloat (unexported, same package access)
// ---------------------------------------------------------------------------

func TestMathAbsFloat(t *testing.T) {
	tests := []struct {
		name string
		x    float64
		want float64
	}{
		{name: "positive", x: 3.14, want: 3.14},
		{name: "negative", x: -2.71, want: 2.71},
		{name: "zero", x: 0, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := absFloat(tt.x)
			if !floatEqual(got, tt.want) {
				t.Errorf("absFloat(%v) = %v, want %v", tt.x, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Integration: PageRank sums to 1
// ---------------------------------------------------------------------------

func TestMathPageRankSumToOne(t *testing.T) {
	adjList := map[string][]string{
		"A": {"B", "C"},
		"B": {"C"},
		"C": {"A"},
	}
	pr := ComputePageRank(adjList, PageRankConfig{})

	sum := 0.0
	for _, v := range pr {
		sum += v
	}
	if math.Abs(sum-1.0) > 1e-4 {
		t.Errorf("PageRank values sum to %v, want 1.0", sum)
	}
}

// ---------------------------------------------------------------------------
// Integration: disconnected graph - closeness reflects unreachable nodes
// ---------------------------------------------------------------------------

func TestMathDisconnectedGraphCloseness(t *testing.T) {
	adjList := map[string][]string{
		"A": {"B"},
		"B": {"A"},
		"C": {"D"},
		"D": {"C"},
	}
	cc := ComputeClosenessCentrality(adjList)

	// Each component is a pair, so each node reaches exactly 1 other
	// at distance 1. closeness = reachable/totalDist = 1/1 = 1.0
	for _, node := range []string{"A", "B", "C", "D"} {
		if !floatEqual(cc[node], 1.0) {
			t.Errorf("ComputeClosenessCentrality()[%q] = %v, want 1.0 (disconnected pair)", node, cc[node])
		}
	}
}
