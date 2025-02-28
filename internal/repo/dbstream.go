package repo

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/ilia-tolliu-go-event-store/estypes"
	"time"
)

const RecordTypeStream = "stream"
const streamIndexName = "StreamIndex"

type DbStream struct {
	Pk             string    `dynamodbav:"PK"`
	Sk             int       `dynamodbav:"SK"`
	RecordType     string    `dynamodbav:"RecordType"`
	StreamType     string    `dynamodbav:"StreamType"`
	StreamRevision int       `dynamodbav:"StreamRevision"`
	UpdatedAt      time.Time `dynamodbav:"UpdatedAt"`
}

type dbStreamKey struct {
	Pk string `dynamodbav:"PK"`
	Sk int    `dynamodbav:"SK"`
}

func FromStream(stream estypes.Stream) DbStream {
	updatedAtUtc := stream.UpdatedAt.UTC()

	return DbStream{
		Pk:             stream.StreamId.String(),
		Sk:             0,
		RecordType:     RecordTypeStream,
		StreamType:     stream.StreamType,
		StreamRevision: stream.Revision,
		UpdatedAt:      updatedAtUtc,
	}
}

func IntoStream(dbStream DbStream) (estypes.Stream, error) {
	streamId, err := uuid.Parse(dbStream.Pk)
	if err != nil {
		return estypes.Stream{}, fmt.Errorf("failed to parse streamId: %w", err)
	}

	stream := estypes.Stream{
		StreamId:   streamId,
		StreamType: dbStream.StreamType,
		Revision:   dbStream.StreamRevision,
		UpdatedAt:  dbStream.UpdatedAt,
	}

	return stream, nil
}
