package pubsub

import (
	"errors"
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/cloud/pubsub"
)

func TestGCPSubscriber(t *testing.T) {
	msgs := []*testGCPMessage{
		&testGCPMessage{data: []byte("1")},
		&testGCPMessage{data: []byte("2")},
		&testGCPMessage{data: []byte("3")},
		&testGCPMessage{data: []byte("4")},
		&testGCPMessage{data: []byte("5")},
		&testGCPMessage{data: []byte("6")},
		&testGCPMessage{data: []byte("7")},
	}
	gcpSub := testGCPSubscription{
		iter: &testGCPIterator{msgs: msgs},
	}

	testSub := &GCPSubscriber{sub: gcpSub, stop: make(chan chan error, 1)}

	pipe := testSub.Start()

	for _, wantMsg := range msgs {
		gotMsg := <-pipe
		if string(gotMsg.Message()) != string(wantMsg.data) {
			t.Errorf("expected subscriber message to contain %q, got %q",
				string(wantMsg.data), string(gotMsg.Message()))
		}
		gotMsg.Done()
	}

	testSub.Stop()

	msg, ok := <-pipe
	if ok {
		t.Errorf("expected subscriber channel to be closed, but it wasn't. Msg: %s", msg)
	}
}

func TestGCPSubscriberWithErr(t *testing.T) {
	gcpSub := testGCPSubscription{
		iter:     &testGCPIterator{},
		givenErr: errors.New("something's wrong"),
	}

	testSub := &GCPSubscriber{sub: gcpSub, stop: make(chan chan error, 1)}
	pipe := testSub.Start()

	msg, ok := <-pipe
	if ok {
		t.Errorf("expected subscriber channel to be closed, but it wasn't. Msg: %s", msg)
	}

	testSub.Stop()

	if testSub.Err() == nil {
		t.Error("expected subscriber to have global error, but didn't find one")
	}
}

type (
	testGCPMessage struct {
		data  []byte
		doned bool
	}

	testGCPIterator struct {
		index   int
		msgs    []*testGCPMessage
		stopped bool
	}

	testGCPSubscription struct {
		iter     *testGCPIterator
		givenErr error
	}
)

func (m *testGCPMessage) ID() string {
	return "test"
}

func (m *testGCPMessage) MsgData() []byte {
	return m.data
}

func (m *testGCPMessage) Done() {
	m.doned = true
}

func (i *testGCPIterator) Next() (gcpMessage, error) {
	if i.index >= len(i.msgs) {
		return nil, errors.New("no more messages in test iterator")
	}

	msg := i.msgs[i.index]
	i.index++
	return msg, nil
}

func (i *testGCPIterator) Stop() {
	i.stopped = true
}

func (s testGCPSubscription) Pull(ctx context.Context, opts ...pubsub.PullOption) (gcpIterator, error) {
	return s.iter, s.givenErr
}
