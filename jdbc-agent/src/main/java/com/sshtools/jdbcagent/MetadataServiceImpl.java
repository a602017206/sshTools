package com.sshtools.jdbcagent;

import com.sshtools.jdbcagent.proto.CloseSessionRequest;
import com.sshtools.jdbcagent.proto.CloseSessionResponse;
import com.sshtools.jdbcagent.proto.Column;
import com.sshtools.jdbcagent.proto.ListColumnsRequest;
import com.sshtools.jdbcagent.proto.ListColumnsResponse;
import com.sshtools.jdbcagent.proto.ListSchemasRequest;
import com.sshtools.jdbcagent.proto.ListSchemasResponse;
import com.sshtools.jdbcagent.proto.ListRoutinesRequest;
import com.sshtools.jdbcagent.proto.ListRoutinesResponse;
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
    private static final String COLUMN_DEFAULT = "COLUMN_DEF";
    private final ConnectionRegistry registry;

    public MetadataServiceImpl(String token, ConnectionRegistry registry, DriverLoader driverLoader) {
        super(token, registry, driverLoader);
        this.registry = registry;
    }

    @Override
    public void listRoutines(ListRoutinesRequest request, StreamObserver<ListRoutinesResponse> responseObserver) {
        if (!isAuthorized(request.getToken(), responseObserver)) return;
        try {
            ListRoutinesResponse response = registry.withConnection(request.getSessionId(), connection -> {
                DatabaseMetaData metaData = connection.getMetaData();
                ListRoutinesResponse.Builder builder = ListRoutinesResponse.newBuilder();
                try (ResultSet routines = request.getFunctions()
                        ? metaData.getFunctions(emptyToNull(request.getCatalog()), emptyToNull(request.getSchema()), null)
                        : metaData.getProcedures(emptyToNull(request.getCatalog()), emptyToNull(request.getSchema()), null)) {
                    String nameColumn = request.getFunctions() ? "FUNCTION_NAME" : "PROCEDURE_NAME";
                    while (routines.next()) builder.addRoutines(routines.getString(nameColumn));
                }
                return builder.build();
            });
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(Status.fromThrowable(e).withDescription(e.getMessage()).withCause(e).asRuntimeException());
        }
    }

    @Override
    public void listSchemas(ListSchemasRequest request, StreamObserver<ListSchemasResponse> responseObserver) {
        if (!isAuthorized(request.getToken(), responseObserver)) {
            return;
        }

        try {
            ListSchemasResponse response = registry.withConnection(request.getSessionId(), connection -> {
                DatabaseMetaData metaData = connection.getMetaData();
                ListSchemasResponse.Builder builder = ListSchemasResponse.newBuilder();
                try (ResultSet schemas = metaData.getSchemas(emptyToNull(request.getCatalog()), null)) {
                    while (schemas.next()) {
                        String schema = schemas.getString("TABLE_SCHEM");
                        if (schema != null && !schema.isBlank()) {
                            builder.addSchemas(schema);
                        }
                    }
                }
                return builder.build();
            });
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(Status.fromThrowable(e)
                    .withDescription(e.getMessage())
                    .withCause(e)
                    .asRuntimeException());
        }
    }

    @Override
    public void listTables(ListTablesRequest request, StreamObserver<ListTablesResponse> responseObserver) {
        if (!isAuthorized(request.getToken(), responseObserver)) {
            return;
        }

        try {
            ListTablesResponse response = registry.withConnection(request.getSessionId(), connection -> {
                DatabaseMetaData metaData = connection.getMetaData();
                ListTablesResponse.Builder builder = ListTablesResponse.newBuilder();
                try (ResultSet tables = metaData.getTables(
                        emptyToNull(request.getCatalog()),
                        emptyToNull(request.getSchema()),
                        null,
                        request.getTypesCount() == 0
                                ? new String[]{"TABLE", "SYSTEM TABLE"}
                                : request.getTypesList().toArray(new String[0]))) {
                    while (tables.next()) {
                        builder.addTables(tables.getString("TABLE_NAME"));
                    }
                }
                return builder.build();
            });
            responseObserver.onNext(response);
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
            ListColumnsResponse response = registry.withConnection(request.getSessionId(), connection -> {
                DatabaseMetaData metaData = connection.getMetaData();
                Set<String> primaryKeys = loadPrimaryKeys(metaData, request);
                ListColumnsResponse.Builder builder = ListColumnsResponse.newBuilder();
                try (ResultSet columns = metaData.getColumns(
                        emptyToNull(request.getCatalog()),
                        emptyToNull(request.getSchema()),
                        request.getTable(),
                        null)) {
                    while (columns.next()) {
                        String name = columns.getString("COLUMN_NAME");
                        builder.addColumns(Column.newBuilder()
                                .setName(name)
                                .setType(columns.getString("TYPE_NAME"))
                                .setNullable(columns.getInt("NULLABLE") == DatabaseMetaData.columnNullable)
                                .setPrimaryKey(primaryKeys.stream().anyMatch(key -> sameIdentifier(key, name)))
                                .setColumnSize(columns.getInt("COLUMN_SIZE"))
                                .setDecimalDigits(columns.getInt("DECIMAL_DIGITS"))
                                .setHasDefault(columns.getObject(COLUMN_DEFAULT) != null)
                                .setDefaultValue(columns.getObject(COLUMN_DEFAULT) == null ? "" : columns.getString(COLUMN_DEFAULT))
                                .setDescription(columns.getString("REMARKS") == null ? "" : columns.getString("REMARKS"))
                                .build());
                    }
                }
                return builder.build();
            });
            responseObserver.onNext(response);
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

    static boolean sameIdentifier(String left, String right) {
        return left != null && right != null && left.equalsIgnoreCase(right);
    }

    private static String emptyToNull(String value) {
        return value == null || value.isBlank() ? null : value;
    }
}
