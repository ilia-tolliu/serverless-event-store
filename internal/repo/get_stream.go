package repo

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/google/uuid"
	"github.com/ilia-tolliu-go-event-store/internal/estypes"
)

func (r *EsRepo) GetStream(ctx context.Context, streamType string, streamId uuid.UUID) (estypes.Stream, error) {
	streamGet, err := PrepareGetDbStream(r.tableName, streamId)
	if err != nil {
		return estypes.Stream{}, fmt.Errorf("failed to prepare GetDbStream: %w", err)
	}

	output, err := r.dynamoDb.GetItem(ctx, streamGet)
	if err != nil {
		return estypes.Stream{}, fmt.Errorf("failed to get stream from DB: %w", err)
	}

	var dbStream DbStream
	err = attributevalue.UnmarshalMap(output.Item, &dbStream)

	if dbStream.StreamType != streamType {
		return estypes.Stream{}, fmt.Errorf("stream type mismatch")
	}

	stream, err := IntoStream(dbStream)
	if err != nil {
		return estypes.Stream{}, fmt.Errorf("failed to convert DbStream into Stream [%s]: %w", streamId, err)
	}

	return stream, nil
}
