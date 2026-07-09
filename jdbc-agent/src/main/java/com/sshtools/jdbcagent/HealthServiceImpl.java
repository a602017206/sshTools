package com.sshtools.jdbcagent;

import com.sshtools.jdbcagent.proto.HealthRequest;
import com.sshtools.jdbcagent.proto.HealthResponse;
import com.sshtools.jdbcagent.proto.JdbcAgentGrpc;
import io.grpc.Status;
import io.grpc.stub.StreamObserver;

public class HealthServiceImpl extends JdbcAgentGrpc.JdbcAgentImplBase {
    private static final String AGENT_VERSION = "0.1.0";

    private final String token;

    public HealthServiceImpl(String token) {
        this.token = token;
    }

    @Override
    public void health(HealthRequest request, StreamObserver<HealthResponse> responseObserver) {
        if (!token.equals(request.getToken())) {
            responseObserver.onError(Status.UNAUTHENTICATED
                    .withDescription("invalid token")
                    .asRuntimeException());
            return;
        }

        responseObserver.onNext(HealthResponse.newBuilder()
                .setStatus("OK")
                .setAgentVersion(AGENT_VERSION)
                .setJavaVersion(System.getProperty("java.version", "unknown"))
                .build());
        responseObserver.onCompleted();
    }
}
