package com.sshtools.jdbcagent;

import com.sshtools.jdbcagent.proto.CloseSessionRequest;
import com.sshtools.jdbcagent.proto.CloseSessionResponse;
import com.sshtools.jdbcagent.proto.Column;
import com.sshtools.jdbcagent.proto.DriverProfile;
import com.sshtools.jdbcagent.proto.ExecuteQueryRequest;
import com.sshtools.jdbcagent.proto.ListColumnsRequest;
import com.sshtools.jdbcagent.proto.ListColumnsResponse;
import com.sshtools.jdbcagent.proto.ListSchemasRequest;
import com.sshtools.jdbcagent.proto.ListSchemasResponse;
import com.sshtools.jdbcagent.proto.ListRoutinesRequest;
import com.sshtools.jdbcagent.proto.ListRoutinesResponse;
import com.sshtools.jdbcagent.proto.ListTablesRequest;
import com.sshtools.jdbcagent.proto.ListTablesResponse;
import com.sshtools.jdbcagent.proto.OpenSessionRequest;
import com.sshtools.jdbcagent.proto.OpenSessionResponse;
import com.sshtools.jdbcagent.proto.QueryResult;
import io.grpc.Status;
import io.grpc.StatusRuntimeException;
import io.grpc.stub.StreamObserver;
import org.h2.Driver;
import org.junit.jupiter.api.Test;

import java.lang.reflect.InvocationHandler;
import java.lang.reflect.Method;
import java.lang.reflect.Proxy;
import java.nio.file.Path;
import java.sql.DatabaseMetaData;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertInstanceOf;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

class MetadataServiceImplTest {
    @Test
    void primaryKeyIdentifierComparisonIgnoresCase() {
        assertTrue(MetadataServiceImpl.sameIdentifier("id", "ID"));
    }

    @Test
    void readColumnMetadataReadsRemarksBeforeDefaultAndOnlyOnce() throws Exception {
        RecordingColumnResultSet rows = RecordingColumnResultSet.sample();
        Column column = MetadataServiceImpl.readColumnMetadata(rows.proxy(), java.util.Set.of("ID_"));

        assertEquals(java.util.List.of(
                "COLUMN_NAME", "TYPE_NAME", "COLUMN_SIZE", "DECIMAL_DIGITS", "NULLABLE", "REMARKS", "COLUMN_DEF"
        ), rows.accessedColumns());
        assertEquals(1, rows.accessCount("COLUMN_DEF"));
        assertEquals(1, rows.accessCount("REMARKS"));
        assertEquals("ID_", column.getName());
        assertEquals("NUMBER", column.getType());
        assertTrue(column.getPrimaryKey());
        assertEquals("显示名称", column.getDescription());
        assertEquals("0", column.getDefaultValue());
        assertTrue(column.getHasDefault());
    }

    @Test
    void readColumnMetadataDoesNotRereadOracleLongDefault() throws Exception {
        RecordingColumnResultSet rows = RecordingColumnResultSet.sample();
        rows.failOnSecondAccess("COLUMN_DEF");
        Column column = MetadataServiceImpl.readColumnMetadata(rows.proxy(), java.util.Set.of());
        assertEquals("0", column.getDefaultValue());
        assertTrue(column.getHasDefault());
    }

    @Test
    void listMetadataAndCloseSession() {
        MetadataServiceImpl service = new MetadataServiceImpl("secret", new ConnectionRegistry(), new DriverLoader());
        openH2Session(service);

        RecordingObserver<QueryResult> createObserver = new RecordingObserver<>();
        service.executeQuery(ExecuteQueryRequest.newBuilder()
                .setToken("secret")
                .setSessionId("h2-meta")
                .setSql("create table users (id int primary key, name varchar(32))")
                .build(), createObserver);
        assertNotNull(createObserver.value);

        RecordingObserver<QueryResult> commentObserver = new RecordingObserver<>();
        service.executeQuery(ExecuteQueryRequest.newBuilder()
                .setToken("secret")
                .setSessionId("h2-meta")
                .setSql("comment on column users.name is '显示名称'")
                .build(), commentObserver);
        assertNotNull(commentObserver.value);

        RecordingObserver<ListTablesResponse> tablesObserver = new RecordingObserver<>();
        service.listTables(ListTablesRequest.newBuilder()
                .setToken("secret")
                .setSessionId("h2-meta")
                .setSchema("PUBLIC")
                .build(), tablesObserver);
        assertTrue(tablesObserver.value.getTablesList().contains("USERS"));

        RecordingObserver<QueryResult> viewObserver = new RecordingObserver<>();
        service.executeQuery(ExecuteQueryRequest.newBuilder()
                .setToken("secret")
                .setSessionId("h2-meta")
                .setSql("create view user_names as select name from users")
                .build(), viewObserver);

        RecordingObserver<ListTablesResponse> viewsObserver = new RecordingObserver<>();
        service.listTables(ListTablesRequest.newBuilder()
                .setToken("secret")
                .setSessionId("h2-meta")
                .setSchema("PUBLIC")
                .addTypes("VIEW")
                .build(), viewsObserver);
        assertTrue(viewsObserver.value.getTablesList().contains("USER_NAMES"));

        RecordingObserver<ListRoutinesResponse> proceduresObserver = new RecordingObserver<>();
        service.listRoutines(ListRoutinesRequest.newBuilder()
                .setToken("secret")
                .setSessionId("h2-meta")
                .setSchema("PUBLIC")
                .build(), proceduresObserver);
        assertNotNull(proceduresObserver.value);

        RecordingObserver<ListSchemasResponse> schemasObserver = new RecordingObserver<>();
        service.listSchemas(ListSchemasRequest.newBuilder()
                .setToken("secret")
                .setSessionId("h2-meta")
                .build(), schemasObserver);
        assertTrue(schemasObserver.value.getSchemasList().contains("PUBLIC"));

        RecordingObserver<ListColumnsResponse> columnsObserver = new RecordingObserver<>();
        service.listColumns(ListColumnsRequest.newBuilder()
                .setToken("secret")
                .setSessionId("h2-meta")
                .setSchema("PUBLIC")
                .setTable("USERS")
                .build(), columnsObserver);
        Map<String, Column> columns = columnsObserver.value.getColumnsList().stream()
                .collect(Collectors.toMap(Column::getName, column -> column));
        assertEquals("INTEGER", columns.get("ID").getType());
        assertTrue(columns.get("ID").getPrimaryKey());
        assertEquals("CHARACTER VARYING", columns.get("NAME").getType());
        assertEquals(32, columns.get("NAME").getColumnSize());
        assertEquals("显示名称", columns.get("NAME").getDescription());

        RecordingObserver<CloseSessionResponse> closeObserver = new RecordingObserver<>();
        service.closeSession(CloseSessionRequest.newBuilder()
                .setToken("secret")
                .setSessionId("h2-meta")
                .build(), closeObserver);
        assertTrue(closeObserver.value.getClosed());

        RecordingObserver<QueryResult> afterCloseObserver = new RecordingObserver<>();
        service.executeQuery(ExecuteQueryRequest.newBuilder()
                .setToken("secret")
                .setSessionId("h2-meta")
                .setSql("select 1")
                .build(), afterCloseObserver);
        StatusRuntimeException err = assertInstanceOf(StatusRuntimeException.class, afterCloseObserver.error);
        assertEquals(Status.Code.NOT_FOUND, err.getStatus().getCode());
    }

