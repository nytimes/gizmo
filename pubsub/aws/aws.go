package aws // import "github.com/NYTimes/gizmo/pubsub/aws"

import (
	"encoding/base64"
	"errors"
	"sync/atomic"
	"time"

	"github.com/NYTimes/gizmo/pubsub"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
)

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type key int

// msgAttrsKey is the context key for the SNS Message Attributes.  Its value of zero is
// arbitrary. If this package defined other context keys, they would have
// different integer values.
const msgAttrsKey key = 0

// publisher will accept AWS credentials and an SNS topic name
// and it will emit any publish events to it.
type publisher struct {
	sns   snsiface.SNSAPI
	topic string
}

// NewPublisher will initiate the SNS client.
// If no credentials are passed in with the config,
// the publisher is instantiated with the AWS_ACCESS_KEY
// and the AWS_SECRET_KEY environment variables.
func NewPublisher(cfg SNSConfig) (pubsub.Publisher, error) {
	p := &publisher{}

	if cfg.Topic == "" {
		return p, errors.New("SNS topic name is required")
	}
	p.topic = cfg.Topic

	if cfg.Region == "" {
		return p, errors.New("SNS region is required")
	}

	sess, err := session.NewSession()
	if err != nil {
		return p, err
	}

	var creds *credentials.Credentials
	if cfg.AccessKey != "" {
		creds = credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, cfg.SessionToken)
	} else if cfg.RoleARN != "" {
		var err error
		creds, err = requestRoleCredentials(sess, cfg.RoleARN, cfg.MFASerialNumber)
		if err != nil {
			return p, err
		}
	}

	p.sns = sns.New(sess, &aws.Config{
		Credentials: creds,
		Region:      &cfg.Region,
		Endpoint:    cfg.EndpointURL,
	})
	return p, nil
}

// Publish will marshal the proto message and emit it to the SNS topic.
// The key will be used as the SNS message subject.
func (p *publisher) Publish(ctx context.Context, key string, m proto.Message) error {
	mb, err := proto.Marshal(m)
	if err != nil {
		return err
	}

	return p.PublishRaw(ctx, key, mb)
}

// PublishRaw will emit the byte array to the SNS topic.
// The key will be used as the SNS message subject.
// You can use func WithMessageAttributes to set SNS message attributes for the message
func (p *publisher) PublishRaw(ctx context.Context, key string, m []byte) error {
	msg := &sns.PublishInput{
		TopicArn: &p.topic,
		Subject:  &key,
		Message:  aws.String(base64.StdEncoding.EncodeToString(m)),
	}

	if v, ok := ctx.Value(msgAttrsKey).(map[string]*sns.MessageAttributeValue); ok {
		msg.MessageAttributes = v
	}

	_, err := p.sns.Publish(msg)
	return err
}

// WithMessageAttributes used to add SNS Message Attributes to the context
// for further usage in publishing messages to sns with provided attributes
func WithMessageAttributes(ctx context.Context, msgAttrs map[string]*sns.MessageAttributeValue) context.Context {
	return context.WithValue(ctx, msgAttrsKey, msgAttrs)
}

var (
	// defaultSQSMaxMessages is default the number of bulk messages
	// the subscriber will attempt to fetch on each
	// receive.
	defaultSQSMaxMessages int64 = 10
	// defaultSQSTimeoutSeconds is the default number of seconds the
	// SQS client will wait before timing out.
	defaultSQSTimeoutSeconds int64 = 2
	// defaultSQSSleepInterval is the default time.Duration the
	// subscriber will wait if it sees no messages
	// on the queue.
	defaultSQSSleepInterval = 2 * time.Second

	// defaultSQSDeleteBufferSize is the default limit of messages
	// allowed in the delete buffer before
	// executing a 'delete batch' request.
	defaultSQSDeleteBufferSize = 0

	defaultSQSConsumeBase64 = true
)

