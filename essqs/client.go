package essqs

import (
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Client struct {
	sqsClient             *sqs.Client
	queueUrl              string
	receiveTimeoutSeconds int
}

func NewClient(sqsClient *sqs.Client, queueUrl string) *Client {
	// todo: add simple validation

	return &Client{
		sqsClient:             sqsClient,
		queueUrl:              queueUrl,
		receiveTimeoutSeconds: 1,
	}
}
