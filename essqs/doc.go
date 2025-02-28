// Package essqs contains AWS SQS client to receive notifications from Event Store
//
// The client is useful when you subscribe to the Event Store notifications with SQS
// and process it e.g. in an ECS or EC2 deployed service.
//
// For AWS Lambda this client is not needed, since you'll get SQS message directly as
// a Lambda input.
//
// To get started you need an AWS SQS client and queue URL:
//
//	import (
//	  "context"
//	  "github.com/aws/aws-sdk-go-v2/config"
//	  "github.com/aws/aws-sdk-go-v2/service/sqs"
//	  "github.com/ilia-tolliu/serverless-event-store/essqs"
//	)
//
//	func bootstrap() (*essqs.Client, error) {
//	  awsConfig, err := config.LoadDefaultConfig(context.Background())
//	  // handle err
//
//	  sqsClient := sqs.NewFromConfig(awsConfig)
//	  esSqsClient := essqs.NewClient(sqsClient, "https://sqs.eu-central-1.amazonaws.com/****/queue-name")
//
//	  return esSqsClient
//	}
package essqs
