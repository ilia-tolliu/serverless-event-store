package essqs

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/ilia-tolliu/serverless-event-store/estypes/esnotification"
)

// AcknowledgeNotification deletes an Event Store notification from SQS after successful processing.
func (c *Client) AcknowledgeNotification(ctx context.Context, notification esnotification.EsNotification) error {
	receiptHandle := notification.GetSqsReceiptHandle()

	_, err := c.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &c.queueUrl,
		ReceiptHandle: &receiptHandle,
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge notification in SQS: %w", err)
	}

	return nil
}
