package config

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"
)

// CopilotModelProfile is one selectable model under a provider.
type CopilotModelProfile struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	ModelID string `json:"model_id"`
}

// CopilotProvider is an OpenAI-compatible endpoint with one or more models.
type CopilotProvider struct {
	ID      string                `json:"id"`
	Kind    string                `json:"kind"`
	Name    string                `json:"name"`
	BaseURL string                `json:"base_url"`
	Models  []CopilotModelProfile `json:"models"`
}

// MigrateLegacyCopilotSettings turns a single Base URL + model into one custom provider.
func MigrateLegacyCopilotSettings(settings AppSettings) (AppSettings, bool) {
	if len(settings.CopilotProviders) > 0 {
		return settings, false
	}
	base := strings.TrimSpace(settings.CopilotBaseURL)
	model := strings.TrimSpace(settings.CopilotModel)
	if base == "" && model == "" {
		return settings, false
	}
	modelName := model
	if modelName == "" {
		modelName = "默认模型"
	}
	modelID := uuid.NewString()
	settings.CopilotProviders = []CopilotProvider{{
		ID:      uuid.NewString(),
		Kind:    "custom",
		Name:    "默认",
		BaseURL: base,
		Models: []CopilotModelProfile{{
			ID:      modelID,
			Name:    modelName,
			ModelID: model,
		}},
	}}
	settings.CopilotActiveModelID = modelID
	return settings, true
}

// SyncLegacyCopilotFields mirrors the active model into the old single-model fields.
func SyncLegacyCopilotFields(settings *AppSettings) {
	if settings == nil || len(settings.CopilotProviders) == 0 {
		return
	}
	activeID := strings.TrimSpace(settings.CopilotActiveModelID)
	for _, provider := range settings.CopilotProviders {
		for _, model := range provider.Models {
			if activeID != "" && strings.TrimSpace(model.ID) != activeID {
				continue
			}
			settings.CopilotBaseURL = strings.TrimSpace(provider.BaseURL)
			settings.CopilotModel = strings.TrimSpace(model.ModelID)
			if settings.CopilotActiveModelID == "" {
				settings.CopilotActiveModelID = strings.TrimSpace(model.ID)
			}
			return
		}
	}
}

// ParseCopilotProviders decodes a Wails/JSON providers payload and fills missing IDs.
func ParseCopilotProviders(raw interface{}) ([]CopilotProvider, error) {
	if raw == nil {
		return []CopilotProvider{}, nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	var providers []CopilotProvider
	if err := json.Unmarshal(data, &providers); err != nil {
		return nil, err
	}
	sanitized := make([]CopilotProvider, 0, len(providers))
	for _, provider := range providers {
		provider.ID = strings.TrimSpace(provider.ID)
		provider.Kind = strings.TrimSpace(provider.Kind)
		provider.Name = strings.TrimSpace(provider.Name)
		provider.BaseURL = strings.TrimSpace(provider.BaseURL)
		if provider.ID == "" {
			provider.ID = uuid.NewString()
		}
		if provider.Kind == "" {
			provider.Kind = "custom"
		}
		models := make([]CopilotModelProfile, 0, len(provider.Models))
		for _, model := range provider.Models {
			model.ID = strings.TrimSpace(model.ID)
			model.Name = strings.TrimSpace(model.Name)
			model.ModelID = strings.TrimSpace(model.ModelID)
			if model.ID == "" {
				model.ID = uuid.NewString()
			}
			models = append(models, model)
		}
		provider.Models = models
		sanitized = append(sanitized, provider)
	}
	return sanitized, nil
}
