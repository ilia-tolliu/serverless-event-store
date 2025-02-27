package sqsclient

import (
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type EsSqsClient struct {
	sqsClient             *sqs.Client
	queueUrl              string
	receiveTimeoutSeconds int
}

func New(sqsClient *sqs.Client, queueUrl string) *EsSqsClient {
	// todo: add simple validation

	return &EsSqsClient{
		sqsClient:             sqsClient,
		queueUrl:              queueUrl,
		receiveTimeoutSeconds: 1,
	}
}
