package copilot

import "testing"

func TestValidateConfigRejectsEmptyAPIKey(t *testing.T) {
	err := ValidateConfig("https://api.example.com/v1", "")
	if err == nil {
		t.Fatal("expected error for empty API key")
	}
	const want = "请先在设置中填写 Base URL 和 API Key"
	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err.Error(), want)
	}
}