    private static void openH2Session(MetadataServiceImpl service) {
        String h2JarPath = Path.of(Driver.class.getProtectionDomain().getCodeSource().getLocation().getPath()).toString();
        RecordingObserver<OpenSessionResponse> openObserver = new RecordingObserver<>();
        service.openSession(OpenSessionRequest.newBuilder()
                .setToken("secret")
                .setSessionId("h2-meta")
                .setProfile(DriverProfile.newBuilder()
                        .setId("h2-2.2.224")
                        .setDriverClass("org.h2.Driver")
                        .setUrlTemplate("jdbc:h2:mem:{database};DB_CLOSE_DELAY=-1")
                        .addJarPaths(h2JarPath)
                        .build())
                .setDatabase("metadata_test")
                .build(), openObserver);
        assertNotNull(openObserver.value);
    }

    private static final class RecordingObserver<T> implements StreamObserver<T> {
        private T value;
        private Throwable error;

        @Override
        public void onNext(T value) {
            this.value = value;
        }

        @Override
        public void onError(Throwable t) {
            this.error = t;
        }

        @Override
        public void onCompleted() {
            // Test observers only need the value or error callback.
        }
    }

    private static final class RecordingColumnResultSet implements InvocationHandler {
        private final Map<String, Object> values;
        private final List<String> accessedColumns = new ArrayList<>();
        private final Map<String, Integer> accessCounts = new HashMap<>();
        private String failOnSecond;
        private boolean wasNull;

        private RecordingColumnResultSet(Map<String, Object> values) {
            this.values = values;
        }

        static RecordingColumnResultSet sample() {
            Map<String, Object> values = new LinkedHashMap<>();
            values.put("COLUMN_NAME", "ID_");
            values.put("TYPE_NAME", "NUMBER");
            values.put("COLUMN_SIZE", 19);
            values.put("DECIMAL_DIGITS", 0);
            values.put("NULLABLE", DatabaseMetaData.columnNoNulls);
            values.put("REMARKS", "显示名称");
            values.put("COLUMN_DEF", "0");
            return new RecordingColumnResultSet(values);
        }

        ResultSet proxy() {
            return (ResultSet) Proxy.newProxyInstance(ResultSet.class.getClassLoader(), new Class<?>[]{ResultSet.class}, this);
        }

        void failOnSecondAccess(String column) {
            failOnSecond = column;
        }

        List<String> accessedColumns() {
            return List.copyOf(accessedColumns);
        }

        int accessCount(String column) {
            return accessCounts.getOrDefault(column, 0);
        }

        @Override
        public Object invoke(Object proxy, Method method, Object[] args) throws SQLException {
            if ("wasNull".equals(method.getName())) {
                return wasNull;
            }
            if (args == null || args.length == 0 || !(args[0] instanceof String column)) {
                throw new UnsupportedOperationException(method.getName());
            }
            accessedColumns.add(column);
            int count = accessCounts.merge(column, 1, Integer::sum);
            if (column.equals(failOnSecond) && count > 1) {
                throw new SQLException("ORA-17027: 流已被关闭", "HY000", 17027);
            }
            Object value = values.get(column);
            wasNull = value == null;
            if ("getInt".equals(method.getName())) {
                return ((Number) value).intValue();
            }
            return value;
        }
    }
}
