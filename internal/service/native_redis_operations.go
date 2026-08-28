package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const redisCollectionPreviewLimit = 50

type redisSetPayload struct {
	Value      string `json:"value"`
	TTLSeconds *int64 `json:"ttlSeconds"`
}

type redisSavePayload struct {
	Type       string             `json:"type"`
	Value      string             `json:"value"`
	Fields     map[string]string  `json:"fields"`
	Items      []string           `json:"items"`
	Members    []string           `json:"members"`
	Entries    []redisZSetEntry   `json:"entries"`
	TTLSeconds *int64             `json:"ttlSeconds"`
}

type redisZSetEntry struct {
	Member string  `json:"member"`
	Score  float64 `json:"score"`
}

func (s *redisNativeSession) MutateResource(ctx context.Context, parent, name, operation, payload string) (NativeMutationResult, error) {
	database, err := redisDatabaseNumber(parent)
	if err != nil {
		return NativeMutationResult{}, err
	}
	client := s.client
	closeClient := false
	if database != s.database {
		client, err = s.factory.New(s.config, database)
		if err != nil {
			return NativeMutationResult{}, err
		}
		closeClient = true
	}
	if closeClient {
		defer client.Close()
	}
	switch strings.TrimSpace(operation) {
	case "set":
		return client.SetKey(ctx, name, payload)
	case "save":
		return client.SaveKeyValue(ctx, name, payload)
	case "delete":
		return client.DeleteKey(ctx, name)
	default:
		return NativeMutationResult{}, fmt.Errorf("不支持的 Redis 操作: %s", operation)
	}
}

func parseRedisSetPayload(payload string) (redisSetPayload, error) {
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return redisSetPayload{}, fmt.Errorf("set 操作 payload 不能为空")
	}
	var parsed redisSetPayload
	if err := json.Unmarshal([]byte(payload), &parsed); err != nil {
		return redisSetPayload{}, fmt.Errorf("set payload 必须是合法 JSON: %w", err)
	}
	return parsed, nil
}

func (c *redisGoClient) SetKey(ctx context.Context, name, payload string) (NativeMutationResult, error) {
	parsed, err := parseRedisSetPayload(payload)
	if err != nil {
		return NativeMutationResult{}, err
	}
	ttl := time.Duration(0)
	if parsed.TTLSeconds != nil {
		if *parsed.TTLSeconds < 0 {
			return NativeMutationResult{}, fmt.Errorf("ttlSeconds 不能为负数")
		}
		if *parsed.TTLSeconds > 0 {
			ttl = time.Duration(*parsed.TTLSeconds) * time.Second
		}
	}
	if err := c.client.Set(ctx, name, parsed.Value, ttl).Err(); err != nil {
		return NativeMutationResult{}, err
	}
	summary := "键值已保存"
	if ttl > 0 {
		summary = fmt.Sprintf("键值已保存，TTL %d 秒", int64(ttl.Seconds()))
	}
	return NativeMutationResult{Summary: summary}, nil
}

func (c *redisGoClient) DeleteKey(ctx context.Context, name string) (NativeMutationResult, error) {
	deleted, err := c.client.Del(ctx, name).Result()
	if err != nil {
		return NativeMutationResult{}, err
	}
	return NativeMutationResult{Summary: fmt.Sprintf("已删除 %d 个键", deleted)}, nil
}

func parseRedisSavePayload(payload string) (redisSavePayload, error) {
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return redisSavePayload{}, fmt.Errorf("save 操作 payload 不能为空")
	}
	var parsed redisSavePayload
	if err := json.Unmarshal([]byte(payload), &parsed); err != nil {
		return redisSavePayload{}, fmt.Errorf("save payload 必须是合法 JSON: %w", err)
	}
	parsed.Type = strings.TrimSpace(parsed.Type)
	if parsed.Type == "" {
		return redisSavePayload{}, fmt.Errorf("save payload 缺少 type")
	}
	return parsed, nil
}

