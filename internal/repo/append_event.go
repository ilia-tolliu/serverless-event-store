package repo

import (
	"context"
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

func (r *EsRepo) AppendEvent(ctx context.Context, streamType string, streamId uuid.UUID, revision int, newEvent estypes.NewEsEvent) (estypes.Stream, error) {
	log := logger.FromContext(ctx)
	now := time.Now()
	stream := estypes.Stream{
		StreamId:   streamId,
		StreamType: streamType,
		Revision:   revision,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	event := estypes.NewEvent(streamId, revision, newEvent, now)

	streamPut, err := PreparePutDbStream(r.tableName, stream)
	if err != nil {
		return estypes.Stream{}, err
	}

	StreamShouldHaveRevision(streamPut, revision-1)

	eventPut, err := PreparePutDbEvent(r.tableName, event)
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
		ClientRequestToken: aws.String(uuid.NewString()), // todo: use better idempotency token; should come from client
	})
	if err != nil {
		cancellation := &types.TransactionCanceledException{}
		if errors.As(err, &cancellation) {
			log.Errorw("transaction canceledException",
				"message", cancellation.Message,
				"reasons", cancellation.CancellationReasons,
				"code", cancellation.ErrorCode(),
				"errorMessage", cancellation.ErrorMessage(),
			)
			return estypes.Stream{}, err
		}
		return estypes.Stream{}, fmt.Errorf("failed to append event to stream: %w", err)
	}

	return stream, nil
}
