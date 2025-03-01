package esnotification

import (
	"encoding/json"
	"fmt"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

func NewFromSqsMessage(sqsMessage sqstypes.Message) (EsNotification, error) {
	var snsMessage snsNotification
	err := json.Unmarshal([]byte(*sqsMessage.Body), &snsMessage)
	if err != nil {
		return EsNotification{}, fmt.Errorf("failed to unmarshal snsNotification: %w", err)
	}

	var esNotification EsNotification
	err = json.Unmarshal([]byte(snsMessage.Message), &esNotification)
	if err != nil {
		return EsNotification{}, fmt.Errorf("failed to unmarshal EsNotification: %v", err)
	}

	esNotification.sqsReceiptHandle = *sqsMessage.ReceiptHandle

	return esNotification, nil
}

type snsNotification struct {
	Message string `json:"Message,omitempty"`
}
