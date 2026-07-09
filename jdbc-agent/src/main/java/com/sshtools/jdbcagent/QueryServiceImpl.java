package com.sshtools.jdbcagent;

import com.sshtools.jdbcagent.proto.ExecuteQueryRequest;
import com.sshtools.jdbcagent.proto.OpenSessionRequest;
import com.sshtools.jdbcagent.proto.OpenSessionResponse;
import com.sshtools.jdbcagent.proto.QueryResult;
import com.sshtools.jdbcagent.proto.Row;
import io.grpc.Status;
import io.grpc.stub.StreamObserver;

import java.sql.Connection;
import java.sql.Driver;
import java.sql.ResultSet;
import java.sql.ResultSetMetaData;
import java.sql.Statement;
import java.util.Properties;

public class QueryServiceImpl extends HealthServiceImpl {
    private final ConnectionRegistry registry;
    private final DriverLoader driverLoader;

    public QueryServiceImpl(String token, ConnectionRegistry registry, DriverLoader driverLoader) {
        super(token);
        this.registry = registry;
        this.driverLoader = driverLoader;
    }

    @Override
    public void openSession(OpenSessionRequest request, StreamObserver<OpenSessionResponse> responseObserver) {
        if (!isAuthorized(request.getToken(), responseObserver)) {
            return;
        }

        try {
            Driver driver = driverLoader.load(request.getProfile());
            Properties properties = new Properties();
            properties.putAll(request.getPropertiesMap());
            if (!request.getUser().isBlank()) {
                properties.setProperty("user", request.getUser());
            }
            if (!request.getPassword().isBlank()) {
                properties.setProperty("password", request.getPassword());
            }

            String url = renderUrl(request);
            Connection connection = driver.connect(url, properties);
            if (connection == null) {
                responseObserver.onError(Status.INVALID_ARGUMENT
                        .withDescription("driver rejected url")
                        .asRuntimeException());
                return;
            }

            registry.put(request.getSessionId(), connection);
            responseObserver.onNext(OpenSessionResponse.newBuilder()
                    .setSessionId(request.getSessionId())
                    .build());
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(Status.INTERNAL
                    .withDescription(e.getMessage())
                    .withCause(e)
                    .asRuntimeException());
        }
    }

    @Override
    public void executeQuery(ExecuteQueryRequest request, StreamObserver<QueryResult> responseObserver) {
        if (!isAuthorized(request.getToken(), responseObserver)) {
            return;
        }

        try (Statement statement = registry.get(request.getSessionId()).createStatement()) {
            boolean hasResultSet = statement.execute(request.getSql());
            QueryResult.Builder result = QueryResult.newBuilder();
            if (hasResultSet) {
                appendRows(statement.getResultSet(), result);
            } else {
                result.setAffected(statement.getUpdateCount());
            }
            responseObserver.onNext(result.build());
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(Status.fromThrowable(e)
                    .withDescription(e.getMessage())
                    .withCause(e)
                    .asRuntimeException());
        }
    }

    private static void appendRows(ResultSet resultSet, QueryResult.Builder result) throws Exception {
        ResultSetMetaData metaData = resultSet.getMetaData();
        int columnCount = metaData.getColumnCount();
        for (int i = 1; i <= columnCount; i++) {
            result.addColumns(metaData.getColumnLabel(i));
        }

        while (resultSet.next()) {
            Row.Builder row = Row.newBuilder();
            for (int i = 1; i <= columnCount; i++) {
                Object value = resultSet.getObject(i);
                row.addValues(value == null ? "" : String.valueOf(value));
            }
            result.addRows(row);
        }
    }

    private static String renderUrl(OpenSessionRequest request) {
        return request.getProfile().getUrlTemplate()
                .replace("{host}", request.getHost())
                .replace("{port}", Integer.toString(request.getPort()))
                .replace("{database}", request.getDatabase());
    }
}
