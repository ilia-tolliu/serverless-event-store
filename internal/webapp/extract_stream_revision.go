package webapp

import (
	"fmt"
	"net/http"
	"strconv"
)

func ExtractStreamRevision(r *http.Request) (int, error) {
	streamRevisionStr := r.PathValue("streamRevision")
	streamRevision, err := strconv.Atoi(streamRevisionStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse stream revision: %w", err)
	}

	return streamRevision, nil
}
