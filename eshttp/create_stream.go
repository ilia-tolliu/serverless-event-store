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

func (c *EsHttpClient) CreateStream(streamType string, initialEvent estypes.NewEsEvent) (*estypes.Stream, error) {
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

func (c *EsHttpClient) formatCreateStreamUrl(streamType string) string {
	return c.baseUrl.JoinPath("streams", streamType).String()
}
