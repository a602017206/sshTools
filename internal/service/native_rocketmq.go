package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

type RocketMQNativeClient interface {
	Ping(context.Context) error
	ListTopics(context.Context) ([]string, error)
	DescribeTopic(context.Context, string) (NativeResourceDetails, error)
	Close() error
}

type RocketMQNativeClientFactory interface {
	New(context.Context, NativeDatabaseConfig) (RocketMQNativeClient, error)
}

type RocketMQNativeProvider struct{ factory RocketMQNativeClientFactory }

func NewRocketMQNativeProvider(factory RocketMQNativeClientFactory) *RocketMQNativeProvider {
	return &RocketMQNativeProvider{factory: factory}
}

func NewDefaultRocketMQNativeProvider() *RocketMQNativeProvider {
	return NewRocketMQNativeProvider(rocketMQDefaultFactory{})
}

func (p *RocketMQNativeProvider) Test(ctx context.Context, cfg NativeDatabaseConfig) error {
	client, err := p.factory.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer client.Close()
	return client.Ping(ctx)
}

func (p *RocketMQNativeProvider) Connect(ctx context.Context, cfg NativeDatabaseConfig) (NativeDatabaseClient, error) {
	client, err := p.factory.New(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &rocketMQNativeSession{client: client}, nil
}

type rocketMQNativeSession struct{ client RocketMQNativeClient }

func (s *rocketMQNativeSession) ListPrimaryResources(ctx context.Context) ([]NativeResource, error) {
	topics, err := s.client.ListTopics(ctx)
	if err != nil {
		return nil, err
	}
	sort.Strings(topics)
	resources := make([]NativeResource, 0, len(topics))
	for _, topic := range topics {
		resources = append(resources, NativeResource{Kind: NativeResourceKindCollection, Name: topic})
	}
	return resources, nil
}

func (*rocketMQNativeSession) ListSecondaryResources(context.Context, string) ([]NativeResource, error) {
	return []NativeResource{}, nil
}

func (s *rocketMQNativeSession) Close() error { return s.client.Close() }

func (s *rocketMQNativeSession) DescribeResource(ctx context.Context, _ string, name string) (NativeResourceDetails, error) {
	return s.client.DescribeTopic(ctx, name)
}

type rocketMQDefaultFactory struct{}

func (rocketMQDefaultFactory) New(_ context.Context, cfg NativeDatabaseConfig) (RocketMQNativeClient, error) {
	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		host = "127.0.0.1"
	}
	port := cfg.Port
	if port == 0 {
		port = 9876
	}
	addr := host + ":" + strconv.Itoa(port)
	opts := []admin.AdminOption{
		admin.WithResolver(primitive.NewPassthroughResolver([]string{addr})),
	}
	user := strings.TrimSpace(cfg.User)
	password := strings.TrimSpace(cfg.Password)
	if user != "" || password != "" {
		opts = append(opts, admin.WithCredentials(primitive.Credentials{
			AccessKey: user,
			SecretKey: password,
		}))
	}
	adm, err := admin.NewAdmin(opts...)
	if err != nil {
		return nil, fmt.Errorf("创建 RocketMQ Admin 失败: %w", err)
	}
	return &rocketMQAdminClient{admin: adm, nameserver: addr}, nil
}

type rocketMQAdminClient struct {
	admin      admin.Admin
	nameserver string
}

func (c *rocketMQAdminClient) Ping(ctx context.Context) error {
	_, err := c.admin.FetchAllTopicList(ctx)
	if err != nil {
		return fmt.Errorf("连接 RocketMQ NameServer 失败: %w", err)
	}
	return nil
}

func (c *rocketMQAdminClient) ListTopics(ctx context.Context) ([]string, error) {
	result, err := c.admin.FetchAllTopicList(ctx)
	if err != nil {
		return nil, fmt.Errorf("列出 RocketMQ Topic 失败: %w", err)
	}
	topics := make([]string, 0, len(result.TopicList))
	for _, topic := range result.TopicList {
		name := strings.TrimSpace(topic)
		if name == "" || strings.HasPrefix(name, "%") || strings.HasPrefix(name, "RMQ_SYS_") {
			continue
		}
		topics = append(topics, name)
	}
	return topics, nil
}

func (c *rocketMQAdminClient) DescribeTopic(ctx context.Context, name string) (NativeResourceDetails, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return NativeResourceDetails{}, fmt.Errorf("RocketMQ Topic 名称不能为空")
	}
	queues, err := c.admin.FetchPublishMessageQueues(ctx, name)
	if err != nil {
		content, marshalErr := json.Marshal(map[string]any{
			"topic":      name,
			"nameserver": c.nameserver,
			"error":      err.Error(),
		})
		if marshalErr != nil {
			return NativeResourceDetails{}, marshalErr
		}
		return NativeResourceDetails{
			Kind:    NativeResourceKindCollection,
			Name:    name,
			Summary: "Topic 存在，但读取队列路由失败",
			Content: string(content),
		}, nil
	}
	queuePayload := make([]map[string]any, 0, len(queues))
	for _, queue := range queues {
		if queue == nil {
			continue
		}
		queuePayload = append(queuePayload, map[string]any{
			"topic":      queue.Topic,
			"brokerName": queue.BrokerName,
			"queueId":    queue.QueueId,
		})
	}
	content, err := json.Marshal(map[string]any{
		"topic":      name,
		"nameserver": c.nameserver,
		"queues":     queuePayload,
	})
	if err != nil {
		return NativeResourceDetails{}, err
	}
	return NativeResourceDetails{
		Kind:    NativeResourceKindCollection,
		Name:    name,
		Summary: fmt.Sprintf("%d 个写队列", len(queuePayload)),
		Content: string(content),
	}, nil
}

func (c *rocketMQAdminClient) Close() error {
	if c.admin == nil {
		return nil
	}
	return c.admin.Close()
}
