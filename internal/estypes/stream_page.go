package estypes

type StreamPage struct {
	Streams     []Stream `json:"streams"`
	HasMore     bool     `json:"hasMore"`
	NextPageKey *string  `json:"nextPageKey,omitempty"`
}
