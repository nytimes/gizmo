package http // import "github.com/NYTimes/gizmo/pubsub/http"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/NYTimes/gizmo/pubsub"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
)

// Publisher implements the pubsub.Publisher and MultiPublisher interfaces for use in a
// plain HTTP environment.
type Publisher struct {
	url    string
	client *http.Client
}

// GCPPublisher publishes data in the same format as a GCP push-style payload.
type GCPPublisher struct {
	pubsub.Publisher
}

// NewPublisher will return a pubsub.Publisher that simply posts the payload to
// the given URL. If no http.Client is provided, the default one has a 5 second
// timeout.
func NewPublisher(url string, client *http.Client) Publisher {
	if client == nil {
		client = &http.Client{
			Timeout: 5 * time.Second,
		}
	}
	return Publisher{url: url, client: client}
}

// NewGCPStylePublisher will return a pubsub.Publisher that wraps the payload
// in a GCP pubsub.Message-like object that will make this publisher emulate
// Google's PubSub posting messages to a server.
// If no http.Client is provided, the default one has a 5 second
// timeout.
func NewGCPStylePublisher(url string, client *http.Client) GCPPublisher {
	return GCPPublisher{NewPublisher(url, client)}
}

// Publish will serialize the given message and pass it to PublishRaw.
func (p Publisher) Publish(ctx context.Context, key string, msg proto.Message) error {
	payload, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return p.PublishRaw(ctx, key, payload)
}

// PublishMulti will serialize the given messages and pass them to PublishMultiRaw.
func (p Publisher) PublishMulti(ctx context.Context, keys []string, msgs []proto.Message) error {
	bmsgs := make([][]byte, len(msgs))
	for i, msg := range msgs {
		payload, err := proto.Marshal(msg)
		if err != nil {
			return err
		}
		bmsgs[i] = payload
	}
	return p.PublishMultiRaw(ctx, keys, bmsgs)
}

// PublishMultiRaw will call PublishRaw for each message given.
func (p Publisher) PublishMultiRaw(ctx context.Context, _ []string, msgs [][]byte) error {
	for _, msg := range msgs {
		err := p.PublishRaw(ctx, "", msg)
		if err != nil {
			return err
		}
	}
	return nil
}

// PublishRaw will POST the given message payload at the URL provided in the Publisher
// construct.
func (p Publisher) PublishRaw(_ context.Context, _ string, payload []byte) error {
	req, err := http.NewRequest("POST", p.url, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		respBody, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("received an error response (%d): %s",
			resp.StatusCode,
			string(respBody),
		)
	}

	return nil
}

type gcpPayload struct {
	Message message `json:"message"`
}

type message struct {
	Data []byte `json:"data"`
}

// Publish will serialize the given message and pass it to PublishRaw.
func (p GCPPublisher) Publish(ctx context.Context, key string, msg proto.Message) error {
	payload, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return p.PublishRaw(ctx, key, payload)
}

// PublishRaw will wrap the given message in a struct similar to GCP's push-style PubSub
// subscriptions and then POST the message payload at the URL provided in the construct.
func (p GCPPublisher) PublishRaw(ctx context.Context, key string, msg []byte) error {
	payload, err := json.Marshal(gcpPayload{Message: message{Data: msg}})
	if err != nil {
		return err
	}
	return p.Publisher.PublishRaw(ctx, key, payload)
}

// PublishMulti will serialize the given messages and pass them to PublishMultiRaw.
func (p GCPPublisher) PublishMulti(ctx context.Context, keys []string, msgs []proto.Message) error {
	bmsgs := make([][]byte, len(msgs))
	for i, msg := range msgs {
		payload, err := proto.Marshal(msg)
		if err != nil {
			return err
		}
		bmsgs[i] = payload
	}
	return p.PublishMultiRaw(ctx, keys, bmsgs)
}

// PublishMultiRaw will call PublishRaw for each message given.
func (p GCPPublisher) PublishMultiRaw(ctx context.Context, _ []string, msgs [][]byte) error {
	for _, msg := range msgs {
		err := p.PublishRaw(ctx, "", msg)
		if err != nil {
			return err
		}
	}
	return nil
}
