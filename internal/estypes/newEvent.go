package estypes

type NewEsEvent struct {
	EventType string `json:"eventType"`
	Payload   any    `json:"payload"`
}
