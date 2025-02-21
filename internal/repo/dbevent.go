package repo

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ilia-tolliu-go-event-store/internal/estypes"
	"time"
)

const RecordTypeEvent = "event"

type DbEvent struct {
	Pk         string    `dynamodbav:"PK"`
	Sk         int       `dynamodbav:"SK"`
	RecordType string    `dynamodbav:"RecordType"`
	EventType  string    `dynamodbav:"EventType"`
	Payload    any       `dynamodbav:"Payload"`
	CreatedAt  time.Time `dynamodbav:"CreatedAt"`
}

func fromEvent(event estypes.Event) (DbEvent, error) {
	payload, err := json.MarshalIndent(event.Payload, "", "  ")
	if err != nil {
		return DbEvent{}, fmt.Errorf("failed to marshal event payload: %w", err)
	}

	dbEvent := DbEvent{
		Pk:         event.StreamId.String(),
		Sk:         event.Revision,
		RecordType: RecordTypeEvent,
		EventType:  event.EventType,
		Payload:    string(payload),
		CreatedAt:  event.CreatedAt,
	}

	return dbEvent, nil
}

func PreparePutDbEvent(tableName string, event estypes.Event) (*types.Put, error) {
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
