package webapp

import (
	"context"
	"fmt"
	"github.com/ilia-tolliu-go-event-store/internal/eserror"
	"github.com/ilia-tolliu-go-event-store/internal/estypes"
	"net/http"
)

type getStreamDetailsResponse struct {
	Stream estypes.Stream `json:"stream"`
}

func (a *WebApp) HandleGetStreamDetails(ctx context.Context, r *http.Request) (Response, error) {
	streamType, err := ExtractStreamType(r)
	if err != nil {
		return Response{}, err
	}

	streamId, err := ExtractStreamId(r)
	if err != nil {
		return Response{}, err
	}

	stream, err := a.esRepo.GetStream(ctx, streamId)
	if err != nil {
		return Response{}, fmt.Errorf("failed to get stream details: %w", err)
	}

	err = stream.ShouldHaveType(streamType)
	if err != nil {
		return Response{}, eserror.NewNotFoundError(err)
	}

	responseBody := getStreamDetailsResponse{
		Stream: stream,
	}
	response := NewResponse(Status(http.StatusOK), Json(responseBody))

	return response, nil
}
