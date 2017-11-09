package aws

import (
	"encoding/base64"
	"errors"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/NYTimes/gizmo/pubsub"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/golang/protobuf/proto"
)

func TestSubscriberNoBase64(t *testing.T) {
	test1 := "hey hey hey!"
	test2 := "ho ho ho!"
	test3 := "yessir!"
	test4 := "nope!"
	sqstest := &TestSQSAPI{
		Messages: [][]*sqs.Message{
			{
				{
					Body:          &test1,
					ReceiptHandle: &test1,
				},
				{
					Body:          &test2,
					ReceiptHandle: &test2,
				},
			},
			{
				{
					Body:          &test3,
					ReceiptHandle: &test3,
				},
				{
					Body:          &test4,
					ReceiptHandle: &test4,
				},
			},
		},
	}

	fals := false
	cfg := SQSConfig{ConsumeBase64: &fals}
	defaultSQSConfig(&cfg)
	sub := &subscriber{
		sqs:      sqstest,
		cfg:      cfg,
		toDelete: make(chan *deleteRequest),
		stop:     make(chan chan error, 1),
	}

	queue := sub.Start()
	verifySQSSub(t, queue, sqstest, test1, 0)
	verifySQSSub(t, queue, sqstest, test2, 1)
	verifySQSSub(t, queue, sqstest, test3, 2)
	verifySQSSub(t, queue, sqstest, test4, 3)
	sub.Stop()

}
func TestSQSReceiveError(t *testing.T) {
	wantErr := errors.New("my sqs error")
	sqstest := &TestSQSAPI{
		Err: wantErr,
	}

	fals := false
	cfg := SQSConfig{ConsumeBase64: &fals}
	defaultSQSConfig(&cfg)
	sub := &subscriber{
		sqs:      sqstest,
		cfg:      cfg,
		toDelete: make(chan *deleteRequest),
		stop:     make(chan chan error, 1),
	}

	queue := sub.Start()
	_, ok := <-queue
	if ok {
		t.Error("no message should've gotten to us, the channel should be closed")
		return
	}
	sub.Stop()

	if sub.Err() != wantErr {
		t.Errorf("expected subscriber to return error '%s'; got '%s'",
			wantErr, sub.Err())
	}

}
func TestSQSDoneAfterStop(t *testing.T) {
	test := "it stopped??"
	sqstest := &TestSQSAPI{
		Messages: [][]*sqs.Message{
			{
				{
					Body:          &test,
					ReceiptHandle: &test,
				},
			},
		},
	}

	fals := false
	cfg := SQSConfig{ConsumeBase64: &fals}
	defaultSQSConfig(&cfg)
	sub := &subscriber{
		sqs:      sqstest,
		cfg:      cfg,
		toDelete: make(chan *deleteRequest),
		stop:     make(chan chan error, 1),
	}

	queue := sub.Start()
	// verify we can receive a message, stop and still mark the message as 'done'
	gotRaw := <-queue
	sub.Stop()
	gotRaw.Done()
	// do all the other normal verifications
	if len(sqstest.Deleted) != 1 {
		t.Errorf("subscriber expected %d deleted message, got: %d", 1, len(sqstest.Deleted))
	}

	if *sqstest.Deleted[0].ReceiptHandle != test {
		t.Errorf("subscriber expected receipt handle of \"%s\" , got:+ \"%s\"",
			test,
			*sqstest.Deleted[0].ReceiptHandle)
	}
}
func TestExtendDoneTimeout(t *testing.T) {
	test := "some test"
	sqstest := &TestSQSAPI{
		Messages: [][]*sqs.Message{
			{
				{
					Body:          &test,
					ReceiptHandle: &test,
				},
			},
		},
	}

	fals := false
	cfg := SQSConfig{ConsumeBase64: &fals}
	defaultSQSConfig(&cfg)
	sub := &subscriber{
		sqs:      sqstest,
		cfg:      cfg,
		toDelete: make(chan *deleteRequest),
		stop:     make(chan chan error, 1),
	}

	queue := sub.Start()
	defer sub.Stop()
	gotRaw := <-queue
	gotRaw.ExtendDoneDeadline(time.Hour)
	if len(sqstest.Extended) != 1 {
		t.Errorf("subscriber expected %d extended message, got %d", 1, len(sqstest.Extended))
	}

	if *sqstest.Extended[0].ReceiptHandle != test {
		t.Errorf("subscriber expected receipt handle of %q , got:+ %q", test, *sqstest.Extended[0].ReceiptHandle)
	}
}

