package webapp

import (
	"context"
	"fmt"
	"github.com/ilia-tolliu/serverless-event-store/estypes"
	"github.com/ilia-tolliu/serverless-event-store/internal/eserror"
	"github.com/ilia-tolliu/serverless-event-store/internal/webapp/types/resp"
	"net/http"
)

type getStreamDetailsResponse struct {
	Stream estypes.Stream `json:"stream"`
}

func (a *WebApp) HandleGetStreamDetails(ctx context.Context, r *http.Request) (resp.EsResponse, error) {
	streamType, err := ExtractStreamType(r)
	if err != nil {
		return resp.EsResponse{}, err
	}

	streamId, err := ExtractStreamId(r)
	if err != nil {
		return resp.EsResponse{}, err
	}

	stream, err := a.esRepo.GetStream(ctx, streamId)
	if err != nil {
		return resp.EsResponse{}, fmt.Errorf("failed to get stream details: %w", err)
	}

	err = stream.ShouldHaveType(streamType)
	if err != nil {
		return resp.EsResponse{}, eserror.NewNotFoundError(err)
	}

	responseBody := getStreamDetailsResponse{
		Stream: stream,
	}
	response := resp.New(resp.WithStatus(http.StatusOK), resp.WithJson(responseBody))

	return response, nil
}
