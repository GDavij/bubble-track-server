package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bubbletrack/server/internal/domain"
	"github.com/bubbletrack/server/internal/infrastructure/repository"
)

type GraphAnalysisEngine struct {
	engine          *repository.GraphEngine
	personRepo      domain.PersonRepository
	relationshipRepo domain.RelationshipRepository
	logger          *slog.Logger
}

func NewGraphAnalysisEngine(
	engine *repository.GraphEngine,
	personRepo domain.PersonRepository,
	relationshipRepo domain.RelationshipRepository,
	logger *slog.Logger,
) *GraphAnalysisEngine {
	return &GraphAnalysisEngine{
		engine:          engine,
		personRepo:      personRepo,
		relationshipRepo: relationshipRepo,
		logger:          logger,
	}
}

func (g *GraphAnalysisEngine) ComputeAllMetrics(ctx context.Context, userID string) (map[string]*domain.NodeMetrics, error) {
	adjList, err := g.buildAdjacencyList(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(adjList) == 0 {
		return map[string]*domain.NodeMetrics{}, nil
	}

	degreeCentrality := domain.ComputeDegreeCentralityMap(adjList)
	betweennessCentrality := domain.ComputeBetweennessCentrality(adjList)
	closenessCentrality := domain.ComputeClosenessCentrality(adjList)
	pagerank := domain.ComputePageRank(adjList, domain.PageRankConfig{})
	communities := domain.DetectCommunitiesLouvain(adjList)

	roleDistribution := make(map[domain.SocialRole]int)
	for _, role := range communities {
		roleDistribution[domain.SocialRole(fmt.Sprintf("%d", role))]++
	}

	metrics := make(map[string]*domain.NodeMetrics)
	now := time.Now().UTC()

	for nodeID := range adjList {
		clustering := domain.ComputeClusteringCoefficientLocal(adjList, nodeID)
		neighbors := adjList[nodeID]

		internalEdges := 0
		for _, n1 := range neighbors {
			for _, n2 := range neighbors {
				if n1 >= n2 {
					continue
				}
				for _, nn := range adjList[n1] {
					if nn == n2 {
						internalEdges++
					}
				}
			}
		}
		externalEdges := len(neighbors) - internalEdges

		communityID := ""
		if c, ok := communities[nodeID]; ok {
			communityID = string(rune('A' + c%26))
		}

		embeddedness := 0.0
		if len(neighbors) > 0 {
			embeddedness = float64(internalEdges) / float64(len(neighbors))
		}
		bridgeScore := 0.0
		if len(neighbors) > 0 {
			bridgeScore = 1.0 - embeddedness
		}

		metrics[nodeID] = &domain.NodeMetrics{
			PersonID:    nodeID,
			UserID:      userID,
			ComputedAt:  now,
			TimeWindow:  "all",
			Degree:      len(neighbors),
			Centrality: domain.CentralityScores{
				Degree:         degreeCentrality[nodeID],
				Betweenness:    betweennessCentrality[nodeID],
				Closeness:      closenessCentrality[nodeID],
				Eigenvector:    degreeCentrality[nodeID],
				PageRank:       pagerank[nodeID],
				ClusteringCoef: clustering,
			},
			Community: domain.CommunityMetrics{
				CommunityID:   communityID,
				CommunityRole: "member",
				InternalEdges: internalEdges,
				ExternalEdges: externalEdges,
				Embeddedness:  embeddedness,
				BridgeScore:    bridgeScore,
				GatewayScore:   bridgeScore,
			},
		}
	}

	return metrics, nil
}

func (g *GraphAnalysisEngine) ComputeGraphSnapshot(ctx context.Context, userID string) (*domain.GraphSnapshot, error) {
	adjList, err := g.buildAdjacencyList(ctx, userID)
	if err != nil {
		return nil, err
	}

	nodeCount := len(adjList)
	edgeCount := 0
	for _, neighbors := range adjList {
		edgeCount += len(neighbors)
	}
	edgeCount /= 2

	density := domain.ComputeDensity(nodeCount, edgeCount)
	avgDegree := domain.ComputeAvgDegree(nodeCount, edgeCount)

	globalClustering := 0.0
	if nodeCount > 0 {
		totalClustering := 0.0
		for node := range adjList {
			totalClustering += domain.ComputeClusteringCoefficientLocal(adjList, node)
		}
		globalClustering = totalClustering / float64(nodeCount)
	}

	communities := domain.DetectCommunitiesLouvain(adjList)
	_ = domain.ComputeModularity(adjList, communities)

	uniqueCommunities := make(map[int]bool)
	for _, c := range communities {
		uniqueCommunities[c] = true
	}

	_ = domain.ComputeModularity(adjList, communities)

	return &domain.GraphSnapshot{
		UserID:             userID,
		TakenAt:            time.Now().UTC(),
		NodeCount:          nodeCount,
		EdgeCount:          edgeCount,
		Density:            density,
		AvgDegree:          avgDegree,
		GlobalClustering:   globalClustering,
		ComponentCount:     len(uniqueCommunities),
		LargestComponent:   nodeCount,
	}, nil
}

func (g *GraphAnalysisEngine) buildAdjacencyList(ctx context.Context, userID string) (map[string][]string, error) {
	rels, err := g.relationshipRepo.GetByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	adjList := make(map[string][]string)
	for _, rel := range rels {
		adjList[rel.SourcePersonID] = append(adjList[rel.SourcePersonID], rel.TargetPersonID)
		adjList[rel.TargetPersonID] = append(adjList[rel.TargetPersonID], rel.SourcePersonID)
	}

	people, err := g.personRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, p := range people {
		if _, ok := adjList[p.ID]; !ok {
			adjList[p.ID] = []string{}
		}
	}

	return adjList, nil
}

func (g *GraphAnalysisEngine) PersistMetrics(ctx context.Context, metrics map[string]*domain.NodeMetrics) error {
	for _, m := range metrics {
		if err := g.engine.UpsertNodeMetrics(ctx, m); err != nil {
			g.logger.Error("failed to persist metrics", "person_id", m.PersonID, "error", err)
		}
	}
	return nil
}

func (g *GraphAnalysisEngine) GetMetricHistory(ctx context.Context, personID, metricName string, limit int) ([]domain.MetricPoint, error) {
	return g.engine.GetMetricHistory(ctx, personID, metricName, limit)
}
