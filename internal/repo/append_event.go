package repo

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/ilia-tolliu-go-event-store/internal/estypes"
	"time"
)

func (r *EsRepo) AppendEvent(ctx context.Context, streamType string, streamId uuid.UUID, revision int, newEvent estypes.NewEsEvent) (estypes.Stream, error) {
	now := time.Now()
	stream := estypes.Stream{
		StreamId:   streamId,
		StreamType: streamType,
		Revision:   revision,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	event := estypes.NewEvent(streamId, revision, newEvent, now)

	streamUpdate, err := prepareStreamUpdate(r.tableName, stream)
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
				Update: streamUpdate,
			},
			{
				Put: eventPut,
			},
		},
		ClientRequestToken: aws.String(uuid.NewString()), // todo: use better idempotency token; should come from client
	})
	if err != nil {
		return estypes.Stream{}, fmt.Errorf("failed to complete DB transaction: %w", err)
	}

	return stream, nil
}

func prepareStreamUpdate(tableName string, stream estypes.Stream) (*types.Update, error) {
	updateExpr, err := expression.NewBuilder().WithUpdate(
		expression.
			Set(expression.Name("StreamRevision"), expression.Value(stream.Revision)).
			Set(expression.Name("UpdatedAt"), expression.Value(stream.UpdatedAt)),
	).
		WithCondition(expression.Name("StreamRevision").Equal(expression.Value(stream.Revision - 1))).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build update expression: %w", err)
	}

	streamKey := keyFromStream(stream)
	streamKeyValue, err := attributevalue.MarshalMap(streamKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal stream key: %w", err)
	}

	update := types.Update{
		Key:                       streamKeyValue,
		TableName:                 aws.String(tableName),
		UpdateExpression:          updateExpr.Update(),
		ExpressionAttributeNames:  updateExpr.Names(),
		ExpressionAttributeValues: updateExpr.Values(),
		ConditionExpression:       updateExpr.Condition(),
	}

	return &update, nil
}

func keyFromStream(stream estypes.Stream) dbStreamKey {
	dbStreamKey := dbStreamKey{
		Pk: stream.StreamId.String(),
		Sk: 0,
	}

	return dbStreamKey
}
