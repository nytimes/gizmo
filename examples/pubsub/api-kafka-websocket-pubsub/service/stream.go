package service

import (
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/NYTimes/gizmo/pubsub"
	"github.com/NYTimes/gizmo/pubsub/kafka"
	"github.com/NYTimes/gizmo/server"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func zeroOffset() int64          { return 0 }
func discardOffset(offset int64) {}

// Stream will init a new pubsub.Publisher and pubsub.Subscriber
// then upgrade the current request to a websocket connection. Any messages
// consumed from Kafka will be published to the web socket and vice versa.
func (s *StreamService) Stream(w http.ResponseWriter, r *http.Request) {
	cfg := *s.cfg
	cfg.Topic = topicName(server.GetInt64Var(r, "stream_id"))
	server.LogWithFields(r).WithField("topic", cfg.Topic).Info("new stream req")

	sub, err := kafka.NewSubscriber(&cfg, zeroOffset, discardOffset)
	if err != nil {
		server.LogWithFields(r).WithField("error", err).Error("unable to create sub")
		http.Error(w, "unable to create subscriber: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer func() {
		if err := sub.Stop(); err != nil {
			server.LogWithFields(r).WithField("error", err).Error("unable to stop sub")
		}
	}()

	var pub pubsub.Publisher
	pub, err = kafka.NewPublisher(&cfg)
	if err != nil {
		server.LogWithFields(r).WithField("error", err).Error("unable to create pub")
		http.Error(w, "unable to create publisher: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer func() {
		kpub, ok := pub.(*kafka.Publisher)
		if ok {
			if err := kpub.Stop(); err != nil {
				server.LogWithFields(r).WithField("error", err).Error("unable to stop pub")
			}
		}
	}()
	var ws *websocket.Conn
	ws, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		server.LogWithFields(r).WithField("error", err).Error("unable to create websocket")
		http.Error(w, "unable to create websocket: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer func() {
		if err := ws.Close(); err != nil {
			server.LogWithFields(r).WithField("error", err).Error("unable to close ws")
		}
	}()

	// start consumer, emit to ws
	noopTicker := time.NewTicker(time.Second * 5)
	subscriberDone := make(chan bool, 1)
	stopSubscriber := make(chan bool, 1)
	go func() {
		defer func() { subscriberDone <- true }()
		var (
			payload []byte
			msgs    = sub.Start()
		)
		for {
			select {
			case msg := <-msgs:
				payload = msg.Message()
				err = ws.SetWriteDeadline(time.Now().Add(time.Second * 30))
				if err != nil {
					server.LogWithFields(r).WithField("error", err).Error("unable to write deadline")
				}
				err = ws.WriteMessage(websocket.TextMessage, payload)
				if err != nil {
					server.LogWithFields(r).WithField("error", err).Error("unable to write ws message")
				}
				err = msg.Done()
				if err != nil {
					server.LogWithFields(r).WithField("error", err).Error("unable to mark gizmo message as done")
				}
			case <-noopTicker.C:
				err = ws.SetWriteDeadline(time.Now().Add(time.Second * 30))
				if err != nil {
					server.LogWithFields(r).WithField("error", err).Error("unable to write deadline")
				}
				if err := ws.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
					server.LogWithFields(r).WithField("error", err).Error("error writing ws message")
					return
				}
			case <-stopSubscriber:
				return
			}
		}
	}()

	// start producer, emit to kafka
	producerDone := make(chan bool, 1)
	go func() {
		defer func() {
			producerDone <- true
			stopSubscriber <- true
			server.LogWithFields(r).WithField("topic", cfg.Topic).Info("closing stream req")
		}()
		var (
			messageType int
			payload     []byte
			err         error
			read        io.Reader
		)
		for {
			messageType, read, err = ws.NextReader()
			if err != nil {
				if err != io.EOF {
					server.LogWithFields(r).WithField("error", err).Error("error reading message")
				}
				return
			}

			switch messageType {
			case websocket.TextMessage:
				payload, err = ioutil.ReadAll(read)
				if err != nil {
					server.LogWithFields(r).WithField("error", err).Error("unable to read payload")
					return
				}
				err = pub.PublishRaw(nil, cfg.Topic, payload)
				if err != nil {
					server.LogWithFields(r).WithField("error", err).Error("unable to publish payload")
					return
				}
			case websocket.PingMessage, websocket.PongMessage, websocket.BinaryMessage:
				server.LogWithFields(r).Info("discarding message type: ", messageType)
			case websocket.CloseMessage:
				server.LogWithFields(r).Info("closing websocket")
				return
			}
		}
	}()

	<-subscriberDone
	<-producerDone
	noopTicker.Stop()
	server.LogWithFields(r).WithField("topic", cfg.Topic).Info("leaving stream req")
}
