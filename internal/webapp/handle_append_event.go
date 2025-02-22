package webapp

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ilia-tolliu-go-event-store/internal/eserror"
	"github.com/ilia-tolliu-go-event-store/internal/estypes"
	"net/http"
)

type appendEventRequest struct {
	Event estypes.NewEsEvent `json:"event"`
}

type appendEventResponse struct {
	Stream estypes.Stream `json:"stream"`
}

func (a *WebApp) HandleAppendEvent(ctx context.Context, r *http.Request) (Response, error) {
	streamType, err := ExtractStreamType(r)
	if err != nil {
		return Response{}, err
	}

	streamId, err := ExtractStreamId(r)
	if err != nil {
		return Response{}, err
	}

	streamRevision, err := ExtractStreamRevision(r)
	if err != nil {
		return Response{}, err
	}

	var reqBody appendEventRequest
	err = json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		return Response{}, fmt.Errorf("failed to parse request body: %w", err)
	}

	stream, err := a.esRepo.GetStream(ctx, streamId)
	if err != nil {
		return Response{}, fmt.Errorf("failed to get stream from event store: %w", err)
	}

	err = stream.ShouldHaveType(streamType)
	if err != nil {
		return Response{}, eserror.NewNotFoundError(err)
	}

	err = stream.ShouldHaveRevision(streamRevision - 1)
	if err != nil {
		return Response{}, eserror.NewDataConflictError(err)
	}

	stream, err = a.esRepo.AppendEvent(ctx, streamType, streamId, streamRevision, reqBody.Event)
	if err != nil {
		return Response{}, fmt.Errorf("failed to append event to stream: %w", err)
	}

	responseBody := appendEventResponse{
		Stream: stream,
	}
	response := NewResponse(Status(http.StatusCreated), Json(responseBody))

	return response, nil
}
