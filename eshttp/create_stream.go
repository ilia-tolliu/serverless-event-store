package eshttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ilia-tolliu/serverless-event-store/estypes"
	"github.com/ilia-tolliu/serverless-event-store/internal/eserror"
	"net/http"
)

type createStreamResponse struct {
	Stream estypes.Stream `json:"stream"`
}

// CreateStream persists new event-sourced stream together with initial event.
//
// Stream type means a kind of workflow that the stream will follow.
// Streams of the same type support the same types of events and are subject to the same processing.
//
// When subscribing to an Event Store updates, stream type can be used as a criteria in SNS subscription filter.
func (c *Client) CreateStream(streamType string, initialEvent estypes.NewEsEvent) (*estypes.Stream, error) {
	esUrl := c.formatCreateStreamUrl(streamType)

	body, err := json.Marshal(map[string]any{
		"initialEvent": initialEvent,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal initial event: %v", err)
	}

	resp, err := http.Post(esUrl, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed POST to Event Store: %w", err)
	}

	defer eserror.Ignore(resp.Body.Close)

	if resp.StatusCode != http.StatusCreated {
		return nil, ErrorFromHttpResponse(resp, "failed to create stream")
	}

	var respBody createStreamResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response as stream: %w", err)
	}

	return &respBody.Stream, nil
}

func (c *Client) formatCreateStreamUrl(streamType string) string {
	return c.baseUrl.JoinPath("streams", streamType).String()
}
