package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

type InfluxDBNativeClient interface {
	Ping(context.Context) error
	ListBuckets(context.Context) ([]string, error)
	Close() error
}
type InfluxDBNativeClientFactory interface {
	New(context.Context, NativeDatabaseConfig) (InfluxDBNativeClient, error)
}
type InfluxDBNativeProvider struct{ factory InfluxDBNativeClientFactory }

func NewInfluxDBNativeProvider(factory InfluxDBNativeClientFactory) *InfluxDBNativeProvider {
	return &InfluxDBNativeProvider{factory: factory}
}
func NewDefaultInfluxDBNativeProvider() *InfluxDBNativeProvider {
	return NewInfluxDBNativeProvider(influxDBHTTPClientFactory{})
}
func (p *InfluxDBNativeProvider) Test(ctx context.Context, cfg NativeDatabaseConfig) error {
	client, err := p.factory.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer client.Close()
	return client.Ping(ctx)
}
func (p *InfluxDBNativeProvider) Connect(ctx context.Context, cfg NativeDatabaseConfig) (NativeDatabaseClient, error) {
	client, err := p.factory.New(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &influxDBNativeSession{client: client}, nil
}

type influxDBNativeSession struct{ client InfluxDBNativeClient }

func (s *influxDBNativeSession) ListPrimaryResources(ctx context.Context) ([]NativeResource, error) {
	buckets, err := s.client.ListBuckets(ctx)
	if err != nil {
		return nil, err
	}
	sort.Strings(buckets)
	resources := make([]NativeResource, 0, len(buckets))
	for _, bucket := range buckets {
		resources = append(resources, NativeResource{Kind: NativeResourceKindDatabase, Name: bucket})
	}
	return resources, nil
}
func (*influxDBNativeSession) ListSecondaryResources(context.Context, string) ([]NativeResource, error) {
	return []NativeResource{}, nil
}
func (s *influxDBNativeSession) Close() error { return s.client.Close() }

type influxDBHTTPClientFactory struct{}

func (influxDBHTTPClientFactory) New(_ context.Context, cfg NativeDatabaseConfig) (InfluxDBNativeClient, error) {
	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		host = "127.0.0.1"
	}
	port := cfg.Port
	if port == 0 {
		port = 8086
	}
	baseURL, err := url.Parse("http://" + host + ":" + strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	return &influxDBHTTPClient{baseURL: baseURL, httpClient: &http.Client{Timeout: cfg.Timeout}, user: cfg.User, token: cfg.Password}, nil
}

type influxDBHTTPClient struct {
	baseURL     *url.URL
	httpClient  *http.Client
	user, token string
}

func (c *influxDBHTTPClient) Ping(ctx context.Context) error {
	response, err := c.request(ctx, "/health")
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return influxDBResponseError(response)
}
func (c *influxDBHTTPClient) ListBuckets(ctx context.Context) ([]string, error) {
	response, err := c.request(ctx, "/api/v2/buckets?limit=1000")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if err := influxDBResponseError(response); err != nil {
		return nil, err
	}
	var payload struct {
		Buckets []struct {
			Name string `json:"name"`
		} `json:"buckets"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("解析 InfluxDB bucket 列表失败: %w", err)
	}
	buckets := make([]string, 0, len(payload.Buckets))
	for _, bucket := range payload.Buckets {
		if bucket.Name != "" {
			buckets = append(buckets, bucket.Name)
		}
	}
	return buckets, nil
}
func (c *influxDBHTTPClient) request(ctx context.Context, path string) (*http.Response, error) {
	endpoint := c.baseURL.ResolveReference(&url.URL{Path: path})
	if strings.Contains(path, "?") {
		parsed, err := url.Parse(path)
		if err != nil {
			return nil, err
		}
		endpoint.RawQuery = parsed.RawQuery
		endpoint.Path = parsed.Path
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		request.Header.Set("Authorization", "Token "+c.token)
	} else if c.user != "" {
		request.SetBasicAuth(c.user, "")
	}
	return c.httpClient.Do(request)
}
func (c *influxDBHTTPClient) Close() error { return nil }
func influxDBResponseError(response *http.Response) error {
	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
		return nil
	}
	return fmt.Errorf("InfluxDB 返回 HTTP %d", response.StatusCode)
}
