package docker

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const viaCLI = "cli"

func BuildDockerImagesCommand() string {
	return "docker images --format '{{json .}}'"
}

func BuildDockerNetworksCommand() string {
	return "docker network ls --format '{{json .}}'"
}

func BuildDockerVolumesCommand() string {
	return "docker volume ls --format '{{json .}}'"
}

func BuildDockerRmiCommand(ref string, force bool) string {
	if force {
		return "docker rmi -f " + shellQuote(ref)
	}
	return "docker rmi " + shellQuote(ref)
}

func BuildDockerNetworkRmCommand(id string) string {
	return "docker network rm " + shellQuote(id)
}

func BuildDockerVolumeRmCommand(name string, force bool) string {
	if force {
		return "docker volume rm -f " + shellQuote(name)
	}
	return "docker volume rm " + shellQuote(name)
}

func BuildDockerImageInspectCommand(ref string) string {
	return "docker image inspect " + shellQuote(ref)
}

func BuildDockerNetworkInspectCommand(id string) string {
	return "docker network inspect " + shellQuote(id)
}

func BuildDockerVolumeInspectCommand(name string) string {
	return "docker volume inspect " + shellQuote(name)
}

// jsonObjectLines returns trimmed lines that look like JSON objects; banners,
// warnings and other noise are dropped.
func jsonObjectLines(stdout string) []string {
	var out []string
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "{") {
			out = append(out, line)
		}
	}
	return out
}

type dockerImageLine struct {
	ID         string `json:"ID"`
	Repository string `json:"Repository"`
	Tag        string `json:"Tag"`
	CreatedAt  string `json:"CreatedAt"`
	Size       string `json:"Size"`
}

// ParseDockerImagesJSONLines parses docker images --format '{{json .}}' output.
func ParseDockerImagesJSONLines(stdout string) ([]ImageSummary, error) {
	lines := jsonObjectLines(stdout)
	out := make([]ImageSummary, 0, len(lines))
	indexByID := make(map[string]int, len(lines))
	for _, line := range lines {
		var raw dockerImageLine
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			return nil, fmt.Errorf("docker: parse images: %w", err)
		}
		tags := []string{}
		if raw.Repository != "" && raw.Repository != "<none>" {
			tag := raw.Tag
			if tag == "" || tag == "<none>" {
				tags = append(tags, raw.Repository)
			} else {
				tags = append(tags, raw.Repository+":"+tag)
			}
		}
		var created int64
		if raw.CreatedAt != "" {
			if ts, err := time.Parse(dockerPSCreatedAtLayout, raw.CreatedAt); err == nil {
				created = ts.Unix()
			}
		}
		// docker images prints one row per tag; merge rows sharing an image ID.
		if idx, ok := indexByID[raw.ID]; ok && raw.ID != "" {
			for _, t := range tags {
				if !containsString(out[idx].Tags, t) {
					out[idx].Tags = append(out[idx].Tags, t)
				}
			}
			continue
		}
		if raw.ID != "" {
			indexByID[raw.ID] = len(out)
		}
		out = append(out, ImageSummary{
			ID:      raw.ID,
			Tags:    tags,
			Size:    parseHumanSize(raw.Size),
			Created: created,
			Via:     viaCLI,
		})
	}
	return out, nil
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// parseHumanSize roughly converts docker's human readable sizes (e.g. "187MB",
// "1.2GB", "512kB") to bytes. Unparsable input yields 0.
func parseHumanSize(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	i := 0
	for i < len(s) && (s[i] == '.' || (s[i] >= '0' && s[i] <= '9')) {
		i++
	}
	if i == 0 {
		return 0
	}
	num, err := strconv.ParseFloat(s[:i], 64)
	if err != nil {
		return 0
	}
	var mult float64
	switch strings.ToUpper(strings.TrimSpace(s[i:])) {
	case "B":
		mult = 1
	case "KB", "KIB":
		mult = 1000
	case "MB", "MIB":
		mult = 1000 * 1000
	case "GB", "GIB":
		mult = 1000 * 1000 * 1000
	case "TB", "TIB":
		mult = 1000 * 1000 * 1000 * 1000
	default:
		return 0
	}
	return int64(num * mult)
}

type dockerNetworkLine struct {
	ID     string `json:"ID"`
	Name   string `json:"Name"`
	Driver string `json:"Driver"`
	Scope  string `json:"Scope"`
}

// ParseDockerNetworksJSONLines parses docker network ls --format '{{json .}}' output.
func ParseDockerNetworksJSONLines(stdout string) ([]NetworkSummary, error) {
	lines := jsonObjectLines(stdout)
	out := make([]NetworkSummary, 0, len(lines))
	for _, line := range lines {
		var raw dockerNetworkLine
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			return nil, fmt.Errorf("docker: parse networks: %w", err)
		}
		out = append(out, NetworkSummary{ID: raw.ID, Name: raw.Name, Driver: raw.Driver, Scope: raw.Scope, Via: viaCLI})
	}
	return out, nil
}

type dockerVolumeLine struct {
	Name       string `json:"Name"`
	Driver     string `json:"Driver"`
	Mountpoint string `json:"Mountpoint"`
}

// ParseDockerVolumesJSONLines parses docker volume ls --format '{{json .}}' output.
func ParseDockerVolumesJSONLines(stdout string) ([]VolumeSummary, error) {
	lines := jsonObjectLines(stdout)
	out := make([]VolumeSummary, 0, len(lines))
	for _, line := range lines {
		var raw dockerVolumeLine
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			return nil, fmt.Errorf("docker: parse volumes: %w", err)
		}
		out = append(out, VolumeSummary{Name: raw.Name, Driver: raw.Driver, Mountpoint: raw.Mountpoint, Via: viaCLI})
	}
	return out, nil
}

