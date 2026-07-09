package com.sshtools.jdbcagent;

import com.sshtools.jdbcagent.proto.CloseSessionRequest;
import com.sshtools.jdbcagent.proto.CloseSessionResponse;
import com.sshtools.jdbcagent.proto.Column;
import com.sshtools.jdbcagent.proto.ListColumnsRequest;
import com.sshtools.jdbcagent.proto.ListColumnsResponse;
import com.sshtools.jdbcagent.proto.ListTablesRequest;
import com.sshtools.jdbcagent.proto.ListTablesResponse;
import io.grpc.Status;
import io.grpc.stub.StreamObserver;

import java.sql.Connection;
import java.sql.DatabaseMetaData;
import java.sql.ResultSet;
import java.util.HashSet;
import java.util.Set;

public class MetadataServiceImpl extends QueryServiceImpl {
    private final ConnectionRegistry registry;

    public MetadataServiceImpl(String token, ConnectionRegistry registry, DriverLoader driverLoader) {
        super(token, registry, driverLoader);
        this.registry = registry;
    }

    @Override
    public void listTables(ListTablesRequest request, StreamObserver<ListTablesResponse> responseObserver) {
        if (!isAuthorized(request.getToken(), responseObserver)) {
            return;
        }

        try {
            Connection connection = registry.get(request.getSessionId());
            DatabaseMetaData metaData = connection.getMetaData();
            ListTablesResponse.Builder response = ListTablesResponse.newBuilder();
            try (ResultSet tables = metaData.getTables(
                    emptyToNull(request.getCatalog()),
                    emptyToNull(request.getSchema()),
                    null,
                    new String[]{"TABLE"})) {
                while (tables.next()) {
                    response.addTables(tables.getString("TABLE_NAME"));
                }
            }
            responseObserver.onNext(response.build());
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(Status.fromThrowable(e)
                    .withDescription(e.getMessage())
                    .withCause(e)
                    .asRuntimeException());
        }
    }

    @Override
    public void listColumns(ListColumnsRequest request, StreamObserver<ListColumnsResponse> responseObserver) {
        if (!isAuthorized(request.getToken(), responseObserver)) {
            return;
        }

        try {
            Connection connection = registry.get(request.getSessionId());
            DatabaseMetaData metaData = connection.getMetaData();
            Set<String> primaryKeys = loadPrimaryKeys(metaData, request);
            ListColumnsResponse.Builder response = ListColumnsResponse.newBuilder();
            try (ResultSet columns = metaData.getColumns(
                    emptyToNull(request.getCatalog()),
                    emptyToNull(request.getSchema()),
                    request.getTable(),
                    null)) {
                while (columns.next()) {
                    String name = columns.getString("COLUMN_NAME");
                    response.addColumns(Column.newBuilder()
                            .setName(name)
                            .setType(columns.getString("TYPE_NAME"))
                            .setNullable(columns.getInt("NULLABLE") == DatabaseMetaData.columnNullable)
                            .setPrimaryKey(primaryKeys.contains(name))
                            .build());
                }
            }
            responseObserver.onNext(response.build());
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(Status.fromThrowable(e)
                    .withDescription(e.getMessage())
                    .withCause(e)
                    .asRuntimeException());
        }
    }

    @Override
    public void closeSession(CloseSessionRequest request, StreamObserver<CloseSessionResponse> responseObserver) {
        if (!isAuthorized(request.getToken(), responseObserver)) {
            return;
        }

        try {
            boolean closed = registry.close(request.getSessionId());
            if (!closed) {
                responseObserver.onError(Status.NOT_FOUND
                        .withDescription("session not found: " + request.getSessionId())
                        .asRuntimeException());
                return;
            }
            responseObserver.onNext(CloseSessionResponse.newBuilder().setClosed(true).build());
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(Status.INTERNAL
                    .withDescription(e.getMessage())
                    .withCause(e)
                    .asRuntimeException());
        }
    }

    private static Set<String> loadPrimaryKeys(DatabaseMetaData metaData, ListColumnsRequest request) throws Exception {
        Set<String> keys = new HashSet<>();
        try (ResultSet primaryKeys = metaData.getPrimaryKeys(
                emptyToNull(request.getCatalog()),
                emptyToNull(request.getSchema()),
                request.getTable())) {
            while (primaryKeys.next()) {
                keys.add(primaryKeys.getString("COLUMN_NAME"));
            }
        }
        return keys;
    }

    private static String emptyToNull(String value) {
        return value == null || value.isBlank() ? null : value;
    }
}
