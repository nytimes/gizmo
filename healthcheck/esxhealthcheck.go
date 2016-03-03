package healthcheck

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
)

var (
	// ESXShutdownTimeout is the hard cut off kill the server while the ESX is waiting
	// for the server to be inactive.
	ESXShutdownTimeout = 180 * time.Second
	// ESXShutdownPollInterval sets the duration for how long ESX will wait between
	// each NumActiveRequests poll in WaitForZero.
	ESXShutdownPollInterval = 1 * time.Second
	// ESXLoadBalancerNotReadyDuration is the amount of time ESX will wait after
	// sending a 'bad' status to the LB during a graceful shutdown.
	ESXLoadBalancerNotReadyDuration = 15 * time.Second
)

// ESX will manage the health checks and manage
// a server's load balanacer status. On Stop, it will block
// until all LBs have received a 'bad' status.
type ESX struct {
	// ready flag health checks and graceful shutdown
	// uint32 so we can use sync/atomic and no defers
	ready uint32

	// last LB status to know if LB knows we're inactive
	lbNotReadyTime   map[string]*time.Time
	lbNotReadyTimeMu sync.RWMutex

	monitor *ActivityMonitor
}

// NewESX returns a new instance of ESX.
func NewESX() *ESX {
	return &ESX{
		lbNotReadyTime: map[string]*time.Time{},
	}
}

// Path returns the default ESX health path.
func (e *ESX) Path() string {
	return "/status.txt"
}

// Start will set the monitor and flip the ready flag to 'True'.
func (e *ESX) Start(monitor *ActivityMonitor) error {
	e.monitor = monitor
	atomic.StoreUint32(&e.ready, 1)
	return nil
}

// Stop will set the flip the 'ready' flag and wait block until the server has removed itself
// from all load balancers.
func (e *ESX) Stop() error {
	// fill the flag and wait
	atomic.StoreUint32(&e.ready, 0)
	if err := e.waitForZero(); err != nil {
		return err
	}
	return nil
}

// ServeHTTP will handle the health check requests on the server. ESX
// will return with an "ok" status as long as the ready flag is set to True.
// If a `deployer` query parameter is included, the request will not be counted
// as a load balancer.
func (e *ESX) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadUint32(&e.ready) == 1 {
		io.WriteString(w, "ok")
		e.updateReadyTime(r, true)
	} else {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		e.updateReadyTime(r, false)
	}
}

// GetIP returns the IP address for the given request.
func getIP(r *http.Request) (string, error) {
	ip, ok := mux.Vars(r)["ip"]
	if ok {
		return ip, nil
	}

	// check real ip header first
	ip = r.Header.Get("X-Real-IP")
	if len(ip) > 0 {
		return ip, nil
	}

	// no nginx reverse proxy?
	// get IP old fashioned way
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", fmt.Errorf("%q is not IP:port", r.RemoteAddr)
	}

	userIP := net.ParseIP(ip)
	if userIP == nil {
		return "", fmt.Errorf("%q is not IP:port", r.RemoteAddr)
	}
	return userIP.String(), nil
}

func (e *ESX) updateReadyTime(r *http.Request, ready bool) {
	ip, err := getIP(r)
	if err != nil {
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
			}
		}
		e.lbNotReadyTimeMu.Unlock()
	}
}

func (e *ESX) lbActive() (active bool) {
	e.lbNotReadyTimeMu.RLock()
	defer e.lbNotReadyTimeMu.RUnlock()
	for _, notReadyTime := range e.lbNotReadyTime {
		if notReadyTime == nil || time.Since(*notReadyTime) < ESXLoadBalancerNotReadyDuration {
			return true
		}
	}
	return false
}

// waitForZero will continously query Active and NumActiveRequests at the ShutdownPollInterval until the
// LB has seen a bad status, the server is not Actve and NumActiveRequests returns 0 or the timeout
// is reached. It will return error in case of timeout.
func (e *ESX) waitForZero() error {
	to := time.After(ESXShutdownTimeout)
	done := make(chan struct{}, 1)
	go func() {
		for {
			if e.monitor.Active() || e.lbActive() {
			} else {
				reqs := e.monitor.NumActiveRequests()
				if reqs == 0 {
					done <- struct{}{}
					break
				} else {
				}
			}
			time.Sleep(ESXShutdownPollInterval)
		}
	}()

	select {
	case <-done:
		return nil
	case <-to:
		return fmt.Errorf("server is still active after %s", ESXShutdownTimeout)
	}
}
