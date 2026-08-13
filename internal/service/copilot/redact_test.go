package copilot

import (
	"strings"
	"testing"
)

func TestRedactPasswordAssignment(t *testing.T) {
	got := Redact("password=secret")
	if strings.Contains(got, "secret") {
		t.Fatalf("password value must be redacted, got %q", got)
	}
	if got == "password=secret" {
		t.Fatal("password=secret must be replaced")
	}
}

func TestRedactPEMPrivateKey(t *testing.T) {
	pem := "-----BEGIN RSA PRIVATE KEY-----\nMIISECRETKEYDATA\n-----END RSA PRIVATE KEY-----"
	got := Redact(pem)
	if strings.Contains(got, "MIISECRETKEYDATA") {
		t.Fatal("PEM body must be redacted")
	}
	if got == pem {
		t.Fatal("PEM private key block must be replaced")
	}
}
