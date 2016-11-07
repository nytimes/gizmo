package gcp

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	v1pubsub "google.golang.org/api/pubsub/v1"

	"github.com/NYTimes/gizmo/pubsub"
)

var _ pubsub.MultiPublisher = &httpPublisher{}
var _ pubsub.Publisher = &httpPublisher{}

type httpPublisher struct {
	svc   *v1pubsub.ProjectsTopicsService
	topic string
}

// NewHTTPPublisher will instantiate a new GCP MultiPublisher that utilizes the HTTP client.
// This client is useful mainly for the App Engine standard environment as the gRPC client
// counts against the socket quota for some reason.
func NewHTTPPublisher(ctx context.Context, projID, topic string, src oauth2.TokenSource) (pubsub.MultiPublisher, error) {
	client := oauth2.NewClient(ctx, src)
	svc, err := v1pubsub.New(client)
	if err != nil {
		return nil, err
	}
	return &httpPublisher{
		topic: fmt.Sprintf("projects/%s/topics/%s", projID, topic),
		svc:   v1pubsub.NewProjectsTopicsService(svc),
	}, nil
}

// Publish will marshal the proto message and publish it to GCP pubsub.
func (p *httpPublisher) Publish(ctx context.Context, key string, msg proto.Message) error {
	mb, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return p.PublishRaw(ctx, key, mb)
}

// PublishRaw will publish the message to GCP pubsub.
func (p *httpPublisher) PublishRaw(ctx context.Context, key string, m []byte) error {
	call := p.svc.Publish(p.topic, &v1pubsub.PublishRequest{
		Messages: []*v1pubsub.PubsubMessage{
			{
				Data:       base64.StdEncoding.EncodeToString(m),
				Attributes: map[string]string{"key": key},
			},
		},
	})
	_, err := call.Do()
	return err
}

// PublishMulti will publish multiple messages to GCP pubsub in a single request.
func (p *httpPublisher) PublishMulti(ctx context.Context, keys []string, messages []proto.Message) error {
	if len(keys) != len(messages) {
		return errors.New("keys and messages must be equal length")
	}

	a := make([][]byte, len(messages))
	for i := range messages {
		b, err := proto.Marshal(messages[i])
		if err != nil {
			return err
		}
		a[i] = b
	}
	return p.PublishMultiRaw(ctx, keys, a)
}

// PublishMultiRaw will publish multiple raw byte array messages to GCP pubsub in a single request.
func (p *httpPublisher) PublishMultiRaw(ctx context.Context, keys []string, messages [][]byte) error {
	if len(keys) != len(messages) {
		return errors.New("keys and messages must be equal length")
	}

	a := make([]*v1pubsub.PubsubMessage, len(messages))
	for i := range messages {
		a[i] = &v1pubsub.PubsubMessage{
			Data:       base64.StdEncoding.EncodeToString(messages[i]),
			Attributes: map[string]string{"key": keys[i]},
		}
	}

	call := p.svc.Publish(p.topic, &v1pubsub.PublishRequest{
		Messages: a,
	})
	_, err := call.Do()
	return err
}
