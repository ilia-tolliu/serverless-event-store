package web

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type livenessCheckResponse struct {
	Status string    `json:"status"`
	At     time.Time `json:"at"`
}

func (h *EsHandler) HandleLivenessCheck(w http.ResponseWriter, _ *http.Request) {
	response := livenessCheckResponse{
		Status: "ok",
		At:     time.Now(),
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	err := encoder.Encode(response)
	if err != nil {
		log.Printf("Failed to send response: %s", err)
	}
}
