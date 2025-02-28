// Package eshttp contains client to the HTTP API of serverless-event-store
//
// The client wraps the following Event Store operations:
//   - create event stream with initial event
//   - append event to stream
//   - get stream details
//   - list streams
//   - get stream events
//
// To get started you need a base URL of the Event Store:
//
//	esHttpClient := eshttp.NewClient("https://****.lambda-url.****.on.aws/")
package eshttp
