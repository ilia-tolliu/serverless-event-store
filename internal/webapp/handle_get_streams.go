package webapp

import (
	"context"
	"fmt"
	"github.com/ilia-tolliu/serverless-event-store/estypes"
	"github.com/ilia-tolliu/serverless-event-store/internal/webapp/types/resp"
	"net/http"
	"time"
)

type getStreamsResponse struct {
	StreamPage estypes.StreamPage `json:"streamPage"`
}

func (a *WebApp) HandleGetStreams(ctx context.Context, r *http.Request) (resp.EsResponse, error) {
	streamType, err := ExtractStreamType(r)
	if err != nil {
		return resp.EsResponse{}, err
	}

	updatedAfter, err := extractUpdatedAfter(r)
	if err != nil {
		return resp.EsResponse{}, err
	}

	nextPageKey, err := extractStreamNextPageKey(r)
	if err != nil {
		return resp.EsResponse{}, err
	}

	streamPage, err := a.esRepo.GetStreams(ctx, streamType, updatedAfter, nextPageKey)
	if err != nil {
		return resp.EsResponse{}, fmt.Errorf("failed to get streams: %w", err)
	}

	responseBody := getStreamsResponse{
		StreamPage: streamPage,
	}
	response := resp.New(resp.WithStatus(http.StatusOK), resp.WithJson(responseBody))

	return response, nil
}

func extractUpdatedAfter(r *http.Request) (time.Time, error) {
	zero := time.Unix(0, 0)

	updatedAfterStr := r.URL.Query().Get("updated-after")
	if updatedAfterStr == "" {
		return zero, nil
	}

	updatedAfter, err := time.Parse(time.RFC3339Nano, updatedAfterStr)
	if err != nil {
		return zero, fmt.Errorf("invalid updated-after value: %w", err)
	}

	return updatedAfter, nil
}

func extractStreamNextPageKey(r *http.Request) (string, error) {
	nextPageKey := r.URL.Query().Get("stream-next-page-key")

	return nextPageKey, nil
}
