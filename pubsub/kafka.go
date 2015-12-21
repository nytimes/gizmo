package pubsub

import (
	"errors"
	"log"

	"github.com/NYTimes/gizmo/config"

	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
)

var (
	// KafkaRequiredAcks will be used in Kafka configs
	// to set the 'RequiredAcks' value.
	KafkaRequiredAcks = sarama.WaitForAll
)

// KafkaPublisher is an experimental publisher that provides an implementation for
// Kafka using the Shopify/sarama library.
type KafkaPublisher struct {
	producer sarama.SyncProducer
	topic    string
}

// NewKafkaPublisher will initiate a new experimental Kafka publisher.
func NewKafkaPublisher(cfg *config.Kafka) (*KafkaPublisher, error) {
	var err error
	p := &KafkaPublisher{}

	if len(cfg.Topic) == 0 {
		return p, errors.New("topic name is required")
	}
	p.topic = cfg.Topic

	sconfig := sarama.NewConfig()
	sconfig.Producer.Retry.Max = cfg.MaxRetry
	sconfig.Producer.RequiredAcks = KafkaRequiredAcks
	p.producer, err = sarama.NewSyncProducer(cfg.BrokerHosts, sconfig)
	return p, err
}

// Publish will marshal the proto message and emit it to the Kafka topic.
func (p *KafkaPublisher) Publish(key string, m proto.Message) error {
	mb, err := proto.Marshal(m)
	if err != nil {
		return err
	}
	return p.PublishRaw(key, mb)
}

// PublishRaw will emit the byte array to the Kafka topic.
func (p *KafkaPublisher) PublishRaw(key string, m []byte) error {
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
func (p *KafkaPublisher) Stop() error {
	return p.producer.Close()
}

type (
	// KafkaSubscriber is an experimental subscriber implementation for Kafka. It is only capable of consuming a
	// single partition so multiple may be required depending on your setup.
	KafkaSubscriber struct {
		cnsmr     sarama.Consumer
		topic     string
		partition int32

		offset          func() int64
		broadcastOffset func(int64)

		kerr error

		stop chan chan error
	}

	// KafkaSubMessage is an SubscriberMessage implementation
	// that will broadcast the message's offset when Done().
	KafkaSubMessage struct {
		message         *sarama.ConsumerMessage
		broadcastOffset func(int64)
	}
)

// Message will return the message payload.
func (m *KafkaSubMessage) Message() []byte {
	return m.message.Value
}

// Done will emit the message's offset.
func (m *KafkaSubMessage) Done() error {
	m.broadcastOffset(m.message.Offset)
	return nil
}

// NewKafkaSubscriber will initiate a the experimental Kafka consumer.
func NewKafkaSubscriber(cfg *config.Kafka, offsetProvider func() int64, offsetBroadcast func(int64)) (*KafkaSubscriber, error) {
	var (
		err error
	)
	s := &KafkaSubscriber{
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

	sconfig := sarama.NewConfig()
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
func (s *KafkaSubscriber) Start() <-chan SubscriberMessage {
	output := make(chan SubscriberMessage)

	pCnsmr, err := s.cnsmr.ConsumePartition(s.topic, s.partition, s.offset())
	if err != nil {
		// TODO: what should we do here?
		log.Print("unable to create partition consumer: ", err)
		close(output)
		return output
	}

	go func(s *KafkaSubscriber, c sarama.PartitionConsumer, output chan SubscriberMessage) {
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
				output <- &KafkaSubMessage{
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
func (s *KafkaSubscriber) Stop() error {
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
func (s *KafkaSubscriber) Err() error {
	return s.kerr
}

// GetKafkaPartitions is a helper function to look up which partitions are available
// via the given brokers for the given topic. This should be called only on startup.
func GetKafkaPartitions(brokerHosts []string, topic string) (partitions []int32, err error) {
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
