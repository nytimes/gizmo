package pubsubtest

import "github.com/golang/protobuf/proto"

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
	TestPublishMsg struct {
		Key  string
		Body []byte
	}
)

func (t *TestPublisher) Publish(key string, msg proto.Message) error {
	data, err := proto.Marshal(msg)
	t.FoundError = err
	return t.PublishRaw(key, data)
}

func (t *TestPublisher) PublishRaw(key string, msg []byte) error {
	t.Published = append(t.Published, TestPublishMsg{key, msg})
	return t.GivenError
}
