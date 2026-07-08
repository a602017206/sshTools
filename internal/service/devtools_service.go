package service

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tjfoc/gmsm/sm4"
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
// 加密/解密
// ============================================================================

// EncryptText 对文本进行加密，返回 Base64 密文
func (s *DevToolsService) EncryptText(input, algorithm, keyHex, ivHex string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("输入不能为空")
	}

	key, err := decodeHexBytes(keyHex, "密钥")
	if err != nil {
		return "", err
	}

	iv, err := decodeHexBytes(ivHex, "IV/Nonce")
	if err != nil {
		return "", err
	}

	plaintext := []byte(input)

	switch strings.ToLower(algorithm) {
	case "aes-gcm":
		if err := validateKeyLength(key, []int{16, 24, 32}, "AES密钥"); err != nil {
			return "", err
		}
		block, err := aes.NewCipher(key)
		if err != nil {
			return "", fmt.Errorf("AES初始化失败: %v", err)
		}
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return "", fmt.Errorf("AES-GCM初始化失败: %v", err)
		}
		if len(iv) != gcm.NonceSize() {
			return "", fmt.Errorf("Nonce长度必须为 %d 字节", gcm.NonceSize())
		}
		ciphertext := gcm.Seal(nil, iv, plaintext, nil)
		return base64.StdEncoding.EncodeToString(ciphertext), nil
	case "aes-cbc":
		if err := validateKeyLength(key, []int{16, 24, 32}, "AES密钥"); err != nil {
			return "", err
		}
		if len(iv) != aes.BlockSize {
			return "", fmt.Errorf("IV长度必须为 %d 字节", aes.BlockSize)
		}
		block, err := aes.NewCipher(key)
		if err != nil {
			return "", fmt.Errorf("AES初始化失败: %v", err)
		}
		padded := pkcs7Pad(plaintext, aes.BlockSize)
		ciphertext := make([]byte, len(padded))
		mode := cipher.NewCBCEncrypter(block, iv)
		mode.CryptBlocks(ciphertext, padded)
		return base64.StdEncoding.EncodeToString(ciphertext), nil
	case "sm4-cbc":
		if err := validateKeyLength(key, []int{16}, "SM4密钥"); err != nil {
			return "", err
		}
		if len(iv) != sm4.BlockSize {
			return "", fmt.Errorf("IV长度必须为 %d 字节", sm4.BlockSize)
		}
		block, err := sm4.NewCipher(key)
		if err != nil {
			return "", fmt.Errorf("SM4初始化失败: %v", err)
		}
		padded := pkcs7Pad(plaintext, sm4.BlockSize)
		ciphertext := make([]byte, len(padded))
		mode := cipher.NewCBCEncrypter(block, iv)
		mode.CryptBlocks(ciphertext, padded)
		return base64.StdEncoding.EncodeToString(ciphertext), nil
	default:
		return "", fmt.Errorf("不支持的算法: %s", algorithm)
	}
}