func verifySQSSub(t *testing.T, queue <-chan pubsub.SubscriberMessage, testsqs *TestSQSAPI, want string, index int) {
	gotRaw := <-queue
	got := string(gotRaw.Message())
	if got != want {
		t.Errorf("subscriber expected:\n%#v\ngot:\n%#v", want, got)
	}
	gotRaw.Done()

	if len(testsqs.Deleted) != (index + 1) {
		t.Errorf("subscriber expected %d deleted message, got: %d", index+1, len(testsqs.Deleted))
	}

	if *testsqs.Deleted[index].ReceiptHandle != want {
		t.Errorf("subscriber expected receipt handle of \"%s\" , got: \"%s\"",
			want,
			*testsqs.Deleted[index].ReceiptHandle)
	}
}

func TestSubscriber(t *testing.T) {
	test1 := &TestProto{"hey hey hey!"}
	test2 := &TestProto{"ho ho ho!"}
	test3 := &TestProto{"yessir!"}
	test4 := &TestProto{"nope!"}
	sqstest := &TestSQSAPI{
		Messages: [][]*sqs.Message{
			{
				{
					Body:          makeB64String(test1),
					ReceiptHandle: &test1.Value,
				},
				{
					Body:          makeB64String(test2),
					ReceiptHandle: &test2.Value,
				},
			},
			{
				{
					Body:          makeB64String(test3),
					ReceiptHandle: &test3.Value,
				},
				{
					Body:          makeB64String(test4),
					ReceiptHandle: &test4.Value,
				},
			},
		},
	}

	cfg := SQSConfig{}
	defaultSQSConfig(&cfg)
	sub := &subscriber{
		sqs:      sqstest,
		cfg:      cfg,
		toDelete: make(chan *deleteRequest),
		stop:     make(chan chan error, 1),
	}

	queue := sub.Start()

	verifySQSSubProto(t, queue, sqstest, test1, 0)
	verifySQSSubProto(t, queue, sqstest, test2, 1)
	verifySQSSubProto(t, queue, sqstest, test3, 2)
	verifySQSSubProto(t, queue, sqstest, test4, 3)

	sub.Stop()
}

func verifySQSSubProto(t *testing.T, queue <-chan pubsub.SubscriberMessage, testsqs *TestSQSAPI, want *TestProto, index int) {
	gotRaw := <-queue
	got := makeProto(gotRaw.Message())
	if !reflect.DeepEqual(got, want) {
		t.Errorf("subscriber expected:\n%#v\ngot:\n%#v", want, got)
	}
	gotRaw.Done()

	if len(testsqs.Deleted) != (index + 1) {
		t.Errorf("subscriber expected %d deleted message, got: %d", index+1, len(testsqs.Deleted))
	}

	if *testsqs.Deleted[index].ReceiptHandle != want.Value {
		t.Errorf("subscriber expected receipt handle of \"%s\" , got: \"%s\"",
			want.Value,
			*testsqs.Deleted[index].ReceiptHandle)
	}
}

func makeB64String(p proto.Message) *string {
	b, _ := proto.Marshal(p)
	s := base64.StdEncoding.EncodeToString(b)
	return &s
}

func makeProto(b []byte) *TestProto {
	t := &TestProto{}
	err := proto.Unmarshal(b, t)
	if err != nil {
		log.Fatalf("unable to unmarshal protobuf: %s", err)
	}
	return t
}

