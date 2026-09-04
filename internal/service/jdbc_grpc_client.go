package service

import (
	"context"
	"fmt"
	"net"

	"AHaSSHTools/internal/service/jdbcproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type grpcJdbcAgentClient struct {
	client jdbcproto.JdbcAgentClient
}

func NewGRPCJdbcAgentClient(ctx context.Context, host string, port int) (JdbcAgentClient, func() error, error) {
	target := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := grpc.DialContext(ctx, target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(64<<20),
			grpc.MaxCallSendMsgSize(64<<20),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("连接 JDBC agent gRPC 失败: %w", err)
	}
	return grpcJdbcAgentClient{client: jdbcproto.NewJdbcAgentClient(conn)}, conn.Close, nil
}

func (c grpcJdbcAgentClient) OpenSession(ctx context.Context, request *jdbcproto.OpenSessionRequest) (*jdbcproto.OpenSessionResponse, error) {
	return c.client.OpenSession(ctx, request)
}

func (c grpcJdbcAgentClient) ExecuteQuery(ctx context.Context, request *jdbcproto.ExecuteQueryRequest) (*jdbcproto.QueryResult, error) {
	return c.client.ExecuteQuery(ctx, request)
}

func (c grpcJdbcAgentClient) ListSchemas(ctx context.Context, request *jdbcproto.ListSchemasRequest) (*jdbcproto.ListSchemasResponse, error) {
	return c.client.ListSchemas(ctx, request)
}

func (c grpcJdbcAgentClient) ListRoutines(ctx context.Context, request *jdbcproto.ListRoutinesRequest) (*jdbcproto.ListRoutinesResponse, error) {
	return c.client.ListRoutines(ctx, request)
}

func (c grpcJdbcAgentClient) ListTables(ctx context.Context, request *jdbcproto.ListTablesRequest) (*jdbcproto.ListTablesResponse, error) {
	return c.client.ListTables(ctx, request)
}

func (c grpcJdbcAgentClient) ListColumns(ctx context.Context, request *jdbcproto.ListColumnsRequest) (*jdbcproto.ListColumnsResponse, error) {
	return c.client.ListColumns(ctx, request)
}

func (c grpcJdbcAgentClient) CloseSession(ctx context.Context, request *jdbcproto.CloseSessionRequest) (*jdbcproto.CloseSessionResponse, error) {
	return c.client.CloseSession(ctx, request)
}
