package com.sshtools.jdbcagent;

import java.io.IOException;
import java.io.InputStream;

/**
 * 监视父进程存活状态。
 *
 * <p>父进程持有一个管道的写端并把读端交给 agent 作为 stdin。父进程一旦退出（包括被强制杀死、
 * 崩溃或调试器中断），操作系统会关闭写端，agent 在读取时得到 EOF，据此自行退出，避免成为孤儿进程。
 */
public final class ParentProcessWatchdog {

    private final InputStream parentChannel;
    private final Runnable onParentExit;

    public ParentProcessWatchdog(InputStream parentChannel, Runnable onParentExit) {
        this.parentChannel = parentChannel;
        this.onParentExit = onParentExit;
    }

    /** 在后台守护线程中监视父进程。 */
    public void start() {
        Thread thread = new Thread(this::run, "parent-process-watchdog");
        thread.setDaemon(true);
        thread.start();
    }

    void run() {
        awaitParentExit();
        onParentExit.run();
    }

    void awaitParentExit() {
        try {
            byte[] buffer = new byte[64];
            while (parentChannel.read(buffer) >= 0) {
                // 父进程仍然持有管道写端；忽略内容，仅把读取当作存活探测。
            }
        } catch (IOException ignored) {
            // 管道断开等同于父进程消失。
        }
    }
}
