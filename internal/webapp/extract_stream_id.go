package webapp

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net/http"
)

func ExtractStreamId(r *http.Request) (uuid.UUID, error) {
	streamIdStr := r.PathValue("streamId")
	if streamIdStr == "" {
		return uuid.Nil, errors.New("streamId is empty")
	}

	streamId, err := uuid.Parse(streamIdStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid streamId: %w", err)
	}

	return streamId, nil
}
