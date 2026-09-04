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
	Type       string            `json:"type"`
	Value      string            `json:"value"`
	Fields     map[string]string `json:"fields"`
	Items      []string          `json:"items"`
	Members    []string          `json:"members"`
	Entries    []redisZSetEntry  `json:"entries"`
	TTLSeconds *int64            `json:"ttlSeconds"`
}

type redisZSetEntry struct {
	Member string  `json:"member"`
	Score  float64 `json:"score"`
}

type redisDeleteKeysPayload struct {
	Keys []string `json:"keys"`
}

type redisQueryEnvelope struct {
	Mode     string `json:"mode"`
	Command  string `json:"command"`
	Pattern  string `json:"pattern"`
	Cursor   string `json:"cursor"`
	Count    int    `json:"count"`
	ReadOnly *bool  `json:"readOnly"`
}

var redisCLIDeniedPrefixes = []string{
	"FLUSH", "DEBUG", "CONFIG", "SHUTDOWN", "SLAVEOF", "REPLICAOF", "MIGRATE", "CLUSTER", "MODULE", "SCRIPT", "ACL",
}

var redisCLIReadCommands = map[string]struct{}{
	"GET": {}, "MGET": {}, "STRLEN": {}, "TTL": {}, "PTTL": {}, "TYPE": {}, "EXISTS": {}, "SCAN": {},
	"HGET": {}, "HGETALL": {}, "HKEYS": {}, "HVALS": {}, "HLEN": {}, "HEXISTS": {},
	"LRANGE": {}, "LLEN": {}, "LINDEX": {},
	"SMEMBERS": {}, "SCARD": {}, "SISMEMBER": {}, "SSCAN": {},
	"ZRANGE": {}, "ZRANGEBYSCORE": {}, "ZCARD": {}, "ZSCORE": {}, "ZSCAN": {},
	"DBSIZE": {}, "INFO": {}, "PING": {}, "ECHO": {}, "KEYS": {}, "OBJECT": {}, "MEMORY": {},
	"GETRANGE": {}, "DUMP": {},
}

func (s *redisNativeSession) MutateResource(ctx context.Context, parent, name, operation, payload string) (NativeMutationResult, error) {
	client, cleanup, err := s.redisClientFor(ctx, parent)
	if err != nil {
		return NativeMutationResult{}, err
	}
	defer cleanup()
	switch strings.TrimSpace(operation) {
	case "set":
		return client.SetKey(ctx, name, payload)
	case "save", "create_key":
		if strings.TrimSpace(name) == "" {
			return NativeMutationResult{}, fmt.Errorf("键名不能为空")
		}
		return client.SaveKeyValue(ctx, name, payload)
	case "delete":
		return client.DeleteKey(ctx, name)
	case "delete_keys":
		keys, parseErr := parseRedisDeleteKeysPayload(payload, name)
		if parseErr != nil {
			return NativeMutationResult{}, parseErr
		}
		return client.DeleteKeys(ctx, keys)
	default:
		return NativeMutationResult{}, fmt.Errorf("不支持的 Redis 操作: %s", operation)
	}
}

func parseRedisDeleteKeysPayload(payload, fallbackName string) ([]string, error) {
	payload = strings.TrimSpace(payload)
	keys := make([]string, 0)
	if payload != "" && payload != "{}" {
		var parsed redisDeleteKeysPayload
		if err := json.Unmarshal([]byte(payload), &parsed); err != nil {
			return nil, fmt.Errorf("delete_keys payload 必须是合法 JSON: %w", err)
		}
		keys = append(keys, parsed.Keys...)
	}
	if name := strings.TrimSpace(fallbackName); name != "" {
		keys = append(keys, name)
	}
	unique := make([]string, 0, len(keys))
	seen := map[string]struct{}{}
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, key)
	}
	if len(unique) == 0 {
		return nil, fmt.Errorf("delete_keys 需要至少一个键名")
	}
	if len(unique) > redisBatchDeleteLimit {
		return nil, fmt.Errorf("单次最多删除 %d 个键", redisBatchDeleteLimit)
	}
	return unique, nil
}

