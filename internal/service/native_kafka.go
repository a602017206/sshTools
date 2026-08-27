package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"
)

type KafkaNativeClient interface {
	Ping(context.Context) error
	ListTopics(context.Context) ([]string, error)
	DescribeTopic(context.Context, string) (NativeResourceDetails, error)
	Close() error
}
type KafkaNativeClientFactory interface {
	New(context.Context, NativeDatabaseConfig) (KafkaNativeClient, error)
}
type KafkaNativeProvider struct{ factory KafkaNativeClientFactory }

func NewKafkaNativeProvider(factory KafkaNativeClientFactory) *KafkaNativeProvider {
	return &KafkaNativeProvider{factory: factory}
}
func NewDefaultKafkaNativeProvider() *KafkaNativeProvider {
	return NewKafkaNativeProvider(kafkaFranzClientFactory{})
}
func (p *KafkaNativeProvider) Test(ctx context.Context, cfg NativeDatabaseConfig) error {
	client, err := p.factory.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer client.Close()
	return client.Ping(ctx)
}
func (p *KafkaNativeProvider) Connect(ctx context.Context, cfg NativeDatabaseConfig) (NativeDatabaseClient, error) {
	client, err := p.factory.New(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &kafkaNativeSession{client: client}, nil
}

type kafkaNativeSession struct{ client KafkaNativeClient }

func (s *kafkaNativeSession) ListPrimaryResources(ctx context.Context) ([]NativeResource, error) {
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
func (*kafkaNativeSession) ListSecondaryResources(context.Context, string) ([]NativeResource, error) {
	return []NativeResource{}, nil
}
func (s *kafkaNativeSession) Close() error { return s.client.Close() }
func (s *kafkaNativeSession) DescribeResource(ctx context.Context, _ string, name string) (NativeResourceDetails, error) {
	return s.client.DescribeTopic(ctx, name)
}

type kafkaFranzClientFactory struct{}

func (kafkaFranzClientFactory) New(_ context.Context, cfg NativeDatabaseConfig) (KafkaNativeClient, error) {
	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		host = "127.0.0.1"
	}
	port := cfg.Port
	if port == 0 {
		port = 9092
	}
	client, err := kgo.NewClient(kgo.SeedBrokers(host + ":" + strconv.Itoa(port)))
	if err != nil {
		return nil, err
	}
	return &kafkaFranzClient{client: client}, nil
}

type kafkaFranzClient struct{ client *kgo.Client }

func (c *kafkaFranzClient) Ping(ctx context.Context) error { return c.client.Ping(ctx) }
func (c *kafkaFranzClient) ListTopics(ctx context.Context) ([]string, error) {
	response, err := c.client.Request(ctx, kmsg.NewPtrMetadataRequest())
	if err != nil {
		return nil, err
	}
	metadata := response.(*kmsg.MetadataResponse)
	topics := make([]string, 0, len(metadata.Topics))
	for _, topic := range metadata.Topics {
		if topic.ErrorCode == 0 && !topic.IsInternal && topic.Topic != nil {
			topics = append(topics, *topic.Topic)
		}
	}
	return topics, nil
}
func (c *kafkaFranzClient) DescribeTopic(ctx context.Context, name string) (NativeResourceDetails, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return NativeResourceDetails{}, fmt.Errorf("Kafka Topic 名称不能为空")
	}
	request := kmsg.NewPtrMetadataRequest()
	request.Topics = append(request.Topics, kmsg.MetadataRequestTopic{Topic: &name})
	response, err := c.client.Request(ctx, request)
	if err != nil {
		return NativeResourceDetails{}, err
	}
	metadata := response.(*kmsg.MetadataResponse)
	if len(metadata.Topics) != 1 {
		return NativeResourceDetails{}, fmt.Errorf("未找到 Kafka Topic: %s", name)
	}
	topic := metadata.Topics[0]
	if topic.ErrorCode != 0 {
		return NativeResourceDetails{}, fmt.Errorf("读取 Kafka Topic %s 失败: %s", name, kerr.ErrorForCode(topic.ErrorCode))
	}
	partitions := make([]map[string]any, 0, len(topic.Partitions))
	for _, partition := range topic.Partitions {
		partitions = append(partitions, map[string]any{"id": partition.Partition, "leader": partition.Leader, "replicas": partition.Replicas, "isr": partition.ISR})
	}
	content, err := json.Marshal(map[string]any{"partitions": partitions})
	if err != nil {
		return NativeResourceDetails{}, err
	}
	return NativeResourceDetails{Kind: NativeResourceKindCollection, Name: name, Summary: fmt.Sprintf("%d 个分区", len(partitions)), Content: string(content)}, nil
}
func (c *kafkaFranzClient) Close() error { c.client.Close(); return nil }
