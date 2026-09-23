package copilot

import (
	"testing"

	"AHaSSHTools/internal/config"
)

func TestProviderAPIKeyCredentialID(t *testing.T) {
	if got := ProviderAPIKeyCredentialID("abc"); got != "copilot:provider:abc:api_key" {
		t.Fatalf("got %q", got)
	}
	if got := ProviderAPIKeyCredentialID("  "); got != APIKeyCredentialID {
		t.Fatalf("empty provider should fall back to legacy key id, got %q", got)
	}
}

func TestResolveEndpointUsesProfileThenActiveThenFirst(t *testing.T) {
	settings := config.AppSettings{
		CopilotActiveModelID: "m2",
		CopilotProviders: []config.CopilotProvider{
			{
				ID:      "p1",
				Kind:    "deepseek",
				Name:    "DeepSeek",
				BaseURL: "https://api.deepseek.com/v1",
				Models: []config.CopilotModelProfile{
					{ID: "m1", Name: "Chat", ModelID: "deepseek-chat"},
					{ID: "m2", Name: "Reasoner", ModelID: "deepseek-reasoner"},
				},
			},
		},
	}

	got, err := ResolveEndpoint(settings, "m1")
	if err != nil {
		t.Fatal(err)
	}
	if got.ModelID != "deepseek-chat" || got.ProviderID != "p1" || got.ModelProfileID != "m1" {
		t.Fatalf("unexpected request override: %+v", got)
	}

	got, err = ResolveEndpoint(settings, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.ModelID != "deepseek-reasoner" {
		t.Fatalf("expected active model, got %+v", got)
	}

	got, err = ResolveEndpoint(settings, "missing")
	if err != nil {
		t.Fatal(err)
	}
	if got.ModelID != "deepseek-chat" {
		t.Fatalf("expected first model fallback, got %+v", got)
	}
}

func TestResolveEndpointFallsBackToLegacyFields(t *testing.T) {
	settings := config.AppSettings{
		CopilotBaseURL: "https://api.openai.com/v1",
		CopilotModel:   "gpt-4o-mini",
	}
	got, err := ResolveEndpoint(settings, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.BaseURL != "https://api.openai.com/v1" || got.ModelID != "gpt-4o-mini" {
		t.Fatalf("unexpected legacy resolve: %+v", got)
	}
}

func TestResolveEndpointRequiresConfiguration(t *testing.T) {
	_, err := ResolveEndpoint(config.AppSettings{}, "")
	if err == nil {
		t.Fatal("expected error when no models are configured")
	}
}

func TestValidateEndpointAllowsOllamaWithoutKey(t *testing.T) {
	ep := ResolvedEndpoint{ProviderKind: ProviderKindOllama, BaseURL: "http://127.0.0.1:11434/v1", ModelID: "llama3.1"}
	if err := ValidateEndpoint(ep, ""); err != nil {
		t.Fatalf("ollama should allow empty key: %v", err)
	}
}

func TestValidateEndpointRejectsEmptyKeyForCloudProviders(t *testing.T) {
	ep := ResolvedEndpoint{ProviderKind: ProviderKindDeepSeek, BaseURL: "https://api.deepseek.com/v1", ModelID: "deepseek-chat"}
	if err := ValidateEndpoint(ep, ""); err == nil {
		t.Fatal("expected error for empty cloud API key")
	}
}
