package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/ilia-tolliu-go-event-store/internal/estypes"
	"github.com/ilia-tolliu-go-event-store/internal/logger"
	"time"
)

type EsRepo struct {
	dynamoDb  *dynamodb.Client
	tableName string
}

func NewEsRepo(dynamoDb *dynamodb.Client, tableName string) *EsRepo {
	return &EsRepo{
		dynamoDb:  dynamoDb,
		tableName: tableName,
	}
}

func (r *EsRepo) CreateStream(ctx context.Context, streamType string, initialEvent estypes.NewEvent) (estypes.Stream, error) {
	log := logger.FromContext(ctx)
	streamId := uuid.New()
	now := time.Now()
	stream := estypes.Stream{
		StreamId:   streamId,
		StreamType: streamType,
		Revision:   1,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	streamPut, err := r.preparePutStream(stream)
	if err != nil {
		return estypes.Stream{}, err
	}

	eventPut, err := r.preparePutInitialEvent(streamId, initialEvent, now)
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
		ClientRequestToken: aws.String(streamId.String()),
	})
	if err != nil {
		cancellation := &types.TransactionCanceledException{}
		if errors.As(err, &cancellation) {
			log.Errorw("transaction canceledException",
				"message", cancellation.Message,
				"reasons", cancellation.CancellationReasons,
				"code", cancellation.ErrorCode(),
			)
			return estypes.Stream{}, err
		}
		return estypes.Stream{}, fmt.Errorf("failed to create stream: %w", err)
	}

	return stream, nil
}

func (r *EsRepo) preparePutStream(stream estypes.Stream) (*types.Put, error) {
	value := map[string]types.AttributeValue{
		"PK":         &types.AttributeValueMemberS{Value: stream.StreamId.String()},
		"SK":         &types.AttributeValueMemberN{Value: "0"},
		"streamType": &types.AttributeValueMemberS{Value: stream.StreamType},
		"revision":   &types.AttributeValueMemberN{Value: "1"},
		"createdAt":  &types.AttributeValueMemberS{Value: stream.CreatedAt.Format(time.RFC3339)},
		"updatedAt":  &types.AttributeValueMemberS{Value: stream.UpdatedAt.Format(time.RFC3339)},
	}

	put := types.Put{
		Item:                value,
		TableName:           aws.String(r.tableName),
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	}

	return &put, nil
}

func (r *EsRepo) preparePutInitialEvent(streamId uuid.UUID, initialEvent estypes.NewEvent, now time.Time) (*types.Put, error) {
	payload, err := json.Marshal(initialEvent.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal initial event payload: %w", err)
	}
	value := map[string]types.AttributeValue{
		"PK":        &types.AttributeValueMemberS{Value: streamId.String()},
		"SK":        &types.AttributeValueMemberN{Value: "1"},
		"eventType": &types.AttributeValueMemberS{Value: initialEvent.EventType},
		"payload":   &types.AttributeValueMemberS{Value: string(payload)},
		"createdAt": &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
	}

	put := types.Put{
		Item:      value,
		TableName: aws.String(r.tableName),
	}

	return &put, nil
}
