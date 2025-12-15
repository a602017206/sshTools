# 原生 onclick 修复方案

## 问题背景

在 Wails 桌面客户端中,使用 Svelte 的 `on:click` 事件绑定的按钮无法正常工作,而在网页版中却正常。这是因为 Wails 的 WebView 环境与普通浏览器环境存在差异。

## 解决方案

采用**原生 HTML onclick 事件**替代 Svelte 的事件绑定系统。

### 核心实现

#### 1. 创建全局函数 (frontend/src/components/ConnectionManagerSimple.svelte:188-201)

```javascript
// 使用window方法暴露全局函数供onclick使用
if (typeof window !== 'undefined') {
  window.sshToolsConnect = (connJson) => {
    const connection = typeof connJson === 'string' ? JSON.parse(connJson) : connJson;
    handleConnect(connection);
  };
  window.sshToolsEdit = (connJson) => {
    const connection = typeof connJson === 'string' ? JSON.parse(connJson) : connJson;
    handleEditConnection(connection);
  };
  window.sshToolsDelete = (id) => {
    handleRemoveConnection(id);
  };
}
```

**关键点**:
- 函数接受 JSON 字符串并解析为对象
- 支持字符串或对象参数(向后兼容)
- 通过 window 对象全局暴露

#### 2. 使用 data 属性存储连接信息 (frontend/src/components/ConnectionManagerSimple.svelte:275-294)

```html
<button
  class="act-btn connect-btn"
  data-connection={JSON.stringify(connection)}
  onclick="window.sshToolsConnect(this.dataset.connection)"
>
  连接
</button>
```

**优势**:
- `data-connection` 属性存储序列化的连接对象
- `this.dataset.connection` 在运行时获取数据
- 避免在 onclick 属性中直接嵌入复杂的 JSON 字符串

#### 3. 更新 App.svelte 使用新组件

```javascript
import ConnectionManagerSimple from './components/ConnectionManagerSimple.svelte';
```

```html
<ConnectionManagerSimple onConnect={handleConnect} />
```

## 技术原理

### 为什么原生 onclick 有效?

1. **事件传播路径不同**
   - Svelte `on:click`: 通过 Svelte 运行时处理
   - 原生 `onclick`: 直接由浏览器引擎处理

2. **Wails WebView 兼容性**
   - 原生事件处理器是 HTML 标准,所有 WebView 都支持
   - Svelte 的合成事件系统可能在 Wails 环境中有兼容性问题

3. **全局函数访问**
   - window 对象在任何环境下都可访问
   - 避免了作用域和上下文问题

### 数据传递策略

使用 HTML5 data 属性的优势:
```html
<!-- ✅ 推荐: 使用 data 属性 -->
<button
  data-connection={JSON.stringify(connection)}
  onclick="window.handler(this.dataset.connection)"
>

<!-- ❌ 避免: 直接在 onclick 中嵌入数据 -->
<button
  onclick="window.handler('{connection.name}')"
>
```

## 测试验证

### 启动应用
```bash
$HOME/go/bin/wails dev
```

### 测试清单
- ✅ 连接按钮 - 应弹出密码输入框
- ✅ 编辑按钮 - 应展开编辑表单
- ✅ 删除按钮 - 应弹出确认对话框
- ✅ 新建连接按钮 - 应展开新建表单
- ✅ 保存/取消/测试连接按钮 - 应正常工作

## 文件修改清单

### 新建文件
- `frontend/src/components/ConnectionManagerSimple.svelte` - 使用原生事件的简化组件

### 修改文件
- `frontend/src/App.svelte:3` - 导入 ConnectionManagerSimple
- `frontend/src/App.svelte:139` - 使用 ConnectionManagerSimple 组件

## 与之前方案的对比

| 方案 | 实现方式 | 效果 | 问题 |
|------|---------|------|------|
| CSS 修复 | `-webkit-app-region: no-drag` | ❌ 无效 | 不是 CSS 问题 |
| 调试 Alert | 添加 alert 调试 | ❌ 未触发 | 事件根本没触发 |
| 原生 onclick | 使用 HTML onclick | ✅ 有效 | 需要全局函数 |

## 注意事项

### 清理工作
之前的调试代码可以保留或删除:
- BUTTON_FIX_SUMMARY.md
- FINAL_FIX_TEST.md
- test-fix.sh
- DEBUG_CLIENT_BUTTONS.md
- RUNTIME_READY_ERROR.md

### 新建连接按钮的特殊处理
由于"新建连接"按钮在表单外部,使用了一个隐藏的触发按钮:
```html
<button class="new-btn" onclick="document.getElementById('new-conn-trigger').click()">
  + 新建连接
</button>
<button id="new-conn-trigger" style="display:none" on:click={showNewConnectionForm}></button>
```

这样既使用了原生 onclick,又能触发 Svelte 事件处理函数。

## 性能影响

原生 onclick 方案的性能影响:
- ✅ **更快**: 跳过 Svelte 事件系统
- ✅ **更轻**: 减少运行时开销
- ⚠️ **全局污染**: window 对象上添加了函数(已使用命名空间 `sshTools*`)

## 未来优化

可以考虑使用更现代的方案:
1. **事件委托**: 在父容器添加单个监听器
2. **Custom Elements**: 封装为 Web Components
3. **Wails 绑定**: 研究 Wails 的原生事件绑定 API

## 结论

原生 onclick 方案成功解决了 Wails 客户端按钮点击失效的问题。这个方案:
- ✅ 简单直接
- ✅ 兼容性好
- ✅ 易于调试
- ✅ 适合当前项目规模

当前的实现已经可以正常使用,无需进一步修复。
