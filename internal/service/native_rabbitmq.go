package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQNativeClient interface {
	Ping(context.Context) error
	ListQueues(context.Context) ([]string, error)
	DescribeQueue(context.Context, string) (NativeResourceDetails, error)
	Close() error
}

type RabbitMQNativeClientFactory interface {
	New(context.Context, NativeDatabaseConfig) (RabbitMQNativeClient, error)
}

type RabbitMQNativeProvider struct{ factory RabbitMQNativeClientFactory }

func NewRabbitMQNativeProvider(factory RabbitMQNativeClientFactory) *RabbitMQNativeProvider {
	return &RabbitMQNativeProvider{factory: factory}
}

func NewDefaultRabbitMQNativeProvider() *RabbitMQNativeProvider {
	return NewRabbitMQNativeProvider(rabbitMQDefaultFactory{})
}

func (p *RabbitMQNativeProvider) Test(ctx context.Context, cfg NativeDatabaseConfig) error {
	client, err := p.factory.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer client.Close()
	return client.Ping(ctx)
}

func (p *RabbitMQNativeProvider) Connect(ctx context.Context, cfg NativeDatabaseConfig) (NativeDatabaseClient, error) {
	client, err := p.factory.New(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &rabbitMQNativeSession{client: client}, nil
}

type rabbitMQNativeSession struct{ client RabbitMQNativeClient }

func (s *rabbitMQNativeSession) ListPrimaryResources(ctx context.Context) ([]NativeResource, error) {
	queues, err := s.client.ListQueues(ctx)
	if err != nil {
		return nil, err
	}
	sort.Strings(queues)
	resources := make([]NativeResource, 0, len(queues))
	for _, name := range queues {
		resources = append(resources, NativeResource{Kind: NativeResourceKindCollection, Name: name})
	}
	return resources, nil
}

func (*rabbitMQNativeSession) ListSecondaryResources(context.Context, string) ([]NativeResource, error) {
	return []NativeResource{}, nil
}

func (s *rabbitMQNativeSession) Close() error { return s.client.Close() }

func (s *rabbitMQNativeSession) DescribeResource(ctx context.Context, _ string, name string) (NativeResourceDetails, error) {
	return s.client.DescribeQueue(ctx, name)
}

type rabbitMQDefaultFactory struct{}

func (rabbitMQDefaultFactory) New(_ context.Context, cfg NativeDatabaseConfig) (RabbitMQNativeClient, error) {
	host, amqpPort, mgmtPort := rabbitMQEndpoints(cfg)
	user := strings.TrimSpace(cfg.User)
	password := cfg.Password
	if user == "" {
		user = "guest"
	}
	if password == "" {
		password = "guest"
	}
	vhost := strings.TrimSpace(cfg.Database)
	if vhost == "" {
		vhost = "/"
	}
	return &rabbitMQClient{
		host:     host,
		amqpPort: amqpPort,
		mgmtPort: mgmtPort,
		user:     user,
		password: password,
		vhost:    vhost,
		timeout:  rabbitMQTimeout(cfg),
	}, nil
}

func rabbitMQEndpoints(cfg NativeDatabaseConfig) (host string, amqpPort, mgmtPort int) {
	host = strings.TrimSpace(cfg.Host)
	if host == "" {
		host = "127.0.0.1"
	}
	amqpPort = cfg.Port
	if amqpPort == 0 {
		amqpPort = 5672
	}
	mgmtPort = 15672
	if amqpPort == 15672 {
		amqpPort = 5672
		mgmtPort = 15672
	}
	return host, amqpPort, mgmtPort
}

func rabbitMQTimeout(cfg NativeDatabaseConfig) time.Duration {
	if cfg.Timeout > 0 {
		return cfg.Timeout
	}
	return 8 * time.Second
}

type rabbitMQClient struct {
	host     string
	amqpPort int
	mgmtPort int
	user     string
	password string
	vhost    string
	timeout  time.Duration
	conn     *amqp.Connection
}

func (c *rabbitMQClient) amqpURL() string {
	user := url.QueryEscape(c.user)
	pass := url.QueryEscape(c.password)
	if c.vhost == "" || c.vhost == "/" {
		return fmt.Sprintf("amqp://%s:%s@%s:%d/", user, pass, c.host, c.amqpPort)
	}
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s", user, pass, c.host, c.amqpPort, url.PathEscape(c.vhost))
}

func (c *rabbitMQClient) Ping(ctx context.Context) error {
	_ = ctx
	conn, err := amqp.DialConfig(c.amqpURL(), amqp.Config{Dial: amqp.DefaultDial(c.timeout)})
	if err != nil {
		return fmt.Errorf("连接 RabbitMQ AMQP 失败: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("打开 RabbitMQ Channel 失败: %w", err)
	}
	_ = ch.Close()
	c.conn = conn
	return nil
}

func (c *rabbitMQClient) ListQueues(ctx context.Context) ([]string, error) {
	queues, err := c.fetchQueues(ctx)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(queues))
	for _, item := range queues {
		if name := strings.TrimSpace(item.Name); name != "" {
			names = append(names, name)
		}
	}
	return names, nil
}

func (c *rabbitMQClient) DescribeQueue(ctx context.Context, name string) (NativeResourceDetails, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return NativeResourceDetails{}, fmt.Errorf("RabbitMQ Queue 名称不能为空")
	}
	queues, err := c.fetchQueues(ctx)
	if err != nil {
		return NativeResourceDetails{}, err
	}
	for _, item := range queues {
		if item.Name != name {
			continue
		}
		content, err := json.Marshal(map[string]any{
			"name":       item.Name,
			"vhost":      item.Vhost,
			"durable":    item.Durable,
			"autoDelete": item.AutoDelete,
			"messages":   item.Messages,
			"consumers":  item.Consumers,
		})
		if err != nil {
			return NativeResourceDetails{}, err
		}
		return NativeResourceDetails{
			Kind:    NativeResourceKindCollection,
			Name:    name,
			Summary: fmt.Sprintf("消息 %d · 消费者 %d", item.Messages, item.Consumers),
			Content: string(content),
		}, nil
	}
	return NativeResourceDetails{}, fmt.Errorf("未找到 RabbitMQ Queue: %s", name)
}

func (c *rabbitMQClient) Close() error {
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

type rabbitMQQueueInfo struct {
	Name       string `json:"name"`
	Vhost      string `json:"vhost"`
	Durable    bool   `json:"durable"`
	AutoDelete bool   `json:"auto_delete"`
	Messages   int    `json:"messages"`
	Consumers  int    `json:"consumers"`
}

func (c *rabbitMQClient) fetchQueues(ctx context.Context) ([]rabbitMQQueueInfo, error) {
	endpoint := fmt.Sprintf(
		"http://%s:%d/api/queues/%s",
		c.host,
		c.mgmtPort,
		url.PathEscape(c.vhost),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.user, c.password)
	client := &http.Client{Timeout: c.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("读取 RabbitMQ Management API 失败（默认端口 %d）: %w", c.mgmtPort, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("RabbitMQ Management API 返回 %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var queues []rabbitMQQueueInfo
	if err := json.Unmarshal(body, &queues); err != nil {
		return nil, fmt.Errorf("解析 RabbitMQ Queue 列表失败: %w", err)
	}
	return queues, nil
}
