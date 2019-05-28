package kafka // import "github.com/NYTimes/gizmo/pubsub/kafka"

import (
	"errors"
	"log"
	"time"

	"github.com/NYTimes/gizmo/pubsub"

	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
)

var (
	// RequiredAcks will be used in Kafka configs
	// to set the 'RequiredAcks' value.
	RequiredAcks = sarama.WaitForAll
)

// Publisher is an experimental publisher that provides an implementation for
// Kafka using the Shopify/sarama library.
type Publisher struct {
	producer sarama.SyncProducer
	topic    string
}

// NewPublisher will initiate a new experimental Kafka publisher.
func NewPublisher(cfg *Config) (pubsub.Publisher, error) {
	var err error
	p := &Publisher{}

	if len(cfg.Topic) == 0 {
		return p, errors.New("topic name is required")
	}
	p.topic = cfg.Topic

	sconfig := cfg.Config
	if sconfig == nil {
		sconfig = sarama.NewConfig()
		sconfig.Producer.Retry.Max = cfg.MaxRetry
		sconfig.Producer.RequiredAcks = RequiredAcks
	}
	// we always want successes to return
	sconfig.Producer.Return.Successes = true
	p.producer, err = sarama.NewSyncProducer(cfg.BrokerHosts, sconfig)
	return p, err
}

// Publish will marshal the proto message and emit it to the Kafka topic.
func (p *Publisher) Publish(ctx context.Context, key string, m proto.Message) error {
	mb, err := proto.Marshal(m)
	if err != nil {
		return err
	}
	return p.PublishRaw(ctx, key, mb)
}

// PublishRaw will emit the byte array to the Kafka topic.
func (p *Publisher) PublishRaw(_ context.Context, key string, m []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(m),
	}
	// TODO: do something with this partition/offset values
	_, _, err := p.producer.SendMessage(msg)
	return err
}

// Stop will close the pub connection.
func (p *Publisher) Stop() error {
	return p.producer.Close()
}

type (
	// subscriber is an experimental subscriber implementation for Kafka. It is only capable of consuming a
	// single partition so multiple may be required depending on your setup.
	subscriber struct {
		cnsmr     sarama.Consumer
		topic     string
		partition int32

		offset          func() int64
		broadcastOffset func(int64)

		kerr error

		stop chan chan error
	}

	// subMessage is an SubscriberMessage implementation
	// that will broadcast the message's offset when Done().
	subMessage struct {
		message         *sarama.ConsumerMessage
		broadcastOffset func(int64)
	}
)

// Message will return the message payload.
func (m *subMessage) Message() []byte {
	return m.message.Value
}

// ExtendDoneDeadline has no effect on subMessage.
func (m *subMessage) ExtendDoneDeadline(time.Duration) error {
	return nil
}

// Done will emit the message's offset.
func (m *subMessage) Done() error {
	m.broadcastOffset(m.message.Offset)
	return nil
}

// NewSubscriber will initiate a the experimental Kafka consumer.
func NewSubscriber(cfg *Config, offsetProvider func() int64, offsetBroadcast func(int64)) (pubsub.Subscriber, error) {
	var (
		err error
	)
	s := &subscriber{
		offset:          offsetProvider,
		broadcastOffset: offsetBroadcast,
		partition:       cfg.Partition,
		stop:            make(chan chan error, 1),
	}

	if len(cfg.BrokerHosts) == 0 {
		return s, errors.New("at least 1 broker host is required")
	}

	if len(cfg.Topic) == 0 {
		return s, errors.New("topic name is required")
	}
	s.topic = cfg.Topic

	sconfig := cfg.Config
	if sconfig == nil {
		sconfig = sarama.NewConfig()
	}
	// we always want to see errors, no matter what
	sconfig.Consumer.Return.Errors = true
	s.cnsmr, err = sarama.NewConsumer(cfg.BrokerHosts, sconfig)
	return s, err
}

// Start will start consuming message on the Kafka topic
// partition and emit any messages to the returned channel.
// On start up, it will call the offset func provider to the subscriber
// to lookup the offset to start at.
// If it encounters any issues, it will populate the Err() error
// and close the returned channel.
func (s *subscriber) Start() <-chan pubsub.SubscriberMessage {
	output := make(chan pubsub.SubscriberMessage)

	pCnsmr, err := s.cnsmr.ConsumePartition(s.topic, s.partition, s.offset())
	if err != nil {
		// TODO: what should we do here?
		log.Print("unable to create partition consumer: ", err)
		close(output)
		return output
	}

	go func(s *subscriber, c sarama.PartitionConsumer, output chan pubsub.SubscriberMessage) {
		defer close(output)
		var msg *sarama.ConsumerMessage
		errs := c.Errors()
		msgs := c.Messages()
		for {
			select {
			case exit := <-s.stop:
				exit <- c.Close()
				return
			case kerr := <-errs:
				s.kerr = kerr
				return
			case msg = <-msgs:
				output <- &subMessage{
					message:         msg,
					broadcastOffset: s.broadcastOffset,
				}
			}
		}
	}(s, pCnsmr, output)

	return output
}

// Stop willablock until the consumer has stopped consuming messages
// and return any errors seen on consumer close.
func (s *subscriber) Stop() error {
	exit := make(chan error)
	s.stop <- exit
	// close result from the partition consumer
	err := <-exit
	if err != nil {
		return err
	}
	return s.cnsmr.Close()
}

