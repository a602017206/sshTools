package copilot

import (
	"bytes"
	"encoding/json"
	"strings"
)

var artifactTypes = map[string]struct{}{
	"sql":             {},
	"shell":           {},
	"native_query":    {},
	"native_mutation": {},
}

// ParseArtifact extracts a typed JSON artifact from a model reply.
// It scans every '{' so a leading bare ES DSL block does not hide the trailing
// {"type":"native_query",...} payload. Content may be a string or any JSON value.
func ParseArtifact(raw string) (*Artifact, bool) {
	art, _, ok := parseArtifactAt(raw)
	return art, ok
}

// ExtractArtifact parses a typed artifact (or promotes a bare query JSON for
// native DB types) and returns a display reply with the artifact envelope removed.
func ExtractArtifact(raw, dbType string) (*Artifact, string) {
	art, end, ok := parseArtifactAt(raw)
	if ok {
		art = NormalizeArtifactForDBType(dbType, art)
		return art, stripRange(raw, end.start, end.end)
	}
	if IsNativeDBType(dbType) {
		if dsl, dslEnd, found := findBareQueryJSON(raw); found {
			return &Artifact{
				Type:        "native_query",
				Content:     dsl,
				Summary:     "可执行查询",
				Destructive: false,
			}, strings.TrimSpace(raw[:dslEnd.start] + raw[dslEnd.end:])
		}
	}
	return nil, raw
}

type span struct{ start, end int }

func parseArtifactAt(raw string) (*Artifact, span, bool) {
	for idx := 0; idx < len(raw); {
		rel := strings.Index(raw[idx:], "{")
		if rel < 0 {
			return nil, span{}, false
		}
		start := idx + rel
		dec := json.NewDecoder(strings.NewReader(raw[start:]))
		var envelope struct {
			Type        string          `json:"type"`
			Content     json.RawMessage `json:"content"`
			Summary     string          `json:"summary"`
			Destructive bool            `json:"destructive"`
		}
		if err := dec.Decode(&envelope); err != nil {
			idx = start + 1
			continue
		}
		artType := strings.TrimSpace(envelope.Type)
		if _, known := artifactTypes[artType]; !known {
			idx = start + 1
			continue
		}
		content, ok := normalizeArtifactContent(envelope.Content)
		if !ok || strings.TrimSpace(content) == "" {
			idx = start + 1
			continue
		}
		consumed := int(dec.InputOffset())
		art := &Artifact{
			Type:        artType,
			Content:     content,
			Summary:     envelope.Summary,
			Destructive: envelope.Destructive,
		}
		if art.Type == "native_mutation" {
			lower := strings.ToLower(art.Content)
			art.Destructive = strings.Contains(lower, "delete") || strings.Contains(lower, `"operation":"delete`)
		} else if art.Type == "native_query" {
			art.Destructive = false
		} else {
			art.Destructive = Classify(art.Type, art.Content).Destructive
		}
		return art, span{start: start, end: start + consumed}, true
	}
	return nil, span{}, false
}

func findBareQueryJSON(raw string) (content string, s span, ok bool) {
	for idx := 0; idx < len(raw); {
		rel := strings.Index(raw[idx:], "{")
		if rel < 0 {
			return "", span{}, false
		}
		start := idx + rel
		dec := json.NewDecoder(strings.NewReader(raw[start:]))
		var body map[string]json.RawMessage
		if err := dec.Decode(&body); err != nil {
			idx = start + 1
			continue
		}
		if _, hasQuery := body["query"]; !hasQuery {
			idx = start + 1
			continue
		}
		if _, hasType := body["type"]; hasType {
			idx = start + 1
			continue
		}
		consumed := int(dec.InputOffset())
		snippet := strings.TrimSpace(raw[start : start+consumed])
		var buf bytes.Buffer
		if err := json.Compact(&buf, []byte(snippet)); err == nil {
			snippet = buf.String()
		}
		return snippet, span{start: start, end: start + consumed}, true
	}
	return "", span{}, false
}

func stripRange(raw string, start, end int) string {
	if start < 0 || end > len(raw) || start >= end {
		return strings.TrimSpace(raw)
	}
	out := strings.TrimSpace(raw[:start] + raw[end:])
	out = strings.TrimSpace(strings.Trim(out, "`"))
	// Drop empty fences left behind.
	out = strings.ReplaceAll(out, "```json\n```", "")
	out = strings.ReplaceAll(out, "```\n```", "")
	return strings.TrimSpace(out)
}

func normalizeArtifactContent(raw json.RawMessage) (string, bool) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return "", false
	}
	if raw[0] == '"' {
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return "", false
		}
		return s, true
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		return string(raw), true
	}
	return buf.String(), true
}

// NormalizeArtifactForDBType coerces sql/shell artifacts into native_query when
// talking to Redis / Elasticsearch style sources so the UI can show 填入/执行.
func NormalizeArtifactForDBType(dbType string, art *Artifact) *Artifact {
	if art == nil || !IsNativeDBType(dbType) {
		return art
	}
	if art.Type == "sql" || art.Type == "shell" {
		clone := *art
		clone.Type = "native_query"
		clone.Destructive = false
		return &clone
	}
	return art
}
