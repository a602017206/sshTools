package service

import (
	"context"
	"errors"
	"fmt"
)

var ErrNativeOperationUnsupported = errors.New("当前原生数据库类型不支持该操作")

// NativeQueryResult carries structured query output as JSON in Content.
type NativeQueryResult struct {
	Summary string `json:"summary"`
	Content string `json:"content"`
}

// NativeMutationResult describes the outcome of a write/delete operation.
type NativeMutationResult struct {
	Summary string `json:"summary"`
	Content string `json:"content"`
}

type NativeQueryExecutor interface {
	ExecuteQuery(context.Context, string, string, string) (NativeQueryResult, error)
}

type NativeResourceMutator interface {
	MutateResource(context.Context, string, string, string, string) (NativeMutationResult, error)
}

func (s *NativeDatabaseService) ExecuteQuery(ctx context.Context, sessionID, parent, name, query string) (NativeQueryResult, error) {
	session, err := s.session(sessionID)
	if err != nil {
		return NativeQueryResult{}, err
	}
	executor, ok := session.client.(NativeQueryExecutor)
	if !ok {
		return NativeQueryResult{}, fmt.Errorf("%w: 查询", ErrNativeOperationUnsupported)
	}
	result, err := executor.ExecuteQuery(ctx, parent, name, query)
	if err != nil {
		return NativeQueryResult{}, fmt.Errorf("执行 %s 查询失败: %w", nativeDatabaseTypeName(session.Config.Type), err)
	}
	return result, nil
}

func (s *NativeDatabaseService) MutateResource(ctx context.Context, sessionID, parent, name, operation, payload string) (NativeMutationResult, error) {
	session, err := s.session(sessionID)
	if err != nil {
		return NativeMutationResult{}, err
	}
	mutator, ok := session.client.(NativeResourceMutator)
	if !ok {
		return NativeMutationResult{}, fmt.Errorf("%w: 变更", ErrNativeOperationUnsupported)
	}
	result, err := mutator.MutateResource(ctx, parent, name, operation, payload)
	if err != nil {
		return NativeMutationResult{}, fmt.Errorf("执行 %s 变更失败: %w", nativeDatabaseTypeName(session.Config.Type), err)
	}
	return result, nil
}
