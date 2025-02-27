package esnotification

import (
	"github.com/google/uuid"
)

type EsNotification struct {
	StreamId       uuid.UUID `json:"StreamId"`
	StreamType     string    `json:"StreamType"`
	StreamRevision int       `json:"StreamRevision,string"`
}