func (s *redisNativeSession) ExecuteQuery(ctx context.Context, parent, _, query string) (NativeQueryResult, error) {
	client, cleanup, err := s.redisClientFor(ctx, parent)
	if err != nil {
		return NativeQueryResult{}, err
	}
	defer cleanup()
	envelope, command, err := parseRedisQuery(query)
	if err != nil {
		return NativeQueryResult{}, err
	}
	mode := strings.ToLower(strings.TrimSpace(envelope.Mode))
	if mode == "scan" || (mode == "" && envelope.Pattern != "" && command == "") {
		page, pageErr := s.ListSecondaryResourcesPage(ctx, parent, envelope.Pattern, envelope.Cursor, envelope.Count)
		if pageErr != nil {
			return NativeQueryResult{}, pageErr
		}
		content, marshalErr := json.Marshal(page)
		if marshalErr != nil {
			return NativeQueryResult{}, marshalErr
		}
		return NativeQueryResult{
			Summary: fmt.Sprintf("SCAN 返回 %d 个键", len(page.Items)),
			Content: string(content),
		}, nil
	}
	args, parseErr := tokenizeRedisCLI(command)
	if parseErr != nil {
		return NativeQueryResult{}, parseErr
	}
	readOnlyOnly := envelope.ReadOnly != nil && *envelope.ReadOnly
	if err := validateRedisCLI(args, readOnlyOnly); err != nil {
		return NativeQueryResult{}, err
	}
	return client.DoCommand(ctx, args)
}

func parseRedisQuery(query string) (redisQueryEnvelope, string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return redisQueryEnvelope{}, "", fmt.Errorf("Redis 查询不能为空")
	}
	if strings.HasPrefix(query, "{") {
		var envelope redisQueryEnvelope
		if err := json.Unmarshal([]byte(query), &envelope); err != nil {
			return redisQueryEnvelope{}, "", fmt.Errorf("Redis 查询 JSON 无效: %w", err)
		}
		command := strings.TrimSpace(envelope.Command)
		if command == "" && strings.EqualFold(envelope.Mode, "cli") {
			return envelope, "", fmt.Errorf("CLI 模式需要 command")
		}
		return envelope, command, nil
	}
	return redisQueryEnvelope{Mode: "cli"}, query, nil
}

func tokenizeRedisCLI(command string) ([]any, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return nil, fmt.Errorf("Redis 命令不能为空")
	}
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return nil, fmt.Errorf("Redis 命令不能为空")
	}
	args := make([]any, len(fields))
	for i, field := range fields {
		args[i] = field
	}
	return args, nil
}

func validateRedisCLI(args []any, readOnlyOnly bool) error {
	if len(args) == 0 {
		return fmt.Errorf("Redis 命令不能为空")
	}
	cmd := strings.ToUpper(fmt.Sprint(args[0]))
	for _, prefix := range redisCLIDeniedPrefixes {
		if strings.HasPrefix(cmd, prefix) {
			return fmt.Errorf("拒绝执行危险 Redis 命令: %s", cmd)
		}
	}
	if readOnlyOnly {
		if _, ok := redisCLIReadCommands[cmd]; !ok {
			return fmt.Errorf("只读模式下不允许命令: %s", cmd)
		}
	}
	return nil
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

func (c *redisGoClient) DeleteKeys(ctx context.Context, keys []string) (NativeMutationResult, error) {
	if len(keys) == 0 {
		return NativeMutationResult{}, fmt.Errorf("没有可删除的键")
	}
	deleted, err := c.client.Del(ctx, keys...).Result()
	if err != nil {
		return NativeMutationResult{}, err
	}
	return NativeMutationResult{Summary: fmt.Sprintf("已删除 %d 个键", deleted)}, nil
}

func (c *redisGoClient) DoCommand(ctx context.Context, args []any) (NativeQueryResult, error) {
	raw, err := c.client.Do(ctx, args...).Result()
	if err != nil && err != redis.Nil {
		return NativeQueryResult{}, err
	}
	content, marshalErr := json.Marshal(map[string]any{
		"command": args,
		"result":  redisCommandResult(raw),
	})
	if marshalErr != nil {
		return NativeQueryResult{}, marshalErr
	}
	return NativeQueryResult{
		Summary: fmt.Sprintf("已执行 %s", strings.ToUpper(fmt.Sprint(args[0]))),
		Content: string(content),
	}, nil
}

func redisCommandResult(value any) any {
	switch typed := value.(type) {
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = redisCommandResult(item)
		}
		return out
	case []byte:
		return string(typed)
	default:
		return typed
	}
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
	if err := writeRedisValue(ctx, pipe, name, parsed); err != nil {
		return NativeMutationResult{}, err
	}
	if parsed.Type != "string" && parsed.TTLSeconds != nil && *parsed.TTLSeconds > 0 {
		pipe.Expire(ctx, name, time.Duration(*parsed.TTLSeconds)*time.Second)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return NativeMutationResult{}, err
	}
	return NativeMutationResult{Summary: fmt.Sprintf("%s 键已保存", parsed.Type)}, nil
}

