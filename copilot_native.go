package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"AHaSSHTools/internal/service"
	"AHaSSHTools/internal/service/copilot"
)

type nativeCopilotReader struct {
	svc *service.NativeDatabaseService
}

func (r nativeCopilotReader) ListResources(sessionID string) ([]copilot.NativeResourceInfo, error) {
	if r.svc == nil {
		return nil, fmt.Errorf("native database service unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	items, err := r.svc.ListPrimaryResources(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return toCopilotNativeResources(items), nil
}

func (r nativeCopilotReader) ListChildResources(sessionID, parent string) ([]copilot.NativeResourceInfo, error) {
	if r.svc == nil {
		return nil, fmt.Errorf("native database service unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	items, err := r.svc.ListSecondaryResources(ctx, sessionID, parent)
	if err != nil {
		return nil, err
	}
	return toCopilotNativeResources(items), nil
}

func (r nativeCopilotReader) DescribeResource(sessionID, parent, name string) (*copilot.NativeResourceView, error) {
	if r.svc == nil {
		return nil, fmt.Errorf("native database service unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	details, err := r.svc.DescribeResource(ctx, sessionID, parent, name)
	if err != nil {
		return nil, err
	}
	return &copilot.NativeResourceView{
		Kind:    string(details.Kind),
		Name:    details.Name,
		Summary: details.Summary,
		Content: details.Content,
	}, nil
}

func (r nativeCopilotReader) ExecuteQuery(sessionID, parent, name, query string) (string, error) {
	if r.svc == nil {
		return "", fmt.Errorf("native database service unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	result, err := r.svc.ExecuteQuery(ctx, sessionID, parent, name, query)
	if err != nil {
		return "", err
	}
	raw, marshalErr := json.Marshal(map[string]string{
		"summary": result.Summary,
		"content": result.Content,
	})
	if marshalErr != nil {
		return result.Summary + "\n" + result.Content, nil
	}
	return string(raw), nil
}

func toCopilotNativeResources(items []service.NativeResource) []copilot.NativeResourceInfo {
	out := make([]copilot.NativeResourceInfo, 0, len(items))
	for _, item := range items {
		out = append(out, copilot.NativeResourceInfo{Kind: string(item.Kind), Name: item.Name})
	}
	return out
}
