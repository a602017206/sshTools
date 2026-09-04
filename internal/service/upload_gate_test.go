package service

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSessionUploadLockAllowsOnlyOneAtATime(t *testing.T) {
	s := NewSFTPService(nil, nil)
	var current, max atomic.Int32
	var wg sync.WaitGroup
	run := func() {
		defer wg.Done()
		unlock := s.lockSessionUpload("sess-1")
		defer unlock()
		n := current.Add(1)
		if n > max.Load() {
			max.Store(n)
		}
		time.Sleep(15 * time.Millisecond)
		current.Add(-1)
	}
	wg.Add(3)
	go run()
	go run()
	go run()
	wg.Wait()
	if max.Load() != 1 {
		t.Fatalf("max concurrent uploads = %d, want 1", max.Load())
	}
}
