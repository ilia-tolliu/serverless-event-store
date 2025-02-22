package webapp

import (
	"context"
	"fmt"
	"github.com/ilia-tolliu-go-event-store/internal/estypes"
	"github.com/ilia-tolliu-go-event-store/internal/esvalidate"
	"net/http"
)

type createStreamRequest struct {
	InitialEvent *estypes.NewEsEvent `json:"initialEvent,omitempty" validate:"required"`
}

type createStreamResponse struct {
	Stream estypes.Stream `json:"stream"`
}

func (a *WebApp) HandleCreateStream(ctx context.Context, r *http.Request) (Response, error) {
	streamType, err := ExtractStreamType(r)
	if err != nil {
		return Response{}, err
	}

	var reqBody createStreamRequest
	err = ExtractRequestBody(r, &reqBody)
	if err != nil {
		return Response{}, err
	}

	err = esvalidate.Validate(reqBody)
	if err != nil {
		return Response{}, err
	}

	stream, err := a.esRepo.CreateStream(ctx, streamType, *reqBody.InitialEvent)
	if err != nil {
		return Response{}, fmt.Errorf("failed to create stream: %w", err)
	}

	responseBody := createStreamResponse{
		Stream: stream,
	}
	response := NewResponse(Status(http.StatusCreated), Json(responseBody))

	return response, nil
}
