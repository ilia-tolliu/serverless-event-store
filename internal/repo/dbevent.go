package repo

import (
	"encoding/json"
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

const RecordTypeEvent = "event"

type DbEvent struct {
	Pk         string    `dynamodbav:"PK"`
	Sk         int       `dynamodbav:"SK"`
	RecordType string    `dynamodbav:"RecordType"`
	EventType  string    `dynamodbav:"EventType"`
	Payload    string    `dynamodbav:"Payload"`
	CreatedAt  time.Time `dynamodbav:"CreatedAt"`
}

func fromEvent(event estypes.Event) (DbEvent, error) {
	payload, err := json.MarshalIndent(event.Payload, "", "  ")
	if err != nil {
		return DbEvent{}, fmt.Errorf("failed to marshal event payload: %w", err)
	}
	payloadStr := string(payload)

	dbEvent := DbEvent{
		Pk:         event.StreamId.String(),
		Sk:         event.Revision,
		RecordType: RecordTypeEvent,
		EventType:  event.EventType,
		Payload:    payloadStr,
		CreatedAt:  event.CreatedAt,
	}

	return dbEvent, nil
}

func PrepareDbEventPut(tableName string, event estypes.Event) (*types.Put, error) {
	dbEvent, err := fromEvent(event)
	if err != nil {
		return nil, err
	}

	value, err := attributevalue.MarshalMap(dbEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal db event: %w", err)
	}

	put := types.Put{
		Item:                value,
		TableName:           aws.String(tableName),
		ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
	}

	return &put, nil
}

func PrepareDbEventsQuery(tableName string, streamId uuid.UUID, afterRevision int) (*dynamodb.QueryInput, error) {
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

func IntoEvent(dbEvent DbEvent) (estypes.Event, error) {
	streamId, err := uuid.Parse(dbEvent.Pk)
	if err != nil {
		return estypes.Event{}, fmt.Errorf("failed to parse streamId: %w", err)
	}

	payload := make(map[string]interface{})
	err = json.Unmarshal([]byte(dbEvent.Payload), &payload)
	if err != nil {
		return estypes.Event{}, fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	event := estypes.Event{
		StreamId:  streamId,
		Revision:  dbEvent.Sk,
		EventType: dbEvent.EventType,
		Payload:   payload,
		CreatedAt: dbEvent.CreatedAt,
	}

	return event, nil
}
