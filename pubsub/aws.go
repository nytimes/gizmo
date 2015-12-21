package pubsub

import (
	"encoding/base64"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/golang/protobuf/proto"

	"github.com/NYTimes/gizmo/config"
)

// SNSPublisher will accept AWS credentials and an SNS topic name
// and it will emit any publish events to it.
type SNSPublisher struct {
	sns   snsiface.SNSAPI
	topic string
}

// NewSNSPublisher will initiate the SNS client.
// If no credentials are passed in with the config,
// the publisher is instantiated with the AWS_ACCESS_KEY
// and the AWS_SECRET_KEY environment variables.
func NewSNSPublisher(cfg *config.SNS) (*SNSPublisher, error) {
	p := &SNSPublisher{}

	if cfg.Topic == "" {
		return p, errors.New("SNS topic name is required")
	}
	p.topic = cfg.Topic

	if cfg.Region == "" {
		return p, errors.New("SNS region is required")
	}

	var creds *credentials.Credentials
	if cfg.AccessKey != "" {
		creds = credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, "")
	} else {
		creds = credentials.NewEnvCredentials()
	}

	p.sns = sns.New(session.New(&aws.Config{
		Credentials: creds,
		Region:      &cfg.Region,
	}))
	return p, nil
}

// Publish will marshal the proto message and emit it to the SNS topic.
// The key will be used as the SNS message subject.
func (p *SNSPublisher) Publish(key string, m proto.Message) error {
	mb, err := proto.Marshal(m)
	if err != nil {
		return err
	}

	return p.PublishRaw(key, mb)
}

// PublishRaw will emit the byte array to the SNS topic.
// The key will be used as the SNS message subject.
func (p *SNSPublisher) PublishRaw(key string, m []byte) error {
	msg := &sns.PublishInput{
		TopicArn: &p.topic,
		Subject:  &key,
		Message:  aws.String(base64.StdEncoding.EncodeToString(m)),
	}

	_, err := p.sns.Publish(msg)
	return err
}

var (
	// defaultSQSMaxMessages is default the number of bulk messages
	// the SQSSubscriber will attempt to fetch on each
	// receive.
	defaultSQSMaxMessages int64 = 10
	// defaultSQSTimeoutSeconds is the default number of seconds the
	// SQS client will wait before timing out.
	defaultSQSTimeoutSeconds int64 = 2
	// SQSSleepInterval is the default time.Duration the
	// SQSSubscriber will wait if it sees no messages
	// on the queue.
	defaultSQSSleepInterval = 2 * time.Second

	// SQSDeleteBufferSize is the default limit of messages
	// allowed in the delete buffer before
	// executing a 'delete batch' request.
	defaultSQSDeleteBufferSize = 0
)

func defaultSQSConfig(cfg *config.SQS) {
	if cfg.MaxMessages == nil {
		cfg.MaxMessages = &defaultSQSMaxMessages
	}

	if cfg.TimeoutSeconds == nil {
		cfg.TimeoutSeconds = &defaultSQSTimeoutSeconds
	}

	if cfg.SleepInterval == nil {
		cfg.SleepInterval = &defaultSQSSleepInterval
	}

	if cfg.DeleteBufferSize == nil {
		cfg.DeleteBufferSize = &defaultSQSDeleteBufferSize
	}
}

type (
	// SQSSubscriber is an SQS client that allows a user to
	// consume messages via the pubsub.Subscriber interface.
	SQSSubscriber struct {
		sqs sqsiface.SQSAPI

		cfg      *config.SQS
		queueURL *string

		toDelete   chan *deleteRequest
		deleteDone chan bool

		stop   chan chan error
		sqsErr error
	}

	// SQSMessage is the SQS implementation of `SubscriberMessage`.
	SQSMessage struct {
		sub     *SQSSubscriber
		message *sqs.Message
	}

	deleteRequest struct {
		entry   *sqs.DeleteMessageBatchRequestEntry
		receipt chan error
	}
)

// NewSQSSubscriber will initiate a new Decrypter for the subscriber
// if a key file is provided. It will also fetch the SQS Queue Url
// and set up the SQS client.
func NewSQSSubscriber(cfg *config.SQS) (*SQSSubscriber, error) {
	var err error
	defaultSQSConfig(cfg)
	s := &SQSSubscriber{
		cfg:        cfg,
		toDelete:   make(chan *deleteRequest),
		deleteDone: make(chan bool),
		stop:       make(chan chan error, 1),
	}

	if len(cfg.QueueName) == 0 {
		return s, errors.New("sqs queue name is required")
	}

	var creds *credentials.Credentials
	if cfg.AccessKey != "" {
		creds = credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, "")
	} else {
		creds = credentials.NewEnvCredentials()
	}
	s.sqs = sqs.New(session.New(&aws.Config{
		Credentials: creds,
		Region:      &cfg.Region,
	}))

	var urlResp *sqs.GetQueueUrlOutput
	urlResp, err = s.sqs.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &cfg.QueueName,
	})

	if err != nil {
		return s, err
	}

	s.queueURL = urlResp.QueueUrl
	return s, nil
}

