# Runtime Ready 错误说明

## 错误信息
```
ERR | process message error: runtime:ready -> Unknown message from front end: runtime:ready
ERR | Unknown message from front end: runtime:ready
```

## 这是什么？

这是 **Wails v2 的一个已知警告信息**，出现的原因：

1. **时序问题**：前端在后端完全初始化之前尝试发送 "runtime:ready" 消息
2. **框架行为**：Wails 框架的内部通信机制导致的
3. **无害警告**：通常不会影响应用的实际功能

## 重要提示 ⚠️

**这个错误通常可以安全忽略！**

- ✅ 不影响按钮点击
- ✅ 不影响SSH连接
- ✅ 不影响数据保存
- ✅ 不影响终端功能

## 验证应用是否正常工作

请测试以下功能，如果都正常，就可以忽略这个错误：

### 基础功能测试
```
[ ] 点击 "+ 新建连接" 按钮 - 表单显示
[ ] 填写表单并保存 - 连接出现在列表
[ ] 点击 "连接" 按钮 - 弹出密码框
[ ] 连接SSH服务器 - 终端显示
[ ] 在终端输入命令 - 有输出
```

如果以上都正常，**无需修复此错误**。

## 如果确实想修复

### 方法1：更新 Wails（推荐）
```bash
# 更新到最新版本
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 检查版本
$HOME/go/bin/wails version
```

### 方法2：修改前端初始化
在 `frontend/src/main.js` 中延迟初始化：

```javascript
// 当前代码
import App from './App.svelte'

const app = new App({
  target: document.getElementById('app'),
})

export default app
```

改为：

```javascript
import App from './App.svelte'

// 等待 DOM 完全加载
window.addEventListener('DOMContentLoaded', () => {
  const app = new App({
    target: document.getElementById('app'),
  })
})

export default app
```

### 方法3：忽略此类错误日志

如果你不想看到这些警告，可以在后端添加日志过滤。

在 `main.go` 中，找到 `wails.Run` 配置，添加：

```go
LogLevel: logger.WARNING,  // 只显示警告级别以上的日志
```

或者完全使用自定义日志：

```go
import (
    "log"
    "github.com/wailsapp/wails/v2/pkg/logger"
)

// 创建自定义logger
type CustomLogger struct {
    logger.Logger
}

func (l *CustomLogger) Print(message string) {
    // 过滤掉 runtime:ready 错误
    if !strings.Contains(message, "runtime:ready") {
        log.Println(message)
    }
}

// 在 wails.Run 中使用
err := wails.Run(&options.App{
    // ... 其他配置
    Logger: &CustomLogger{},
})
```

## 相关资源

- [Wails GitHub Issues - runtime:ready](https://github.com/wailsapp/wails/issues?q=runtime%3Aready)
- [Wails 文档 - 日志配置](https://wails.io/docs/reference/options#logger)

## 推荐做法

**除非这个错误导致功能异常，否则建议忽略它。**

这是 Wails 框架的内部行为，在未来版本中可能会被修复。只要应用功能正常，这个错误不需要处理。

## 需要帮助？

如果遇到以下情况才需要修复：
- ❌ 按钮确实无法点击
- ❌ 应用启动后立即崩溃
- ❌ 某些功能完全无法使用
- ❌ 错误信息不断重复出现（每秒多次）

否则，这个错误可以安全忽略。
