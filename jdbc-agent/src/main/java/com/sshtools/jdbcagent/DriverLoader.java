package com.sshtools.jdbcagent;

import com.sshtools.jdbcagent.proto.DriverProfile;

import java.net.URL;
import java.net.URLClassLoader;
import java.nio.file.Path;
import java.sql.Driver;
import java.util.ArrayList;
import java.util.List;

public class DriverLoader {
    public Driver load(DriverProfile profile) throws Exception {
        List<URL> urls = new ArrayList<>();
        for (String jarPath : profile.getJarPathsList()) {
            urls.add(Path.of(jarPath).toUri().toURL());
        }
        ClassLoader parent = Thread.currentThread().getContextClassLoader();
        URLClassLoader loader = new URLClassLoader(urls.toArray(URL[]::new), parent);
        Class<?> driverClass = Class.forName(profile.getDriverClass(), true, loader);
        return (Driver) driverClass.getDeclaredConstructor().newInstance();
    }
}
