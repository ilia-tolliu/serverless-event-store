package eshttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/ilia-tolliu/serverless-event-store/estypes"
	"github.com/ilia-tolliu/serverless-event-store/internal/eserror"
	"net/http"
	"strconv"
)

type appendEventResponse struct {
	Stream estypes.Stream `json:"stream"`
}

// AppendEvent persists another event at the tail of the stream.
//
// To successfully append an event to a stream, you should not at which revision N the stream currently is.
// Then you can append an event exactly with revision N + 1.
// Attempt to append an event with inconsistent revision will cause an error.
//
// This is one of the guarantees of an Event Store.
func (c *Client) AppendEvent(streamType string, streamId uuid.UUID, revision int, event estypes.NewEsEvent) (*estypes.Stream, error) {
	url := c.formatAppendEventUrl(streamType, streamId, revision)

	body, err := json.Marshal(map[string]any{
		"event": event,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %v", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create PUT request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed PUT to Event Store: %w", err)
	}

	defer eserror.Ignore(resp.Body.Close)

	if resp.StatusCode != http.StatusCreated {
		return nil, ErrorFromHttpResponse(resp, "failed to append event")
	}

	var respBody appendEventResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response as stream: %w", err)
	}

	return &respBody.Stream, nil
}

func (c *Client) formatAppendEventUrl(streamType string, streamId uuid.UUID, revision int) string {
	url := c.baseUrl.JoinPath("streams", streamType, streamId.String(), "events", strconv.Itoa(revision))

	return url.String()
}
