package service

import (
	"testing"
)

func TestParseFileURIListKeepsLocalPaths(t *testing.T) {
	got := ParseFileURIList("file:///tmp/a.txt\n# comment\nfile:///tmp/My%20Docs\nhttp://example.com/x")
	if len(got) != 2 || got[0] != "/tmp/a.txt" || got[1] != "/tmp/My Docs" {
		t.Fatalf("got %#v", got)
	}
	win := ParseFileURIList("file:///C:/Users/a.txt")
	if len(win) != 1 || win[0] != "C:/Users/a.txt" {
		t.Fatalf("windows uri = %#v", win)
	}
}

func TestNormalizeClipboardFilePathsDropsEmpty(t *testing.T) {
	got := NormalizeClipboardFilePaths([]string{" /tmp/a ", "", "/tmp/a", "/tmp/b"})
	if len(got) != 2 || got[0] != "/tmp/a" || got[1] != "/tmp/b" {
		t.Fatalf("got %#v", got)
	}
}
