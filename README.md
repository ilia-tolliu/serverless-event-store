# go-event-store - Serverless Event Store

## Technologies

* Go language
* AWS Lambda to run
* AWS DynamoDB to store data
* AWS SNS to notify about updates
* AWS CloudWatch to collect logs
* AWS CDK for infrastructure-as-a-code

Usage of 

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

