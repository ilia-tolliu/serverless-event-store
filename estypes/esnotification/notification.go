package esnotification

import (
	"github.com/google/uuid"
)

type EsNotification struct {
	StreamId         uuid.UUID `json:"StreamId"`
	StreamType       string    `json:"StreamType"`
	StreamRevision   int       `json:"StreamRevision,string"`
	sqsReceiptHandle string
}

func (n *EsNotification) SetSqsReceiptHandle(receiptHandle string) {
	n.sqsReceiptHandle = receiptHandle
}

func (n *EsNotification) GetSqsReceiptHandle() string {
	return n.sqsReceiptHandle
}
