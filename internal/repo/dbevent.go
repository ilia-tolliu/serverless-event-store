package repo

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/ilia-tolliu-go-event-store/estypes"
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

func FromEvent(event estypes.Event) DbEvent {
	return DbEvent{
		Pk:         event.StreamId.String(),
		Sk:         event.Revision,
		RecordType: RecordTypeEvent,
		EventType:  event.EventType,
		Payload:    event.Payload,
		CreatedAt:  event.CreatedAt,
	}
}

func IntoEvent(dbEvent DbEvent) (estypes.Event, error) {
	streamId, err := uuid.Parse(dbEvent.Pk)
	if err != nil {
		return estypes.Event{}, fmt.Errorf("failed to parse streamId: %w", err)
	}

	event := estypes.Event{
		StreamId:  streamId,
		Revision:  dbEvent.Sk,
		EventType: dbEvent.EventType,
		Payload:   dbEvent.Payload,
		CreatedAt: dbEvent.CreatedAt,
	}

	return event, nil
}

func PreparePutEventQuery(tableName string, event estypes.Event) (*types.Put, error) {
	dbEvent := FromEvent(event)

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
