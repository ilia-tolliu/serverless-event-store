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
	"strconv"
	"strings"
	"time"
)

const RecordTypeStream = "stream"
const streamIndexName = "StreamIndex"

type DbStream struct {
	Pk             string    `dynamodbav:"PK"`
	Sk             int       `dynamodbav:"SK"`
	RecordType     string    `dynamodbav:"RecordType"`
	StreamType     string    `dynamodbav:"StreamType"`
	StreamRevision int       `dynamodbav:"StreamRevision"`
	CreatedAt      time.Time `dynamodbav:"CreatedAt"`
	UpdatedAt      time.Time `dynamodbav:"UpdatedAt"`
}

func fromStream(stream estypes.Stream) DbStream {
	dbStream := DbStream{
		Pk:             stream.StreamId.String(),
		Sk:             0,
		RecordType:     RecordTypeStream,
		StreamType:     stream.StreamType,
		StreamRevision: stream.Revision,
		CreatedAt:      stream.CreatedAt,
		UpdatedAt:      stream.UpdatedAt,
	}

	return dbStream
}

func PreparePutDbStream(tableName string, stream estypes.Stream) (*types.Put, error) {
	dbStream := fromStream(stream)

	value, err := attributevalue.MarshalMap(dbStream)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal db stream: %w", err)
	}

	put := types.Put{
		Item:      value,
		TableName: aws.String(tableName),
	}

	return &put, nil
}

func StreamShouldNotExist(put *types.Put) {
	put.ConditionExpression = aws.String("attribute_not_exists(PK)")
}

func StreamShouldHaveRevision(put *types.Put, revision int) {
	put.ConditionExpression = aws.String("#StreamRevision = :StreamRevision")
	put.ExpressionAttributeNames = map[string]string{
		"#StreamRevision": "StreamRevision",
	}
	put.ExpressionAttributeValues = map[string]types.AttributeValue{
		":StreamRevision": &types.AttributeValueMemberN{
			Value: strconv.Itoa(revision),
		},
	}
}

func PrepareGetDbStream(tableName string, streamId uuid.UUID) (*dynamodb.GetItemInput, error) {
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

func IntoStream(dbStream DbStream) (estypes.Stream, error) {
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

type streamNextPageKey struct {
	Pk         string `dynamodbav:"PK"`
	Sk         int    `dynamodbav:"SK"`
	StreamType string
	UpdatedAt  string
}

func ParseNextPageKey(nextPageKey string) (map[string]types.AttributeValue, error) {
	parts := strings.Split(nextPageKey, "|")
	streamNextPageKey := streamNextPageKey{
		Pk:         parts[0],
		Sk:         0,
		StreamType: parts[1],
		UpdatedAt:  parts[2],
	}

	key, err := attributevalue.MarshalMap(streamNextPageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal next page key: %w", err)
	}

	return key, nil
}

func FormatNextPageKey(lastEvaluatedKey map[string]types.AttributeValue) (string, error) {
	var streamNextPageKey streamNextPageKey
	err := attributevalue.UnmarshalMap(lastEvaluatedKey, &streamNextPageKey)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal next page key: %w", err)
	}

	parts := []string{
		streamNextPageKey.Pk,
		streamNextPageKey.StreamType,
		streamNextPageKey.UpdatedAt,
	}

	return strings.Join(parts, "|"), nil
}
