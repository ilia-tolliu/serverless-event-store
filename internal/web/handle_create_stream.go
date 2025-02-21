package web

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ilia-tolliu-go-event-store/internal/estypes"
	"net/http"
)

type createStreamRequest struct {
	InitialEvent estypes.NewEsEvent `json:"initialEvent"`
}

type createStreamResponse struct {
	Stream estypes.Stream `json:"stream"`
}

func (a *EsWebApp) HandleCreateStream(ctx context.Context, r *http.Request) (Response, error) {
	streamType := r.PathValue("streamType")
	if streamType == "" {
		return Response{}, fmt.Errorf("no streamType specified")
	}

	var reqBody createStreamRequest
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		return Response{}, fmt.Errorf("failed to parse request body: %w", err)
	}

	stream, err := a.esRepo.CreateStream(ctx, streamType, reqBody.InitialEvent)
	if err != nil {
		return Response{}, fmt.Errorf("failed to create stream: %w", err)
	}

	responseBody := createStreamResponse{
		Stream: stream,
	}
	response := NewResponse(Status(http.StatusCreated), Json(responseBody))

	return response, nil
}
