package webapp

import (
	"context"
	"github.com/ilia-tolliu/serverless-event-store/internal/webapp/types/resp"
	"net/http"
	"time"
)

type livenessCheckResponse struct {
	Status string    `json:"status"`
	At     time.Time `json:"at"`
}

func (a *WebApp) HandleLivenessCheck(_ context.Context, _ *http.Request) (resp.EsResponse, error) {
	responseBody := livenessCheckResponse{
		Status: "ok",
		At:     time.Now(),
	}
	response := resp.New(resp.WithStatus(http.StatusOK), resp.WithJson(responseBody))

	return response, nil
}
