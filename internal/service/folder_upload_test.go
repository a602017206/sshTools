package service

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPartitionLocalUploadItemsSeparatesDirsAndFiles(t *testing.T) {
	dirs, files := partitionLocalUploadItems([]LocalUploadItem{
		{LocalPath: "/a", RelPath: "a", IsDir: true},
		{LocalPath: "/a/f.txt", RelPath: "a/f.txt", IsDir: false},
		{LocalPath: "/a/b", RelPath: "a/b", IsDir: true},
		{LocalPath: "/a/b/g.txt", RelPath: "a/b/g.txt", IsDir: false},
	})
	if len(dirs) != 2 || dirs[0].RelPath != "a" || dirs[1].RelPath != "a/b" {
		t.Fatalf("dirs = %#v", dirs)
	}
	if len(files) != 2 || files[0].RelPath != "a/f.txt" || files[1].RelPath != "a/b/g.txt" {
		t.Fatalf("files = %#v", files)
	}
}

func TestRunFolderUploadIsSerialAndMakesDirsFirst(t *testing.T) {
	var mu sync.Mutex
	var current, max int32
	var log []string
	record := func(op string) {
		mu.Lock()
		log = append(log, op)
		mu.Unlock()
	}

	err := runFolderUpload(folderUploadRunner{
		mkdir: func(remotePath string) error {
			n := atomic.AddInt32(&current, 1)
			if n > atomic.LoadInt32(&max) {
				atomic.StoreInt32(&max, n)
			}
			record("mkdir:" + remotePath)
			time.Sleep(5 * time.Millisecond)
			atomic.AddInt32(&current, -1)
			return nil
		},
		upload: func(localPath, remotePath string) error {
			n := atomic.AddInt32(&current, 1)
			if n > atomic.LoadInt32(&max) {
				atomic.StoreInt32(&max, n)
			}
			record("upload:" + remotePath)
			time.Sleep(5 * time.Millisecond)
			atomic.AddInt32(&current, -1)
			return nil
		},
		cancelled: func() bool { return false },
	}, "/home", []LocalUploadItem{
		{LocalPath: "/tmp/a", RelPath: "a", IsDir: true},
		{LocalPath: "/tmp/a/one.txt", RelPath: "a/one.txt"},
		{LocalPath: "/tmp/a/two.txt", RelPath: "a/two.txt"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt32(&max) > 1 {
		t.Fatalf("folder upload ran %d operations at once; sshd/SFTP cannot take unbounded fan-out", max)
	}
	mu.Lock()
	defer mu.Unlock()
	want := []string{"mkdir:/home/a", "upload:/home/a/one.txt", "upload:/home/a/two.txt"}
	if len(log) != len(want) {
		t.Fatalf("log = %#v", log)
	}
	for i := range want {
		if log[i] != want[i] {
			t.Fatalf("step %d: got %s want %s", i, log[i], want[i])
		}
	}
}

func TestRunFolderUploadStopsAfterCancel(t *testing.T) {
	uploads := 0
	err := runFolderUpload(folderUploadRunner{
		mkdir: func(string) error { return nil },
		upload: func(string, string) error {
			uploads++
			return nil
		},
		cancelled: func() bool { return uploads >= 1 },
	}, "/r", []LocalUploadItem{
		{LocalPath: "/f1", RelPath: "f1"},
		{LocalPath: "/f2", RelPath: "f2"},
		{LocalPath: "/f3", RelPath: "f3"},
	})
	if !errors.Is(err, errFolderUploadCancelled) {
		t.Fatalf("err = %v", err)
	}
	if uploads != 1 {
		t.Fatalf("uploads = %d, want 1 after cancel", uploads)
	}
}
