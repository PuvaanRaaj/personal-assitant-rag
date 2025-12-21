package storage

import (
	"context"
	"fmt"

	"github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// QdrantClient wraps Qdrant vector database operations
type QdrantClient struct {
	client qdrant.CollectionsClient
	conn   *grpc.ClientConn
}

// NewQdrantClient creates a new Qdrant client
func NewQdrantClient(url string) (*QdrantClient, error) {
	conn, err := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Qdrant: %w", err)
	}

	client := qdrant.NewCollectionsClient(conn)

	return &QdrantClient{
		client: client,
		conn:   conn,
	}, nil
}

// Close closes the connection to Qdrant
func (q *QdrantClient) Close() error {
	return q.conn.Close()
}

// CreateCollection creates a new collection for a user
func (q *QdrantClient) CreateCollection(ctx context.Context, collectionName string, vectorSize uint64) error {
	_, err := q.client.Create(ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: &qdrant.VectorsConfig{
			Config: &qdrant.VectorsConfig_Params{
				Params: &qdrant.VectorParams{
					Size:     vectorSize,
					Distance: qdrant.Distance_Cosine,
				},
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	return nil
}

// CollectionExists checks if a collection exists
func (q *QdrantClient) CollectionExists(ctx context.Context, collectionName string) (bool, error) {
	response, err := q.client.List(ctx, &qdrant.ListCollectionsRequest{})
	if err != nil {
		return false, fmt.Errorf("failed to list collections: %w", err)
	}

	for _, collection := range response.Collections {
		if collection.Name == collectionName {
			return true, nil
		}
	}

	return false, nil
}

// DeleteCollection deletes a collection
func (q *QdrantClient) DeleteCollection(ctx context.Context, collectionName string) error {
	_, err := q.client.Delete(ctx, &qdrant.DeleteCollection{
		CollectionName: collectionName,
	})

	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}

	return nil
}
