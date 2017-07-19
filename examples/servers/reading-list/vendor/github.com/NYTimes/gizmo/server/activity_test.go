package server

import "testing"

func TestActivityMonitorRequestCount(t *testing.T) {
	a := NewActivityMonitor()
	// test the active request count
	a.CountRequest() // 1
	a.CountRequest() // 2
	a.CountRequest() // 3

	if !a.Active() {
		t.Error("ActivityMonitor is inactive when there should be 3 active requests")
	}

	if active := a.NumActiveRequests(); active != 3 {
		t.Errorf("ActivityMonitor expected 3 active request, got %d", active)
	}

	a.UncountRequest() // 2

	if active := a.NumActiveRequests(); active != 2 {
		t.Errorf("ActivityMonitor expected 2 active request, got %d", active)
	}

	if !a.Active() {
		t.Error("ActivityMonitor is inactive when there should be 2 active requests")
	}

	a.UncountRequest() // 1

	if active := a.NumActiveRequests(); active != 1 {
		t.Errorf("ActivityMonitor expected 1 active request, got %d", active)
	}

	if !a.Active() {
		t.Error("ActivityMonitor is inactive when there should be 1 active request")
	}

	a.UncountRequest() // 0

	if a.Active() {
		t.Error("ActivityMonitor is active when there should be 0 active request")
	}

}
