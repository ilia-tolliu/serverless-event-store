package estypes

type NewEsEvent struct {
	EventType string         `json:"eventType" validate:"required"`
	Payload   map[string]any `json:"payload,omitempty" validate:"required"`
}
