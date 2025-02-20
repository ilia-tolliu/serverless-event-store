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
