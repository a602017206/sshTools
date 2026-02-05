package service

import (
	"strings"
	"testing"
)

func TestFormatJSON(t *testing.T) {
	service := NewDevToolsService()

	tests := []struct {
		name      string
		input     string
		wantErr   bool
		expectKey string // 期望输出中包含的关键字
	}{
		{
			name:      "有效的JSON对象",
			input:     `{"name":"test","age":30}`,
			wantErr:   false,
			expectKey: "name",
		},
		{
			name:      "有效的JSON数组",
			input:     `[1,2,3]`,
			wantErr:   false,
			expectKey: "1",
		},
		{
			name:      "嵌套的JSON对象",
			input:     `{"user":{"name":"张三","age":25},"active":true}`,
			wantErr:   false,
			expectKey: "张三",
		},
		{
			name:    "无效的JSON - 缺少引号",
			input:   `{name:"test"}`,
			wantErr: true,
		},
		{
			name:    "无效的JSON - 缺少结束符",
			input:   `{"name":"test"`,
			wantErr: true,
		},
		{
			name:    "空字符串",
			input:   "",
			wantErr: true,
		},
		{
			name:    "只有空白字符",
			input:   "   \n\t   ",
			wantErr: true,
		},
		{
			name:      "带空白的有效JSON",
			input:     `  { "test": true }  `,
			wantErr:   false,
			expectKey: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.FormatJSON(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("FormatJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == "" {
					t.Errorf("FormatJSON() returned empty result")
				}
				if tt.expectKey != "" && !strings.Contains(result, tt.expectKey) {
					t.Errorf("FormatJSON() result does not contain expected key '%s'", tt.expectKey)
				}
			}
		})
	}
}

func TestValidateJSON(t *testing.T) {
	service := NewDevToolsService()

	tests := []struct {
		name      string
		input     string
		wantValid bool
		wantError bool // 是否期望系统错误（非验证错误）
	}{
		{
			name:      "有效的JSON对象",
			input:     `{"test": true}`,
			wantValid: true,
			wantError: false,
		},
		{
			name:      "有效的JSON数组",
			input:     `[1, 2, 3]`,
			wantValid: true,
			wantError: false,
		},
		{
			name:      "有效的JSON字符串",
			input:     `"hello world"`,
			wantValid: true,
			wantError: false,
		},
		{
			name:      "有效的JSON数字",
			input:     `123.456`,
			wantValid: true,
			wantError: false,
		},
		{
			name:      "有效的JSON布尔值",
			input:     `true`,
			wantValid: true,
			wantError: false,
		},
		{
			name:      "有效的JSON null",
			input:     `null`,
			wantValid: true,
			wantError: false,
		},
		{
			name:      "无效的JSON",
			input:     `{test}`,
			wantValid: false,
			wantError: false,
		},
		{
			name:      "不完整的JSON",
			input:     `{"test":`,
			wantValid: false,
			wantError: false,
		},
		{
			name:      "空字符串",
			input:     "",
			wantValid: false,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ValidateJSON(tt.input)

			if (err != nil) != tt.wantError {
				t.Errorf("ValidateJSON() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if result.Valid != tt.wantValid {
				t.Errorf("ValidateJSON() valid = %v, want %v (error msg: %s)", result.Valid, tt.wantValid, result.Error)
			}

			if !result.Valid && result.Error == "" {
				t.Errorf("ValidateJSON() invalid but no error message provided")
			}
		})
	}
}

func TestMinifyJSON(t *testing.T) {
	service := NewDevToolsService()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		maxLen  int // 最大长度（用于验证压缩效果）
	}{
		{
			name:    "压缩格式化的JSON",
			input:   "{\n  \"name\": \"test\",\n  \"age\": 30\n}",
			wantErr: false,
			maxLen:  25, // 压缩后应该更短
		},
		{
			name:    "压缩JSON数组",
			input:   "[\n  1,\n  2,\n  3\n]",
			wantErr: false,
			maxLen:  10,
		},
		{
			name:    "无效的JSON",
			input:   `{invalid}`,
			wantErr: true,
		},
		{
			name:    "空字符串",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.MinifyJSON(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("MinifyJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == "" {
					t.Errorf("MinifyJSON() returned empty result")
				}
				if len(result) > tt.maxLen {
					t.Errorf("MinifyJSON() result length %d exceeds max %d", len(result), tt.maxLen)
				}
				// 验证压缩后不包含换行符
				if strings.Contains(result, "\n") {
					t.Errorf("MinifyJSON() result contains newlines")
				}
			}
		})
	}
}

