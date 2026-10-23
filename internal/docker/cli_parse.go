package docker

import (
	"encoding/json"
	"strings"
	"time"
)

const dockerPSCreatedAtLayout = "2006-01-02 15:04:05 -0700 MST"

type dockerPSLine struct {
	ID        string          `json:"ID"`
	Id        string          `json:"Id"`
	Names     json.RawMessage `json:"Names"`
	Image     string          `json:"Image"`
	Status    string          `json:"Status"`
	Ports     json.RawMessage `json:"Ports"`
	CreatedAt string          `json:"CreatedAt"`
}

// ParseDockerPSJSONLines parses newline-delimited JSON from docker ps --format.
// IDs are shortened to 12 chars and names lose the leading "/" so the result
// matches what Client.ListContainers returns for the Engine API.
func ParseDockerPSJSONLines(stdout string) ([]ContainerSummary, error) {
	lines := strings.Split(stdout, "\n")
	out := make([]ContainerSummary, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var raw dockerPSLine
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			return nil, err
		}
		id := raw.ID
		if id == "" {
			id = raw.Id
		}
		created := int64(0)
		if raw.CreatedAt != "" {
			if ts, err := time.Parse(dockerPSCreatedAtLayout, raw.CreatedAt); err == nil {
				created = ts.Unix()
			}
		}
		out = append(out, ContainerSummary{
			ID:      shortContainerID(id),
			Name:    firstContainerName(decodeStringOrList(raw.Names)),
			Image:   raw.Image,
			Status:  raw.Status,
			Ports:   strings.Join(decodeStringOrList(raw.Ports), ", "),
			Created: created,
		})
	}
	return out, nil
}

// decodeStringOrList accepts either a JSON string or an array of strings.
func decodeStringOrList(raw json.RawMessage) []string {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if s == "" {
			return nil
		}
		return []string{s}
	}
	var list []string
	if err := json.Unmarshal(raw, &list); err == nil {
		return list
	}
	return nil
}

// firstContainerName picks the first name from a (possibly comma separated) list.
func firstContainerName(values []string) string {
	if len(values) == 0 {
		return ""
	}
	first := values[0]
	if i := strings.Index(first, ","); i >= 0 {
		first = first[:i]
	}
	return containerName([]string{strings.TrimSpace(first)})
}
