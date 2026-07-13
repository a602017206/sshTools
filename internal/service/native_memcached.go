package service

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
)

type MemcachedNativeClient interface {
	Stats(context.Context) (map[string]string, error)
	Close() error
}

type MemcachedNativeClientFactory interface {
	New(context.Context, NativeDatabaseConfig) (MemcachedNativeClient, error)
}

type MemcachedNativeProvider struct{ factory MemcachedNativeClientFactory }

func NewMemcachedNativeProvider(factory MemcachedNativeClientFactory) *MemcachedNativeProvider {
	return &MemcachedNativeProvider{factory: factory}
}

func NewDefaultMemcachedNativeProvider() *MemcachedNativeProvider {
	return NewMemcachedNativeProvider(memcachedTCPClientFactory{})
}

func (p *MemcachedNativeProvider) Test(ctx context.Context, cfg NativeDatabaseConfig) error {
	client, err := p.factory.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer client.Close()
	_, err = client.Stats(ctx)
	return err
}

func (p *MemcachedNativeProvider) Connect(ctx context.Context, cfg NativeDatabaseConfig) (NativeDatabaseClient, error) {
	client, err := p.factory.New(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if _, err := client.Stats(ctx); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &memcachedNativeSession{client: client}, nil
}

type memcachedNativeSession struct{ client MemcachedNativeClient }

func (s *memcachedNativeSession) ListPrimaryResources(ctx context.Context) ([]NativeResource, error) {
	stats, err := s.client.Stats(ctx)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(stats))
	for key := range stats {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	resources := make([]NativeResource, 0, len(keys))
	for _, key := range keys {
		resources = append(resources, NativeResource{Kind: NativeResourceKindStatistic, Name: key + "=" + stats[key]})
	}
	return resources, nil
}

func (*memcachedNativeSession) ListSecondaryResources(context.Context, string) ([]NativeResource, error) {
	return []NativeResource{}, nil
}
func (s *memcachedNativeSession) Close() error { return s.client.Close() }

type memcachedTCPClientFactory struct{}

func (memcachedTCPClientFactory) New(ctx context.Context, cfg NativeDatabaseConfig) (MemcachedNativeClient, error) {
	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		host = "127.0.0.1"
	}
	port := cfg.Port
	if port == 0 {
		port = 11211
	}
	connection, err := (&net.Dialer{}).DialContext(ctx, "tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return nil, err
	}
	return &memcachedTCPClient{connection: connection}, nil
}

type memcachedTCPClient struct{ connection net.Conn }

func (c *memcachedTCPClient) Stats(ctx context.Context) (map[string]string, error) {
	if deadline, ok := ctx.Deadline(); ok {
		_ = c.connection.SetDeadline(deadline)
	}
	if _, err := c.connection.Write([]byte("stats\r\n")); err != nil {
		return nil, err
	}
	stats := make(map[string]string)
	scanner := bufio.NewScanner(c.connection)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "END" {
			return stats, nil
		}
		fields := strings.Fields(line)
		if len(fields) == 3 && fields[0] == "STAT" {
			stats[fields[1]] = fields[2]
			continue
		}
		return nil, fmt.Errorf("Memcached 返回未知响应: %s", line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("Memcached 在返回 END 前关闭连接")
}
func (c *memcachedTCPClient) Close() error { return c.connection.Close() }
