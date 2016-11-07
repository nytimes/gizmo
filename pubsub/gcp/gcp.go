package gcp

import (
	"errors"
	"sync"
	"time"

	"google.golang.org/api/option"

	gpubsub "cloud.google.com/go/pubsub"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"

	"github.com/NYTimes/gizmo/pubsub"
)

// subscriber is a Google Cloud Platform PubSub client that allows a user to
// consume messages via the pubsub.Subscriber interface.
type subscriber struct {
	sub  subscription
	ctx  context.Context
	name string

	mtxStop sync.Mutex
	stop    chan chan error
	stopped bool
	err     error
}

// NewSubscriber will instantiate a new Subscriber that wraps
// a pubsub.Iterator.
func NewSubscriber(ctx context.Context, projID, subscription string, opts ...option.ClientOption) (pubsub.Subscriber, error) {
	client, err := gpubsub.NewClient(ctx, projID, opts...)
	if err != nil {
		return nil, err
	}
	return &subscriber{
		sub:  subscriptionImpl{sub: client.Subscription(subscription)},
		ctx:  ctx,
		name: subscription,
		stop: make(chan chan error, 1),
	}, nil
}

var (
	defaultMaxMessages  = gpubsub.MaxPrefetch(10)
	defaultMaxExtension = gpubsub.MaxExtension(60 * time.Second)
)

// Start will start pulling from pubsub via a pubsub.Iterator.
func (s *subscriber) Start() <-chan pubsub.SubscriberMessage {
	output := make(chan pubsub.SubscriberMessage)
	go func(s *subscriber, output chan pubsub.SubscriberMessage) {
		defer close(output)
		var (
			iter iterator
			msg  message
			err  error
		)

		iter, err = s.sub.Pull(s.ctx, defaultMaxMessages, defaultMaxExtension)
		if err != nil {
			go s.Stop()
			s.err = err
		}

		for {
			select {
			case exit := <-s.stop:
				if iter != nil {
					iter.Stop()
				}
				exit <- nil
				return
			default:
				// something's wrong and we're on the way to stopping
				if iter == nil {
					continue
				}

				msg, err = iter.Next()
				if err != nil {
					s.err = err
					go s.Stop()
					continue
				}

				output <- &subMessage{
					msg: msg,
					sub: s.name,
					ctx: s.ctx,
				}
			}
		}
	}(s, output)
	return output

}

// Err will contain any error the Subscriber has encountered while processing.
func (s *subscriber) Err() error {
	return s.err
}

// Stop will block until the consumer has stopped consuming messages.
func (s *subscriber) Stop() error {
	s.mtxStop.Lock()
	defer s.mtxStop.Unlock()
	if s.stopped {
		return nil
	}
	s.stopped = true
	exit := make(chan error)
	s.stop <- exit
	return <-exit
}

// subMessage pubsub implementation of pubsub.SubscriberMessage.
type subMessage struct {
	msg message
	ctx context.Context
	sub string
}

// Message will return the data of the pubsub Message.
func (m *subMessage) Message() []byte {
	return m.msg.MsgData()
}

// ExtendDoneDeadline will call the deprecated ModifyAckDeadline for a pubsub
// Message. This likely should not be called.
func (m *subMessage) ExtendDoneDeadline(dur time.Duration) error {
	return errors.New("not suppported")
}

// Done will acknowledge the pubsub Message.
func (m *subMessage) Done() error {
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
func NewPublisher(ctx context.Context, projID, topic string, opts ...option.ClientOption) (pubsub.MultiPublisher, error) {
	client, err := gpubsub.NewClient(ctx, projID, opts...)
	if err != nil {
		return nil, err
	}

	return &publisher{
		topic: client.Topic(topic),
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
	_, err := p.topic.Publish(ctx, &gpubsub.Message{
		Data:       m,
		Attributes: map[string]string{"key": key},
	})
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

	a := make([]*gpubsub.Message, len(messages))
	for i := range messages {
		a[i] = &gpubsub.Message{
			Data:       messages[i],
			Attributes: map[string]string{"key": keys[i]},
		}
	}

	_, err := p.topic.Publish(ctx, a...)
	return err
}

// interfaces and types to make this more testable
type (
	subscription interface {
		Pull(ctx context.Context, opts ...gpubsub.PullOption) (iterator, error)
	}
	iterator interface {
		Next() (message, error)
		Stop()
	}
	message interface {
		ID() string
		MsgData() []byte
		Done()
	}

	subscriptionImpl struct {
		sub *gpubsub.Subscription
	}
	iteratorImpl struct {
		iter *gpubsub.MessageIterator
	}
	messageImpl struct {
		msg *gpubsub.Message
	}
)

func (m messageImpl) ID() string {
	return m.msg.ID
}

func (m messageImpl) MsgData() []byte {
	return m.msg.Data
}

func (m messageImpl) Done() {
	m.msg.Done(true)
}

func (i iteratorImpl) Next() (message, error) {
	msg, err := i.iter.Next()
	return messageImpl{msg}, err
}

func (i iteratorImpl) Stop() {
	i.iter.Stop()
}

func (s subscriptionImpl) Pull(ctx context.Context, opts ...gpubsub.PullOption) (iterator, error) {
	iter, err := s.sub.Pull(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return iteratorImpl{iter: iter}, nil
}