/*
  500000	     13969 ns/op	    1494 B/op	      31 allocs/op
 1000000	     14248 ns/op	    1491 B/op	      31 allocs/op
 2000000	     14138 ns/op	    1489 B/op	      31 allocs/op
*/
func BenchmarkSubscriber_Proto(b *testing.B) {
	test1 := &TestProto{"hey hey hey!"}
	sqstest := &TestSQSAPI{
		Messages: [][]*sqs.Message{
			{
				{
					Body:          makeB64String(test1),
					ReceiptHandle: &test1.Value,
				},
			},
		},
	}

	for i := 0; i < b.N/2; i++ {
		sqstest.Messages = append(sqstest.Messages, []*sqs.Message{
			{
				Body:          makeB64String(test1),
				ReceiptHandle: &test1.Value,
			},
			{
				Body:          makeB64String(test1),
				ReceiptHandle: &test1.Value,
			},
		})

	}

	cfg := SQSConfig{}
	defaultSQSConfig(&cfg)
	sub := &subscriber{
		sqs:      sqstest,
		cfg:      cfg,
		toDelete: make(chan *deleteRequest),
		stop:     make(chan chan error, 1),
	}
	queue := sub.Start()
	for i := 0; i < b.N; i++ {
		gotRaw := <-queue
		// get message, forcing base64 decode
		gotRaw.Message()
		// send delete message
		gotRaw.Done()
	}
	go sub.Stop()
}

type TestSQSAPI struct {
	Offset   int
	Messages [][]*sqs.Message
	Deleted  []*sqs.DeleteMessageBatchRequestEntry
	Extended []*sqs.ChangeMessageVisibilityInput
	Err      error
}

var _ sqsiface.SQSAPI = &TestSQSAPI{}

func (s *TestSQSAPI) ReceiveMessage(*sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	if s.Offset >= len(s.Messages) {
		return &sqs.ReceiveMessageOutput{}, s.Err
	}
	out := s.Messages[s.Offset]
	s.Offset++
	return &sqs.ReceiveMessageOutput{Messages: out}, s.Err
}

func (s *TestSQSAPI) DeleteMessageBatch(i *sqs.DeleteMessageBatchInput) (*sqs.DeleteMessageBatchOutput, error) {
	s.Deleted = append(s.Deleted, i.Entries...)
	return nil, errNotImpl
}

func (s *TestSQSAPI) ChangeMessageVisibility(i *sqs.ChangeMessageVisibilityInput) (*sqs.ChangeMessageVisibilityOutput, error) {
	s.Extended = append(s.Extended, i)
	return nil, nil
}

///////////
// ALL METHODS BELOW HERE ARE EMPTY AND JUST SATISFYING THE SQSAPI interface
///////////

func (s *TestSQSAPI) DeleteMessage(d *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) DeleteMessageWithContext(aws.Context, *sqs.DeleteMessageInput, ...request.Option) (*sqs.DeleteMessageOutput, error) {
	return nil, errNotImpl
}

func (s *TestSQSAPI) DeleteMessageBatchRequest(i *sqs.DeleteMessageBatchInput) (*request.Request, *sqs.DeleteMessageBatchOutput) {
	return nil, nil
}
func (s *TestSQSAPI) DeleteMessageBatchWithContext(aws.Context, *sqs.DeleteMessageBatchInput, ...request.Option) (*sqs.DeleteMessageBatchOutput, error) {
	return nil, errNotImpl
}

func (s *TestSQSAPI) AddPermissionRequest(*sqs.AddPermissionInput) (*request.Request, *sqs.AddPermissionOutput) {
	return nil, nil
}
func (s *TestSQSAPI) AddPermission(*sqs.AddPermissionInput) (*sqs.AddPermissionOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) AddPermissionWithContext(aws.Context, *sqs.AddPermissionInput, ...request.Option) (*sqs.AddPermissionOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) ChangeMessageVisibilityRequest(*sqs.ChangeMessageVisibilityInput) (*request.Request, *sqs.ChangeMessageVisibilityOutput) {
	return nil, nil
}
func (s *TestSQSAPI) ChangeMessageVisibilityWithContext(aws.Context, *sqs.ChangeMessageVisibilityInput, ...request.Option) (*sqs.ChangeMessageVisibilityOutput, error) {
	return nil, errNotImpl
}

