package estypes

// NewEsEvent is an event to be appended to a stream.
//
// Event type will tell an event handler how to parse and handle the payload.
// Payload can be any string, but usually supposed to be a serialized JSON.
type NewEsEvent struct {
	EventType string `json:"eventType" validate:"required"`
	Payload   string `json:"payload,omitempty" validate:"required"`
}
