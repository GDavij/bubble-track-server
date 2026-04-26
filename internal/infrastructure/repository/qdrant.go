package repository

import (
	"context"
	"fmt"

	"github.com/bubbletrack/server/internal/config"
	"github.com/bubbletrack/server/internal/domain"
	pb "github.com/qdrant/go-client/qdrant"
)

type QdrantMemoryRepository struct {
	client         *pb.Client
	collectionName string
	vectorSize     uint64
}

func NewQdrantMemoryRepository(client *pb.Client, cfg config.QdrantConfig) *QdrantMemoryRepository {
	return &QdrantMemoryRepository{
		client:         client,
		collectionName: cfg.Collection,
		vectorSize:     768,
	}
}

func (r *QdrantMemoryRepository) EnsureCollection(ctx context.Context) error {
	exists, err := r.collectionExists(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	return r.client.CreateCollection(ctx, &pb.CreateCollection{
		CollectionName: r.collectionName,
		VectorsConfig: pb.NewVectorsConfig(&pb.VectorParams{
			Size:     r.vectorSize,
			Distance: pb.Distance_Cosine,
		}),
	})
}

func (r *QdrantMemoryRepository) Store(ctx context.Context, userID string, interactionID string, embedding []float32, metadata map[string]any) error {
	pointID := pb.NewID(interactionID)
	payload := pb.NewValueMap(metadata)
	payload["user_id"] = pb.NewValueString(userID)
	payload["interaction_id"] = pb.NewValueString(interactionID)

	wait := true
	_, err := r.client.Upsert(ctx, &pb.UpsertPoints{
		CollectionName: r.collectionName,
		Points: []*pb.PointStruct{
			{
				Id:      pointID,
				Vectors: pb.NewVectorsDense(embedding),
				Payload: payload,
			},
		},
		Wait: &wait,
	})
	return err
}

func (r *QdrantMemoryRepository) SearchFiltered(ctx context.Context, userID string, queryEmbedding []float32, limit int, filter domain.MemoryFilter) ([]domain.MemoryResult, error) {
	if limit <= 0 {
		limit = 5
	}

	conditions := []*pb.Condition{
		pb.NewMatchKeyword("user_id", userID),
	}

	if filter.Person != "" {
		conditions = append(conditions, pb.NewMatchKeyword("people", filter.Person))
	}
	if filter.Session != "" {
		conditions = append(conditions, pb.NewMatchKeyword("session", filter.Session))
	}

	qdrantFilter := &pb.Filter{Must: conditions}

	limitVal := uint64(limit)
	results, err := r.client.Query(ctx, &pb.QueryPoints{
		CollectionName: r.collectionName,
		Query:          pb.NewQueryDense(queryEmbedding),
		Limit:          &limitVal,
		Filter:         qdrantFilter,
		WithPayload: pb.NewWithPayload(true),
	})
	if err != nil {
		return nil, fmt.Errorf("qdrant query: %w", err)
	}

	memoryResults := make([]domain.MemoryResult, 0, len(results))
	for _, point := range results {
		metadata := extractPayload(point.Payload)
		memoryResults = append(memoryResults, domain.MemoryResult{
			InteractionID: fmt.Sprintf("%v", metadata["interaction_id"]),
			Score:         point.Score,
			Metadata:      metadata,
		})
	}

	return memoryResults, nil
}

func (r *QdrantMemoryRepository) Search(ctx context.Context, userID string, queryEmbedding []float32, limit int) ([]domain.MemoryResult, error) {
	if limit <= 0 {
		limit = 5
	}

	limitVal := uint64(limit)
	results, err := r.client.Query(ctx, &pb.QueryPoints{
		CollectionName: r.collectionName,
		Query:          pb.NewQueryDense(queryEmbedding),
		Limit:          &limitVal,
		Filter: &pb.Filter{
			Must: []*pb.Condition{
				pb.NewMatchKeyword("user_id", userID),
			},
		},
		WithPayload: pb.NewWithPayload(true),
	})
	if err != nil {
		return nil, fmt.Errorf("qdrant query: %w", err)
	}

	memoryResults := make([]domain.MemoryResult, 0, len(results))
	for _, point := range results {
		metadata := extractPayload(point.Payload)
		memoryResults = append(memoryResults, domain.MemoryResult{
			InteractionID: fmt.Sprintf("%v", metadata["interaction_id"]),
			Score:         point.Score,
			Metadata:      metadata,
		})
	}

	return memoryResults, nil
}

func (r *QdrantMemoryRepository) collectionExists(ctx context.Context) (bool, error) {
	collections, err := r.client.ListCollections(ctx)
	if err != nil {
		return false, err
	}
	for _, name := range collections {
		if name == r.collectionName {
			return true, nil
		}
	}
	return false, nil
}

func extractPayload(payload map[string]*pb.Value) map[string]any {
	m := make(map[string]any)
	for k, v := range payload {
		if v != nil {
			m[k] = v.GetKind()
		}
	}
	return m
}
