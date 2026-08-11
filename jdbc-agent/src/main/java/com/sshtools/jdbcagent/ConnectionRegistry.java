package com.sshtools.jdbcagent;

import io.grpc.Status;

import java.sql.Connection;
import java.sql.SQLException;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

public class ConnectionRegistry {
    private final Map<String, Connection> connections = new ConcurrentHashMap<>();
    private final Map<String, Object> locks = new ConcurrentHashMap<>();

    public void put(String sessionId, Connection connection) {
        connections.put(sessionId, connection);
        locks.computeIfAbsent(sessionId, ignored -> new Object());
    }

    public Connection get(String sessionId) {
        Connection connection = connections.get(sessionId);
        if (connection == null) {
            throw Status.NOT_FOUND
                    .withDescription("session not found: " + sessionId)
                    .asRuntimeException();
        }
        return connection;
    }

    public Object lockFor(String sessionId) {
        get(sessionId);
        return locks.computeIfAbsent(sessionId, ignored -> new Object());
    }

    public <T> T withConnection(String sessionId, ConnectionCallback<T> callback) throws Exception {
        Connection connection = get(sessionId);
        synchronized (lockFor(sessionId)) {
            return callback.apply(connection);
        }
    }

    public boolean close(String sessionId) throws SQLException {
        Object lock = locks.computeIfAbsent(sessionId, ignored -> new Object());
        synchronized (lock) {
            Connection connection = connections.remove(sessionId);
            locks.remove(sessionId);
            if (connection == null) {
                return false;
            }
            connection.close();
            return true;
        }
    }

    @FunctionalInterface
    public interface ConnectionCallback<T> {
        T apply(Connection connection) throws Exception;
    }
}
