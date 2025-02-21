package webapp

import (
	"fmt"
	"net/http"
)

func ExtractStreamType(r *http.Request) (string, error) {
	streamType := r.PathValue("streamType")
	if streamType == "" {
		return "", fmt.Errorf("no streamType specified")
	}

	return streamType, nil
}
