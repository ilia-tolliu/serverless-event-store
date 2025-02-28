package httpclient

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/ilia-tolliu-go-event-store/estypes"
	"github.com/ilia-tolliu-go-event-store/internal/eserror"
	"net/http"
)

type getStreamDetailsResponse struct {
	Stream estypes.Stream `json:"stream"`
}

func (c *EsHttpClient) GetStreamDetails(streamType string, streamId uuid.UUID) (*estypes.Stream, error) {
	esUrl := c.formatGetStreamDetailsUrl(streamType, streamId)

	resp, err := http.Get(esUrl)
	if err != nil {
		return nil, fmt.Errorf("failed GET stream details from Event Store: %w", err)
	}

	defer eserror.Ignore(resp.Body.Close)

	if resp.StatusCode != http.StatusOK {
		return nil, ErrorFromHttpResponse(resp, "failed to get stream details")
	}

	var respBody getStreamDetailsResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response as stream: %w", err)
	}

	return &respBody.Stream, nil
}

func (c *EsHttpClient) formatGetStreamDetailsUrl(streamType string, streamId uuid.UUID) string {
	return c.baseUrl.JoinPath("streams", streamType, streamId.String(), "details").String()
}