// DecryptText 对 Base64 密文进行解密
func (s *DevToolsService) DecryptText(input, algorithm, keyHex, ivHex string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("输入不能为空")
	}

	key, err := decodeHexBytes(keyHex, "密钥")
	if err != nil {
		return "", err
	}

	iv, err := decodeHexBytes(ivHex, "IV/Nonce")
	if err != nil {
		return "", err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %v", err)
	}

	switch strings.ToLower(algorithm) {
	case "aes-gcm":
		if err := validateKeyLength(key, []int{16, 24, 32}, "AES密钥"); err != nil {
			return "", err
		}
		block, err := aes.NewCipher(key)
		if err != nil {
			return "", fmt.Errorf("AES初始化失败: %v", err)
		}
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return "", fmt.Errorf("AES-GCM初始化失败: %v", err)
		}
		if len(iv) != gcm.NonceSize() {
			return "", fmt.Errorf("Nonce长度必须为 %d 字节", gcm.NonceSize())
		}
		plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
		if err != nil {
			return "", fmt.Errorf("解密失败: %v", err)
		}
		return string(plaintext), nil
	case "aes-cbc":
		if err := validateKeyLength(key, []int{16, 24, 32}, "AES密钥"); err != nil {
			return "", err
		}
		if len(iv) != aes.BlockSize {
			return "", fmt.Errorf("IV长度必须为 %d 字节", aes.BlockSize)
		}
		block, err := aes.NewCipher(key)
		if err != nil {
			return "", fmt.Errorf("AES初始化失败: %v", err)
		}
		if len(ciphertext)%aes.BlockSize != 0 {
			return "", fmt.Errorf("密文长度不正确")
		}
		plaintext := make([]byte, len(ciphertext))
		mode := cipher.NewCBCDecrypter(block, iv)
		mode.CryptBlocks(plaintext, ciphertext)
		unpadded, err := pkcs7Unpad(plaintext, aes.BlockSize)
		if err != nil {
			return "", err
		}
		return string(unpadded), nil
	case "sm4-cbc":
		if err := validateKeyLength(key, []int{16}, "SM4密钥"); err != nil {
			return "", err
		}
		if len(iv) != sm4.BlockSize {
			return "", fmt.Errorf("IV长度必须为 %d 字节", sm4.BlockSize)
		}
		block, err := sm4.NewCipher(key)
		if err != nil {
			return "", fmt.Errorf("SM4初始化失败: %v", err)
		}
		if len(ciphertext)%sm4.BlockSize != 0 {
			return "", fmt.Errorf("密文长度不正确")
		}
		plaintext := make([]byte, len(ciphertext))
		mode := cipher.NewCBCDecrypter(block, iv)
		mode.CryptBlocks(plaintext, ciphertext)
		unpadded, err := pkcs7Unpad(plaintext, sm4.BlockSize)
		if err != nil {
			return "", err
		}
		return string(unpadded), nil
	default:
		return "", fmt.Errorf("不支持的算法: %s", algorithm)
	}
}

func decodeHexBytes(input, name string) ([]byte, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("%s不能为空", name)
	}
	data, err := hex.DecodeString(input)
	if err != nil {
		return nil, fmt.Errorf("%s必须为十六进制: %v", name, err)
	}
	return data, nil
}