// cliOutput runs cmd and returns stdout, mapping errors like cliRun.
func cliOutput(sess CLISession, cmd string) ([]byte, error) {
	if sess == nil {
		return nil, fmt.Errorf("docker: CLI 模式缺少 SSH 会话")
	}
	out, err := sess.CombinedOutput(cmd)
	if err != nil {
		return nil, cliCommandError(out, err)
	}
	return out, nil
}

func CLIListImages(sess CLISession) ([]ImageSummary, error) {
	out, err := cliOutput(sess, BuildDockerImagesCommand())
	if err != nil {
		return nil, err
	}
	return ParseDockerImagesJSONLines(string(out))
}

func CLIListNetworks(sess CLISession) ([]NetworkSummary, error) {
	out, err := cliOutput(sess, BuildDockerNetworksCommand())
	if err != nil {
		return nil, err
	}
	return ParseDockerNetworksJSONLines(string(out))
}

func CLIListVolumes(sess CLISession) ([]VolumeSummary, error) {
	out, err := cliOutput(sess, BuildDockerVolumesCommand())
	if err != nil {
		return nil, err
	}
	return ParseDockerVolumesJSONLines(string(out))
}

// ParseDockerImageInspectJSON parses docker image inspect output (JSON array, first element).
func ParseDockerImageInspectJSON(data []byte) (ImageDetail, error) {
	var arr []imageInspectJSON
	if err := json.Unmarshal(bytesTrimSpace(data), &arr); err != nil {
		return ImageDetail{}, fmt.Errorf("docker: parse image inspect JSON: %w", err)
	}
	if len(arr) == 0 {
		return ImageDetail{}, fmt.Errorf("docker: inspect returned no images")
	}
	raw := arr[0]
	tags := raw.RepoTags
	if tags == nil {
		tags = []string{}
	}
	return ImageDetail{
		ImageSummary: ImageSummary{
			ID:      raw.ID,
			Tags:    tags,
			Size:    raw.Size,
			Created: inspectCreatedUnix(raw.Created),
		},
		Architecture: raw.Architecture,
		Os:           raw.Os,
		Author:       raw.Author,
		Comment:      raw.Comment,
	}, nil
}

// ParseDockerNetworkInspectJSON parses docker network inspect output (JSON array, first element).
func ParseDockerNetworkInspectJSON(data []byte) (NetworkDetail, error) {
	var arr []networkJSON
	if err := json.Unmarshal(bytesTrimSpace(data), &arr); err != nil {
		return NetworkDetail{}, fmt.Errorf("docker: parse network inspect JSON: %w", err)
	}
	if len(arr) == 0 {
		return NetworkDetail{}, fmt.Errorf("docker: inspect returned no networks")
	}
	raw := arr[0]
	var parts []string
	for _, cfg := range raw.IPAM.Config {
		switch {
		case cfg.Subnet != "" && cfg.Gateway != "":
			parts = append(parts, cfg.Subnet+" gw "+cfg.Gateway)
		case cfg.Subnet != "":
			parts = append(parts, cfg.Subnet)
		case cfg.Gateway != "":
			parts = append(parts, "gw "+cfg.Gateway)
		}
	}
	return NetworkDetail{
		NetworkSummary: raw.summary(),
		Internal:       raw.Internal,
		Attachable:     raw.Attachable,
		Ingress:        raw.Ingress,
		IPAM:           strings.Join(parts, ", "),
	}, nil
}

// ParseDockerVolumeInspectJSON parses docker volume inspect output (JSON array, first element).
func ParseDockerVolumeInspectJSON(data []byte) (VolumeDetail, error) {
	var arr []volumeJSON
	if err := json.Unmarshal(bytesTrimSpace(data), &arr); err != nil {
		return VolumeDetail{}, fmt.Errorf("docker: parse volume inspect JSON: %w", err)
	}
	if len(arr) == 0 {
		return VolumeDetail{}, fmt.Errorf("docker: inspect returned no volumes")
	}
	raw := arr[0]
	labels := raw.Labels
	if labels == nil {
		labels = map[string]string{}
	}
	return VolumeDetail{VolumeSummary: raw.summary(), Labels: labels, Scope: raw.Scope}, nil
}

func CLIInspectImage(sess CLISession, ref string) (ImageDetail, error) {
	out, err := cliOutput(sess, BuildDockerImageInspectCommand(ref))
	if err != nil {
		return ImageDetail{}, err
	}
	return ParseDockerImageInspectJSON(out)
}

func CLIInspectNetwork(sess CLISession, id string) (NetworkDetail, error) {
	out, err := cliOutput(sess, BuildDockerNetworkInspectCommand(id))
	if err != nil {
		return NetworkDetail{}, err
	}
	return ParseDockerNetworkInspectJSON(out)
}

func CLIInspectVolume(sess CLISession, name string) (VolumeDetail, error) {
	out, err := cliOutput(sess, BuildDockerVolumeInspectCommand(name))
	if err != nil {
		return VolumeDetail{}, err
	}
	return ParseDockerVolumeInspectJSON(out)
}

func CLIRemoveImage(sess CLISession, ref string, force bool) error {
	return cliRun(sess, BuildDockerRmiCommand(ref, force))
}

func CLIRemoveNetwork(sess CLISession, id string) error {
	return cliRun(sess, BuildDockerNetworkRmCommand(id))
}

func CLIRemoveVolume(sess CLISession, name string, force bool) error {
	return cliRun(sess, BuildDockerVolumeRmCommand(name, force))
}
