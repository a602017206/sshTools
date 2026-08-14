package copilot

import (
	"fmt"
	"strings"
)

func ValidateConfig(baseURL, apiKey string) error {
	if strings.TrimSpace(baseURL) == "" || strings.TrimSpace(apiKey) == "" {
		return fmt.Errorf("请先在设置中填写 Base URL 和 API Key")
	}
	return nil
}
