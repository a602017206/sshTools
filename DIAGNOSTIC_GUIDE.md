# 按钮点击问题诊断指南

## 🔴 紧急测试步骤

应用已重新启动，现在界面上有一个**红色的"测试"按钮**（在"SSH 连接"标题旁边）。

### 第一步：测试最简单的按钮

**请点击红色的"测试"按钮**

#### 预期结果A：看到 Alert
✅ 如果弹出 alert："测试按钮被点击了！"
- **说明**：onclick 事件可以工作
- **下一步**：测试连接列表中的按钮

#### 预期结果B：没有任何反应
❌ 如果点击后没有任何反应
- **说明**：问题不在事件绑定，而是按钮本身无法点击
- **下一步**：执行深度诊断（见下方）

---

## 📋 深度诊断步骤

如果红色测试按钮也无法点击，请按以下步骤操作：

### 诊断1：检查按钮是否可见

在 Wails 应用窗口中按 `F12` 或 `Cmd+Option+I` 打开开发者工具。

在 Console 中运行：
```javascript
// 检查红色测试按钮
const testBtn = document.querySelector('.new-btn[onclick*="testClick"]');
console.log('测试按钮元素:', testBtn);
console.log('按钮文本:', testBtn ? testBtn.textContent : '未找到');
console.log('按钮可见性:', testBtn ? getComputedStyle(testBtn).visibility : 'N/A');
console.log('按钮显示:', testBtn ? getComputedStyle(testBtn).display : 'N/A');
console.log('按钮透明度:', testBtn ? getComputedStyle(testBtn).opacity : 'N/A');
```

**告诉我输出结果**

### 诊断2：检查按钮是否被遮挡

在 Console 中运行：
```javascript
// 检查按钮位置的元素层级
const testBtn = document.querySelector('.new-btn[onclick*="testClick"]');
if (testBtn) {
  const rect = testBtn.getBoundingClientRect();
  const centerX = rect.left + rect.width / 2;
  const centerY = rect.top + rect.height / 2;
  const elementAtCenter = document.elementFromPoint(centerX, centerY);

  console.log('按钮位置:', rect);
  console.log('按钮中心点元素:', elementAtCenter);
  console.log('是否是按钮本身:', elementAtCenter === testBtn);

  if (elementAtCenter !== testBtn) {
    console.log('❌ 按钮被遮挡了！遮挡元素:', elementAtCenter);
    console.log('遮挡元素的样式:', getComputedStyle(elementAtCenter));
  }
}
```

**告诉我输出结果**

### 诊断3：检查 CSS pointer-events

在 Console 中运行：
```javascript
// 检查所有可能影响点击的CSS属性
const testBtn = document.querySelector('.new-btn[onclick*="testClick"]');
if (testBtn) {
  const style = getComputedStyle(testBtn);
  console.log('pointer-events:', style.pointerEvents);
  console.log('cursor:', style.cursor);
  console.log('-webkit-app-region:', style.webkitAppRegion);
  console.log('z-index:', style.zIndex);
  console.log('position:', style.position);

  // 检查父元素
  let parent = testBtn.parentElement;
  let level = 1;
  while (parent && level <= 3) {
    const pStyle = getComputedStyle(parent);
    console.log(`父元素${level}:`, parent.className);
    console.log(`  pointer-events:`, pStyle.pointerEvents);
    console.log(`  -webkit-app-region:`, pStyle.webkitAppRegion);
    parent = parent.parentElement;
    level++;
  }
}
```

**告诉我输出结果**

### 诊断4：手动触发点击

在 Console 中运行：
```javascript
// 尝试通过代码触发点击
const testBtn = document.querySelector('.new-btn[onclick*="testClick"]');
if (testBtn) {
  console.log('尝试手动点击...');
  testBtn.click();

  // 也尝试直接调用函数
  console.log('尝试直接调用函数...');
  if (typeof window.testClick === 'function') {
    window.testClick();
  } else {
    console.log('❌ window.testClick 函数不存在！');
  }
}
```

**告诉我是否弹出了 alert**

### 诊断5：检查事件监听器

在开发者工具中：
1. 切换到 **Elements** 标签
2. 找到红色测试按钮（使用选择器工具）
3. 在右侧面板找到 **Event Listeners** 标签
4. 展开查看有哪些事件监听器

**告诉我看到了哪些事件监听器**

---

## 🔧 临时强制修复测试

如果上述诊断显示按钮被遮挡或 pointer-events 有问题，请在 Console 中运行：

```javascript
// 暴力修复：强制所有按钮可点击
document.querySelectorAll('button').forEach(btn => {
  btn.style.pointerEvents = 'auto';
  btn.style.webkitAppRegion = 'no-drag';
  btn.style.position = 'relative';
  btn.style.zIndex = '9999';
});

console.log('✅ 已强制启用所有按钮');
console.log('现在请尝试点击红色测试按钮');
```

运行后，再次尝试点击红色测试按钮。

**告诉我是否有效**

---

## 🎯 下一步行动

根据上述测试结果，我需要你告诉我：

### 情况A：红色测试按钮可以点击
```
✅ 红色测试按钮：可以点击，弹出了 alert
□ 连接按钮：[可以/不可以] 点击
□ 编辑按钮：[可以/不可以] 点击
□ 删除按钮：[可以/不可以] 点击
```

### 情况B：红色测试按钮也无法点击
```
❌ 红色测试按钮：无法点击
诊断1 结果：[粘贴 console 输出]
诊断2 结果：[粘贴 console 输出]
诊断3 结果：[粘贴 console 输出]
诊断4 结果：[是否弹出 alert]
强制修复后：[是否有效]
```

---

## 💡 可能的根本原因

基于之前的经验，可能的原因包括：

1. **窗口拖拽区域**
   - Wails 默认可能启用了整个窗口作为拖拽区域
   - 需要在 Go 后端配置中禁用

2. **WebView 特有问题**
   - macOS 的 WKWebView 可能有特殊限制
   - 可能需要特定的配置

3. **CSS 层叠问题**
   - 某个透明的覆盖层阻止了点击
   - z-index 导致的层级问题

4. **Wails 配置问题**
   - wails.json 中的窗口设置
   - Go 后端的 WindowOptions

让我知道诊断结果，我会根据具体情况提供解决方案！
