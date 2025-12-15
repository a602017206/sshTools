# 快速测试步骤

## ⚡ 立即测试

应用已重新启动：**http://localhost:34115**

### 步骤1：看到红色按钮了吗？

在应用窗口顶部，"SSH 连接"标题旁边，应该有一个**红色的"测试"按钮**。

- ✅ 如果看到了：进行步骤2
- ❌ 如果没看到：告诉我，可能是界面没有正确加载

### 步骤2：点击红色测试按钮

点击红色的"测试"按钮。

#### 结果A：弹出 Alert
✅ **太好了！** 说明 onclick 可以工作。
- 现在尝试点击连接列表中的"连接"、"编辑"、"删除"按钮
- 告诉我这些按钮是否可以点击

#### 结果B：没有任何反应
❌ 说明按钮无法点击。请继续步骤3。

### 步骤3：打开开发者工具

在 Wails 应用窗口中：
- **Mac**: 按 `Cmd + Option + I` 或 `F12`
- **Windows/Linux**: 按 `F12` 或 `Ctrl + Shift + I`

如果无法打开开发者工具：
- 尝试右键点击窗口，选择"检查"或"Inspect"
- 告诉我是否可以打开开发者工具

### 步骤4：运行自动诊断

在开发者工具的 **Console** 标签中，复制粘贴以下代码并按回车：

```javascript
// === 自动诊断脚本 ===
console.log('========================================');
console.log('按钮点击问题自动诊断');
console.log('========================================\n');

// 1. 查找红色测试按钮
console.log('1. 查找红色测试按钮...');
const testBtn = document.querySelector('.new-btn[onclick*="testClick"]');
if (!testBtn) {
  console.log('❌ 未找到测试按钮！');
} else {
  console.log('✅ 找到测试按钮:', testBtn.textContent.trim());

  // 2. 检查基本属性
  console.log('\n2. 检查基本属性...');
  const style = getComputedStyle(testBtn);
  console.log('  - 可见性:', style.visibility);
  console.log('  - 显示:', style.display);
  console.log('  - 透明度:', style.opacity);
  console.log('  - pointer-events:', style.pointerEvents);
  console.log('  - cursor:', style.cursor);
  console.log('  - -webkit-app-region:', style.webkitAppRegion);
  console.log('  - z-index:', style.zIndex);

  // 3. 检查是否被遮挡
  console.log('\n3. 检查是否被遮挡...');
  const rect = testBtn.getBoundingClientRect();
  const centerX = rect.left + rect.width / 2;
  const centerY = rect.top + rect.height / 2;
  const elementAtCenter = document.elementFromPoint(centerX, centerY);

  console.log('  - 按钮位置:', {
    top: rect.top,
    left: rect.left,
    width: rect.width,
    height: rect.height
  });
  console.log('  - 按钮中心点元素:', elementAtCenter);
  console.log('  - 是否被遮挡:', elementAtCenter !== testBtn ? '❌ 是' : '✅ 否');

  if (elementAtCenter !== testBtn) {
    console.log('  - 遮挡元素:', elementAtCenter.tagName, elementAtCenter.className);
  }

  // 4. 检查window函数
  console.log('\n4. 检查window函数...');
  console.log('  - window.testClick 存在:', typeof window.testClick === 'function' ? '✅ 是' : '❌ 否');

  // 5. 尝试手动触发
  console.log('\n5. 尝试手动触发...');
  try {
    console.log('  - 调用 testBtn.click()...');
    testBtn.click();
  } catch (e) {
    console.log('  - ❌ 错误:', e.message);
  }
}

console.log('\n========================================');
console.log('诊断完成！');
console.log('请截图以上输出并提供给我');
console.log('========================================');
```

### 步骤5：尝试强制修复

如果上面的诊断显示按钮被遮挡或 pointer-events 有问题，请在 Console 中运行：

```javascript
// === 强制修复脚本 ===
console.log('正在强制修复所有按钮...');

document.querySelectorAll('button, .btn, .act-btn, .new-btn').forEach((btn, index) => {
  btn.style.pointerEvents = 'auto';
  btn.style.webkitAppRegion = 'no-drag';
  btn.style.cursor = 'pointer';
  btn.style.position = 'relative';
  btn.style.zIndex = '9999';
});

// 确保没有覆盖层
document.querySelectorAll('*').forEach(el => {
  const style = getComputedStyle(el);
  if (style.webkitAppRegion === 'drag') {
    el.style.webkitAppRegion = 'no-drag';
  }
});

console.log('✅ 修复完成！');
console.log('现在请尝试点击红色测试按钮');
```

运行后，再次点击红色测试按钮，告诉我是否弹出了 alert。

---

## 📊 请告诉我的信息

请测试后复制以下模板并填写：

```
=== 测试结果 ===

步骤1 - 看到红色按钮：[是/否]

步骤2 - 点击红色按钮：
  □ 弹出了 alert
  □ 没有任何反应

步骤3 - 开发者工具：
  □ 可以打开
  □ 无法打开

步骤4 - 自动诊断输出：
[粘贴 Console 输出或截图]

步骤5 - 强制修复后：
  □ 红色按钮可以点击了
  □ 仍然无法点击
  □ 其他按钮（连接/编辑/删除）：[可以/不可以] 点击

其他观察：
[任何其他你注意到的情况]
```

---

## 🚨 如果开发者工具无法打开

如果无法打开开发者工具，我们需要修改 main.go 启用调试模式。告诉我，我会帮你修改配置。
