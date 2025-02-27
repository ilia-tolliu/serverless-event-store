package repo

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	estypes2 "github.com/ilia-tolliu-go-event-store/estypes"
	"time"
)

func (r *EsRepo) GetStreams(ctx context.Context, streamType string, updatedAfter time.Time, streamNextPageKey string) (estypes2.StreamPage, error) {
	streamsQuery, err := prepareStreamsQuery(r.tableName, streamType, updatedAfter, streamNextPageKey)
	if err != nil {
		return estypes2.StreamPage{}, fmt.Errorf("failed to prepare DbStreamsQuery: %w", err)
	}

	output, err := r.dynamoDb.Query(ctx, streamsQuery)
	if err != nil {
		return estypes2.StreamPage{}, fmt.Errorf("failed to get streams from DB: %w", err)
	}

	streams := make([]estypes2.Stream, 0, len(output.Items))
	lastEvaluatedKey := output.LastEvaluatedKey
	var newNextPageKey string
	if lastEvaluatedKey != nil {
		newNextPageKey, err = FormatNextPageKey(lastEvaluatedKey)
		if err != nil {
			return estypes2.StreamPage{}, fmt.Errorf("failed to format next page key: %w", err)
		}
	}

	for _, item := range output.Items {
		var dbStream DbStream
		err = attributevalue.UnmarshalMap(item, &dbStream)
		if err != nil {
			return estypes2.StreamPage{}, fmt.Errorf("failed to unmarshal stream from DB: %w", err)
		}

		stream, err := IntoStream(dbStream)
		if err != nil {
			return estypes2.StreamPage{}, fmt.Errorf("failed to convert DbStream into Stream [%s]: %w", dbStream.Pk, err)
		}

		streams = append(streams, stream)
	}

	page := estypes2.StreamPage{
		Streams: streams,
		HasMore: output.LastEvaluatedKey != nil,
	}
	if newNextPageKey != "" {
		page.NextPageKey = &newNextPageKey
	}

	return page, nil
}

func prepareStreamsQuery(tableName string, streamType string, updatedAfter time.Time, nextPageKey string) (*dynamodb.QueryInput, error) {
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
