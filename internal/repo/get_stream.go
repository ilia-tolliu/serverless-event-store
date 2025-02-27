package repo

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	"github.com/ilia-tolliu-go-event-store/estypes"
	"github.com/ilia-tolliu-go-event-store/internal/eserror"
)

func (r *EsRepo) GetStream(ctx context.Context, streamId uuid.UUID) (estypes.Stream, error) {
	streamGet, err := prepareStreamGet(r.tableName, streamId)
	if err != nil {
		return estypes.Stream{}, fmt.Errorf("failed to prepare GetDbStream: %w", err)
	}

	output, err := r.dynamoDb.GetItem(ctx, streamGet)
	if err != nil {
		return estypes.Stream{}, fmt.Errorf("failed to get stream from DB: %w", err)
	}

	if output.Item == nil {
		err = fmt.Errorf("stream not found")
		return estypes.Stream{}, eserror.NewNotFoundError(err)
	}

	var dbStream DbStream
	err = attributevalue.UnmarshalMap(output.Item, &dbStream)
	if err != nil {
		return estypes.Stream{}, fmt.Errorf("failed to unmarshal stream from DB: %w", err)
	}

	stream, err := IntoStream(dbStream)
	if err != nil {
		return estypes.Stream{}, fmt.Errorf("failed to convert DbStream into Stream [%s]: %w", streamId, err)
	}

	return stream, nil
}

func prepareStreamGet(tableName string, streamId uuid.UUID) (*dynamodb.GetItemInput, error) {
	keySrc := map[string]any{
		"PK": streamId.String(),
		"SK": 0,
	}

	key, err := attributevalue.MarshalMap(keySrc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal stream key: %w", err)
	}

	get := &dynamodb.GetItemInput{
		Key:       key,
		TableName: aws.String(tableName),
	}

	return get, nil
}
