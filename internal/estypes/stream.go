package estypes

import (
	"github.com/google/uuid"
	"time"
)

type Stream struct {
	StreamId   uuid.UUID `json:"streamId" dynamodbav:"PK,string"`
	StreamType string    `json:"streamType" dynamodbav:"streamType"`
	Revision   int       `json:"revision" dynamodbav:"revision"`
	CreatedAt  time.Time `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt" dynamodbav:"updatedAt"`
}

func NewStream(streamId uuid.UUID, streamType string, now time.Time) Stream {
	return Stream{
		StreamId:   streamId,
		StreamType: streamType,
		Revision:   1,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
