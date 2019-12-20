package gcp // import "github.com/NYTimes/gizmo/pubsub/gcp"

import (
	"errors"
	"sync"
	"time"

	gpubsub "cloud.google.com/go/pubsub"
	"github.com/NYTimes/gizmo/pubsub"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
)

// Subscriber is a Google Cloud Platform PubSub client that allows a user to
// consume messages via the pubsub.Subscriber interface.
type Subscriber struct {
	sub subscription
	ctx context.Context

	mtxStop sync.Mutex
	stopped bool
	cancel  func()

	err error
}

// NewSubscriber will instantiate a new Subscriber that wraps a pubsub.Iterator.
func NewSubscriber(ctx context.Context, projID, subscription string, opts ...option.ClientOption) (*Subscriber, error) {
	client, err := gpubsub.NewClient(ctx, projID, opts...)
	if err != nil {
		return &Subscriber{}, err
	}

	sub := client.Subscription(subscription)
	sub.ReceiveSettings = gpubsub.ReceiveSettings{
		MaxExtension:           defaultMaxExtension,
		MaxOutstandingMessages: defaultMaxMessages,
	}
	return &Subscriber{
		ctx: ctx,
		sub: subscriptionImpl{Sub: sub},
	}, nil
}

var (
	defaultMaxMessages  = 10
	defaultMaxExtension = 60 * time.Second
)

// Start will start pulling from pubsub via a pubsub.Iterator.
func (s *Subscriber) Start() <-chan pubsub.SubscriberMessage {
	output := make(chan pubsub.SubscriberMessage)
	go func(s *Subscriber, output chan pubsub.SubscriberMessage) {
		defer close(output)

		s.ctx, s.cancel = context.WithCancel(s.ctx)
		err := s.sub.Receive(s.ctx, func(ctx context.Context, msg message) {
			sm := &SubMessage{
				msg: msg,
			}
			if mi, ok := msg.(messageImpl); ok {
				sm.Attributes = mi.Msg.Attributes
			}
			output <- sm
		})
		if err != nil {
			s.Stop()
			s.err = err
		}
	}(s, output)
	return output
}

// Err will contain any error the Subscriber has encountered while processing.
func (s *Subscriber) Err() error {
	return s.err
}

// Stop will block until the consumer has stopped consuming messages.
func (s *Subscriber) Stop() error {
	s.mtxStop.Lock()
	defer s.mtxStop.Unlock()
	if s.stopped {
		return nil
	}
	s.stopped = true
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}

// SetReceiveSettings sets the ReceivedSettings on the google pubsub Subscription.
// Should be called before Start().
func (s *Subscriber) SetReceiveSettings(settings gpubsub.ReceiveSettings) {
	s.sub.(subscriptionImpl).Sub.ReceiveSettings = settings
}

// SubMessage pubsub implementation of pubsub.SubscriberMessage.
type SubMessage struct {
	msg        message
	Attributes map[string]string
}

// Message will return the data of the pubsub Message.
func (m *SubMessage) Message() []byte {
	return m.msg.MsgData()
}

// ExtendDoneDeadline will call the deprecated ModifyAckDeadline for a pubsub
// Message. This likely should not be called.
func (m *SubMessage) ExtendDoneDeadline(dur time.Duration) error {
	return errors.New("not suppported")
}

// Done will acknowledge the pubsub Message.
func (m *SubMessage) Done() error {
	m.msg.Done()
	return nil
}

// publisher is a Google Cloud Platform PubSub client that allows a user to
// consume messages via the pubsub.MultiPublisher interface.
type publisher struct {
	topic *gpubsub.Topic
}

var _ pubsub.Publisher = &publisher{}
var _ pubsub.MultiPublisher = &publisher{}

// NewPublisher will instantiate a new GCP MultiPublisher.
func NewPublisher(ctx context.Context, cfg Config, opts ...option.ClientOption) (pubsub.MultiPublisher, error) {
	if cfg.ProjectID == "" {
		return nil, errors.New("project id is required")
	}
	if cfg.Topic == "" {
		return nil, errors.New("topic name is required")
	}

	c, err := gpubsub.NewClient(ctx, cfg.ProjectID, opts...)
	if err != nil {
		return nil, err
	}
	t := c.Topic(cfg.Topic)
	// Update PublishSettings from cfg.PublishSettings
	// but never set thresholds to 0.
	if cfg.PublishSettings.DelayThreshold > 0 {
		t.PublishSettings.DelayThreshold = cfg.PublishSettings.DelayThreshold
	}
	if cfg.PublishSettings.CountThreshold > 0 {
		t.PublishSettings.CountThreshold = cfg.PublishSettings.CountThreshold
	}
	if cfg.PublishSettings.ByteThreshold > 0 {
		t.PublishSettings.ByteThreshold = cfg.PublishSettings.ByteThreshold
	}
	if cfg.PublishSettings.NumGoroutines > 0 {
		t.PublishSettings.NumGoroutines = cfg.PublishSettings.NumGoroutines
	}
	if cfg.PublishSettings.Timeout > 0 {
		t.PublishSettings.Timeout = cfg.PublishSettings.Timeout
	}
	return &publisher{
		topic: t,
	}, nil
}

// Publish will marshal the proto message and publish it to GCP pubsub.
func (p *publisher) Publish(ctx context.Context, key string, msg proto.Message) error {
	mb, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return p.PublishRaw(ctx, key, mb)
}

// PublishRaw will publish the message to GCP pubsub.
func (p *publisher) PublishRaw(ctx context.Context, key string, m []byte) error {
	res := p.topic.Publish(ctx, &gpubsub.Message{
		Data:       m,
		Attributes: map[string]string{"key": key},
	})
	_, err := res.Get(ctx)
	return err
}

// PublishMulti will publish multiple messages to GCP pubsub in a single request.
func (p *publisher) PublishMulti(ctx context.Context, keys []string, messages []proto.Message) error {
	if len(keys) != len(messages) {
		return errors.New("keys and messages must be equal length")
	}

	a := make([][]byte, len(messages))
	for i := range messages {
		b, err := proto.Marshal(messages[i])
		if err != nil {
			return err
		}
		a[i] = b
	}
	return p.PublishMultiRaw(ctx, keys, a)
}

// PublishMultiRaw will publish multiple raw byte array messages to GCP pubsub in a single request.
func (p *publisher) PublishMultiRaw(ctx context.Context, keys []string, messages [][]byte) error {
	if len(keys) != len(messages) {
		return errors.New("keys and messages must be equal length")
	}

	for i := range messages {
		err := p.PublishRaw(ctx, keys[i], messages[i])
		if err != nil {
			return err
		}
	}

	return nil
}

// interfaces and types to make this more testable
type (
	subscription interface {
		Receive(ctx context.Context, f func(context.Context, message)) error
	}
	message interface {
		ID() string
		MsgData() []byte
		Done()
	}

	messageImpl struct {
		Msg *gpubsub.Message
	}

	subscriptionImpl struct {
		Sub *gpubsub.Subscription
	}
)

func (m messageImpl) ID() string {
	return m.Msg.ID
}

func (m messageImpl) MsgData() []byte {
	return m.Msg.Data
}

func (m messageImpl) Done() {
	m.Msg.Ack()
}

func (s subscriptionImpl) Receive(ctx context.Context, f func(context.Context, message)) error {
	return s.Sub.Receive(ctx, func(ctx context.Context, msg *gpubsub.Message) {
		f(ctx, messageImpl{msg})
	})
}
