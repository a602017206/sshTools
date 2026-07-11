package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
)

func TestAdoptiumRuntimeProviderSelectsCurrentPlatformPackage(t *testing.T) {
	expectedOS, err := adoptiumOperatingSystem(runtime.GOOS)
	if err != nil {
		t.Fatal(err)
	}
	expectedArch, err := adoptiumArchitecture(runtime.GOARCH)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v3/assets/latest/21/hotspot" {
			t.Fatalf("unexpected API path: %s", request.URL.Path)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(writer, `[
          {"binary":{"architecture":"x64","os":"other","image_type":"jre","package":{"checksum":"wrong","link":"https://example.invalid/wrong.zip","name":"wrong.zip"}},"version_data":{"semver":"21.0.1+1"}},
		  {"binary":{"architecture":%q,"os":%q,"image_type":"jre","package":{"checksum":%q,"link":"https://example.invalid/temurin.zip","name":"temurin.zip"}},"version_data":{"semver":"21.0.7+6"}}
		]`, expectedArch, expectedOS, strings.Repeat("a", 64))
	}))
	defer server.Close()
	provider := NewAdoptiumRuntimeProvider(server.Client(), server.URL)

	result, err := provider.Latest(context.Background(), 21)
	if err != nil {
		t.Fatalf("latest runtime failed: %v", err)
	}
	if result.Version != "21.0.7+6" || result.URL != "https://example.invalid/temurin.zip" || result.SHA256 != strings.Repeat("a", 64) || result.Name != "temurin.zip" {
		t.Fatalf("unexpected runtime package: %+v", result)
	}
}
