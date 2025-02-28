package eshttp

import (
	"encoding/json"
	"fmt"
	"github.com/ilia-tolliu/serverless-event-store/estypes"
	"github.com/ilia-tolliu/serverless-event-store/internal/eserror"
	"iter"
	"net/http"
	"net/url"
	"time"
)

type getStreamsResponse struct {
	StreamPage estypes.StreamPage `json:"streamPage"`
}

func (c *Client) GetStreams(streamType string, updatedAfter time.Time) iter.Seq2[*estypes.Stream, error] {
	var nextPageKey *string

	streamIter := func(yield func(*estypes.Stream, error) bool) {
		for {
			streamPage, err := c.requestStreamPage(streamType, updatedAfter, nextPageKey)
			if err != nil {
				yield(nil, err)
				return
			}

			for _, stream := range streamPage.Streams {
				if !yield(&stream, nil) {
					return
				}
			}

			if !streamPage.HasMore {
				return
			}

			nextPageKey = streamPage.NextPageKey
		}
	}

	return streamIter
}

func (c *Client) formatGetStreamsUrl(streamType string, updatedAfter time.Time, nextPageKey *string) string {
	esUrl := c.baseUrl.JoinPath("streams", streamType)

	updatedAfterUtc := updatedAfter.UTC()

	queryValues := url.Values{
		"updated-after": []string{updatedAfterUtc.Format(time.RFC3339Nano)},
	}
	if nextPageKey != nil {
		queryValues.Add("nextPageKey", *nextPageKey)
	}
	query := queryValues.Encode()
	esUrl.RawQuery = query

	return esUrl.String()
}

func (c *Client) requestStreamPage(streamType string, updatedAfter time.Time, nextPageKey *string) (*estypes.StreamPage, error) {
	esUrl := c.formatGetStreamsUrl(streamType, updatedAfter, nextPageKey)
	println("esUrl: %s", esUrl)

	resp, err := http.Get(esUrl)
	if err != nil {
		return nil, fmt.Errorf("failed GET streams from Event Store: %w", err)
	}

	defer eserror.Ignore(resp.Body.Close)

	if resp.StatusCode != http.StatusOK {
		return nil, ErrorFromHttpResponse(resp, "failed to request streams")
	}

	var respBody getStreamsResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response as stream page: %w", err)
	}

	return &respBody.StreamPage, nil
}
