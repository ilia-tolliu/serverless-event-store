package repo

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/ilia-tolliu-go-event-store/internal/estypes"
	"time"
)

func (r *EsRepo) GetStreams(ctx context.Context, streamType string, updatedAfter time.Time, streamNextPageKey string) (estypes.StreamPage, error) {
	streamsQuery, err := PrepareDbStreamsQuery(r.tableName, streamType, updatedAfter, streamNextPageKey)
	if err != nil {
		return estypes.StreamPage{}, fmt.Errorf("failed to prepare DbStreamsQuery: %w", err)
	}

	output, err := r.dynamoDb.Query(ctx, streamsQuery)
	if err != nil {
		return estypes.StreamPage{}, fmt.Errorf("failed to get streams from DB: %w", err)
	}

	streams := make([]estypes.Stream, 0, len(output.Items))
	lastEvaluatedKey := output.LastEvaluatedKey
	var newNextPageKey string
	if lastEvaluatedKey != nil {
		newNextPageKey, err = FormatNextPageKey(lastEvaluatedKey)
		if err != nil {
			return estypes.StreamPage{}, fmt.Errorf("failed to format next page key: %w", err)
		}
	}

	for _, item := range output.Items {
		var dbStream DbStream
		err = attributevalue.UnmarshalMap(item, &dbStream)
		if err != nil {
			return estypes.StreamPage{}, fmt.Errorf("failed to unmarshal stream from DB: %w", err)
		}

		stream, err := IntoStream(dbStream)
		if err != nil {
			return estypes.StreamPage{}, fmt.Errorf("failed to convert DbStream into Stream [%s]: %w", dbStream.Pk, err)
		}

		streams = append(streams, stream)
	}

	page := estypes.StreamPage{
		Streams: streams,
		HasMore: output.LastEvaluatedKey != nil,
	}
	if newNextPageKey != "" {
		page.NextPageKey = &newNextPageKey
	}

	return page, nil
}
