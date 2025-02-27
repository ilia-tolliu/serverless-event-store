package estypes

type NewEsEvent struct {
	EventType string `json:"eventType" validate:"required"`
	Payload   string `json:"payload,omitempty" validate:"required"`
}
