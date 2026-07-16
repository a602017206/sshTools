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

import java.nio.file.Path;
import java.util.Map;
import java.util.stream.Collectors;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertInstanceOf;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

class MetadataServiceImplTest {
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
        }
    }
}