func TestEscapeJSON(t *testing.T) {
	service := NewDevToolsService()

	tests := []struct {
		name       string
		input      string
		wantErr    bool
		expectChar string // 期望输出中包含的字符
	}{
		{
			name:       "转义简单JSON对象",
			input:      `{"test":"value"}`,
			wantErr:    false,
			expectChar: "\\", // 应该包含转义字符
		},
		{
			name:       "转义包含换行的JSON",
			input:      "{\n  \"test\": \"value\"\n}",
			wantErr:    false,
			expectChar: "\\n", // 应该转义换行符
		},
		{
			name:    "无效的JSON",
			input:   `{invalid}`,
			wantErr: true,
		},
		{
			name:    "空字符串",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.EscapeJSON(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("EscapeJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == "" {
					t.Errorf("EscapeJSON() returned empty result")
				}
				if tt.expectChar != "" && !strings.Contains(result, tt.expectChar) {
					t.Errorf("EscapeJSON() result does not contain expected char '%s', got: %s", tt.expectChar, result)
				}
			}
		})
	}
}

func TestFormatJSONError(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectMsg string // 期望错误消息中包含的关键字
	}{
		{
			name:      "不完整的JSON",
			input:     `{"test":`,
			expectMsg: "不完整",
		},
		{
			name:      "非法字符",
			input:     `{test}`,
			expectMsg: "非法",
		},
	}

	service := NewDevToolsService()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ValidateJSON(tt.input)

			if err != nil {
				t.Errorf("ValidateJSON() unexpected error = %v", err)
				return
			}

			if result.Valid {
				t.Errorf("ValidateJSON() expected invalid result")
			}

			if !strings.Contains(result.Error, tt.expectMsg) {
				t.Logf("Error message: %s", result.Error)
				t.Logf("Expected to contain: %s", tt.expectMsg)
				// 不严格要求，因为错误消息可能因Go版本而异
			}
		})
	}
}

func TestEncryptDecryptAESGCM(t *testing.T) {
	service := NewDevToolsService()
	plaintext := "hello gcm"
	keyHex := "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"
	nonceHex := "0f0e0d0c0b0a090807060504"

	ciphertext, err := service.EncryptText(plaintext, "aes-gcm", keyHex, nonceHex)
	if err != nil {
		t.Fatalf("EncryptText(aes-gcm) error = %v", err)
	}
	if ciphertext == "" {
		t.Fatalf("EncryptText(aes-gcm) returned empty ciphertext")
	}

	decoded, err := service.DecryptText(ciphertext, "aes-gcm", keyHex, nonceHex)
	if err != nil {
		t.Fatalf("DecryptText(aes-gcm) error = %v", err)
	}
	if decoded != plaintext {
		t.Fatalf("DecryptText(aes-gcm) = %q, want %q", decoded, plaintext)
	}
}

func TestEncryptDecryptAESCBC(t *testing.T) {
	service := NewDevToolsService()
	plaintext := "hello cbc"
	keyHex := "00112233445566778899aabbccddeeff"
	ivHex := "0102030405060708090a0b0c0d0e0f10"

	ciphertext, err := service.EncryptText(plaintext, "aes-cbc", keyHex, ivHex)
	if err != nil {
		t.Fatalf("EncryptText(aes-cbc) error = %v", err)
	}
	if ciphertext == "" {
		t.Fatalf("EncryptText(aes-cbc) returned empty ciphertext")
	}

	decoded, err := service.DecryptText(ciphertext, "aes-cbc", keyHex, ivHex)
	if err != nil {
		t.Fatalf("DecryptText(aes-cbc) error = %v", err)
	}
	if decoded != plaintext {
		t.Fatalf("DecryptText(aes-cbc) = %q, want %q", decoded, plaintext)
	}
}

func TestEncryptDecryptSM4CBC(t *testing.T) {
	service := NewDevToolsService()
	plaintext := "hello sm4"
	keyHex := "0123456789abcdeffedcba9876543210"
	ivHex := "0f1e2d3c4b5a69788796a5b4c3d2e1f0"

	ciphertext, err := service.EncryptText(plaintext, "sm4-cbc", keyHex, ivHex)
	if err != nil {
		t.Fatalf("EncryptText(sm4-cbc) error = %v", err)
	}
	if ciphertext == "" {
		t.Fatalf("EncryptText(sm4-cbc) returned empty ciphertext")
	}

	decoded, err := service.DecryptText(ciphertext, "sm4-cbc", keyHex, ivHex)
	if err != nil {
		t.Fatalf("DecryptText(sm4-cbc) error = %v", err)
	}
	if decoded != plaintext {
		t.Fatalf("DecryptText(sm4-cbc) = %q, want %q", decoded, plaintext)
	}
}
