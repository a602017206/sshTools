import { copyFile, mkdir, stat } from 'node:fs/promises';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawn } from 'node:child_process';

const scriptDir = dirname(fileURLToPath(import.meta.url));
const rootDir = dirname(scriptDir);
const agentDir = join(rootDir, 'jdbc-agent');
const sourceJar = join(agentDir, 'build', 'libs', 'sshtools-jdbc-agent-all.jar');
const targetJar = join(rootDir, 'frontend', 'build', 'jdbc-agent.jar');
const gradleCommand = process.platform === 'win32' ? 'gradlew.bat' : './gradlew';

await new Promise((resolve, reject) => {
  const child = spawn(gradleCommand, ['shadowJar'], {
    cwd: agentDir,
    shell: process.platform === 'win32',
    stdio: 'inherit'
  });
  child.once('error', reject);
  child.once('exit', (code) => {
    if (code === 0) resolve();
    else reject(new Error(`JDBC agent Gradle 构建失败，退出码: ${code}`));
  });
});

try {
  await stat(sourceJar);
} catch {
  throw new Error(`未找到 JDBC agent 构建产物: ${sourceJar}`);
}

await mkdir(dirname(targetJar), { recursive: true });
await copyFile(sourceJar, targetJar);
console.log(`已暂存 JDBC agent: ${targetJar}`);
