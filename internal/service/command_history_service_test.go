package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCommandHistoryRecordAndSuggest(t *testing.T) {
	svc := NewCommandHistoryService(t.TempDir())
	_ = svc.Record("c1", "cd /var/log")
	_ = svc.Record("c1", "cd /var/log")
	_ = svc.Record("c1", "cd /tmp")
	got, err := svc.Suggest("c1", "cd", 8)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) < 2 || got[0].Command != "cd /var/log" || got[0].Count != 2 {
		t.Fatalf("unexpected suggest: %+v", got)
	}
	if err := svc.Record("c1", "   "); err != nil {
		t.Fatal(err)
	}
}

func TestCommandHistorySuggestPrefixBeforeContains(t *testing.T) {
	svc := NewCommandHistoryService(t.TempDir())
	_ = svc.Record("c1", "grep error")
	_ = svc.Record("c1", "tail -f error.log")
	_ = svc.Record("c1", "tail -f error.log")

	got, err := svc.Suggest("c1", "tail", 8)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Command != "tail -f error.log" {
		t.Fatalf("expected prefix match only, got %+v", got)
	}

	got, err = svc.Suggest("c1", "error", 8)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected contains matches, got %+v", got)
	}
	if got[0].Command != "tail -f error.log" {
		t.Fatalf("expected count desc sort, got %+v", got)
	}
}

func TestCommandHistoryMaxEntries(t *testing.T) {
	root := t.TempDir()
	svc := NewCommandHistoryService(root)

	for i := 0; i < 510; i++ {
		if err := svc.Record("c1", fmt.Sprintf("cmd-%d", i)); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond)
	}

	data, err := os.ReadFile(filepath.Join(root, "c1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var stored commandHistoryFile
	if err := json.Unmarshal(data, &stored); err != nil {
		t.Fatal(err)
	}
	if len(stored.Entries) > 500 {
		t.Fatalf("expected at most 500 entries, got %d", len(stored.Entries))
	}
}

func TestCommandHistoryEmptySuggest(t *testing.T) {
	svc := NewCommandHistoryService(t.TempDir())
	got, err := svc.Suggest("missing", "cd", 8)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("expected nil for missing connection, got %+v", got)
	}
}