func (s *TestSQSAPI) ChangeMessageVisibilityBatchRequest(*sqs.ChangeMessageVisibilityBatchInput) (*request.Request, *sqs.ChangeMessageVisibilityBatchOutput) {
	return nil, nil
}
func (s *TestSQSAPI) ChangeMessageVisibilityBatch(*sqs.ChangeMessageVisibilityBatchInput) (*sqs.ChangeMessageVisibilityBatchOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) ChangeMessageVisibilityBatchWithContext(aws.Context, *sqs.ChangeMessageVisibilityBatchInput, ...request.Option) (*sqs.ChangeMessageVisibilityBatchOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) CreateQueueRequest(*sqs.CreateQueueInput) (*request.Request, *sqs.CreateQueueOutput) {
	return nil, nil
}
func (s *TestSQSAPI) CreateQueue(*sqs.CreateQueueInput) (*sqs.CreateQueueOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) CreateQueueWithContext(aws.Context, *sqs.CreateQueueInput, ...request.Option) (*sqs.CreateQueueOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) DeleteMessageRequest(*sqs.DeleteMessageInput) (*request.Request, *sqs.DeleteMessageOutput) {
	return nil, nil
}

func (s *TestSQSAPI) DeleteQueueRequest(*sqs.DeleteQueueInput) (*request.Request, *sqs.DeleteQueueOutput) {
	return nil, nil
}
func (s *TestSQSAPI) DeleteQueue(*sqs.DeleteQueueInput) (*sqs.DeleteQueueOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) DeleteQueueWithContext(aws.Context, *sqs.DeleteQueueInput, ...request.Option) (*sqs.DeleteQueueOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) GetQueueAttributesRequest(*sqs.GetQueueAttributesInput) (*request.Request, *sqs.GetQueueAttributesOutput) {
	return nil, nil
}
func (s *TestSQSAPI) GetQueueAttributes(*sqs.GetQueueAttributesInput) (*sqs.GetQueueAttributesOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) GetQueueAttributesWithContext(aws.Context, *sqs.GetQueueAttributesInput, ...request.Option) (*sqs.GetQueueAttributesOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) GetQueueUrlRequest(*sqs.GetQueueUrlInput) (*request.Request, *sqs.GetQueueUrlOutput) {
	return nil, nil
}
func (s *TestSQSAPI) GetQueueUrl(*sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) GetQueueUrlWithContext(aws.Context, *sqs.GetQueueUrlInput, ...request.Option) (*sqs.GetQueueUrlOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) ListDeadLetterSourceQueuesRequest(*sqs.ListDeadLetterSourceQueuesInput) (*request.Request, *sqs.ListDeadLetterSourceQueuesOutput) {
	return nil, nil
}
func (s *TestSQSAPI) ListDeadLetterSourceQueues(*sqs.ListDeadLetterSourceQueuesInput) (*sqs.ListDeadLetterSourceQueuesOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) ListDeadLetterSourceQueuesWithContext(aws.Context, *sqs.ListDeadLetterSourceQueuesInput, ...request.Option) (*sqs.ListDeadLetterSourceQueuesOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) ListQueuesRequest(*sqs.ListQueuesInput) (*request.Request, *sqs.ListQueuesOutput) {
	return nil, nil
}
func (s *TestSQSAPI) ListQueues(*sqs.ListQueuesInput) (*sqs.ListQueuesOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) ListQueuesWithContext(aws.Context, *sqs.ListQueuesInput, ...request.Option) (*sqs.ListQueuesOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) PurgeQueueRequest(*sqs.PurgeQueueInput) (*request.Request, *sqs.PurgeQueueOutput) {
	return nil, nil
}
func (s *TestSQSAPI) PurgeQueue(*sqs.PurgeQueueInput) (*sqs.PurgeQueueOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) PurgeQueueWithContext(aws.Context, *sqs.PurgeQueueInput, ...request.Option) (*sqs.PurgeQueueOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) ReceiveMessageRequest(*sqs.ReceiveMessageInput) (*request.Request, *sqs.ReceiveMessageOutput) {
	return nil, nil
}
func (s *TestSQSAPI) ReceiveMessageWithContext(aws.Context, *sqs.ReceiveMessageInput, ...request.Option) (*sqs.ReceiveMessageOutput, error) {
	return nil, errNotImpl
}

