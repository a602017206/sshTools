package com.sshtools.jdbcagent;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.io.PipedInputStream;
import java.io.PipedOutputStream;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import org.junit.jupiter.api.Test;

class ParentProcessWatchdogTest {

    @Test
    void triggersCallbackWhenParentClosesChannel() throws Exception {
        PipedOutputStream parentSide = new PipedOutputStream();
        PipedInputStream agentSide = new PipedInputStream(parentSide);
        CountDownLatch exited = new CountDownLatch(1);

        new ParentProcessWatchdog(agentSide, exited::countDown).start();

        assertFalse(exited.await(200, TimeUnit.MILLISECONDS), "父进程仍存活时不应退出");

        parentSide.close();

        assertTrue(exited.await(2, TimeUnit.SECONDS), "父进程关闭管道后 agent 应退出");
    }

    @Test
    void keepsRunningWhileParentKeepsChannelOpen() throws Exception {
        PipedOutputStream parentSide = new PipedOutputStream();
        PipedInputStream agentSide = new PipedInputStream(parentSide);
        CountDownLatch exited = new CountDownLatch(1);

        new ParentProcessWatchdog(agentSide, exited::countDown).start();

        parentSide.write('x');
        parentSide.flush();

        assertFalse(exited.await(300, TimeUnit.MILLISECONDS), "收到心跳数据不应触发退出");
    }
}
