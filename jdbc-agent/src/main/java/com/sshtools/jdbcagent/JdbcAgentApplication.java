package com.sshtools.jdbcagent;

import io.grpc.Server;
import io.grpc.ServerBuilder;

public final class JdbcAgentApplication {
    private JdbcAgentApplication() {
    }

    public static void main(String[] args) throws Exception {
        int port = parseIntArg(args, "--port", 0);
        String token = parseStringArg(args, "--token", "");
        if (port <= 0) {
            throw new IllegalArgumentException("--port is required");
        }
        if (token.isBlank()) {
            throw new IllegalArgumentException("--token is required");
        }

        Server server = ServerBuilder.forPort(port)
                .addService(new HealthServiceImpl(token))
                .build()
                .start();

        Runtime.getRuntime().addShutdownHook(new Thread(server::shutdown));
        server.awaitTermination();
    }

    private static int parseIntArg(String[] args, String name, int defaultValue) {
        String value = parseStringArg(args, name, "");
        if (value.isBlank()) {
            return defaultValue;
        }
        return Integer.parseInt(value);
    }

    private static String parseStringArg(String[] args, String name, String defaultValue) {
        for (int i = 0; i < args.length - 1; i++) {
            if (name.equals(args[i])) {
                return args[i + 1];
            }
        }
        return defaultValue;
    }
}
