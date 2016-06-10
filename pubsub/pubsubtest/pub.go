package pubsubtest

import (
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
)

type (
	// TestPublisher is a simple implementation of pubsub.Publisher meant to
	// help mock out any implementations.
	TestPublisher struct {
		// Published will contain a list of all messages that have been published.
		Published []TestPublishMsg

		// GivenError will be returned by the TestPublisher on publish.
		// Good for testing error scenarios.
		GivenError error

		// FoundError will contain any errors encountered while marshalling
		// the protobuf struct.
		FoundError error
	}
	// TestPublishMsg is a test publish message.
	TestPublishMsg struct {
		// Key represents the message key.
		Key string
		// Body represents the message body.
		Body []byte
	}
)

// Publish publishes the message, delegating to PublishRaw.
func (t *TestPublisher) Publish(ctx context.Context, key string, msg proto.Message) error {
	data, err := proto.Marshal(msg)
	t.FoundError = err
	return t.PublishRaw(ctx, key, data)
}

// PublishRaw publishes the raw message byte slice.
func (t *TestPublisher) PublishRaw(_ context.Context, key string, msg []byte) error {
	t.Published = append(t.Published, TestPublishMsg{key, msg})
	return t.GivenError
}
