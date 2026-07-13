package service

import (
	"fmt"
	"strings"
	"sync"
)

type NativeDatabaseRegistry struct {
	providers map[NativeDatabaseType]NativeDatabaseProvider
	mu        sync.RWMutex
}

func NewNativeDatabaseRegistry() *NativeDatabaseRegistry {
	return &NativeDatabaseRegistry{providers: make(map[NativeDatabaseType]NativeDatabaseProvider)}
}

func (r *NativeDatabaseRegistry) Register(databaseType NativeDatabaseType, provider NativeDatabaseProvider) error {
	databaseType = NativeDatabaseType(strings.TrimSpace(string(databaseType)))
	if databaseType == "" {
		return fmt.Errorf("原生数据库 provider ID 不能为空")
	}
	if provider == nil {
		return fmt.Errorf("原生数据库 provider %s 不能为空", databaseType)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.providers[databaseType]; exists {
		return fmt.Errorf("原生数据库 provider 已注册: %s", databaseType)
	}
	r.providers[databaseType] = provider
	return nil
}

func (r *NativeDatabaseRegistry) Provider(databaseType NativeDatabaseType) (NativeDatabaseProvider, error) {
	r.mu.RLock()
	provider, exists := r.providers[databaseType]
	r.mu.RUnlock()
	if !exists || provider == nil {
		return nil, fmt.Errorf("未注册的原生数据库 provider: %s", databaseType)
	}
	return provider, nil
}

func (r *NativeDatabaseRegistry) Providers() map[NativeDatabaseType]NativeDatabaseProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	providers := make(map[NativeDatabaseType]NativeDatabaseProvider, len(r.providers))
	for databaseType, provider := range r.providers {
		providers[databaseType] = provider
	}
	return providers
}
