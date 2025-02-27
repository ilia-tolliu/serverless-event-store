package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	"github.com/ilia-tolliu-go-event-store/esclient/httpclient"
	"github.com/ilia-tolliu-go-event-store/esclient/sqsclient"
	"github.com/ilia-tolliu-go-event-store/estypes"
	"github.com/ilia-tolliu-go-event-store/estypes/esnotification"
	"github.com/ilia-tolliu-go-event-store/internal/config"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestEventStore(t *testing.T) {
	bootstrap(t)

	streamId := testCreateStream(t, "test-stream", estypes.NewEsEvent{
		EventType: "stream-created",
		Payload:   initialPayload,
	})

	testReceiveNotification(t, esnotification.EsNotification{
		StreamId:       streamId,
		StreamType:     "test-stream",
		StreamRevision: 1,
	})

	testAppendEvent(t, "test-stream", streamId, 2, estypes.NewEsEvent{
		EventType: "something-important-happened",
		Payload:   secondPayload,
	})

	testReceiveNotification(t, esnotification.EsNotification{
		StreamId:       streamId,
		StreamType:     "test-stream",
		StreamRevision: 2,
	})
}

type testSqsQueue struct {
	url *string
	arn string
}

type payload1 struct {
	Message string
}

type payload2 struct {
	Message string
	Amount  int
	Moment  time.Time
}

var (
	esHttpClient *httpclient.EsHttpClient
	esSqsClient  *sqsclient.EsSqsClient

	now = time.Now()

	initialPayload = payload1{
		Message: "I am payload1",
	}

	secondPayload = payload2{
		Message: "I am payload2",
		Amount:  123,
		Moment:  now,
	}
)

func bootstrap(t *testing.T) {
	awsConfig, err := awsconfig.LoadDefaultConfig(t.Context())
	if err != nil {
		t.Fatalf("failed to load AWS SDK config, %v", err)
	}

	testConfig := loadTestConfig(t, awsConfig)
	sqsClient := sqs.NewFromConfig(awsConfig)
	snsClient := sns.NewFromConfig(awsConfig)

	queue := makeTestQueue(t, sqsClient)

	subscribeQueueToSns(t, snsClient, testConfig, queue)

	esHttpClient = httpclient.New(testConfig.EsUrl)
	esSqsClient = sqsclient.New(sqsClient, *queue.url)
}

func loadTestConfig(t *testing.T, awsConfig aws.Config) *config.EsTestConfig {
	mode := config.NewFromEnv()

	esTestConfig, err := config.EsTestConfigFromAws(t.Context(), mode, awsConfig)
	if err != nil {
		t.Fatalf("failed to load config from AWS, %v", err)
	}
	t.Logf("using config: %+v", esTestConfig)

	return esTestConfig
}

func makeTestQueue(t *testing.T, sqsClient *sqs.Client) testSqsQueue {
	queueName := fmt.Sprintf("event-sourcing-test-queueOutput-%s", uuid.New().String())

	queueOutput, err := sqsClient.CreateQueue(t.Context(), &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		t.Fatalf("failed to create queueOutput, %v", err)
	}

	t.Cleanup(func() {
		_, err := sqsClient.DeleteQueue(context.Background(), &sqs.DeleteQueueInput{
			QueueUrl: queueOutput.QueueUrl,
		})
		if err != nil {
			t.Fatalf("failed to delete queue, %v", err)
		}
	})

	queueArn := getQueueArn(t, sqsClient, queueOutput.QueueUrl)

	queue := testSqsQueue{
		url: queueOutput.QueueUrl,
		arn: queueArn,
	}

	allowSnsToPublishToQueue(t, sqsClient, queue)

	return queue
}

func getQueueArn(t *testing.T, sqsClient *sqs.Client, queueUrl *string) string {
	queueAttrOutput, err := sqsClient.GetQueueAttributes(t.Context(), &sqs.GetQueueAttributesInput{
		QueueUrl: queueUrl,
		AttributeNames: []sqstypes.QueueAttributeName{
			"QueueArn",
		},
	})
	if err != nil {
		t.Fatalf("failed to get queueOutput attributes, %v", err)
	}
	queueArn := queueAttrOutput.Attributes["QueueArn"]
	t.Logf("SQS queueOutput ARN: %s", queueArn)

	return queueArn
}

func allowSnsToPublishToQueue(t *testing.T, sqsClient *sqs.Client, queue testSqsQueue) {
	policy := formatQueuePolicyJson(t, queue.arn)

	_, err := sqsClient.SetQueueAttributes(t.Context(), &sqs.SetQueueAttributesInput{
		QueueUrl: queue.url,
		Attributes: map[string]string{
			"Policy": policy,
		},
	})
	if err != nil {
		t.Fatalf("failed to allow SNS to publish messages to SQS, %v", err)
	}
}

func formatQueuePolicyJson(t *testing.T, queueArn string) string {
	policyDoc := policyDocument{
		Version: "2012-10-17",
		Statement: []policyStatement{{
			Effect:    "Allow",
			Action:    "sqs:SendMessage",
			Principal: map[string]string{"Service": "sns.amazonaws.com"},
			Resource:  aws.String(queueArn),
		}},
	}

	policyJson, err := json.Marshal(policyDoc)
	if err != nil {
		t.Fatalf("failed to format policy document: %v", err)
	}

	return string(policyJson)
}

func subscribeQueueToSns(t *testing.T, snsClient *sns.Client, testConfig *config.EsTestConfig, queue testSqsQueue) {
	subOutput, err := snsClient.Subscribe(t.Context(), &sns.SubscribeInput{
		Protocol: aws.String("sqs"),
		TopicArn: aws.String(testConfig.EsSnsTopic),
		Endpoint: aws.String(queue.arn),
	})
	if err != nil {
		t.Fatalf("failed to subscribe to stream notifications, %v", err)
	}
	t.Logf("subscribed to stream notifications: %s", *subOutput.SubscriptionArn)
}

type policyDocument struct {
	Version   string
	Statement []policyStatement
}

type policyStatement struct {
	Effect    string
	Action    string
	Principal map[string]string `json:",omitempty"`
	Resource  *string           `json:",omitempty"`
}

func testCreateStream(t *testing.T, streamType string, initialEvent estypes.NewEsEvent) uuid.UUID {
	stream, err := esHttpClient.CreateStream(streamType, initialEvent)
	if err != nil {
		t.Fatalf("failed to create stream: %v", err)
	}

	require.NoError(t, stream.ShouldHaveRevision(1))
	require.NoError(t, stream.ShouldHaveType(streamType))

	return stream.StreamId
}

func testReceiveNotification(t *testing.T, expectedNotification esnotification.EsNotification) {
	notifications, err := esSqsClient.ReceiveNotifications(t.Context())
	if err != nil {
		t.Fatalf("failed to receive notifications: %v", err)
	}

	require.Len(t, notifications, 1)
	require.Equal(t, notifications[0], expectedNotification)
}

func testAppendEvent(t *testing.T, streamType string, streamId uuid.UUID, revision int, newEvent estypes.NewEsEvent) {
	stream, err := esHttpClient.AppendEvent(streamType, streamId, revision, newEvent)
	if err != nil {
		t.Fatalf("failed to append event: %#v", err)
	}
	t.Logf("stream after AppendEvent: %#v", stream)

	require.NoError(t, stream.ShouldHaveRevision(revision))
	require.NoError(t, stream.ShouldHaveType(streamType))
}
