package server

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ESXShutdownTimeout is the hard cut off kill the server while the ESXHealthCheck is waiting
	// for the server to be inactive.
	ESXShutdownTimeout = 180 * time.Second
	// ESXShutdownPollInterval sets the duration for how long ESXHealthCheck will wait between
	// each NumActiveRequests poll in WaitForZero.
	ESXShutdownPollInterval = 1 * time.Second
	// ESXLoadBalancerNotReadyDuration is the amount of time ESXHealthCheck will wait after
	// sending a 'bad' status to the LB during a graceful shutdown.
	ESXLoadBalancerNotReadyDuration = 15 * time.Second
)

// ESXHealthCheck will manage the health checks and manage
// a server's load balanacer status. On Stop, it will block
// until all LBs have received a 'bad' status.
type ESXHealthCheck struct {
	// ready flag health checks and graceful shutdown
	// uint32 so we can use sync/atomic and no defers
	ready uint32

	// last LB status to know if LB knows we're inactive
	lbNotReadyTime   map[string]*time.Time
	lbNotReadyTimeMu sync.RWMutex

	monitor *ActivityMonitor
}

// NewESXHealthCheck returns a new instance of ESXHealthCheck.
func NewESXHealthCheck() *ESXHealthCheck {
	return &ESXHealthCheck{
		lbNotReadyTime: map[string]*time.Time{},
	}
}

// Path returns the default ESX health path.
func (e *ESXHealthCheck) Path() string {
	return "/status.txt"
}

// Start will set the monitor and flip the ready flag to 'True'.
func (e *ESXHealthCheck) Start(monitor *ActivityMonitor) error {
	e.monitor = monitor
	atomic.StoreUint32(&e.ready, 1)
	return nil
}

// Stop will set the flip the 'ready' flag and wait block until the server has removed itself
// from all load balancers.
func (e *ESXHealthCheck) Stop() error {
	// fill the flag and wait
	atomic.StoreUint32(&e.ready, 0)
	if err := e.waitForZero(); err != nil {
		Log.Errorf("server still active after %s, this will not be a graceful shutdown: %s", ESXShutdownTimeout, err)
		return err
	}
	return nil
}

// ServeHTTP will handle the health check requests on the server. ESXHealthCheck
// will return with an "ok" status as long as the ready flag is set to True.
// If a `deployer` query parameter is included, the request will not be counted
// as a load balancer.
func (e *ESXHealthCheck) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadUint32(&e.ready) == 1 {
		if _, err := io.WriteString(w, "ok-"+Name); err != nil {
			LogWithFields(r).Warn("unable to write healthcheck response: ", err)
		}
		e.updateReadyTime(r, true)
	} else {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		e.updateReadyTime(r, false)
	}
}

func (e *ESXHealthCheck) updateReadyTime(r *http.Request, ready bool) {
	ip, err := GetIP(r)
	if err != nil {
		Log.Warnf("status endpoint was unable to get LB IP addr: %s", err)
		return
	}

	if _, ok := r.URL.Query()["deployer"]; !ok {
		e.lbNotReadyTimeMu.Lock()
		if ready {
			e.lbNotReadyTime[ip] = nil
		} else {
			if last := e.lbNotReadyTime[ip]; last == nil {
				now := time.Now()
				e.lbNotReadyTime[ip] = &now
				Log.Infof("status endpoint has returned it's first not-ready status to: %s", ip)
			}
		}
		e.lbNotReadyTimeMu.Unlock()
	}
}

func (e *ESXHealthCheck) lbActive() (active bool) {
	e.lbNotReadyTimeMu.RLock()
	defer e.lbNotReadyTimeMu.RUnlock()
	for ip, notReadyTime := range e.lbNotReadyTime {
		if notReadyTime == nil || time.Since(*notReadyTime) < ESXLoadBalancerNotReadyDuration {
			Log.Infof("load balancer is still active: %s", ip)
			return true
		}
	}
	return false
}

// waitForZero will continously query Active and NumActiveRequests at the ShutdownPollInterval until the
// LB has seen a bad status, the server is not Actve and NumActiveRequests returns 0 or the timeout
// is reached. It will return error in case of timeout.
func (e *ESXHealthCheck) waitForZero() error {
	to := time.After(ESXShutdownTimeout)
	done := make(chan struct{}, 1)
	go func() {
		for {
			if e.monitor.Active() || e.lbActive() {
				Log.Info("server is still active")
			} else {
				Log.Info("server is no longer active")
				reqs := e.monitor.NumActiveRequests()
				if reqs == 0 {
					done <- struct{}{}
					break
				} else {
					Log.Info("server still has requests in flight. waiting longer...")
				}
			}
			time.Sleep(ESXShutdownPollInterval)
		}
	}()

	select {
	case <-done:
		Log.Info("server is no longer receiving traffic")
		return nil
	case <-to:
		return fmt.Errorf("server is still active after %s", ESXShutdownTimeout)
	}
}
