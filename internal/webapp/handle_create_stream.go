package webapp

import (
	"context"
	"fmt"
	"github.com/ilia-tolliu/serverless-event-store/estypes"
	"github.com/ilia-tolliu/serverless-event-store/internal/esvalidate"
	"github.com/ilia-tolliu/serverless-event-store/internal/webapp/types/resp"
	"net/http"
)

type createStreamRequest struct {
	InitialEvent *estypes.NewEsEvent `json:"initialEvent,omitempty" validate:"required"`
}

type createStreamResponse struct {
	Stream estypes.Stream `json:"stream"`
}

func (a *WebApp) HandleCreateStream(ctx context.Context, r *http.Request) (resp.EsResponse, error) {
	streamType, err := ExtractStreamType(r)
	if err != nil {
		return resp.EsResponse{}, err
	}

	var reqBody createStreamRequest
	err = ExtractRequestBody(r, &reqBody)
	if err != nil {
		return resp.EsResponse{}, err
	}

	err = esvalidate.Validate(reqBody)
	if err != nil {
		return resp.EsResponse{}, err
	}

	stream, err := a.esRepo.CreateStream(ctx, streamType, *reqBody.InitialEvent)
	if err != nil {
		return resp.EsResponse{}, fmt.Errorf("failed to create stream: %w", err)
	}

	responseBody := createStreamResponse{
		Stream: stream,
	}
	response := resp.New(resp.WithStatus(http.StatusCreated), resp.WithJson(responseBody))

	return response, nil
}
