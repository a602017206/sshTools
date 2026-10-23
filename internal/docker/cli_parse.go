package docker

import (
	"encoding/json"
	"strings"
	"time"
)

const dockerPSCreatedAtLayout = "2006-01-02 15:04:05 -0700 MST"

type dockerPSLine struct {
	ID        string `json:"ID"`
	Id        string `json:"Id"`
	Names     string `json:"Names"`
	Image     string `json:"Image"`
	Status    string `json:"Status"`
	Ports     string `json:"Ports"`
	CreatedAt string `json:"CreatedAt"`
}

// ParseDockerPSJSONLines parses newline-delimited JSON from docker ps --format.
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
			ID:      id,
			Name:    raw.Names,
			Image:   raw.Image,
			Status:  raw.Status,
			Ports:   raw.Ports,
			Created: created,
		})
	}
	return out, nil
}
