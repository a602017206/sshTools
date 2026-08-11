package com.sshtools.jdbcagent;

import org.h2.Driver;
import org.junit.jupiter.api.Test;

import java.sql.Connection;
import java.sql.DriverManager;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

class ConnectionRegistryTest {
    static {
        try {
            DriverManager.registerDriver(new Driver());
        } catch (Exception e) {
            throw new ExceptionInInitializerError(e);
        }
    }

    @Test
    void withConnectionSerializesAccessForSameSession() throws Exception {
        ConnectionRegistry registry = new ConnectionRegistry();
        Connection connection = DriverManager.getConnection("jdbc:h2:mem:registry_lock;DB_CLOSE_DELAY=-1");
        registry.put("session-1", connection);

        CountDownLatch started = new CountDownLatch(1);
        CountDownLatch releaseFirst = new CountDownLatch(1);
        AtomicInteger concurrent = new AtomicInteger();
        AtomicInteger maxConcurrent = new AtomicInteger();
        List<String> order = Collections.synchronizedList(new ArrayList<>());

        Thread first = new Thread(() -> {
            try {
                registry.withConnection("session-1", conn -> {
                    order.add("first-enter");
                    concurrent.incrementAndGet();
                    maxConcurrent.updateAndGet(value -> Math.max(value, concurrent.get()));
                    started.countDown();
                    releaseFirst.await(2, TimeUnit.SECONDS);
                    concurrent.decrementAndGet();
                    order.add("first-exit");
                    return null;
                });
            } catch (Exception e) {
                throw new RuntimeException(e);
            }
        });

        Thread second = new Thread(() -> {
            try {
                assertTrue(started.await(2, TimeUnit.SECONDS));
                registry.withConnection("session-1", conn -> {
                    order.add("second-enter");
                    concurrent.incrementAndGet();
                    maxConcurrent.updateAndGet(value -> Math.max(value, concurrent.get()));
                    concurrent.decrementAndGet();
                    order.add("second-exit");
                    return null;
                });
            } catch (Exception e) {
                throw new RuntimeException(e);
            }
        });

        first.start();
        second.start();
        assertTrue(started.await(2, TimeUnit.SECONDS));
        releaseFirst.countDown();
        first.join(2000);
        second.join(2000);

        assertEquals(1, maxConcurrent.get());
        assertEquals(List.of("first-enter", "first-exit", "second-enter", "second-exit"), order);
        connection.close();
    }
}
