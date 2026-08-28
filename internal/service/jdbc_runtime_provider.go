package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
)

const adoptiumAPIBaseURL = "https://api.adoptium.net"

type ManagedRuntimePackage struct {
	Version string
	Name    string
	URL     string
	SHA256  string
}

type ManagedRuntimeProvider interface {
	Latest(ctx context.Context, featureVersion int) (ManagedRuntimePackage, error)
}

type ArtifactFetcher interface {
	Download(ctx context.Context, sourceURL, expectedSHA256, target string) error
}

type AdoptiumRuntimeProvider struct {
	client  *http.Client
	baseURL string
}

func NewAdoptiumRuntimeProvider(client *http.Client, baseURL string) *AdoptiumRuntimeProvider {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	if baseURL == "" {
		baseURL = adoptiumAPIBaseURL
	}
	return &AdoptiumRuntimeProvider{client: client, baseURL: strings.TrimRight(baseURL, "/")}
}

func (p *AdoptiumRuntimeProvider) Latest(ctx context.Context, featureVersion int) (ManagedRuntimePackage, error) {
	operatingSystem, architecture, err := adoptiumTarget()
	if err != nil {
		return ManagedRuntimePackage{}, err
	}
	assets, err := p.latestAssets(ctx, featureVersion, operatingSystem, architecture)
	if err != nil {
		return ManagedRuntimePackage{}, err
	}
	if pkg, ok := adoptiumRuntimePackage(assets, operatingSystem, architecture); ok {
		return pkg, nil
	}
	return ManagedRuntimePackage{}, fmt.Errorf("Adoptium 没有适用于 %s/%s 的 Java %d JRE", runtime.GOOS, runtime.GOARCH, featureVersion)
}

func adoptiumTarget() (string, string, error) {
	operatingSystem, err := adoptiumOperatingSystem(runtime.GOOS)
	if err != nil {
		return "", "", err
	}
	architecture, err := adoptiumArchitecture(runtime.GOARCH)
	if err != nil {
		return "", "", err
	}
	return operatingSystem, architecture, nil
}

func (p *AdoptiumRuntimeProvider) latestAssets(ctx context.Context, featureVersion int, operatingSystem, architecture string) ([]adoptiumAsset, error) {
	endpoint, err := p.latestEndpoint(featureVersion, operatingSystem, architecture)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("创建 Adoptium API 请求失败: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "AHaSSHTools-JDBC/1")
	response, err := p.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("查询 Adoptium 运行时失败: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("查询 Adoptium 运行时失败，HTTP 状态: %s", response.Status)
	}
	return decodeAdoptiumAssets(response.Body)
}

func (p *AdoptiumRuntimeProvider) latestEndpoint(featureVersion int, operatingSystem, architecture string) (*url.URL, error) {
	endpoint, err := url.Parse(fmt.Sprintf("%s/v3/assets/latest/%d/hotspot", p.baseURL, featureVersion))
	if err != nil {
		return nil, fmt.Errorf("构造 Adoptium API 地址失败: %w", err)
	}
	query := endpoint.Query()
	query.Set("architecture", architecture)
	query.Set("heap_size", "normal")
	query.Set("image_type", "jre")
	query.Set("jvm_impl", "hotspot")
	query.Set("os", operatingSystem)
	query.Set("vendor", "eclipse")
	endpoint.RawQuery = query.Encode()
	return endpoint, nil
}

func decodeAdoptiumAssets(body io.Reader) ([]adoptiumAsset, error) {
	data, err := io.ReadAll(io.LimitReader(body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("读取 Adoptium API 响应失败: %w", err)
	}
	var assets []adoptiumAsset
	if err := json.Unmarshal(data, &assets); err != nil {
		return nil, fmt.Errorf("解析 Adoptium API 响应失败: %w", err)
	}
	return assets, nil
}

func adoptiumRuntimePackage(assets []adoptiumAsset, operatingSystem, architecture string) (ManagedRuntimePackage, bool) {
	for _, asset := range assets {
		binary := asset.Binary
		if binary.OS != operatingSystem || binary.Architecture != architecture || binary.ImageType != "jre" {
			continue
		}
		pkg := binary.Package
		if pkg.Name == "" || pkg.Link == "" || len(pkg.Checksum) != 64 {
			continue
		}
		return ManagedRuntimePackage{Version: asset.VersionData.Semver, Name: pkg.Name, URL: pkg.Link, SHA256: pkg.Checksum}, true
	}
	return ManagedRuntimePackage{}, false
}

type adoptiumAsset struct {
	Binary struct {
		Architecture string `json:"architecture"`
		OS           string `json:"os"`
		ImageType    string `json:"image_type"`
		Package      struct {
			Checksum string `json:"checksum"`
			Link     string `json:"link"`
			Name     string `json:"name"`
		} `json:"package"`
	} `json:"binary"`
	VersionData struct {
		Semver string `json:"semver"`
	} `json:"version_data"`
}

func adoptiumOperatingSystem(goos string) (string, error) {
	switch goos {
	case "darwin":
		return "mac", nil
	case "linux", "windows", "aix", "solaris":
		return goos, nil
	default:
		return "", fmt.Errorf("Adoptium 不支持当前操作系统: %s", goos)
	}
}

func adoptiumArchitecture(goarch string) (string, error) {
	switch goarch {
	case "amd64":
		return "x64", nil
	case "386":
		return "x86", nil
	case "arm64":
		return "aarch64", nil
	case "arm":
		return "arm", nil
	case "ppc64", "ppc64le", "s390x", "riscv64":
		return goarch, nil
	default:
		return "", fmt.Errorf("Adoptium 不支持当前架构: %s", goarch)
	}
}
