package copilot

import (
	"encoding/json"
	"strings"
)

// ParseArtifact extracts the first JSON artifact from a model reply.
// Markdown fenced JSON is accepted. Classify always overrides Destructive.
func ParseArtifact(raw string) (*Artifact, bool) {
	idx := strings.Index(raw, "{")
	if idx < 0 {
		return nil, false
	}
	dec := json.NewDecoder(strings.NewReader(raw[idx:]))
	var art Artifact
	if err := dec.Decode(&art); err != nil {
		return nil, false
	}
	if art.Type != "sql" && art.Type != "shell" {
		return nil, false
	}
	if strings.TrimSpace(art.Content) == "" {
		return nil, false
	}
	art.Destructive = Classify(art.Type, art.Content).Destructive
	return &art, true
}
