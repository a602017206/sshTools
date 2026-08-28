package com.sshtools.jdbcagent;

import com.sshtools.jdbcagent.proto.DriverProfile;
import com.sshtools.jdbcagent.proto.ExecuteQueryRequest;
import com.sshtools.jdbcagent.proto.OpenSessionRequest;
import com.sshtools.jdbcagent.proto.OpenSessionResponse;
import com.sshtools.jdbcagent.proto.QueryResult;
import io.grpc.stub.StreamObserver;
import org.h2.Driver;
import org.junit.jupiter.api.Test;

import java.nio.file.Path;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;

class QueryServiceImplTest {
    @Test
    void executeQueryReturnsRows() {
        QueryServiceImpl service = new QueryServiceImpl("secret", new ConnectionRegistry(), new DriverLoader());
        String h2JarPath = Path.of(Driver.class.getProtectionDomain().getCodeSource().getLocation().getPath()).toString();
        RecordingObserver<OpenSessionResponse> openObserver = new RecordingObserver<>();

        service.openSession(OpenSessionRequest.newBuilder()
                .setToken("secret")
                .setSessionId("h2-test")
                .setProfile(DriverProfile.newBuilder()
                        .setId("h2-2.2.224")
                        .setDriverClass("org.h2.Driver")
                        .setUrlTemplate("jdbc:h2:mem:{database};DB_CLOSE_DELAY=-1")
                        .addJarPaths(h2JarPath)
                        .build())
                .setDatabase("query_test")
                .build(), openObserver);

        assertNotNull(openObserver.value);
        assertEquals("h2-test", openObserver.value.getSessionId());

        RecordingObserver<QueryResult> queryObserver = new RecordingObserver<>();
        service.executeQuery(ExecuteQueryRequest.newBuilder()
                .setToken("secret")
                .setSessionId("h2-test")
                .setSql("select 1 as id, 'ok' as name")
                .build(), queryObserver);

        assertNotNull(queryObserver.value);
        assertEquals("ID", queryObserver.value.getColumns(0));
        assertEquals("NAME", queryObserver.value.getColumns(1));
        assertEquals("1", queryObserver.value.getRows(0).getValues(0));
        assertEquals("ok", queryObserver.value.getRows(0).getValues(1));
    }

    private static final class RecordingObserver<T> implements StreamObserver<T> {
        private T value;

        @Override
        public void onNext(T value) {
            this.value = value;
        }

        @Override
        public void onError(Throwable t) {
            throw new AssertionError(t);
        }

        @Override
        public void onCompleted() {
            // Test observers only need the value or error callback.
        }
    }
}
