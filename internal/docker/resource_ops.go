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
