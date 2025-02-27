package estypes

import (
	"fmt"
	"github.com/google/uuid"
	"time"
)

type Stream struct {
	StreamId   uuid.UUID `json:"streamId" dynamodbav:"PK,string"`
	StreamType string    `json:"streamType" dynamodbav:"streamType"`
	Revision   int       `json:"revision" dynamodbav:"revision"`
	UpdatedAt  time.Time `json:"updatedAt" dynamodbav:"updatedAt"`
}

func NewStream(streamId uuid.UUID, streamType string, now time.Time) Stream {
	return Stream{
		StreamId:   streamId,
		StreamType: streamType,
		Revision:   1,
		UpdatedAt:  now,
	}
}

func (s *Stream) ShouldHaveType(streamType string) error {
	if s.StreamType != streamType {
		return fmt.Errorf("stream type does not match; streamId: [%s], streamType: [%s], wanted streamType: [%s]", s.StreamId, s.StreamType, streamType)
	}

	return nil
}

func (s *Stream) ShouldHaveRevision(revision int) error {
	if s.Revision != revision {
		return fmt.Errorf("stream revision does not match; streamId: [%s], revision: [%d], wanted revision: [%d]", s.StreamId, s.Revision, revision)
	}

	return nil
}
