package webapp

import (
	"context"
	"net/http"
	"time"
)

type livenessCheckResponse struct {
	Status string    `json:"status"`
	At     time.Time `json:"at"`
}

func (a *WebApp) HandleLivenessCheck(_ context.Context, _ *http.Request) (Response, error) {
	responseBody := livenessCheckResponse{
		Status: "ok",
		At:     time.Now(),
	}
	response := NewResponse(Status(http.StatusOK), Json(responseBody))

	return response, nil
}
