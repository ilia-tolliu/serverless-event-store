package estypes

type NewEsEvent struct {
	EventType string `json:"eventType" validate:"required"`
	Payload   any    `json:"payload,omitempty" validate:"required"`
}
