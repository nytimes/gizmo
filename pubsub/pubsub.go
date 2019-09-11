package pubsub

import (
	"time"

	"golang.org/x/net/context"

	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)

// Log is the structured logger used throughout the package.
var Log = logrus.New()

// Publisher is a generic interface to encapsulate how we want our publishers
// to behave. Until we find reason to change, we're forcing all pubslishers
// to emit protobufs.
type Publisher interface {
	// Publish will publish a message with context.
	Publish(context.Context, string, proto.Message) error
	// Publish will publish a raw byte array as a message with context.
	PublishRaw(context.Context, string, []byte) error
}

// MultiPublisher is an interface for publishers who support sending multiple
// messages in a single request, in addition to individual messages.
type MultiPublisher interface {
	Publisher

	// PublishMulti will publish multiple messages with a context.
	PublishMulti(context.Context, []string, []proto.Message) error
	// PublishMultiRaw will publish multiple raw byte array messages with a context.
	PublishMultiRaw(context.Context, []string, [][]byte) error
}

// Subscriber is a generic interface to encapsulate how we want our subscribers
// to behave. For now the system will auto stop if it encounters any errors. If
// a user encounters a closed channel, they should check the Err() method to see
// what happened.
type Subscriber interface {
	// Start will return a channel of raw messages.
	Start() <-chan SubscriberMessage
	// Err will contain any errors returned from the consumer connection.
	Err() error
	// Stop will initiate a graceful shutdown of the subscriber connection.
	Stop() error
}

// SubscriberMessage is a struct to encapsulate subscriber messages and provide
// a mechanism for acknowledging messages _after_ they've been processed.
type SubscriberMessage interface {
	Message() []byte
	ExtendDoneDeadline(time.Duration) error
	Done() error
}
