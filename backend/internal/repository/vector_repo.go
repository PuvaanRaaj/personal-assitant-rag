package repository

import (
	"context"
	"fmt"

	"github.com/PuvaanRaaj/personal-rag-agent/internal/model"
	"github.com/PuvaanRaaj/personal-rag-agent/internal/storage"
	"github.com/qdrant/go-client/qdrant"
)

// VectorRepository handles vector database operations
type VectorRepository struct {
	client *storage.QdrantClient
}

// NewVectorRepository creates a new vector repository
func NewVectorRepository(client *storage.QdrantClient) *VectorRepository {
	return &VectorRepository{client: client}
}

// GetCollectionName returns the collection name for a user
func (r *VectorRepository) GetCollectionName(userID string) string {
	return fmt.Sprintf("user_%s_docs", userID)
}

// EnsureCollection ensures a collection exists for the user
func (r *VectorRepository) EnsureCollection(ctx context.Context, userID string, vectorSize uint64) error {
	collectionName := r.GetCollectionName(userID)

	exists, err := r.client.CollectionExists(ctx, collectionName)
	if err != nil {
		return err
	}

	if !exists {
		return r.client.CreateCollection(ctx, collectionName, vectorSize)
	}

	return nil
}

// InsertVectors inserts vectors into a user's collection
func (r *VectorRepository) InsertVectors(ctx context.Context, userID string, points []*model.VectorPoint) error {
	_ = r.GetCollectionName(userID) // TODO: use when implementing upsert

	// Convert to Qdrant points
	qdrantPoints := make([]*qdrant.PointStruct, len(points))
	for i, p := range points {
		qdrantPoints[i] = &qdrant.PointStruct{
			Id: &qdrant.PointId{
				PointIdOptions: &qdrant.PointId_Uuid{
					Uuid: p.ID,
				},
			},
			Vectors: &qdrant.Vectors{
				VectorsOptions: &qdrant.Vectors_Vector{
					Vector: &qdrant.Vector{
						Data: p.Vector,
					},
				},
			},
			Payload: convertToQdrantPayload(p.Payload),
		}
	}

	// TODO: Implement upsert vectors to Qdrant
	// This requires the Points client
	_ = qdrantPoints

	return fmt.Errorf("insert vectors not fully implemented yet")
}

// Search performs similarity search
func (r *VectorRepository) Search(ctx context.Context, userID string, vector []float32, limit int) ([]*model.VectorPoint, error) {
	collectionName := r.GetCollectionName(userID)

	// TODO: Implement search
	// This requires the Points client

	_ = collectionName
	_ = vector
	_ = limit

	return nil, fmt.Errorf("search not fully implemented yet")
}

// DeleteByDocumentID deletes all vectors for a document
func (r *VectorRepository) DeleteByDocumentID(ctx context.Context, userID, documentID string) error {
	_ = r.GetCollectionName(userID)

	// TODO: Implement delete by filter using Points client
	// This requires filtering by document_id in the payload
	_ = documentID

	return fmt.Errorf("delete by document ID not fully implemented yet")
}

// convertToQdrantPayload converts a map to Qdrant payload
func convertToQdrantPayload(payload map[string]interface{}) map[string]*qdrant.Value {
	result := make(map[string]*qdrant.Value)

	for key, value := range payload {
		switch v := value.(type) {
		case string:
			result[key] = &qdrant.Value{
				Kind: &qdrant.Value_StringValue{StringValue: v},
			}
		case int:
			result[key] = &qdrant.Value{
				Kind: &qdrant.Value_IntegerValue{IntegerValue: int64(v)},
			}
		case int64:
			result[key] = &qdrant.Value{
				Kind: &qdrant.Value_IntegerValue{IntegerValue: v},
			}
		case float64:
			result[key] = &qdrant.Value{
				Kind: &qdrant.Value_DoubleValue{DoubleValue: v},
			}
		case bool:
			result[key] = &qdrant.Value{
				Kind: &qdrant.Value_BoolValue{BoolValue: v},
			}
		}
	}

	return result
}
