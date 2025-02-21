package repo

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/ilia-tolliu-go-event-store/internal/estypes"
	"strconv"
	"time"
)

const RecordTypeStream = "stream"

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

func IntoStream(dbstream DbStream) (estypes.Stream, error) {
	streamId, err := uuid.Parse(dbstream.Pk)
	if err != nil {
		return estypes.Stream{}, fmt.Errorf("failed to parse UUID: %w", err)
	}

	stream := estypes.Stream{
		StreamId:   streamId,
		StreamType: dbstream.StreamType,
		Revision:   dbstream.StreamRevision,
		CreatedAt:  dbstream.CreatedAt,
		UpdatedAt:  dbstream.UpdatedAt,
	}

	return stream, nil
}
