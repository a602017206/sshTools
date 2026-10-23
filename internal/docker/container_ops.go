package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (c *Client) StartContainer(ctx context.Context, id string) error {
	return c.containerLifecyclePost(ctx, id, "start")
}

func (c *Client) StopContainer(ctx context.Context, id string) error {
	return c.containerLifecyclePost(ctx, id, "stop")
}

func (c *Client) RestartContainer(ctx context.Context, id string) error {
	return c.containerLifecyclePost(ctx, id, "restart")
}

func (c *Client) containerLifecyclePost(ctx context.Context, id, action string) error {
	reqURL := fmt.Sprintf("%s/containers/%s/%s", c.baseURL, id, action)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusNotModified {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	return fmt.Errorf("docker: %s container failed: %s: %s", action, resp.Status, strings.TrimSpace(string(body)))
}

func (c *Client) RemoveContainer(ctx context.Context, id string, force bool) error {
	forceParam := "0"
	if force {
		forceParam = "1"
	}
	reqURL := fmt.Sprintf("%s/containers/%s?force=%s", c.baseURL, id, forceParam)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	return fmt.Errorf("docker: remove container failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
}

type containerInspectJSON struct {
	ID      string `json:"Id"`
	Created string `json:"Created"`
	Name    string `json:"Name"`
	Image   string `json:"Image"`
	Mounts  []struct {
		Source      string `json:"Source"`
		Destination string `json:"Destination"`
	} `json:"Mounts"`
	State struct {
		Status string `json:"Status"`
	} `json:"State"`
	Config struct {
		Image      string   `json:"Image"`
		Env        []string `json:"Env"`
		Cmd        []string `json:"Cmd"`
		Entrypoint []string `json:"Entrypoint"`
	} `json:"Config"`
	HostConfig struct {
		NetworkMode string `json:"NetworkMode"`
	} `json:"HostConfig"`
	NetworkSettings struct {
		Ports map[string][]struct {
			HostIP   string `json:"HostIp"`
			HostPort string `json:"HostPort"`
		} `json:"Ports"`
	} `json:"NetworkSettings"`
}

func (c *Client) InspectContainer(ctx context.Context, id string) (ContainerDetail, error) {
	reqURL := fmt.Sprintf("%s/containers/%s/json", c.baseURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return ContainerDetail{}, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ContainerDetail{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return ContainerDetail{}, fmt.Errorf("docker: inspect container failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var raw containerInspectJSON
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return ContainerDetail{}, err
	}
	return mapContainerDetail(raw), nil
}

func mapContainerDetail(raw containerInspectJSON) ContainerDetail {
	image := raw.Config.Image
	if image == "" {
		image = raw.Image
	}
	status := raw.State.Status
	created := inspectCreatedUnix(raw.Created)
	mounts := make([]string, 0, len(raw.Mounts))
	for _, m := range raw.Mounts {
		mounts = append(mounts, fmt.Sprintf("%s -> %s", m.Source, m.Destination))
	}
	env := raw.Config.Env
	if env == nil {
		env = []string{}
	}
	cmd := raw.Config.Cmd
	if cmd == nil {
		cmd = []string{}
	}
	entrypoint := raw.Config.Entrypoint
	if entrypoint == nil {
		entrypoint = []string{}
	}
	return ContainerDetail{
		ID:          shortContainerID(raw.ID),
		Name:        strings.TrimPrefix(raw.Name, "/"),
		Image:       image,
		State:       status,
		Status:      status,
		Ports:       formatInspectPorts(raw.NetworkSettings.Ports),
		NetworkMode: raw.HostConfig.NetworkMode,
		Created:     created,
		Mounts:      mounts,
		Env:         env,
		Cmd:         cmd,
		Entrypoint:  entrypoint,
	}
}

func inspectCreatedUnix(created string) int64 {
	if created == "" {
		return 0
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if t, err := time.Parse(layout, created); err == nil {
			return t.Unix()
		}
	}
	return 0
}

func formatInspectPorts(ports map[string][]struct {
	HostIP   string `json:"HostIp"`
	HostPort string `json:"HostPort"`
}) string {
	if len(ports) == 0 {
		return ""
	}
	parts := make([]string, 0, len(ports))
	for key, bindings := range ports {
		private, typ := parseInspectPortKey(key)
		if len(bindings) == 0 {
			parts = append(parts, fmt.Sprintf("%d/%s", private, typ))
			continue
		}
		for _, b := range bindings {
			hostPort, _ := strconv.Atoi(b.HostPort)
			if hostPort > 0 {
				parts = append(parts, fmt.Sprintf("%d:%d/%s", hostPort, private, typ))
			} else {
				parts = append(parts, fmt.Sprintf("%d/%s", private, typ))
			}
		}
	}
	return strings.Join(parts, ", ")
}

func parseInspectPortKey(key string) (private int, typ string) {
	typ = "tcp"
	before, after, ok := strings.Cut(key, "/")
	if ok {
		typ = after
	} else {
		before = key
	}
	private, _ = strconv.Atoi(before)
	return private, typ
}
