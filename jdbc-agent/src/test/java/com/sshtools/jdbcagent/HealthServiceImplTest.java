package com.sshtools.jdbcagent;

import com.sshtools.jdbcagent.proto.HealthRequest;
import com.sshtools.jdbcagent.proto.HealthResponse;
import io.grpc.Status;
import io.grpc.StatusRuntimeException;
import io.grpc.stub.StreamObserver;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertInstanceOf;
import static org.junit.jupiter.api.Assertions.assertNotNull;

class HealthServiceImplTest {
    @Test
    void healthReturnsOkForValidToken() {
        HealthServiceImpl service = new HealthServiceImpl("secret");
        RecordingObserver<HealthResponse> observer = new RecordingObserver<>();

        service.health(HealthRequest.newBuilder().setToken("secret").build(), observer);

        assertNotNull(observer.value);
        assertEquals("OK", observer.value.getStatus());
        assertEquals("0.1.0", observer.value.getAgentVersion());
        assertNotNull(observer.value.getJavaVersion());
    }

    @Test
    void healthRejectsInvalidToken() {
        HealthServiceImpl service = new HealthServiceImpl("secret");
        RecordingObserver<HealthResponse> observer = new RecordingObserver<>();

        service.health(HealthRequest.newBuilder().setToken("bad").build(), observer);

        StatusRuntimeException err = assertInstanceOf(StatusRuntimeException.class, observer.error);
        assertEquals(Status.Code.UNAUTHENTICATED, err.getStatus().getCode());
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
