package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type imageListJSON struct {
	ID       string   `json:"Id"`
	RepoTags []string `json:"RepoTags"`
	Size     int64    `json:"Size"`
	Created  int64    `json:"Created"`
}

type imageInspectJSON struct {
	ID           string   `json:"Id"`
	RepoTags     []string `json:"RepoTags"`
	Size         int64    `json:"Size"`
	Created      string   `json:"Created"`
	Architecture string   `json:"Architecture"`
	Os           string   `json:"Os"`
	Author       string   `json:"Author"`
	Comment      string   `json:"Comment"`
}

// imagePath 转义镜像引用；保留 "/"（带仓库前缀的名称），转义其余特殊字符。
func imagePath(ref string) string {
	parts := strings.Split(ref, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return strings.Join(parts, "/")
}

func (c *Client) ListImages(ctx context.Context) ([]ImageSummary, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/images/json", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("docker: list images failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var raw []imageListJSON
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	out := make([]ImageSummary, 0, len(raw))
	for _, r := range raw {
		tags := r.RepoTags
		if tags == nil {
			tags = []string{}
		}
		out = append(out, ImageSummary{ID: r.ID, Tags: tags, Size: r.Size, Created: r.Created})
	}
	return out, nil
}

func (c *Client) InspectImage(ctx context.Context, ref string) (ImageDetail, error) {
	reqURL := fmt.Sprintf("%s/images/%s/json", c.baseURL, imagePath(ref))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return ImageDetail{}, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ImageDetail{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return ImageDetail{}, fmt.Errorf("docker: inspect image failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var raw imageInspectJSON
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return ImageDetail{}, err
	}
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

func (c *Client) RemoveImage(ctx context.Context, ref string, force bool) error {
	reqURL := fmt.Sprintf("%s/images/%s", c.baseURL, imagePath(ref))
	if force {
		reqURL += "?force=1"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	return fmt.Errorf("docker: remove image failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
}

type networkJSON struct {
	ID         string `json:"Id"`
	Name       string `json:"Name"`
	Driver     string `json:"Driver"`
	Scope      string `json:"Scope"`
	Internal   bool   `json:"Internal"`
	Attachable bool   `json:"Attachable"`
	Ingress    bool   `json:"Ingress"`
	IPAM       struct {
		Config []struct {
			Subnet  string `json:"Subnet"`
			Gateway string `json:"Gateway"`
		} `json:"Config"`
	} `json:"IPAM"`
}

type volumeJSON struct {
	Name       string            `json:"Name"`
	Driver     string            `json:"Driver"`
	Mountpoint string            `json:"Mountpoint"`
	CreatedAt  string            `json:"CreatedAt"`
	Labels     map[string]string `json:"Labels"`
	Scope      string            `json:"Scope"`
}

func (n networkJSON) summary() NetworkSummary {
	return NetworkSummary{ID: n.ID, Name: n.Name, Driver: n.Driver, Scope: n.Scope}
}

func (v volumeJSON) summary() VolumeSummary {
	return VolumeSummary{Name: v.Name, Driver: v.Driver, Mountpoint: v.Mountpoint, Created: inspectCreatedUnix(v.CreatedAt)}
}

// doJSON 发起请求并在 200 时解码 JSON；其他状态返回带 op 描述的错误。
func (c *Client) doJSON(ctx context.Context, method, reqURL, op string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("docker: %s failed: %s: %s", op, resp.Status, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) doDelete(ctx context.Context, reqURL, op string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	return fmt.Errorf("docker: %s failed: %s: %s", op, resp.Status, strings.TrimSpace(string(body)))
}

func (c *Client) ListNetworks(ctx context.Context) ([]NetworkSummary, error) {
	var raw []networkJSON
	if err := c.doJSON(ctx, http.MethodGet, c.baseURL+"/networks", "list networks", &raw); err != nil {
		return nil, err
	}
	out := make([]NetworkSummary, 0, len(raw))
	for _, r := range raw {
		out = append(out, r.summary())
	}
	return out, nil
}

func (c *Client) InspectNetwork(ctx context.Context, id string) (NetworkDetail, error) {
	var raw networkJSON
	if err := c.doJSON(ctx, http.MethodGet, c.baseURL+"/networks/"+url.PathEscape(id), "inspect network", &raw); err != nil {
		return NetworkDetail{}, err
	}
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

func (c *Client) RemoveNetwork(ctx context.Context, id string) error {
	return c.doDelete(ctx, c.baseURL+"/networks/"+url.PathEscape(id), "remove network")
}

func (c *Client) ListVolumes(ctx context.Context) ([]VolumeSummary, error) {
	var raw struct {
		Volumes []volumeJSON `json:"Volumes"`
	}
	if err := c.doJSON(ctx, http.MethodGet, c.baseURL+"/volumes", "list volumes", &raw); err != nil {
		return nil, err
	}
	out := make([]VolumeSummary, 0, len(raw.Volumes))
	for _, r := range raw.Volumes {
		out = append(out, r.summary())
	}
	return out, nil
}

func (c *Client) InspectVolume(ctx context.Context, name string) (VolumeDetail, error) {
	var raw volumeJSON
	if err := c.doJSON(ctx, http.MethodGet, c.baseURL+"/volumes/"+url.PathEscape(name), "inspect volume", &raw); err != nil {
		return VolumeDetail{}, err
	}
	labels := raw.Labels
	if labels == nil {
		labels = map[string]string{}
	}
	return VolumeDetail{VolumeSummary: raw.summary(), Labels: labels, Scope: raw.Scope}, nil
}

func (c *Client) RemoveVolume(ctx context.Context, name string, force bool) error {
	reqURL := c.baseURL + "/volumes/" + url.PathEscape(name)
	if force {
		reqURL += "?force=1"
	}
	return c.doDelete(ctx, reqURL, "remove volume")
}
