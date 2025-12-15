# 客户端按钮点击问题 - 调试指南

## 问题描述
网页版（http://localhost:34115）按钮正常，但Wails桌面客户端中按钮无法点击。

## 原因分析
这是Wails桌面应用的常见问题，通常由以下原因导致：

1. **窗口拖拽区域覆盖**：Wails可能将某些区域设置为可拖拽窗口的区域
2. **CSS属性冲突**：`-webkit-app-region: drag` 会阻止鼠标事件
3. **z-index层级问题**：某些覆盖层阻挡了按钮点击

## 已实施的修复

我已经为所有交互元素添加了 `-webkit-app-region: no-drag` CSS属性：

### 修复的元素
- ✅ 所有按钮（新建、保存、取消、测试、连接、编辑、删除）
- ✅ 所有输入框（text, number, password）
- ✅ 下拉选择框（select）
- ✅ 复选框（checkbox）

### CSS添加的属性
```css
button, input, select {
  -webkit-app-region: no-drag;
}
```

## 测试步骤

### 1. 重新构建应用
```bash
# 停止当前运行的 wails dev
# 按 Ctrl+C

# 重新启动
$HOME/go/bin/wails dev
```

### 2. 测试每个按钮

**新建连接按钮**：
```
[ ] 点击 "+ 新建连接" 按钮
    应该：表单展开
```

**表单按钮**：
```
[ ] 点击 "取消" 按钮
    应该：表单关闭
[ ] 点击 "测试连接" 按钮
    应该：显示测试结果
[ ] 点击 "保存" 按钮
    应该：保存连接并关闭表单
```

**连接列表按钮**：
```
[ ] 点击 "连接" 按钮
    应该：弹出密码输入框
[ ] 点击 "编辑" 按钮
    应该：表单展开并预填充数据
[ ] 点击 "删除" 按钮
    应该：弹出确认对话框
```

## 如果问题仍然存在

### 方法1：检查开发者工具
1. 在Wails应用中按 `F12` 打开开发者工具
2. 切换到 "Elements" 标签
3. 使用选择器工具（左上角箭头图标）
4. 点击无法响应的按钮
5. 在右侧 "Styles" 面板中查找：
   - 是否有 `-webkit-app-region: drag`
   - 是否有 `pointer-events: none`
   - 是否有覆盖的z-index

### 方法2：手动添加样式
如果某个按钮仍然无法点击，在开发者工具的Console中运行：

```javascript
// 测试连接按钮
document.querySelectorAll('button').forEach(btn => {
  btn.style.webkitAppRegion = 'no-drag';
  btn.style.pointerEvents = 'auto';
});
```

如果这样能让按钮工作，说明CSS可能没有正确应用。

### 方法3：检查是否有全局拖拽设置
在开发者工具Console中运行：

```javascript
// 检查body或main元素
console.log(getComputedStyle(document.body).webkitAppRegion);
console.log(getComputedStyle(document.querySelector('main')).webkitAppRegion);
```

如果输出是 "drag"，那需要在 App.svelte 中添加：

```css
main {
  -webkit-app-region: drag;
}

/* 然后为所有交互元素添加 */
button, input, select, .clickable {
  -webkit-app-region: no-drag !important;
}
```

### 方法4：完全禁用窗口拖拽
如果以上都不行，可以在 `main.go` 中禁用窗口拖拽：

```go
// 在 wails.Run 配置中添加：
Frameless: false,  // 确保这个是 false
```

## 验证修复
运行这个命令检查修改是否保存：

```bash
grep -n "webkit-app-region" frontend/src/components/ConnectionManager.svelte
```

应该看到多行输出，包含 `no-drag`。

## 替代方案

如果按钮仍然无法点击，可以尝试：

### 1. 添加全局样式
在 `frontend/src/style.css` 中添加：

```css
* {
  -webkit-app-region: no-drag;
}

/* 如果需要窗口拖拽，只在特定区域启用 */
.window-header {
  -webkit-app-region: drag;
}
```

### 2. 使用 !important 强制应用
在 ConnectionManager.svelte 中，将所有按钮样式改为：

```css
-webkit-app-region: no-drag !important;
```

## 报告结果

请测试后告诉我：
1. 哪些按钮现在可以点击了？
2. 哪些按钮仍然无法点击？
3. 在开发者工具中看到了什么信息？

这将帮助我进一步定位问题。

## 相关链接
- [Wails 窗口拖拽文档](https://wails.io/docs/reference/options#frameless)
- [WebKit CSS 区域属性](https://developer.mozilla.org/en-US/docs/Web/CSS/-webkit-app-region)
