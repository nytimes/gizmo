package http

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/golang/protobuf/proto"
)

func TestPublishRaw(t *testing.T) {
	tests := []struct {
		givenPayload []byte
		givenHandler http.HandlerFunc

		wantErr bool
	}{
		{
			[]byte("hi there!"),
			func(w http.ResponseWriter, r *http.Request) {
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Errorf("unable to read published request body: %s", err)
				}

				if string(body) != "hi there!" {
					t.Errorf("expected published request body to be 'hi there!', but was %q", string(body))
				}

				w.WriteHeader(http.StatusOK)
				io.WriteString(w, "good jorb")
			},

			false,
		},
		{
			[]byte("hi there!"),
			func(w http.ResponseWriter, r *http.Request) {
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Errorf("unable to read published request body: %s", err)
				}

				if string(body) != "hi there!" {
					t.Errorf("expected published request body to be 'hi there!', but was %q", string(body))
				}

				w.WriteHeader(http.StatusServiceUnavailable)
				io.WriteString(w, "doh!")
			},

			true,
		},
	}

	for _, test := range tests {
		srv := httptest.NewServer(test.givenHandler)

		pub := NewPublisher(srv.URL, nil)

		gotErr := pub.PublishRaw(nil, "", test.givenPayload)

		if test.wantErr && gotErr == nil {
			t.Errorf("expected error response from publish but got none")
		}
		if !test.wantErr && gotErr != nil {
			t.Errorf("expected no error response from publish but got one: %s", gotErr)
		}
		srv.Close()
	}

}

func TestPublish(t *testing.T) {
	tests := []struct {
		givenPayload proto.Message
		givenHandler http.HandlerFunc

		wantErr bool
	}{
		{
			&TestProto{"hi there!"},
			func(w http.ResponseWriter, r *http.Request) {
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Errorf("unable to read published request body: %s", err)
				}

				var got TestProto
				err = proto.Unmarshal(body, &got)
				if err != nil {
					t.Errorf("unable to proto marshal published request body: %s", err)
				}

				want := &TestProto{"hi there!"}
				if !reflect.DeepEqual(&got, want) {
					t.Errorf("expected published request body to be %#v', but was %#v", want, got)
				}

				w.WriteHeader(http.StatusOK)
				io.WriteString(w, "good jorb")
			},

			false,
		},
		{
			&TestProto{"hi there!"},
			func(w http.ResponseWriter, r *http.Request) {
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Errorf("unable to read published request body: %s", err)
				}

				var got TestProto
				err = proto.Unmarshal(body, &got)
				if err != nil {
					t.Errorf("unable to proto marshal published request body: %s", err)
				}

				want := &TestProto{"hi there!"}
				if !reflect.DeepEqual(&got, want) {
					t.Errorf("expected published request body to be %#v', but was %#v", want, got)
				}

				w.WriteHeader(http.StatusServiceUnavailable)
				io.WriteString(w, "doh!")
			},

			true,
		},
		{
			nil,
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				io.WriteString(w, "good jorb?")
			},

			true,
		},
	}

	for _, test := range tests {
		srv := httptest.NewServer(test.givenHandler)

		pub := NewPublisher(srv.URL, nil)

		gotErr := pub.Publish(nil, "", test.givenPayload)

		if test.wantErr && gotErr == nil {
			t.Errorf("expected error response from publish but got none")
		}
		if !test.wantErr && gotErr != nil {
			t.Errorf("expected no error response from publish but got one: %s", gotErr)
		}
		srv.Close()
	}

}

func TestGCPPublish(t *testing.T) {
	tests := []struct {
		givenPayload proto.Message
		givenHandler http.HandlerFunc

		wantErr bool
	}{
		{
			&TestProto{"hi there!"},
			func(w http.ResponseWriter, r *http.Request) {
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Errorf("unable to read published request body: %s", err)
				}

				var msg gcpPayload
				err = json.Unmarshal(body, &msg)
				if err != nil {
					t.Errorf("unable to json marshal published request body: %s", err)
				}

				var got TestProto
				err = proto.Unmarshal(msg.Message.Data, &got)
				if err != nil {
					t.Errorf("unable to proto marshal published request body: %s", err)
				}

				want := &TestProto{"hi there!"}
				if !reflect.DeepEqual(&got, want) {
					t.Errorf("expected published request body to be %#v', but was %#v", want, got)
				}

				w.WriteHeader(http.StatusOK)
				io.WriteString(w, "good jorb")
			},

			false,
		},
		{
			&TestProto{"hi there!"},
			func(w http.ResponseWriter, r *http.Request) {
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Errorf("unable to read published request body: %s", err)
				}

				var msg gcpPayload
				err = json.Unmarshal(body, &msg)
				if err != nil {
					t.Errorf("unable to json marshal published request body: %s", err)
				}

				var got TestProto
				err = proto.Unmarshal(msg.Message.Data, &got)
				if err != nil {
					t.Errorf("unable to proto marshal published request body: %s", err)
				}

				want := &TestProto{"hi there!"}
				if !reflect.DeepEqual(&got, want) {
					t.Errorf("expected published request body to be %#v', but was %#v", want, got)
				}

				w.WriteHeader(http.StatusServiceUnavailable)
				io.WriteString(w, "doh!")
			},

			true,
		},
		{
			nil,
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				io.WriteString(w, "good jorb?")
			},

			true,
		},
	}

	for _, test := range tests {
		srv := httptest.NewServer(test.givenHandler)

		pub := NewGCPStylePublisher(srv.URL, nil)

		gotErr := pub.Publish(nil, "", test.givenPayload)

		if test.wantErr && gotErr == nil {
			t.Errorf("expected error response from publish but got none")
		}
		if !test.wantErr && gotErr != nil {
			t.Errorf("expected no error response from publish but got one: %s", gotErr)
		}
		srv.Close()
	}

}
