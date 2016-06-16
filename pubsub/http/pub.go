package http

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

type publisher struct {
	url    string
	client *http.Client
}

type gcpPublisher struct {
	pubsub.Publisher
}

// NewPublisher will return a pubsub.Publisher that simply posts the payload to
// the given URL. If no http.Client is provided, the default one has a 5 second
// timeout.
func NewPublisher(url string, client *http.Client) pubsub.Publisher {
	if client == nil {
		client = &http.Client{
			Timeout: 5 * time.Second,
		}
	}
	return publisher{url: url, client: client}
}

// NewGCPStylePublisher will return a pubsub.Publisher that wraps the payload
// in a GCP pubsub.Message-like object that will make this publisher emulate
// Google's PubSub posting messages to a server.
// If no http.Client is provided, the default one has a 5 second
// timeout.
func NewGCPStylePublisher(url string, client *http.Client) pubsub.Publisher {
	return gcpPublisher{NewPublisher(url, client)}
}

func (p publisher) Publish(ctx context.Context, key string, msg proto.Message) error {
	payload, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return p.PublishRaw(ctx, key, payload)

}

func (p publisher) PublishRaw(_ context.Context, _ string, payload []byte) error {
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

func (p gcpPublisher) Publish(ctx context.Context, key string, msg proto.Message) error {
	payload, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return p.PublishRaw(ctx, key, payload)
}

func (p gcpPublisher) PublishRaw(ctx context.Context, key string, msg []byte) error {
	payload, err := json.Marshal(gcpPayload{Message: message{Data: msg}})
	if err != nil {
		return err
	}
	return p.Publisher.PublishRaw(ctx, key, payload)
}