func defaultSQSConfig(cfg *SQSConfig) {
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

	if cfg.ConsumeBase64 == nil {
		cfg.ConsumeBase64 = &defaultSQSConsumeBase64
	}
}

type (
	// subscriber is an SQS client that allows a user to
	// consume messages via the pubsub.Subscriber interface.
	subscriber struct {
		sqs sqsiface.SQSAPI

		cfg      SQSConfig
		queueURL *string

		toDelete chan *deleteRequest
		// inFlight and stopped are signals to manage delete requests
		// at shutdown.
		inFlight uint64
		stopped  uint32

		stop   chan chan error
		sqsErr error
	}

	// SQSMessage is the SQS implementation of `SubscriberMessage`.
	subscriberMessage struct {
		sub     *subscriber
		message *sqs.Message
	}

	deleteRequest struct {
		entry   *sqs.DeleteMessageBatchRequestEntry
		receipt chan error
	}
)

// incrementInflight will increment the add in flight count.
func (s *subscriber) incrementInFlight() {
	atomic.AddUint64(&s.inFlight, 1)
}

// removeInfFlight will decrement the in flight count.
func (s *subscriber) decrementInFlight() {
	atomic.AddUint64(&s.inFlight, ^uint64(0))
}

// inFlightCount returns the number of in-flight requests currently
// running on this server.
func (s *subscriber) inFlightCount() uint64 {
	return atomic.LoadUint64(&s.inFlight)
}

// NewSubscriber will initiate a new Decrypter for the subscriber
// if a key file is provided. It will also fetch the SQS Queue Url
// and set up the SQS client.
func NewSubscriber(cfg SQSConfig) (pubsub.Subscriber, error) {
	var err error
	defaultSQSConfig(&cfg)
	s := &subscriber{
		cfg:      cfg,
		toDelete: make(chan *deleteRequest),
		stop:     make(chan chan error, 1),
	}

	if (len(cfg.QueueName) == 0) && (len(cfg.QueueURL) == 0) {
		return s, errors.New("sqs queue name or url is required")
	}

	sess, err := session.NewSession()
	if err != nil {
		return s, err
	}

	var creds *credentials.Credentials
	if cfg.AccessKey != "" {
		creds = credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, cfg.SessionToken)
	} else if cfg.RoleARN != "" {
		var err error
		creds, err = requestRoleCredentials(sess, cfg.RoleARN, cfg.MFASerialNumber)
		if err != nil {
			return s, err
		}
	}

	s.sqs = sqs.New(sess, &aws.Config{
		Credentials: creds,
		Region:      &cfg.Region,
		Endpoint:    cfg.EndpointURL,
	})

	if len(cfg.QueueURL) == 0 {
		var urlResp *sqs.GetQueueUrlOutput
		urlResp, err = s.sqs.GetQueueUrl(&sqs.GetQueueUrlInput{
			QueueName:              &cfg.QueueName,
			QueueOwnerAWSAccountId: &cfg.QueueOwnerAccountID,
		})

		if err != nil {
			return s, err
		}

		s.queueURL = urlResp.QueueUrl
	} else {
		s.queueURL = &cfg.QueueURL
	}

	return s, nil
}

// Message will decode protobufed message bodies and simply return
// a byte slice containing the message body for all others types.
func (m *subscriberMessage) Message() []byte {
	if !*m.sub.cfg.ConsumeBase64 {
		return []byte(*m.message.Body)
	}

	msgBody, err := base64.StdEncoding.DecodeString(*m.message.Body)
	if err != nil {
		pubsub.Log.Warnf("unable to parse message body: %s", err)
	}
	return msgBody
}

// ExtendDoneDeadline changes the visibility timeout of the underlying SQS
// message. It will set the visibility timeout of the message to the given
// duration.
func (m *subscriberMessage) ExtendDoneDeadline(d time.Duration) error {
	_, err := m.sub.sqs.ChangeMessageVisibility(&sqs.ChangeMessageVisibilityInput{
		QueueUrl:          m.sub.queueURL,
		ReceiptHandle:     m.message.ReceiptHandle,
		VisibilityTimeout: aws.Int64(int64(d.Seconds())),
	})
	return err
}

