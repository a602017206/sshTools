package service

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DevToolsService 提供开发工具相关服务
type DevToolsService struct{}

// NewDevToolsService 创建新的DevTools服务实例
func NewDevToolsService() *DevToolsService {
	return &DevToolsService{}
}

// JSONValidationResult JSON验证结果
type JSONValidationResult struct {
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}

// FormatJSON 格式化JSON字符串
// 参数：
//   - input: 待格式化的JSON字符串
//
// 返回：
//   - string: 格式化后的JSON字符串（4空格缩进）
//   - error: 错误信息
func (s *DevToolsService) FormatJSON(input string) (string, error) {
	// 移除首尾空白
	input = strings.TrimSpace(input)

	if input == "" {
		return "", fmt.Errorf("输入不能为空")
	}

	// 解析JSON
	var obj interface{}
	err := json.Unmarshal([]byte(input), &obj)
	if err != nil {
		// 提供更友好的错误信息
		return "", fmt.Errorf("JSON解析失败: %v", formatJSONError(err))
	}

	// 格式化输出（4个空格缩进）
	formatted, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		return "", fmt.Errorf("JSON格式化失败: %v", err)
	}

	return string(formatted), nil
}

// ValidateJSON 验证JSON字符串的有效性
// 参数：
//   - input: 待验证的JSON字符串
//
// 返回：
//   - JSONValidationResult: 验证结果
//   - error: 系统错误（非JSON格式错误）
func (s *DevToolsService) ValidateJSON(input string) (JSONValidationResult, error) {
	input = strings.TrimSpace(input)

	if input == "" {
		return JSONValidationResult{
			Valid: false,
			Error: "输入不能为空",
		}, nil
	}

	var obj interface{}
	err := json.Unmarshal([]byte(input), &obj)
	if err != nil {
		return JSONValidationResult{
			Valid: false,
			Error: formatJSONError(err),
		}, nil
	}

	return JSONValidationResult{
		Valid: true,
	}, nil
}

// MinifyJSON 压缩JSON（移除所有不必要的空白）
// 参数：
//   - input: 待压缩的JSON字符串
//
// 返回：
//   - string: 压缩后的JSON字符串
//   - error: 错误信息
func (s *DevToolsService) MinifyJSON(input string) (string, error) {
	input = strings.TrimSpace(input)

	if input == "" {
		return "", fmt.Errorf("输入不能为空")
	}

	var obj interface{}
	err := json.Unmarshal([]byte(input), &obj)
	if err != nil {
		return "", fmt.Errorf("JSON解析失败: %v", formatJSONError(err))
	}

	// 紧凑输出
	minified, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("JSON压缩失败: %v", err)
	}

	return string(minified), nil
}

// EscapeJSON 转义JSON字符串（用于在其他JSON中嵌套）
// 参数：
//   - input: 待转义的JSON字符串
//
// 返回：
//   - string: 转义后的字符串
//   - error: 错误信息
func (s *DevToolsService) EscapeJSON(input string) (string, error) {
	input = strings.TrimSpace(input)

	if input == "" {
		return "", fmt.Errorf("输入不能为空")
	}

	// 验证输入是有效的JSON
	var obj interface{}
	err := json.Unmarshal([]byte(input), &obj)
	if err != nil {
		return "", fmt.Errorf("JSON解析失败: %v", formatJSONError(err))
	}

	// 转换为转义的字符串
	escaped, err := json.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("转义失败: %v", err)
	}

	return string(escaped), nil
}

// formatJSONError 格式化JSON错误信息，使其更友好
func formatJSONError(err error) string {
	if err == nil {
		return ""
	}

	errMsg := err.Error()

	// 如果错误消息已包含行号信息，直接返回
	if strings.Contains(errMsg, "line") || strings.Contains(errMsg, "offset") {
		return errMsg
	}

	// 简化常见错误
	switch {
	case strings.Contains(errMsg, "unexpected end of JSON input"):
		return "JSON不完整，缺少结束符号"
	case strings.Contains(errMsg, "invalid character"):
		// 尝试提取字符信息
		if idx := strings.Index(errMsg, "invalid character"); idx >= 0 {
			return errMsg[idx:]
		}
		return "包含非法字符"
	case strings.Contains(errMsg, "unexpected EOF"):
		return "JSON意外结束"
	default:
		return fmt.Sprintf("格式错误: %s", errMsg)
	}
}

// ============================================================================
// Base64 编解码
// ============================================================================

// EncodeBase64 将字符串编码为 Base64
func (s *DevToolsService) EncodeBase64(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("输入不能为空")
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(input))
	return encoded, nil
}

// DecodeBase64 将 Base64 字符串解码
func (s *DevToolsService) DecodeBase64(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("输入不能为空")
	}
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %v", err)
	}
	return string(decoded), nil
}

// ============================================================================
// Hash 计算
// ============================================================================

// CalculateHash 计算字符串的哈希值
func (s *DevToolsService) CalculateHash(input, algorithm string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("输入不能为空")
	}

	var hash []byte

	switch strings.ToLower(algorithm) {
	case "md5":
		h := md5.Sum([]byte(input))
		hash = h[:]
	case "sha256":
		h := sha256.Sum256([]byte(input))
		hash = h[:]
	case "sha512":
		h := sha512.Sum512([]byte(input))
		hash = h[:]
	default:
		return "", fmt.Errorf("不支持的哈希算法: %s (支持: md5, sha256, sha512)", algorithm)
	}

	return hex.EncodeToString(hash), nil
}

// ============================================================================
// 时间戳转换
// ============================================================================

// TimestampToDateTime 将 Unix 时间戳转换为日期时间字符串
func (s *DevToolsService) TimestampToDateTime(timestamp int64, format string) (string, error) {
	if timestamp <= 0 {
		return "", fmt.Errorf("无效的时间戳: %d", timestamp)
	}

	if format == "" {
		format = "2006-01-02 15:04:05"
	}

	t := time.Unix(timestamp, 0)
	return t.Format(format), nil
}

// DateTimeToTimestamp 将日期时间字符串转换为 Unix 时间戳
func (s *DevToolsService) DateTimeToTimestamp(datetime, format string) (int64, error) {
	if datetime == "" {
		return 0, fmt.Errorf("输入不能为空")
	}

	if format == "" {
		format = "2006-01-02 15:04:05"
	}

	t, err := time.Parse(format, datetime)
	if err != nil {
		return 0, fmt.Errorf("日期时间解析失败: %v", err)
	}

	return t.Unix(), nil
}

// GetCurrentTimestamp 获取当前 Unix 时间戳
func (s *DevToolsService) GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// ============================================================================
// UUID 生成
// ============================================================================

// GenerateUUIDv4 生成 UUID v4
func (s *DevToolsService) GenerateUUIDv4() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("生成UUID失败: %v", err)
	}
	return id.String(), nil
}
