package estypes

import (
	"github.com/google/uuid"
	"time"
)

type Event struct {
	StreamId  uuid.UUID `json:"streamId" dynamodbav:"PK,string"`
	Revision  int       `json:"revision" dynamodbav:"SK"`
	EventType string    `json:"eventType" dynamodbav:"eventType"`
	Payload   any       `json:"payload" dynamodbav:"eventType"`
	CreatedAt time.Time `json:"createdAt" dynamodbav:"createdAt"`
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
