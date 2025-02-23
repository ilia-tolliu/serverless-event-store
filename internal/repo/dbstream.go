package repo

import (
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

const RecordTypeStream = "stream"
const streamIndexName = "StreamIndex"

type DbStreamCreate struct {
	Pk             string    `dynamodbav:"PK"`
	Sk             int       `dynamodbav:"SK"`
	RecordType     string    `dynamodbav:"RecordType"`
	StreamType     string    `dynamodbav:"StreamType"`
	StreamRevision int       `dynamodbav:"StreamRevision"`
	CreatedAt      time.Time `dynamodbav:"CreatedAt"`
	UpdatedAt      time.Time `dynamodbav:"UpdatedAt"`
}

type dbStreamKey struct {
	Pk string `dynamodbav:"PK"`
	Sk int    `dynamodbav:"SK"`
}

func createFromStream(stream estypes.Stream) DbStreamCreate {
	dbStreamCreate := DbStreamCreate{
		Pk:             stream.StreamId.String(),
		Sk:             0,
		RecordType:     RecordTypeStream,
		StreamType:     stream.StreamType,
		StreamRevision: stream.Revision,
		CreatedAt:      stream.CreatedAt,
		UpdatedAt:      stream.UpdatedAt,
	}

	return dbStreamCreate
}

func keyFromStream(stream estypes.Stream) dbStreamKey {
	dbStreamKey := dbStreamKey{
		Pk: stream.StreamId.String(),
		Sk: 0,
	}

	return dbStreamKey
}

func PrepareDbStreamCreate(tableName string, stream estypes.Stream) (*types.Put, error) {
	dbStreamCreate := createFromStream(stream)

	value, err := attributevalue.MarshalMap(dbStreamCreate)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal db stream create: %w", err)
	}

	put := types.Put{
		Item:                value,
		TableName:           aws.String(tableName),
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	}

	return &put, nil
}

func PrepareDbStreamUpdate(tableName string, stream estypes.Stream) (*types.Update, error) {
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

func PrepareDbStreamGet(tableName string, streamId uuid.UUID) (*dynamodb.GetItemInput, error) {
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

func PrepareDbStreamsQuery(tableName string, streamType string, updatedAfter time.Time, nextPageKey string) (*dynamodb.QueryInput, error) {
	keyCond, err := expression.NewBuilder().
		WithKeyCondition(
			expression.Key("StreamType").Equal(expression.Value(streamType)).
				And(expression.Key("UpdatedAt").GreaterThanEqual(expression.Value(updatedAfter))),
		).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build key condition: %w", err)
	}

	var exclusiveStartKey map[string]types.AttributeValue
	if nextPageKey != "" {
		exclusiveStartKey, err = ParseNextPageKey(nextPageKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse next page key: %w", err)
		}
	}

	query := &dynamodb.QueryInput{
		KeyConditionExpression:    keyCond.KeyCondition(),
		ExpressionAttributeNames:  keyCond.Names(),
		ExpressionAttributeValues: keyCond.Values(),
		TableName:                 aws.String(tableName),
		IndexName:                 aws.String(streamIndexName),
		ScanIndexForward:          aws.Bool(true),
	}

	if nextPageKey != "" {
		query.ExclusiveStartKey = exclusiveStartKey
	}

	return query, nil
}

func IntoStream(dbStream DbStreamCreate) (estypes.Stream, error) {
	streamId, err := uuid.Parse(dbStream.Pk)
	if err != nil {
		return estypes.Stream{}, fmt.Errorf("failed to parse streamId: %w", err)
	}

	stream := estypes.Stream{
		StreamId:   streamId,
		StreamType: dbStream.StreamType,
		Revision:   dbStream.StreamRevision,
		CreatedAt:  dbStream.CreatedAt,
		UpdatedAt:  dbStream.UpdatedAt,
	}

	return stream, nil
}
