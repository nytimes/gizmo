package pubsub

import (
	"time"

	"github.com/NYTimes/gizmo/config"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/cloud/pubsub"
)

type GCPSubscriber struct {
	sub  *pubsub.Subscription
	ctx  context.Context
	name string

	stop chan chan error
	err  error
}

func NewGCPSubscriber(ctx context.Context, cfg config.PubSub) (Subscriber, error) {
	client, err := pubsub.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		return nil, err
	}
	return &GCPSubscriber{
		sub:  client.Subscription(cfg.Subscription),
		ctx:  ctx,
		name: cfg.Subscription,
	}, nil
}

var (
	defaultGCPMaxMessages  = pubsub.MaxPrefetch(10)
	defaultGCPMaxExtension = pubsub.MaxExtension(60 * time.Second)
)

func (s *GCPSubscriber) Start() <-chan SubscriberMessage {
	output := make(chan SubscriberMessage)
	go func(s *GCPSubscriber, output chan SubscriberMessage) {
		defer close(output)
		var (
			iter *pubsub.Iterator
			msg  *pubsub.Message
			err  error
		)

		iter, err = s.sub.Pull(s.ctx, defaultGCPMaxMessages, defaultGCPMaxExtension)
		if err != nil {
			s.err = err
			return
		}

		for {
			select {
			case exit := <-s.stop:
				iter.Stop()
				exit <- nil
				return
			default:
				msg, err = iter.Next()
				if err != nil {
					s.err = err
					go s.Stop()
					continue
				}

				output <- &GCPSubMessage{
					msg: msg,
					sub: s.name,
					ctx: s.ctx,
				}
			}
		}
	}(s, output)
	return output

}

func (s *GCPSubscriber) Err() error {
	return s.err
}

// Stop will block until the consumer has stopped consuming
// messages.
func (s *GCPSubscriber) Stop() error {
	exit := make(chan error)
	s.stop <- exit
	return <-exit
}

type GCPSubMessage struct {
	msg *pubsub.Message
	ctx context.Context
	sub string
}

func (m *GCPSubMessage) Message() []byte {
	return m.msg.Data
}

func (m *GCPSubMessage) ExtendDoneDeadline(dur time.Duration) error {
	return pubsub.ModifyAckDeadline(m.ctx, m.sub, m.msg.ID, dur)
}

func (m *GCPSubMessage) Done() error {
	m.msg.Done(true)
	return nil
}

type GCPPublisher struct {
	topic string
	ctx   context.Context
}

func NewGCPPublisher(ctx context.Context, topic string) (Publisher, error) {
	return &GCPPublisher{
		topic: string,
		ctx:   ctx,
	}, nil
}

func (p *GCPPublisher) Publish(ctx context.Context, key string, msg proto.Message) error {
	mb, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return p.PublishRaw(ctx, key, mb)
}

func (p *GCPPublisher) PublishRaw(ctx context.Context, key string, m []byte) error {
	_, err := pubsub.Publish(ctx, p.topic, &pubsub.Message{
		Data:       m,
		Attributes: map[string]string{"key": key},
	})
	return err
}