// Message will decode protobufed message bodies and simply return
// a byte slice containing the message body for all others types.
func (m *SQSMessage) Message() []byte {
	if m.sub.cfg.ConsumeProtobuf {
		msgBody, err := base64.StdEncoding.DecodeString(*m.message.Body)
		if err != nil {
			Log.Warnf("unable to parse message body: %s", err)
		}
		return msgBody
	}
	return []byte(*m.message.Body)
}

// Done will queue up a message to be deleted. By default,
// the `SQSDeleteBufferSize` will be 0, so this will block until the
// message has been deleted.
func (m *SQSMessage) Done() error {
	receipt := make(chan error)
	m.sub.toDelete <- &deleteRequest{
		entry: &sqs.DeleteMessageBatchRequestEntry{
			Id:            m.message.MessageId,
			ReceiptHandle: m.message.ReceiptHandle,
		},
		receipt: receipt,
	}
	return <-receipt
}

// Start will start consuming messages on the SQS queue
// and emit any messages to the returned channel.
// If it encounters any issues, it will populate the Err() error
// and close the returned channel.
func (s *SQSSubscriber) Start() <-chan SubscriberMessage {
	output := make(chan SubscriberMessage)
	go s.handleDeletes()
	go func(s *SQSSubscriber, output chan SubscriberMessage) {
		defer close(output)
		var (
			resp *sqs.ReceiveMessageOutput
			err  error
		)
		for {
			select {
			case exit := <-s.stop:
				close(s.toDelete)
				<-s.deleteDone
				exit <- nil
				return
			default:
				// get messages
				Log.Infof("receiving messages")
				resp, err = s.sqs.ReceiveMessage(&sqs.ReceiveMessageInput{
					MaxNumberOfMessages: s.cfg.MaxMessages,
					QueueUrl:            s.queueURL,
					WaitTimeSeconds:     s.cfg.TimeoutSeconds,
				})
				if err != nil {
					// we've encountered a major error
					// this will set the error value and close the channel
					// so the user will stop iterating and check the err
					s.sqsErr = err
					return
				}

				// if we didn't get any messages, lets chill out for a sec
				if len(resp.Messages) == 0 {
					Log.Infof("no messages found. sleeping for %s", s.cfg.SleepInterval)
					time.Sleep(*s.cfg.SleepInterval)
					continue
				}

				Log.Infof("found %d messages", len(resp.Messages))

				// for each message, pass to output
				for _, msg := range resp.Messages {
					output <- &SQSMessage{
						sub:     s,
						message: msg,
					}
				}
			}
		}
	}(s, output)
	return output
}

func (s *SQSSubscriber) handleDeletes() {
	batchInput := &sqs.DeleteMessageBatchInput{
		QueueUrl: s.queueURL,
	}
	var (
		err           error
		entriesBuffer []*sqs.DeleteMessageBatchRequestEntry
	)
	for deleteRequest := range s.toDelete {
		entriesBuffer = append(entriesBuffer, deleteRequest.entry)

		// if buffer is full, send the request
		if len(entriesBuffer) > *s.cfg.DeleteBufferSize {
			batchInput.Entries = entriesBuffer
			_, err = s.sqs.DeleteMessageBatch(batchInput)
			// cleaer buffer
			entriesBuffer = []*sqs.DeleteMessageBatchRequestEntry{}
		}

		deleteRequest.receipt <- err
	}
	// clear any remainders before shutdown
	if len(entriesBuffer) > 0 {
		batchInput.Entries = entriesBuffer
		_, s.sqsErr = s.sqs.DeleteMessageBatch(batchInput)
	}
	s.deleteDone <- true
}

// Stop will block until the consumer has stopped consuming
// messages.
func (s *SQSSubscriber) Stop() error {
	exit := make(chan error)
	s.stop <- exit
	return <-exit
}

// Err will contain any errors that occurred during
// consumption. This method should be checked after
// a user encounters a closed channel.
func (s *SQSSubscriber) Err() error {
	return s.sqsErr
}
