package com.sshtools.jdbcagent;

import io.grpc.Server;
import io.grpc.ServerBuilder;

public final class JdbcAgentApplication {

    /** 显式开关，只有宿主进程传入时才启用 stdin 守护，避免直接手工运行 jar 时立即退出。 */
    private static final String WATCH_STDIN_FLAG = "--watch-stdin";

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
                .addService(new MetadataServiceImpl(token, new ConnectionRegistry(), new DriverLoader()))
                .build()
                .start();

        Runtime.getRuntime().addShutdownHook(new Thread(server::shutdown));

        if (hasFlag(args, WATCH_STDIN_FLAG)) {
            new ParentProcessWatchdog(System.in, () -> System.exit(0)).start();
        }

        server.awaitTermination();
    }

    private static boolean hasFlag(String[] args, String name) {
        for (String arg : args) {
            if (name.equals(arg)) {
                return true;
            }
        }
        return false;
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