// Done will queue up a message to be deleted. By default,
// the `SQSDeleteBufferSize` will be 0, so this will block until the
// message has been deleted.
func (m *subscriberMessage) Done() error {
	defer m.sub.decrementInFlight()
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
func (s *subscriber) Start() <-chan pubsub.SubscriberMessage {
	output := make(chan pubsub.SubscriberMessage)
	go s.handleDeletes()
	go func(s *subscriber, output chan pubsub.SubscriberMessage) {
		defer close(output)
		var (
			resp *sqs.ReceiveMessageOutput
			err  error
		)
		for {
			select {
			case exit := <-s.stop:
				exit <- nil
				return
			default:
				// get messages
				pubsub.Log.Debugf("receiving messages")
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
					go s.Stop()
					continue
				}

				// if we didn't get any messages, lets chill out for a sec
				if len(resp.Messages) == 0 {
					pubsub.Log.Debugf("no messages found. sleeping for %s", s.cfg.SleepInterval)
					time.Sleep(*s.cfg.SleepInterval)
					continue
				}

				pubsub.Log.Debugf("found %d messages", len(resp.Messages))

				// for each message, pass to output
				for _, msg := range resp.Messages {
					output <- &subscriberMessage{
						sub:     s,
						message: msg,
					}
					s.incrementInFlight()
				}
			}
		}
	}(s, output)
	return output
}

func (s *subscriber) handleDeletes() {
	batchInput := &sqs.DeleteMessageBatchInput{
		QueueUrl: s.queueURL,
	}
	var (
		err           error
		entriesBuffer []*sqs.DeleteMessageBatchRequestEntry
		delRequest    *deleteRequest
	)
	for delRequest = range s.toDelete {
		entriesBuffer = append(entriesBuffer, delRequest.entry)
		// if the subber is stopped and this is the last request,
		// flush quit!
		if s.isStopped() && s.inFlightCount() == 1 {
			break
		}
		// if buffer is full, send the request
		if len(entriesBuffer) > *s.cfg.DeleteBufferSize {
			batchInput.Entries = entriesBuffer
			_, err = s.sqs.DeleteMessageBatch(batchInput)
			// cleaer buffer
			entriesBuffer = []*sqs.DeleteMessageBatchRequestEntry{}
		}

		delRequest.receipt <- err
	}
	// clear any remainders before shutdown
	if len(entriesBuffer) > 0 {
		batchInput.Entries = entriesBuffer
		_, err = s.sqs.DeleteMessageBatch(batchInput)
		delRequest.receipt <- err
	}
}

func (s *subscriber) isStopped() bool {
	return atomic.LoadUint32(&s.stopped) == 1
}

// Stop will block until the consumer has stopped consuming
// messages.
func (s *subscriber) Stop() error {
	if s.isStopped() {
		return errors.New("sqs subscriber is already stopped")
	}
	exit := make(chan error)
	s.stop <- exit
	atomic.SwapUint32(&s.stopped, uint32(1))
	return <-exit
}

// Err will contain any errors that occurred during
// consumption. This method should be checked after
// a user encounters a closed channel.
func (s *subscriber) Err() error {
	return s.sqsErr
}

// requestRoleCredentials return the credentials from AssumeRoleProvider to assume the role
// referenced by the roleARN. If MFASerialNumber is specified, prompt for MFA token from stdin.
func requestRoleCredentials(sess *session.Session, roleARN string, MFASerialNumber string) (*credentials.Credentials, error) {
	if roleARN == "" {
		return nil, errors.New("role ARN is required")
	}

	return stscreds.NewCredentials(sess, roleARN, func(provider *stscreds.AssumeRoleProvider) {
		if MFASerialNumber != "" {
			provider.SerialNumber = &MFASerialNumber
			provider.TokenProvider = stscreds.StdinTokenProvider
		}
	}), nil
}
