package repo

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ilia-tolliu-go-event-store/internal/estypes"
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