func writeRedisValue(ctx context.Context, pipe redis.Pipeliner, name string, parsed redisSavePayload) error {
	switch parsed.Type {
	case "string":
		pipe.Set(ctx, name, parsed.Value, redisDuration(parsed.TTLSeconds))
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
		if members := redisZSetMembers(parsed.Entries); len(members) > 0 {
			pipe.ZAdd(ctx, name, members...)
		}
	default:
		return fmt.Errorf("不支持的 Redis 类型: %s", parsed.Type)
	}
	return nil
}

func redisZSetMembers(entries []redisZSetEntry) []redis.Z {
	members := make([]redis.Z, 0, len(entries))
	for _, entry := range entries {
		if strings.TrimSpace(entry.Member) != "" {
			members = append(members, redis.Z{Score: entry.Score, Member: entry.Member})
		}
	}
	return members
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
		return c.readStringPreview(ctx, name, data)
	case "hash":
		return c.readHashPreview(ctx, name, data)
	case "list":
		return c.readListPreview(ctx, name, data)
	case "set":
		return c.readSetPreview(ctx, name, data)
	case "zset":
		return c.readZSetPreview(ctx, name, data)
	default:
		data["unsupportedPreview"] = true
	}
	return nil
}

func (c *redisGoClient) readStringPreview(ctx context.Context, name string, data map[string]any) error {
	value, err := c.client.Get(ctx, name).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	data["length"] = len(value)
	if len(value) > redisPreviewLimit {
		value = value[:redisPreviewLimit]
		data["truncated"] = true
	}
	data["value"] = value
	return nil
}

func (c *redisGoClient) readHashPreview(ctx context.Context, name string, data map[string]any) error {
	fields, err := c.client.HGetAll(ctx, name).Result()
	if err != nil {
		return err
	}
	data["length"] = len(fields)
	if len(fields) > redisCollectionPreviewLimit {
		data["truncated"] = true
	}
	data["fields"] = limitRedisMap(fields, redisCollectionPreviewLimit)
	return nil
}

func (c *redisGoClient) readListPreview(ctx context.Context, name string, data map[string]any) error {
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
	return nil
}

func (c *redisGoClient) readSetPreview(ctx context.Context, name string, data map[string]any) error {
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
	return nil
}

func (c *redisGoClient) readZSetPreview(ctx context.Context, name string, data map[string]any) error {
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

func redisKeySummary(kind string, ttl time.Duration) string {
	if ttl == -1 {
		return fmt.Sprintf("%s · 永不过期", kind)
	}
	if ttl == -2 {
		return fmt.Sprintf("%s · 已不存在", kind)
	}
	return fmt.Sprintf("%s · %d 秒后过期", kind, ttlSeconds(ttl))
}