func validateKeyLength(key []byte, sizes []int, name string) error {
	for _, size := range sizes {
		if len(key) == size {
			return nil
		}
	}
	return fmt.Errorf("%s长度必须为 %v 字节", name, sizes)
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	return append(data, bytes.Repeat([]byte{byte(padding)}, padding)...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, fmt.Errorf("填充不正确")
	}
	padding := int(data[len(data)-1])
	if padding == 0 || padding > blockSize {
		return nil, fmt.Errorf("填充不正确")
	}
	for i := 0; i < padding; i++ {
		if data[len(data)-1-i] != byte(padding) {
			return nil, fmt.Errorf("填充不正确")
		}
	}
	return data[:len(data)-padding], nil
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

// TimestampToDateTimeMs 将 Unix 毫秒时间戳转换为日期时间字符串
func (s *DevToolsService) TimestampToDateTimeMs(timestampMs int64, format string) (string, error) {
	if timestampMs <= 0 {
		return "", fmt.Errorf("无效的毫秒时间戳: %d", timestampMs)
	}

	if format == "" {
		format = "2006-01-02 15:04:05.000"
	}

	t := time.Unix(0, timestampMs*int64(time.Millisecond))
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

// DateTimeToTimestampMs 将日期时间字符串转换为 Unix 毫秒时间戳
func (s *DevToolsService) DateTimeToTimestampMs(datetime, format string) (int64, error) {
	if datetime == "" {
		return 0, fmt.Errorf("输入不能为空")
	}

	if format == "" {
		format = "2006-01-02 15:04:05.000"
	}

	t, err := time.Parse(format, datetime)
	if err != nil {
		return 0, fmt.Errorf("日期时间解析失败: %v", err)
	}

	return t.UnixNano() / int64(time.Millisecond), nil
}

// GetCurrentTimestamp 获取当前 Unix 时间戳
func (s *DevToolsService) GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// GetCurrentTimestampMs 获取当前 Unix 毫秒时间戳
func (s *DevToolsService) GetCurrentTimestampMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
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

// ============================================================================
// URL 编解码
// ============================================================================

// URLEncodeResult URL编码结果
type URLEncodeResult struct {
	Encoded   string `json:"encoded"`
	FullURL   string `json:"fullUrl,omitempty"`
	Component string `json:"component,omitempty"`
}

// URLDecodeResult URL解码结果
type URLDecodeResult struct {
	Decoded   string            `json:"decoded"`
	Params    map[string]string `json:"params,omitempty"`
	Component string            `json:"component,omitempty"`
}

// URLEncode 对字符串进行 URL 编码
// mode: "path" - 编码路径部分, "query" - 编码查询参数值, "fragment" - 编码片段, "full" - 编码整个URL组件
func (s *DevToolsService) URLEncode(input, mode string) (URLEncodeResult, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return URLEncodeResult{}, fmt.Errorf("输入不能为空")
	}

	result := URLEncodeResult{}

	switch mode {
	case "query":
		// 编码查询参数值（空格转为 +）
		result.Encoded = url.QueryEscape(input)
		result.Component = "query"
	case "path":
		// 编码路径部分
		result.Encoded = url.PathEscape(input)
		result.Component = "path"
	case "fragment":
		// 编码URL片段
		result.Encoded = url.QueryEscape(input)
		result.Component = "fragment"
	case "full":
		// 尝试解析为完整URL，然后重新编码
		u, err := url.Parse(input)
		if err != nil {
			// 不是完整URL，作为普通字符串编码
			result.Encoded = url.QueryEscape(input)
		} else {
			result.FullURL = u.String()
			result.Encoded = u.String()
		}
		result.Component = "full"
	default:
		// 默认使用 QueryEscape
		result.Encoded = url.QueryEscape(input)
		result.Component = "query"
	}

	return result, nil
}

// URLDecode 对 URL 编码的字符串进行解码
func (s *DevToolsService) URLDecode(input, mode string) (URLDecodeResult, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return URLDecodeResult{}, fmt.Errorf("输入不能为空")
	}

	result := URLDecodeResult{}

	var decoded string
	var err error

	switch mode {
	case "query":
		// 解码查询参数值（+ 转为空格）
		decoded, err = url.QueryUnescape(input)
		result.Component = "query"
	case "path":
		// 解码路径部分
		decoded, err = url.PathUnescape(input)
		result.Component = "path"
	case "fragment":
		decoded, err = url.QueryUnescape(input)
		result.Component = "fragment"
	case "full":
		// 尝试解析完整URL
		u, parseErr := url.Parse(input)
		if parseErr == nil {
			// 解析查询参数
			result.Params = make(map[string]string)
			for key, values := range u.Query() {
				if len(values) > 0 {
					result.Params[key] = values[0]
				}
			}
		}
		decoded, err = url.QueryUnescape(input)
		result.Component = "full"
	default:
		// 默认尝试 QueryUnescape
		decoded, err = url.QueryUnescape(input)
		result.Component = "query"
	}

	if err != nil {
		return URLDecodeResult{}, fmt.Errorf("URL解码失败: %v", err)
	}

	result.Decoded = decoded
	return result, nil
}

// ParseURL 解析 URL 返回各个组成部分
func (s *DevToolsService) ParseURL(input string) (map[string]interface{}, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("输入不能为空")
	}

	u, err := url.Parse(input)
	if err != nil {
		return nil, fmt.Errorf("URL解析失败: %v", err)
	}

	result := map[string]interface{}{
		"scheme":   u.Scheme,
		"host":     u.Host,
		"hostname": u.Hostname(),
		"port":     u.Port(),
		"path":     u.Path,
		"rawPath":  u.RawPath,
		"rawQuery": u.RawQuery,
		"fragment": u.Fragment,
	}

	// 解析查询参数
	queryParams := make(map[string][]string)
	for key, values := range u.Query() {
		queryParams[key] = values
	}
	result["queryParams"] = queryParams

	return result, nil
}
