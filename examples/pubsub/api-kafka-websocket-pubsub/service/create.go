package service

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/NYTimes/gizmo/server"
)

// CreateStream is a JSON endpoint for creating a new topic in Kafka.
func (s *StreamService) CreateStream(r *http.Request) (int, interface{}, error) {
	id := time.Now().Unix()
	topic := topicName(id)
	err := createTopic(topic)
	if err != nil {
		return http.StatusInternalServerError, nil, jsonErr{err}
	}

	server.LogWithFields(r).WithField("topic", topic).Info("created new topic")

	return http.StatusOK, struct {
		Status   string `json:"status"`
		StreamID int64  `json:"stream_id"`
	}{"success!", id}, nil
}

func topicName(id int64) string {
	return fmt.Sprintf("stream-%d", id)
}

func createTopic(name string) error {
	cmd := exec.Command("kafka-topics.sh",
		"--create",
		"--zookeeper",
		"localhost:2181",
		"--replication-factor",
		"1",
		"--partition",
		"1",
		"--topic",
		name)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

type jsonErr struct {
	Err error `json:"error"`
}

func (e jsonErr) Error() string {
	return e.Err.Error()
}
