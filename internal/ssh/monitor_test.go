package ssh

import "testing"

func TestParsePerCoreUsage(t *testing.T) {
	before := "cpu  100 0 100 800 0 0 0 0 0 0\ncpu0 50 0 50 400 0 0 0 0 0 0\ncpu1 50 0 50 400 0 0 0 0 0 0\n"
	after := "cpu  200 0 200 1600 0 0 0 0 0 0\ncpu0 100 0 100 800 0 0 0 0 0 0\ncpu1 125 0 125 500 0 0 0 0 0 0\n"
	output := before + "__SPLIT__\n" + after

	usage := parsePerCoreUsage(output)
	if len(usage) != 2 {
		t.Fatalf("expected 2 cores, got %v", usage)
	}
	// cpu0: total 500->1000 (+500), idle 400->800 (+400) → busy 20%
	if usage[0] != 20 {
		t.Errorf("cpu0 busy = %v, want 20", usage[0])
	}
	// cpu1: total 500->750 (+250), idle 400->500 (+100) → busy 60%
	if usage[1] != 60 {
		t.Errorf("cpu1 busy = %v, want 60", usage[1])
	}
}

func TestParsePerCoreUsageBadInput(t *testing.T) {
	if got := parsePerCoreUsage("no split marker"); got != nil {
		t.Errorf("expected nil for missing split, got %v", got)
	}
	if got := parsePerCoreUsage("__SPLIT__\n"); got != nil {
		t.Errorf("expected nil for empty snapshots, got %v", got)
	}
}