func (c *redisGoClient) SaveKeyValue(ctx context.Context, name, payload string) (NativeMutationResult, error) {
	parsed, err := parseRedisSavePayload(payload)
	if err != nil {
		return NativeMutationResult{}, err
	}
	pipe := c.client.Pipeline()
	switch parsed.Type {
	case "string":
		ttl := redisDuration(parsed.TTLSeconds)
		pipe.Set(ctx, name, parsed.Value, ttl)
	case "hash":
		pipe.Del(ctx, name)
		if len(parsed.Fields) > 0 {
			pipe.HSet(ctx, name, parsed.Fields)
		}
	case "list":
		pipe.Del(ctx, name)
		if len(parsed.Items) > 0 {
			pipe.RPush(ctx, name, redisStringArgs(parsed.Items)...)
		}
	case "set":
		pipe.Del(ctx, name)
		if len(parsed.Members) > 0 {
			pipe.SAdd(ctx, name, redisStringArgs(parsed.Members)...)
		}
	case "zset":
		pipe.Del(ctx, name)
		if len(parsed.Entries) > 0 {
			members := make([]redis.Z, 0, len(parsed.Entries))
			for _, entry := range parsed.Entries {
				if strings.TrimSpace(entry.Member) == "" {
					continue
				}
				members = append(members, redis.Z{Score: entry.Score, Member: entry.Member})
			}
			if len(members) > 0 {
				pipe.ZAdd(ctx, name, members...)
			}
		}
	default:
		return NativeMutationResult{}, fmt.Errorf("不支持的 Redis 类型: %s", parsed.Type)
	}
	if parsed.Type != "string" && parsed.TTLSeconds != nil && *parsed.TTLSeconds > 0 {
		pipe.Expire(ctx, name, time.Duration(*parsed.TTLSeconds)*time.Second)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return NativeMutationResult{}, err
	}
	return NativeMutationResult{Summary: fmt.Sprintf("%s 键已保存", parsed.Type)}, nil
}

func redisDuration(ttlSeconds *int64) time.Duration {
	if ttlSeconds == nil || *ttlSeconds <= 0 {
		return 0
	}
	return time.Duration(*ttlSeconds) * time.Second
}

func redisStringArgs(values []string) []any {
	args := make([]any, len(values))
	for index, value := range values {
		args[index] = value
	}
	return args
}

func (c *redisGoClient) readKeyPreview(ctx context.Context, kind, name string, data map[string]any) error {
	switch kind {
	case "string":
		value, err := c.client.Get(ctx, name).Result()
		if err != nil && err != redis.Nil {
			return err
		}
		if len(value) > redisPreviewLimit {
			value = value[:redisPreviewLimit]
			data["truncated"] = true
		}
		data["value"] = value
	case "hash":
		fields, err := c.client.HGetAll(ctx, name).Result()
		if err != nil {
			return err
		}
		data["fields"] = limitRedisMap(fields, redisCollectionPreviewLimit)
	case "list":
		items, err := c.client.LRange(ctx, name, 0, redisCollectionPreviewLimit-1).Result()
		if err != nil {
			return err
		}
		length, _ := c.client.LLen(ctx, name).Result()
		data["items"] = items
		data["length"] = length
		if length > int64(len(items)) {
			data["truncated"] = true
		}
	case "set":
		members, err := c.client.SMembers(ctx, name).Result()
		if err != nil {
			return err
		}
		count, _ := c.client.SCard(ctx, name).Result()
		if len(members) > redisCollectionPreviewLimit {
			members = members[:redisCollectionPreviewLimit]
			data["truncated"] = true
		}
		data["members"] = members
		data["length"] = count
	case "zset":
		items, err := c.client.ZRangeWithScores(ctx, name, 0, redisCollectionPreviewLimit-1).Result()
		if err != nil {
			return err
		}
		pairs := make([]map[string]any, 0, len(items))
		for _, item := range items {
			pairs = append(pairs, map[string]any{"member": item.Member, "score": item.Score})
		}
		count, _ := c.client.ZCard(ctx, name).Result()
		data["entries"] = pairs
		data["length"] = count
		if count > int64(len(items)) {
			data["truncated"] = true
		}
	default:
		data["unsupportedPreview"] = true
	}
	return nil
}

func limitRedisMap(values map[string]string, limit int) map[string]string {
	if len(values) <= limit {
		return values
	}
	limited := make(map[string]string, limit)
	count := 0
	for key, value := range values {
		limited[key] = value
		count++
		if count >= limit {
			break
		}
	}
	return limited
}

func limitRedisSlice(values []string, limit int) []string {
	if len(values) <= limit {
		return values
	}
	return values[:limit]
}

func redisKeySummary(kind string, ttl time.Duration) string {
	if ttl == -1 {
		return fmt.Sprintf("%s · 永不过期", kind)
	}
	if ttl == -2 {
		return fmt.Sprintf("%s · 已不存在", kind)
	}
	return fmt.Sprintf("%s · %d 秒后过期", kind, ttlSeconds(ttl))
}
