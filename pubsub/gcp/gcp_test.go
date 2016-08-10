package gcp

import (
	"errors"
	"testing"

	"cloud.google.com/go/pubsub"

	"golang.org/x/net/context"
)

func TestGCPSubscriber(t *testing.T) {
	msgs := []*testMessage{
		&testMessage{data: []byte("1")},
		&testMessage{data: []byte("2")},
		&testMessage{data: []byte("3")},
		&testMessage{data: []byte("4")},
		&testMessage{data: []byte("5")},
		&testMessage{data: []byte("6")},
		&testMessage{data: []byte("7")},
	}
	gcpSub := testSubscription{
		iter: &testIterator{msgs: msgs},
	}

	testSub := &subscriber{sub: gcpSub, stop: make(chan chan error, 1)}

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

func TestSubscriberWithErr(t *testing.T) {
	gcpSub := testSubscription{
		iter:     &testIterator{},
		givenErr: errors.New("something's wrong"),
	}

	testSub := &subscriber{sub: gcpSub, stop: make(chan chan error, 1)}
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
	testMessage struct {
		data  []byte
		doned bool
	}

	testIterator struct {
		index   int
		msgs    []*testMessage
		stopped bool
	}

	testSubscription struct {
		iter     *testIterator
		givenErr error
	}
)

func (m *testMessage) ID() string {
	return "test"
}

func (m *testMessage) MsgData() []byte {
	return m.data
}

func (m *testMessage) Done() {
	m.doned = true
}

func (i *testIterator) Next() (message, error) {
	if i.index >= len(i.msgs) {
		return nil, errors.New("no more messages in test iterator")
	}

	msg := i.msgs[i.index]
	i.index++
	return msg, nil
}

func (i *testIterator) Stop() {
	i.stopped = true
}

func (s testSubscription) Pull(ctx context.Context, opts ...pubsub.PullOption) (iterator, error) {
	return s.iter, s.givenErr
}
