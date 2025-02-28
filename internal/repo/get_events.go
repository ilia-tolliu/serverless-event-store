package repo

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	estypes2 "github.com/ilia-tolliu/serverless-event-store/estypes"
)

func (r *EsRepo) GetEvents(ctx context.Context, streamId uuid.UUID, afterRevision int) (estypes2.EventPage, error) {
	eventsQuery, err := prepareEventsQuery(r.tableName, streamId, afterRevision)
	if err != nil {
		return estypes2.EventPage{}, fmt.Errorf("failed to prepare DbEventsQuery: %w", err)
	}

	output, err := r.dynamoDb.Query(ctx, eventsQuery)
	if err != nil {
		return estypes2.EventPage{}, fmt.Errorf("failed to get events from DB: %w", err)
	}

	events := make([]estypes2.Event, 0, len(output.Items))
	var lastEvaluatedRevision int

	for _, item := range output.Items {
		var dbEvent DbEvent
		err = attributevalue.UnmarshalMap(item, &dbEvent)
		if err != nil {
			return estypes2.EventPage{}, fmt.Errorf("failed to unmarshal event from DB: %w", err)
		}

		event, err := IntoEvent(dbEvent)
		if err != nil {
			return estypes2.EventPage{}, fmt.Errorf("failed to convert DbEvent into Event [%s::%d]: %w", streamId, dbEvent.Sk, err)
		}

		lastEvaluatedRevision = event.Revision
		events = append(events, event)
	}

	page := estypes2.EventPage{
		Events:                events,
		HasMore:               output.LastEvaluatedKey != nil,
		LastEvaluatedRevision: lastEvaluatedRevision,
	}

	return page, nil
}

func prepareEventsQuery(tableName string, streamId uuid.UUID, afterRevision int) (*dynamodb.QueryInput, error) {
	keyCond, err := expression.NewBuilder().
		WithKeyCondition(expression.Key("PK").Equal(expression.Value(streamId.String()))).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build key condition: %w", err)
	}

	exclusiveStartKeySrc := map[string]any{
		"PK": streamId.String(),
		"SK": afterRevision,
	}
	exclusiveStartKey, err := attributevalue.MarshalMap(exclusiveStartKeySrc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal exclusiveStartKey: %w", err)
	}

	query := &dynamodb.QueryInput{
		KeyConditionExpression:    keyCond.KeyCondition(),
		ExpressionAttributeNames:  keyCond.Names(),
		ExpressionAttributeValues: keyCond.Values(),
		ExclusiveStartKey:         exclusiveStartKey,
		TableName:                 aws.String(tableName),
	}

	return query, nil
}
