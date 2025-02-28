package essqs

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/ilia-tolliu/serverless-event-store/estypes/esnotification"
)

// ReceiveNotifications retrieves pending Event Store notifications from SQS queue
//
// The Event Store payload is extracted from SQS message and SNS message.
// Once received notification is processed, you should acknowledge it to delete from SQS.
func (c *Client) ReceiveNotifications(ctx context.Context) ([]esnotification.EsNotification, error) {
	messagesOut, err := c.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:        &c.queueUrl,
		WaitTimeSeconds: int32(c.receiveTimeoutSeconds),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages from SQS: %w", err)
	}

	notifications := make([]esnotification.EsNotification, 0, len(messagesOut.Messages))

	for _, msg := range messagesOut.Messages {
		notification, err := esnotification.NewFromSqsMessage(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to convert SQS message into EsNotification:\n%#v\n %w", msg, err)
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}
