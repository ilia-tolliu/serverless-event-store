package web

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type livenessCheckResponse struct {
	Status string    `json:"status"`
	At     time.Time `json:"at"`
}

func (a *EsWebApp) HandleLivenessCheck(_ context.Context, w http.ResponseWriter, _ *http.Request) error {
	response := livenessCheckResponse{
		Status: "ok",
		At:     time.Now(),
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	return encoder.Encode(response)
}
