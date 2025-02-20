package estypes

type NewEvent struct {
	EventType string `json:"eventType"`
	Payload   any    `json:"payload"`
}
