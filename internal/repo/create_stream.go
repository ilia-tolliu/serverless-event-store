package repo

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/ilia-tolliu-go-event-store/estypes"
	"time"
)

func (r *EsRepo) CreateStream(ctx context.Context, streamType string, initialEvent estypes.NewEsEvent) (estypes.Stream, error) {
	streamId := uuid.New()
	now := time.Now()
	stream := estypes.NewStream(streamId, streamType, now)
	event := estypes.NewEvent(streamId, 1, initialEvent, now)

	streamPut, err := prepareStreamPut(r.tableName, stream)
	if err != nil {
		return estypes.Stream{}, err
	}

	eventPut, err := PreparePutEventQuery(r.tableName, event)
	if err != nil {
		return estypes.Stream{}, err
	}

	_, err = r.dynamoDb.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Put: streamPut,
			},
			{
				Put: eventPut,
			},
		},
		ClientRequestToken: aws.String(streamId.String()), // todo: use better idempotency token; should come from client
	})
	if err != nil {
		return estypes.Stream{}, fmt.Errorf("failed to create stream: %w", err)
	}

	return stream, nil
}

func prepareStreamPut(tableName string, stream estypes.Stream) (*types.Put, error) {
	dbStream := FromStream(stream)

	value, err := attributevalue.MarshalMap(dbStream)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal db stream: %w", err)
	}

	put := types.Put{
		Item:                value,
		TableName:           aws.String(tableName),
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	}

	return &put, nil
}
