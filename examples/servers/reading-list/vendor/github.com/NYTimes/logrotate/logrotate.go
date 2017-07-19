package logrotate

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// File wraps an *os.File and listens for a 'SIGHUP' signal from logrotated
// so it can reopen the new file.
type File struct {
	*os.File
	me     sync.Mutex
	path   string
	sighup chan os.Signal
}

// NewFile creates a File pointer and kicks off the goroutine listening for
// SIGHUP signals.
func NewFile(path string) (*File, error) {

	lr := &File{
		me:     sync.Mutex{},
		path:   path,
		sighup: make(chan os.Signal, 1),
	}

	if err := lr.reopen(); err != nil {
		return nil, err
	}

	go func() {
		signal.Notify(lr.sighup, syscall.SIGHUP)

		for _ = range lr.sighup {
			fmt.Fprintf(os.Stderr, "%s: Reopening %q\n", time.Now(), lr.path)
			if err := lr.reopen(); err != nil {
				fmt.Fprintf(os.Stderr, "%s: Error reopening: %s\n", time.Now(), err)
			}
		}
	}()

	return lr, nil

}

func (lr *File) reopen() (err error) {
	lr.me.Lock()
	defer lr.me.Unlock()
	lr.File.Close()
	lr.File, err = os.OpenFile(lr.path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	return
}

// Write will write to the underlying file. It uses a sync.Mutex to ensure
// uninterrupted writes during logrotates.
func (lr *File) Write(b []byte) (int, error) {
	lr.me.Lock()
	defer lr.me.Unlock()
	return lr.File.Write(b)
}

// Close will stop the goroutine listening for SIGHUP signals and then close
// the underlying os.File.
func (lr *File) Close() error {
	lr.me.Lock()
	defer lr.me.Unlock()
	signal.Stop(lr.sighup)
	close(lr.sighup)
	return lr.File.Close()
}
