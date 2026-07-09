package com.sshtools.jdbcagent;

import io.grpc.Status;

import java.sql.Connection;
import java.sql.SQLException;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

public class ConnectionRegistry {
    private final Map<String, Connection> connections = new ConcurrentHashMap<>();

    public void put(String sessionId, Connection connection) {
        connections.put(sessionId, connection);
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

    public boolean close(String sessionId) throws SQLException {
        Connection connection = connections.remove(sessionId);
        if (connection == null) {
            return false;
        }
        connection.close();
        return true;
    }
}
