package server

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCounterByStatusXX(t *testing.T) {
	tests := []int{111, 222, 333, 444, 555}
	statuses := make(chan int, 1)

	counter := CountedByStatusXX(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := <-statuses
		w.WriteHeader(status)
		if bod, _ := ioutil.ReadAll(r.Body); string(bod) != "blah" {
			t.Errorf("CountedByStatusXX expected the request body to be 'blah', got '%s'", string(bod))
		}
		r.Body.Close()
	}), "counted", nil)

	for _, given := range tests {
		statuses <- given
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "http://yup.com/foo", bytes.NewBufferString("blah"))
		counter.ServeHTTP(w, r)
		if given != w.Code {
			t.Errorf("CountedByStatusXX expected response code of %d, got %d", given, w.Code)
		}
	}

	close(statuses)

	if cnt := counter.counter1xx.Count(); cnt != 1 {
		t.Errorf("CountedByStatusXX expected 1xx counter to have a count of 1, got %d", cnt)
	}
	if cnt := counter.counter2xx.Count(); cnt != 1 {
		t.Errorf("CountedByStatusXX expected 2xx counter to have a count of 1, got %d", cnt)
	}
	if cnt := counter.counter3xx.Count(); cnt != 1 {
		t.Errorf("CountedByStatusXX expected 3xx counter to have a count of 1, got %d", cnt)
	}
	if cnt := counter.counter4xx.Count(); cnt != 1 {
		t.Errorf("CountedByStatusXX expected 4xx counter to have a count of 1, got %d", cnt)
	}
	if cnt := counter.counter5xx.Count(); cnt != 1 {
		t.Errorf("CountedByStatusXX expected 5xx counter to have a count of 1, got %d", cnt)
	}
}

func TestTimer(t *testing.T) {
	r, _ := http.NewRequest("POST", "http://uhhuh.io/", bytes.NewBufferString("yerp"))
	timer := Timed(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		if bod, _ := ioutil.ReadAll(r.Body); string(bod) != "yerp" {
			t.Errorf("Timer expected the request body to be 'yerp', got '%s'", string(bod))
		}
		r.Body.Close()
	}), "timed", nil)
	w := httptest.NewRecorder()
	timer.ServeHTTP(w, r)

	if cnt := timer.Count(); cnt != 1 {
		t.Errorf("Timer expected Count() to return 1, got %d", cnt)
	}

	if dur := timer.Max(); dur < int64(200*time.Millisecond) || dur > int64(300*time.Millisecond) {
		t.Errorf("Timer expected Max() to return between 200 and 300 ms, got %d", dur)
	}
}
