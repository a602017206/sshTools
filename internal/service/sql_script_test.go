package service

import (
	"strings"
	"testing"
)

func TestSplitSQLScriptSplitsOnSemicolonOutsideStrings(t *testing.T) {
	script := "INSERT INTO t VALUES ('a;b');\nINSERT INTO t VALUES (1);"
	got, err := SplitSQLScript(strings.NewReader(script), "mysql")
	if err != nil {
		t.Fatalf("SplitSQLScript: %v", err)
	}
	want := []string{
		"INSERT INTO t VALUES ('a;b')",
		"INSERT INTO t VALUES (1)",
	}
	if len(got) != len(want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("stmt %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSplitSQLScriptIgnoresCommentsAndBlankStatements(t *testing.T) {
	script := "-- header\nCREATE TABLE t (id int);\n/* block */\n;\nSELECT 1;"
	got, err := SplitSQLScript(strings.NewReader(script), "mysql")
	if err != nil {
		t.Fatalf("SplitSQLScript: %v", err)
	}
	if len(got) != 2 || got[0] != "CREATE TABLE t (id int)" || got[1] != "SELECT 1" {
		t.Fatalf("got %#v", got)
	}
}

func TestSplitSQLScriptOracleSlashTerminator(t *testing.T) {
	script := "CREATE TABLE t (id number)\n/\nINSERT INTO t VALUES (1)\n/\n"
	got, err := SplitSQLScript(strings.NewReader(script), "oracle")
	if err != nil {
		t.Fatalf("SplitSQLScript: %v", err)
	}
	if len(got) != 2 || !strings.Contains(got[0], "CREATE TABLE") || !strings.Contains(got[1], "INSERT") {
		t.Fatalf("got %#v", got)
	}
}

func TestSQLFilePreamble(t *testing.T) {
	if got := SQLFilePreamble("mysql", "shop", ""); !strings.Contains(strings.ToLower(got), "use") || !strings.Contains(got, "shop") {
		t.Fatalf("mysql preamble = %q", got)
	}
	if got := SQLFilePreamble("oracle", "ORCL", "PEMS"); !strings.Contains(got, "CURRENT_SCHEMA") || !strings.Contains(got, "PEMS") {
		t.Fatalf("oracle preamble = %q", got)
	}
	if got := SQLFilePreamble("postgresql", "app", "sales"); !strings.Contains(strings.ToLower(got), "search_path") || !strings.Contains(got, "sales") {
		t.Fatalf("postgres preamble = %q", got)
	}
	if got := SQLFilePreamble("oracle", "ORCL", ""); got != "" {
		t.Fatalf("empty schema should skip preamble, got %q", got)
	}
}
