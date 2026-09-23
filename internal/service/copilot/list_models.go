package copilot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// RemoteModel is one entry from an OpenAI-compatible GET /v1/models response.
type RemoteModel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func openAIBase(baseURL string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	base = strings.TrimSuffix(base, "/chat/completions")
	base = strings.TrimRight(base, "/")
	if base == "" {
		return ""
	}
	if !strings.HasSuffix(base, "/v1") {
		base += "/v1"
	}
	return base
}

func modelsURL(baseURL string) string {
	return openAIBase(baseURL) + "/models"
}

// ResolveListModelsAPIKey prefers the unsaved draft key, then the stored key.
// Ollama may list models without a key; other kinds require one.
func ResolveListModelsAPIKey(draftKey, storedKey, kind string) (string, error) {
	if key := strings.TrimSpace(draftKey); key != "" {
		return key, nil
	}
	if key := strings.TrimSpace(storedKey); key != "" {
		return key, nil
	}
	if APIKeyOptional(kind) {
		return "", nil
	}
	return "", fmt.Errorf("请先填写 API Key")
}

// ListModels calls GET {base}/v1/models. Errors never include the API key.
func ListModels(ctx context.Context, baseURL, apiKey string, client *http.Client) ([]RemoteModel, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, fmt.Errorf("请先填写 Base URL")
	}
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, modelsURL(baseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("openai models: create request: %w", err)
	}
	if strings.TrimSpace(apiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai models: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxChatResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("openai models: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, statusError("openai models", resp.StatusCode, body, apiKey)
	}

	var parsed openaiModelsResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("openai models: decode response: %w", err)
	}

	seen := make(map[string]bool, len(parsed.Data))
	out := make([]RemoteModel, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		id := strings.TrimSpace(item.ID)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, RemoteModel{ID: id, Name: id})
	}
	return out, nil
}

type openaiModelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}
