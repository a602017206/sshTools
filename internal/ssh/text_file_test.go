package ssh

import (
	"bytes"
	"strings"
	"testing"
)

func TestValidateRemoteText(t *testing.T) {
	got, err := ValidateRemoteText([]byte("key: value\n"))
	if err != nil {
		t.Fatalf("合法文本被拒绝: %v", err)
	}
	if got != "key: value\n" {
		t.Fatalf("内容被改写: %q", got)
	}

	got, err = ValidateRemoteText(nil)
	if err != nil || got != "" {
		t.Fatalf("空文件应允许编辑, got=%q err=%v", got, err)
	}

	_, err = ValidateRemoteText([]byte{'a', 0, 'b'})
	if err == nil || !strings.Contains(err.Error(), "二进制") {
		t.Fatalf("二进制文件应拒绝, err=%v", err)
	}

	_, err = ValidateRemoteText([]byte{0xff, 0xfe, 0xfd})
	if err == nil || !strings.Contains(err.Error(), "UTF-8") {
		t.Fatalf("非 UTF-8 应拒绝, err=%v", err)
	}

	_, err = ValidateRemoteText(bytes.Repeat([]byte("a"), int(MaxRemoteTextBytes+1)))
	if err == nil || !strings.Contains(err.Error(), "1MB") {
		t.Fatalf("超限文件应拒绝, err=%v", err)
	}

	_, err = ValidateRemoteText(bytes.Repeat([]byte("a"), int(MaxRemoteTextBytes)))
	if err != nil {
		t.Fatalf("刚好 1MB 应允许: %v", err)
	}
}
