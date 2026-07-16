# 数据库侧栏与例程列表错误修复

## 背景

JDBC 驱动没有统一的数据库枚举元数据接口。已连接数据库在左侧回退展示时仍保留失败提示，造成对象可加载但侧栏显示失败。旧版 JDBC agent 也不支持 `ListRoutines`，界面会把兼容性问题显示为连接错误。

## 范围

修复数据库侧栏的当前库回退状态，统一左侧对象分类与 JDBC 元数据调用方式；右侧表对象可直接打开表结构。旧版 agent 的例程请求继续由后端降级为空列表。

## 修改文件

- `frontend/src/lib/databaseObjectTree.js`
- `frontend/src/components/DatabaseSidebarTree.svelte`
- `frontend/src/components/SelectedDatabaseObjects.svelte`
- `frontend/src/components/TerminalPanel.svelte`
- `frontend/test/databaseObjectTree.test.js`

## 验证

执行 `node --test frontend/test/databaseObjectTree.test.js` 和 `cd frontend && npm run build`。桌面应用重新启动后，内嵌 JDBC agent 会替换为包含例程接口的版本。

### 工具链阻塞记录

首次执行 `npm run build` 时，Vite 编译完成，但 `stage-jdbc-agent.mjs` 调用 Gradle 读取本机 `~/.gradle` 缓存锁文件失败，错误为 `gradle-8.5-bin.zip.lck (Operation not permitted)`。最小修复方案是允许该构建命令访问本机 Gradle 缓存后原样重试；不修改构建脚本、不跳过 agent 打包。

## 剩余风险

不同 JDBC 驱动对 `getProcedures` 与 `getFunctions` 的支持程度不同，未返回例程时界面会显示为空列表。对于正在运行的旧 agent，必须重启 agent 或应用才会启用新的 RPC 实现。
