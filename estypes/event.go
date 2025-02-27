package estypes

import (
	"github.com/google/uuid"
	"time"
)

type Event struct {
	StreamId  uuid.UUID `json:"streamId"`
	Revision  int       `json:"revision"`
	EventType string    `json:"eventType"`
	Payload   string    `json:"payload"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewEvent(streamId uuid.UUID, revision int, newEvent NewEsEvent, now time.Time) Event {
	return Event{
		StreamId:  streamId,
		Revision:  revision,
		EventType: newEvent.EventType,
		Payload:   newEvent.Payload,
		CreatedAt: now,
	}
}
