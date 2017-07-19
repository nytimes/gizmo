package logrotate

import (
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"
)

const TestPath = "/tmp/testfile"

func TestRotate(t *testing.T) {

	file, err := NewFile(TestPath)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(TestPath)
	defer os.Remove(TestPath + "2")
	defer file.Close()

	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			fmt.Fprintln(file, "tick")
		}
	}()

	time.Sleep(time.Second / 2)

	os.Rename(TestPath, TestPath+"2")

	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Error(err)
		return
	}

	if err := proc.Signal(syscall.SIGHUP); err != nil {
		t.Error(err)
		return
	}

	time.Sleep(time.Second / 2)

}
