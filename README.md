# go-event-store - Serverless Event Store

## Technologies

* Go language
* AWS Lambda to run
* AWS DynamoDB to store data
* AWS SNS to notify about updates
* AWS CloudWatch to collect logs
* AWS CDK for infrastructure-as-a-code

## Setting up development environment

### Prerequisites

* Go language toolchain
* Node.js of the latest LTS version for using CDK
* AWS account
* AWS CLI configured to your account
* AWS CDK
* Just command runner

### Getting started

* Go to `./_infrastructure/aws-event-store` and create `.env` file based on `.env.example`.
* Bootstrap your AWS environment: `$ cdk bootstrap`
* Build & deploy: `$ just deploy`

The output of deploy step will contain resource references:

```
Outputs:
AwsEventStoreStack.EsDynamoDbTable = AwsEventStoreStack-EsTable********
AwsEventStoreStack.EsLogGroup = AwsEventStoreStack-EsLogs********
AwsEventStoreStack.EsSnsTopic = arn:aws:sns:eu-central-1:********:AwsEventStoreStack-EsTopic********
AwsEventStoreStack.EsUrl = https://********.lambda-url.********.on.aws/
```

You can also check this references later: `$ just describe`

* Done! Event store is deployed and available for use in development.
* Navigate to the Event Store url for interactive API specification.

**NB! For production use Lambda Function URL should be updated to use AWS IAM authentication**

## What is an Event Store

An Event Store is the storage mechanism at the heart of an event-sourced system.

Event sourcing was defined by Martin Fowler and popularised by Greg Young, Adam Dymitruk and others.

Event sourced systems work well for such use cases as automating human processes in regulated industries 
(banking, state services, healthcare, gambling).

Event Store keeps data in append-only event streams. 
Each event stream represents an entity. 
Current state of an entity may be derived from events by processing them in order.

Every event represents some decision made about an entity, which led to update in its state.

An event also represents an atomic and consistent update to an entity.

## Guarantees of an Event Store

* Event streams are append only.
* Events are appended with sequential revision numbers without gaps.
* Conflicting events (with already existing revision number) are rejected.

## Using Event Store in your system

Event store has two apis:

* HTTP API for creating streams, appending events and reading events
* SNS channel that will notify about all streams' updates

Use HTTP API to create streams and append events in command handlers.

Use notifications to trigger updates in your read models and reactors. 
A notification message looks like this:

```json
  {
    "StreamId": "2886e475-d5cc-40f7-aec5-401071388b3c",
    "StreamType": "test-stream-type",
    "StreamRevision": "18"
  }
```

You may subscribe to these notifications with e.g. SQS queue 
and apply SNS subscription filtering by StreamType.

Once your component gets a notification, use 
`GET /streams/{streamType}/{streamId}/event` endpoint to read the stream events.
