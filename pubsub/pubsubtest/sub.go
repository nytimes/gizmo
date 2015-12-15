package pubsubtest

import (
	"encoding/json"

	"github.com/golang/protobuf/proto"
	"github.com/nytimes/gizmo/pubsub"
)

type (
	// TestPublisher is a simple implementation of pubsub.Publisher meant to
	// help mock out any implementations.
	TestSubscriber struct {
		// ProtoMessages will be marshalled into []byte  used to mock out
		// a feed if it is populated.
		ProtoMessages []proto.Message

		// JSONMessages will be marshalled into []byte and used to mock out
		// a feed if it is populated.
		JSONMessages []interface{}

		// GivenErrError will be returned by the TestSubscriber on Err().
		// Good for testing error scenarios.
		GivenErrError error

		// GivenStopError will be returned by the TestSubscriber on Stop().
		// Good for testing error scenarios.
		GivenStopError error

		// FoundError will contain any errors encountered while marshalling
		// the JSON and protobuf struct.
		FoundError error
	}

	TestSubsMessage struct {
		Msg   []byte
		Doned bool
	}
)

func (m *TestSubsMessage) Message() []byte {
	return m.Msg
}

func (m *TestSubsMessage) Done() error {
	m.Doned = true
	return nil
}

// Start will populate and return the test channel for the subscriber
func (t *TestSubscriber) Start() <-chan pubsub.SubscriberMessage {
	msgs := make(chan pubsub.SubscriberMessage, len(t.JSONMessages)+len(t.ProtoMessages))
	for _, pmsg := range t.ProtoMessages {
		msg, err := proto.Marshal(pmsg)
		if err != nil {
			t.FoundError = err
			continue
		}
		msgs <- &TestSubsMessage{Msg: msg}
	}

	for _, jmsg := range t.JSONMessages {
		msg, err := json.Marshal(jmsg)
		if err != nil {
			t.FoundError = err
			continue
		}
		msgs <- &TestSubsMessage{Msg: msg}
	}
	close(msgs)

	return msgs
}

func (t *TestSubscriber) Err() error {
	return t.GivenErrError
}

func (t *TestSubscriber) Stop() error {
	return t.GivenStopError
}
