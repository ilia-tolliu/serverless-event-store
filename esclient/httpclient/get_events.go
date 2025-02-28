package httpclient

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/ilia-tolliu-go-event-store/estypes"
	"github.com/ilia-tolliu-go-event-store/internal/eserror"
	"iter"
	"net/http"
	"net/url"
	"strconv"
)

type getEventsResponse struct {
	EventPage estypes.EventPage `json:"eventPage"`
}

func (c *EsHttpClient) GetEvents(streamType string, streamId uuid.UUID, afterRevision int) iter.Seq2[*estypes.Event, error] {
	currentAfterRevision := afterRevision

	eventIter := func(yield func(*estypes.Event, error) bool) {
		for {
			eventPage, err := c.requestEventPage(streamType, streamId, currentAfterRevision)
			if err != nil {
				yield(nil, err)
				return
			}

			for _, event := range eventPage.Events {
				if !yield(&event, nil) {
					return
				}
			}

			if !eventPage.HasMore {
				return
			}

			currentAfterRevision = eventPage.LastEvaluatedRevision
		}
	}

	return eventIter
}

func (c *EsHttpClient) formatGetEventsUrl(streamType string, streamId uuid.UUID, afterRevision int) string {
	esUrl := c.baseUrl.JoinPath("streams", streamType, streamId.String(), "events")

	queryValues := url.Values{
		"after-revision": []string{strconv.Itoa(afterRevision)},
	}
	query := queryValues.Encode()
	esUrl.RawQuery = query

	return esUrl.String()
}

func (c *EsHttpClient) requestEventPage(streamType string, streamId uuid.UUID, afterRevision int) (*estypes.EventPage, error) {
	esUrl := c.formatGetEventsUrl(streamType, streamId, afterRevision)

	resp, err := http.Get(esUrl)
	if err != nil {
		return nil, fmt.Errorf("failed GET events from Event Store: %w", err)
	}

	defer eserror.Ignore(resp.Body.Close)

	if resp.StatusCode != http.StatusOK {
		return nil, ErrorFromHttpResponse(resp, "failed to request events")
	}

	var respBody getEventsResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response as event page: %w", err)
	}

	return &respBody.EventPage, nil
}
