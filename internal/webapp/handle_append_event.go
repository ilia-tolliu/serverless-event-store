package webapp

import (
	"context"
	"fmt"
	"github.com/ilia-tolliu/serverless-event-store/estypes"
	"github.com/ilia-tolliu/serverless-event-store/internal/eserror"
	"github.com/ilia-tolliu/serverless-event-store/internal/esvalidate"
	"github.com/ilia-tolliu/serverless-event-store/internal/webapp/types/resp"
	"net/http"
)

type appendEventRequest struct {
	Event *estypes.NewEsEvent `json:"event,omitempty" validate:"required"`
}

type appendEventResponse struct {
	Stream estypes.Stream `json:"stream"`
}

func (a *WebApp) HandleAppendEvent(ctx context.Context, r *http.Request) (resp.EsResponse, error) {
	streamType, err := ExtractStreamType(r)
	if err != nil {
		return resp.EsResponse{}, err
	}

	streamId, err := ExtractStreamId(r)
	if err != nil {
		return resp.EsResponse{}, err
	}

	streamRevision, err := ExtractStreamRevision(r)
	if err != nil {
		return resp.EsResponse{}, err
	}

	var reqBody appendEventRequest
	err = ExtractRequestBody(r, &reqBody)
	if err != nil {
		return resp.EsResponse{}, err
	}

	err = esvalidate.Validate(&reqBody)
	if err != nil {
		return resp.EsResponse{}, err
	}

	stream, err := a.esRepo.GetStream(ctx, streamId)
	if err != nil {
		return resp.EsResponse{}, fmt.Errorf("failed to get stream from event store: %w", err)
	}

	err = stream.ShouldHaveType(streamType)
	if err != nil {
		return resp.EsResponse{}, eserror.NewNotFoundError(err)
	}

	err = stream.ShouldHaveRevision(streamRevision - 1)
	if err != nil {
		return resp.EsResponse{}, eserror.NewDataConflictError(err)
	}

	stream, err = a.esRepo.AppendEvent(ctx, streamType, streamId, streamRevision, *reqBody.Event)
	if err != nil {
		return resp.EsResponse{}, fmt.Errorf("failed to append event to stream: %w", err)
	}

	responseBody := appendEventResponse{
		Stream: stream,
	}
	response := resp.New(resp.WithStatus(http.StatusCreated), resp.WithJson(responseBody))

	return response, nil
}