// Err will contain any  errors that occurred during
// consumption. This method should be checked after
// a user encounters a closed channel.
func (s *subscriber) Err() error {
	return s.kerr
}

// GetPartitions is a helper function to look up which partitions are available
// via the given brokers for the given topic. This should be called only on startup.
func GetPartitions(brokerHosts []string, topic string) (partitions []int32, err error) {
	if len(brokerHosts) == 0 {
		return partitions, errors.New("at least 1 broker host is required")
	}

	if len(topic) == 0 {
		return partitions, errors.New("topic name is required")
	}

	var cnsmr sarama.Consumer
	cnsmr, err = sarama.NewConsumer(brokerHosts, sarama.NewConfig())
	if err != nil {
		return partitions, err
	}

	defer func() {
		if cerr := cnsmr.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	return cnsmr.Partitions(topic)
}

type (
	// consumerGroupSubscriber uses the Sarama consumer groups implementation.
	// The Kafka consumer group feature enables applications to parallelize
	// topic consumption by dynamically assigning topic partitions to one or
	// more consumers that belong to the same group. This implementation uses
	// Kafka to store offsets.
	consumerGroupSubscriber struct {
		cnsmr sarama.ConsumerGroup
		commitBroadcastHandler
		kerr  error
		stop  chan bool
		topic string
	}

	// commitBroadcastHandler is invoked immediately after offset commit.
	commitBroadcastHandler func(*sarama.ConsumerMessage)

	// consumerGroupMessage is an SubscriberMessage implementation
	// that commits offsets to Kafka and then invokes the
	// commitBroadcastHandler.
	consumerGroupMessage struct {
		commitBroadcastHandler
		message *sarama.ConsumerMessage
		session sarama.ConsumerGroupSession
	}

	// consumerGroupHandler handles messages received in a consumer group
	// session.
	consumerGroupHandler struct {
		commitBroadcastHandler
		output chan pubsub.SubscriberMessage
		stop   chan bool
	}
)

// NewConsumerGroupSubscriber initializes a Kafka consumer that uses consumer groups.
func NewConsumerGroupSubscriber(cfg *Config, groupID string, cbh commitBroadcastHandler) (pubsub.Subscriber, error) {
	var err error
	s := &consumerGroupSubscriber{
		commitBroadcastHandler: cbh,
		stop:                   make(chan bool, 1),
	}

	if len(cfg.BrokerHosts) == 0 {
		return s, errors.New("at least 1 broker host is required")
	}

	if len(cfg.Topic) == 0 {
		return s, errors.New("topic name is required")
	}
	s.topic = cfg.Topic

	sconfig := cfg.Config
	if sconfig == nil {
		sconfig = sarama.NewConfig()
	}
	// we always want to see errors, no matter what
	sconfig.Consumer.Return.Errors = true
	s.cnsmr, err = sarama.NewConsumerGroup(cfg.BrokerHosts, groupID, sconfig)
	return s, err
}

// Message returns the message payload.
func (cgm *consumerGroupMessage) Message() []byte {
	return cgm.message.Value
}

// ExtendDoneDeadline takes no action.
func (cgm *consumerGroupMessage) ExtendDoneDeadline(time.Duration) error {
	return nil
}

// Done commits message offset then invokes commit broadcast handler.
func (cgm *consumerGroupMessage) Done() error {
	cgm.session.MarkMessage(cgm.message, "")
	if cgm.commitBroadcastHandler != nil {
		cgm.commitBroadcastHandler(cgm.message)
	}
	return nil
}

// Start joins a consumer group session with a handler that streams messages to
// the subscriber message channel. Consumer group errors can be retrieved
// through the Err() function.
func (cgs *consumerGroupSubscriber) Start() <-chan pubsub.SubscriberMessage {
	output := make(chan pubsub.SubscriberMessage)

	cgh := &consumerGroupHandler{
		commitBroadcastHandler: cgs.commitBroadcastHandler,
		output:                 output,
		stop:                   cgs.stop,
	}

	ctx := context.Background()
	go func() {
		for {
			topic := []string{cgs.topic}
			if err := cgs.cnsmr.Consume(ctx, topic, cgh); err == sarama.ErrClosedConsumerGroup {
				return
			} else if err != nil {
				cgs.kerr = err
			}
		}
	}()

	return output
}

// Stop halts consumer group session and blocks until consumer group handler
// completes.
func (cgs *consumerGroupSubscriber) Stop() error {
	close(cgs.stop)
	return cgs.cnsmr.Close()
}

// Err contains any errors raised by a consumer group session.
func (cgs *consumerGroupSubscriber) Err() error {
	return cgs.kerr
}

// Setup takes no action (required for ConsumerGroupHandler interface).
func (cgh *consumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup takes no action (required for ConsumerGroupHandler interface).
func (cgh *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim feeds messages from consumer group session to subscriber
// channel.
func (cgh *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

Loop:
	for {
		select {
		case msg := <-claim.Messages():
			if msg == nil {
				// channel is closed
				break Loop
			}
			cgh.output <- &consumerGroupMessage{
				commitBroadcastHandler: cgh.commitBroadcastHandler,
				message:                msg,
				session:                session,
			}
		case <-cgh.stop:
			break Loop
		}
	}

	return nil
}
