package webapp

import (
	"context"
	"fmt"
	"github.com/ilia-tolliu-go-event-store/estypes"
	"github.com/ilia-tolliu-go-event-store/internal/eserror"
	"net/http"
	"strconv"
)

type getEventsResponse struct {
	EventPage estypes.EventPage `json:"eventPage"`
}

func (a *WebApp) HandleGetStreamEvents(ctx context.Context, r *http.Request) (Response, error) {
	streamType, err := ExtractStreamType(r)
	if err != nil {
		return Response{}, err
	}

	streamId, err := ExtractStreamId(r)
	if err != nil {
		return Response{}, err
	}

	afterRevision, err := extractAfterRevision(r)
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

	eventPage, err := a.esRepo.GetEvents(ctx, streamId, afterRevision)
	if err != nil {
		return Response{}, fmt.Errorf("failed to get events: %w", err)
	}

	responseBody := getEventsResponse{
		EventPage: eventPage,
	}
	response := NewResponse(Status(http.StatusOK), Json(responseBody))

	return response, nil
}

func extractAfterRevision(r *http.Request) (int, error) {
	afterRevisionStr := r.URL.Query().Get("after-revision")
	if afterRevisionStr == "" {
		return 0, nil
	}

	afterRevision, err := strconv.Atoi(afterRevisionStr)
	if err != nil {
		return 0, fmt.Errorf("invalid after-revision value: %w", err)
	}

	return afterRevision, nil
}
