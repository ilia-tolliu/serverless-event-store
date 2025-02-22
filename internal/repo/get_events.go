package repo

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/google/uuid"
	"github.com/ilia-tolliu-go-event-store/internal/estypes"
)

func (r *EsRepo) GetEvents(ctx context.Context, streamId uuid.UUID, afterRevision int) (estypes.EventPage, error) {
	eventsQuery, err := PrepareDbEventsQuery(r.tableName, streamId, afterRevision)
	if err != nil {
		return estypes.EventPage{}, fmt.Errorf("failed to prepare DbEventsQuery: %w", err)
	}

	output, err := r.dynamoDb.Query(ctx, eventsQuery)
	if err != nil {
		return estypes.EventPage{}, fmt.Errorf("failed to get events from DB: %w", err)
	}

	events := make([]estypes.Event, 0, len(output.Items))
	var lastEvaluatedRevision int

	for _, item := range output.Items {
		var dbEvent DbEvent
		err = attributevalue.UnmarshalMap(item, &dbEvent)
		if err != nil {
			return estypes.EventPage{}, fmt.Errorf("failed to unmarshal event from DB: %w", err)
		}

		event, err := IntoEvent(dbEvent)
		if err != nil {
			return estypes.EventPage{}, fmt.Errorf("failed to convert DbEvent into Event [%s::%d]: %w", streamId, dbEvent.Sk, err)
		}

		lastEvaluatedRevision = event.Revision
		events = append(events, event)
	}

	page := estypes.EventPage{
		Events:               events,
		HasMore:              output.LastEvaluatedKey != nil,
		LasEvaluatedRevision: lastEvaluatedRevision,
	}

	return page, nil
}
