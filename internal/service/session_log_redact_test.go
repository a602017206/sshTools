package service

import (
	"strings"
	"testing"
)

func TestRedactSessionLogPassword(t *testing.T) {
	got := RedactSessionLog("export password=secret123")
	if got == "export password=secret123" || strings.Contains(got, "secret123") {
		t.Fatalf("password not redacted: %q", got)
	}
}

func TestRedactSessionLogPasswd(t *testing.T) {
	got := RedactSessionLog("export PASSWD=secret123")
	if strings.Contains(got, "secret123") {
		t.Fatalf("passwd not redacted: %q", got)
	}
}

func TestRedactSessionLogBearerAndPEM(t *testing.T) {
	pem := "-----BEGIN RSA PRIVATE KEY-----\nABC\n-----END RSA PRIVATE KEY-----"
	got := RedactSessionLog("Authorization: Bearer abcdefghijklmnop " + pem)
	if strings.Contains(got, "abcdefghijklmnop") {
		t.Fatal("bearer not redacted")
	}
	if strings.Contains(got, "BEGIN RSA PRIVATE KEY") && strings.Contains(got, "ABC") {
		t.Fatal("pem body should be redacted")
	}
}

func TestRedactSessionLogLongHex(t *testing.T) {
	hex := "deadbeefdeadbeefdeadbeefdeadbeef"
	got := RedactSessionLog("token=" + hex)
	if strings.Contains(got, hex) {
		t.Fatalf("long hex not redacted: %q", got)
	}
}

func TestRedactSessionLogLongBase64(t *testing.T) {
	b64 := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghij0123456789+/=="
	got := RedactSessionLog("data=" + b64)
	if strings.Contains(got, b64) {
		t.Fatalf("long base64 not redacted: %q", got)
	}
}
