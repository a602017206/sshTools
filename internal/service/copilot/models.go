package copilot

import (
	"fmt"
	"strings"

	"AHaSSHTools/internal/config"
)

const (
	ProviderKindDeepSeek = "deepseek"
	ProviderKindOpenAI   = "openai"
	ProviderKindOllama   = "ollama"
	ProviderKindCustom   = "custom"
)

// ResolvedEndpoint is the provider + model used for one chat turn.
type ResolvedEndpoint struct {
	ProviderID     string
	ProviderKind   string
	BaseURL        string
	ModelProfileID string
	DisplayName    string
	ModelID        string
}

// ProviderAPIKeyCredentialID returns the credential store key for a provider.
func ProviderAPIKeyCredentialID(providerID string) string {
	id := strings.TrimSpace(providerID)
	if id == "" {
		return APIKeyCredentialID
	}
	return "copilot:provider:" + id + ":api_key"
}

func APIKeyOptional(kind string) bool {
	return strings.EqualFold(strings.TrimSpace(kind), ProviderKindOllama)
}

// ResolveEndpoint picks a configured model. Request profile ID wins, then the
// saved active model, then the first remaining model, then legacy fields.
func ResolveEndpoint(settings config.AppSettings, profileID string) (ResolvedEndpoint, error) {
	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
		profileID = strings.TrimSpace(settings.CopilotActiveModelID)
	}
	if ep, ok := findModel(settings, profileID); ok {
		return ep, nil
	}
	if ep, ok := firstModel(settings); ok {
		return ep, nil
	}
	base := strings.TrimSpace(settings.CopilotBaseURL)
	model := strings.TrimSpace(settings.CopilotModel)
	if base != "" && model != "" {
		return ResolvedEndpoint{
			ProviderKind: ProviderKindCustom,
			BaseURL:      base,
			ModelID:      model,
			DisplayName:  model,
		}, nil
	}
	return ResolvedEndpoint{}, fmt.Errorf("请先在设置中添加模型")
}

func ValidateEndpoint(ep ResolvedEndpoint, apiKey string) error {
	if strings.TrimSpace(ep.BaseURL) == "" || strings.TrimSpace(ep.ModelID) == "" {
		return fmt.Errorf("请先在设置中填写 Base URL 和模型")
	}
	if APIKeyOptional(ep.ProviderKind) {
		return nil
	}
	return ValidateConfig(ep.BaseURL, apiKey)
}

func findModel(settings config.AppSettings, profileID string) (ResolvedEndpoint, bool) {
	if profileID == "" {
		return ResolvedEndpoint{}, false
	}
	for _, provider := range settings.CopilotProviders {
		for _, model := range provider.Models {
			if strings.TrimSpace(model.ID) != profileID {
				continue
			}
			return ResolvedEndpoint{
				ProviderID:     strings.TrimSpace(provider.ID),
				ProviderKind:   strings.TrimSpace(provider.Kind),
				BaseURL:        strings.TrimSpace(provider.BaseURL),
				ModelProfileID: strings.TrimSpace(model.ID),
				DisplayName:    strings.TrimSpace(model.Name),
				ModelID:        strings.TrimSpace(model.ModelID),
			}, true
		}
	}
	return ResolvedEndpoint{}, false
}

func firstModel(settings config.AppSettings) (ResolvedEndpoint, bool) {
	for _, provider := range settings.CopilotProviders {
		if len(provider.Models) == 0 {
			continue
		}
		model := provider.Models[0]
		return ResolvedEndpoint{
			ProviderID:     strings.TrimSpace(provider.ID),
			ProviderKind:   strings.TrimSpace(provider.Kind),
			BaseURL:        strings.TrimSpace(provider.BaseURL),
			ModelProfileID: strings.TrimSpace(model.ID),
			DisplayName:    strings.TrimSpace(model.Name),
			ModelID:        strings.TrimSpace(model.ModelID),
		}, true
	}
	return ResolvedEndpoint{}, false
}
