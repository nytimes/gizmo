package gcp

import (
	"errors"
	"testing"

	"github.com/NYTimes/gizmo/pubsub"
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
	gcpSub := &testSubscription{
		msgs: msgs,
	}

	testSub := &Subscriber{sub: gcpSub, ctx: context.Background()}

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
	gcpSub := &testSubscription{
		givenErr: errors.New("something's wrong"),
	}

	var testSub pubsub.Subscriber
	testSub = &Subscriber{sub: gcpSub, ctx: context.Background()}
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

	testSubscription struct {
		msgs []*testMessage

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

func (s *testSubscription) Receive(ctx context.Context, f func(context.Context, message)) error {
	// iterate over messages and call f
	for _, msg := range s.msgs {
		f(ctx, msg)
	}
	return s.givenErr
}