func (s *TestSQSAPI) RemovePermissionRequest(*sqs.RemovePermissionInput) (*request.Request, *sqs.RemovePermissionOutput) {
	return nil, nil
}
func (s *TestSQSAPI) RemovePermission(*sqs.RemovePermissionInput) (*sqs.RemovePermissionOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) RemovePermissionWithContext(aws.Context, *sqs.RemovePermissionInput, ...request.Option) (*sqs.RemovePermissionOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) SendMessageRequest(*sqs.SendMessageInput) (*request.Request, *sqs.SendMessageOutput) {
	return nil, nil
}
func (s *TestSQSAPI) SendMessage(*sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) SendMessageWithContext(aws.Context, *sqs.SendMessageInput, ...request.Option) (*sqs.SendMessageOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) SendMessageBatchRequest(*sqs.SendMessageBatchInput) (*request.Request, *sqs.SendMessageBatchOutput) {
	return nil, nil
}
func (s *TestSQSAPI) SendMessageBatch(*sqs.SendMessageBatchInput) (*sqs.SendMessageBatchOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) SendMessageBatchWithContext(aws.Context, *sqs.SendMessageBatchInput, ...request.Option) (*sqs.SendMessageBatchOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) SetQueueAttributesRequest(*sqs.SetQueueAttributesInput) (*request.Request, *sqs.SetQueueAttributesOutput) {
	return nil, nil
}
func (s *TestSQSAPI) SetQueueAttributes(*sqs.SetQueueAttributesInput) (*sqs.SetQueueAttributesOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) SetQueueAttributesWithContext(aws.Context, *sqs.SetQueueAttributesInput, ...request.Option) (*sqs.SetQueueAttributesOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) ListQueueTags(input *sqs.ListQueueTagsInput) (*sqs.ListQueueTagsOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) ListQueueTagsRequest(input *sqs.ListQueueTagsInput) (req *request.Request, output *sqs.ListQueueTagsOutput) {
	return nil, nil
}
func (s *TestSQSAPI) ListQueueTagsWithContext(ctx aws.Context, input *sqs.ListQueueTagsInput, opts ...request.Option) (*sqs.ListQueueTagsOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) TagQueue(input *sqs.TagQueueInput) (*sqs.TagQueueOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) TagQueueRequest(input *sqs.TagQueueInput) (req *request.Request, output *sqs.TagQueueOutput) {
	return nil, nil
}
func (s *TestSQSAPI) TagQueueWithContext(ctx aws.Context, input *sqs.TagQueueInput, opts ...request.Option) (*sqs.TagQueueOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) UntagQueue(input *sqs.UntagQueueInput) (*sqs.UntagQueueOutput, error) {
	return nil, errNotImpl
}
func (s *TestSQSAPI) UntagQueueRequest(input *sqs.UntagQueueInput) (req *request.Request, output *sqs.UntagQueueOutput) {
	return nil, nil
}
func (s *TestSQSAPI) UntagQueueWithContext(ctx aws.Context, input *sqs.UntagQueueInput, opts ...request.Option) (*sqs.UntagQueueOutput, error) {
	return nil, errNotImpl
}
