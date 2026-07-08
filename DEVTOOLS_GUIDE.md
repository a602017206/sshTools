# 开发工具集 - 使用指南

## 📖 概述

本项目成功实现了一个**可扩展的开发工具集**功能，通过统一的UI界面提供各种开发辅助工具。当前已实现**JSON格式化工具**，并设计了良好的架构以便轻松添加更多工具。

## ✨ 主要特性

- ✅ **模块化架构**：每个工具都是独立的Svelte组件
- ✅ **插件化设计**：通过简单注册即可添加新工具
- ✅ **实时验证**：JSON输入实时验证（500ms防抖）
- ✅ **语法高亮**：格式化后的JSON自动高亮显示
- ✅ **友好的错误提示**：清晰指出JSON格式错误
- ✅ **可调整面板**：支持拖拽调整工具面板宽度
- ✅ **主题适配**：自动适配明暗主题

## 🚀 快速开始

### 1. 启动应用

```bash
# 开发模式
wails dev

# 或者构建生产版本
wails build
```

### 2. 使用工具集

1. 点击右上角的 **⚙️ 工具图标**（在主题切换按钮旁边）
2. 工具面板会从右侧滑出
3. 从左侧列表中选择 **JSON 格式化** 工具
4. 在输入区粘贴JSON内容
5. 点击 **✨ 格式化** 按钮
6. 查看格式化结果和语法高亮

### 3. 工具功能

#### JSON 格式化工具

**核心功能：**
- ✨ **格式化**：美化JSON，4空格缩进
- 🗜️ **压缩**：移除所有空白字符
- ✓ **实时验证**：自动检查JSON语法
- 🎨 **语法高亮**：区分键、值、类型
- 📋 **一键复制**：快速复制结果
- 🗑️ **清空**：清除所有内容

**使用示例：**

输入：
```json
{"name":"张三","age":30,"hobbies":["阅读","编程"]}
```

格式化后：
```json
{
    "name": "张三",
    "age": 30,
    "hobbies": [
        "阅读",
        "编程"
    ]
}
```

## 🏗️ 架构说明

### 文件结构

```
sshTools/
├── frontend/src/
│   ├── components/
│   │   ├── DevToolsPanel.svelte          # 工具集主面板
│   │   └── tools/
│   │       └── JsonFormatter.svelte      # JSON格式化工具
│   ├── stores/
│   │   └── devtools.js                   # 工具集状态管理
│   ├── tools/
│   │   └── index.js                      # 工具注册中心
│   ├── App.svelte                        # 集成工具集按钮
│   └── main.js                           # 初始化工具
│
├── internal/service/
│   ├── devtools_service.go               # 后端服务
│   └── devtools_service_test.go          # 单元测试
│
└── app.go                                # Wails应用主文件
```

### 技术栈

**前端：**
- Svelte 3 - 响应式UI框架
- Svelte Store - 状态管理
- CSS Variables - 主题系统

**后端：**
- Go 1.x - 高性能后端
- encoding/json - JSON处理
- Wails v2 - 桌面应用框架

### 数据流

```
用户输入
   ↓
JsonFormatter.svelte (前端组件)
   ↓
FormatJSON(input) [Wails Binding]
   ↓
devtools_service.go (后端服务)
   ↓
json.MarshalIndent() [Go标准库]
   ↓
返回格式化结果
   ↓
前端展示 + 语法高亮
```

## 🔧 添加新工具

### 步骤 1：创建工具组件

在 `frontend/src/components/tools/` 创建新组件，例如 `Base64Tool.svelte`：

```svelte
<script>
  import { EncodeBase64, DecodeBase64 } from '../../../wailsjs/go/main/App.js';

  let inputText = '';
  let outputText = '';
  let mode = 'encode';

  async function handleConvert() {
    try {
      if (mode === 'encode') {
        outputText = await EncodeBase64(inputText);
      } else {
        outputText = await DecodeBase64(inputText);
      }
    } catch (err) {
      alert(`转换失败: ${err}`);
    }
  }
</script>

<div class="base64-tool">
  <div class="toolbar">
    <select bind:value={mode}>
      <option value="encode">编码</option>
      <option value="decode">解码</option>
    </select>
    <button on:click={handleConvert}>转换</button>
  </div>

  <textarea bind:value={inputText} placeholder="输入内容..."></textarea>
  <textarea value={outputText} readonly placeholder="输出结果..."></textarea>
</div>

<style>
  /* 样式... */
</style>
```

### 步骤 2：添加后端服务方法

在 `internal/service/devtools_service.go` 添加：

```go
import "encoding/base64"

// EncodeBase64 Base64编码
func (s *DevToolsService) EncodeBase64(input string) (string, error) {
	encoded := base64.StdEncoding.EncodeToString([]byte(input))
	return encoded, nil
}

// DecodeBase64 Base64解码
func (s *DevToolsService) DecodeBase64(input string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %v", err)
	}
	return string(decoded), nil
}
```

### 步骤 3：在 App 中暴露方法

在 `app.go` 添加：

```go
// EncodeBase64 Base64编码
func (a *App) EncodeBase64(input string) (string, error) {
	return a.devToolsService.EncodeBase64(input)
}

// DecodeBase64 Base64解码
func (a *App) DecodeBase64(input string) (string, error) {
	return a.devToolsService.DecodeBase64(input)
}
```

### 步骤 4：注册工具

在 `frontend/src/tools/index.js` 添加：

```javascript
import Base64Tool from '../components/tools/Base64Tool.svelte';

registerTool({
  id: 'base64',
  name: 'Base64',
  description: 'Base64编码和解码',
  icon: '🔐',
  component: Base64Tool,
  category: 'encoder',
  order: 2
});
```

### 步骤 5：测试

运行 `wails dev`，新工具会自动出现在工具列表中！

## 📝 API 文档

### 前端 API

#### devToolsStore

```javascript
import { devToolsStore } from './stores/devtools.js';

// 打开工具面板
devToolsStore.open();

// 关闭工具面板
devToolsStore.close();

// 切换面板状态
devToolsStore.toggle();

// 设置激活的工具
devToolsStore.setActiveTool('json-formatter');

// 设置面板宽度
devToolsStore.setWidth(600);
```

#### registerTool

```javascript
import { registerTool } from './stores/devtools.js';

registerTool({
  id: 'tool-id',              // 必需：唯一标识符
  name: '工具名称',            // 必需：显示名称
  description: '工具描述',     // 可选：工具说明
  icon: '🔧',                 // 必需：图标（emoji或SVG）
  component: ToolComponent,   // 必需：Svelte组件
  category: 'category',       // 可选：分类
  order: 1                    // 可选：排序权重
});
```

### 后端 API

#### DevToolsService 方法

```go
// FormatJSON 格式化JSON（4空格缩进）
func (s *DevToolsService) FormatJSON(input string) (string, error)

// ValidateJSON 验证JSON有效性
func (s *DevToolsService) ValidateJSON(input string) (JSONValidationResult, error)

// MinifyJSON 压缩JSON（移除空白）
func (s *DevToolsService) MinifyJSON(input string) (string, error)

// EscapeJSON 转义JSON字符串
func (s *DevToolsService) EscapeJSON(input string) (string, error)
```

#### JSONValidationResult 结构

```go
type JSONValidationResult struct {
	Valid bool   `json:"valid"`      // 是否有效
	Error string `json:"error"`      // 错误信息（如果无效）
}
```

## 🧪 测试

### 运行后端测试

```bash
# 运行所有测试
go test ./internal/service -v

# 运行特定测试
go test ./internal/service -v -run TestFormatJSON

# 测试覆盖率
go test ./internal/service -cover
```

### 测试结果

```
=== RUN   TestFormatJSON
    ✓ 有效的JSON对象
    ✓ 有效的JSON数组
    ✓ 嵌套的JSON对象
    ✓ 无效的JSON - 缺少引号
    ✓ 无效的JSON - 缺少结束符
    ✓ 空字符串
    ✓ 只有空白字符
    ✓ 带空白的有效JSON
--- PASS: TestFormatJSON

=== RUN   TestValidateJSON
    ✓ 所有验证测试通过
--- PASS: TestValidateJSON

PASS
ok  	sshTools/internal/service	0.467s
```

## 💡 最佳实践

### 1. 工具设计原则

- **单一职责**：每个工具专注于一个功能
- **用户友好**：提供清晰的提示和错误信息
- **性能优化**：使用防抖、节流等技术
- **一致性**：遵循现有工具的UI/UX模式

### 2. 错误处理

```javascript
// 前端错误处理示例
async function handleOperation() {
  try {
    const result = await BackendMethod(input);
    // 处理成功结果
  } catch (err) {
    errorMessage = `操作失败: ${err}`;
    // 显示错误给用户
  }
}
```

```go
// 后端错误处理示例
func (s *Service) Method(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("输入不能为空")
	}

	// 处理逻辑...

	return result, nil
}
```

### 3. 性能优化建议

- **防抖验证**：避免频繁调用后端（当前500ms）
- **按需加载**：使用动态组件加载
- **缓存结果**：对相同输入缓存结果
- **Web Worker**：大数据处理可使用Worker

## 🎨 自定义样式

工具组件会自动继承应用的主题变量：

```css
/* 可用的CSS变量 */
--bg-primary          /* 主背景色 */
--bg-secondary        /* 次背景色 */
--bg-hover            /* 悬停背景色 */
--text-primary        /* 主文本色 */
--text-secondary      /* 次文本色 */
--text-tertiary       /* 三级文本色 */
--border-primary      /* 边框色 */
--accent-primary      /* 强调色 */
```

## 🐛 故障排除

### 问题：工具没有显示在列表中

**解决方案：**
1. 检查是否在 `tools/index.js` 中正确注册
2. 确认组件导入路径正确
3. 查看浏览器控制台是否有错误

### 问题：后端方法调用失败

**解决方案：**
1. 确认方法在 `app.go` 中已暴露
2. 检查方法名首字母是否大写（必须导出）
3. 运行 `wails dev` 重新生成绑定

### 问题：TypeScript 绑定不存在

**解决方案：**
运行 `wails dev` 会自动生成 `wailsjs/go/main/App.js` 绑定文件

## 📊 未来扩展建议

推荐添加的工具（按优先级）：

| 优先级 | 工具名称 | 功能描述 | 难度 |
|--------|---------|---------|------|
| 🔥 高 | URL 编解码 | URL encode/decode | ⭐ 简单 |
| 🔥 高 | 时间戳转换 | Unix时间戳 ⇄ 日期 | ⭐ 简单 |
| 🔥 高 | UUID 生成器 | 生成UUID v1/v4 | ⭐ 简单 |
| 🔶 中 | Base64 工具 | Base64编解码 | ⭐ 简单 |
| 🔶 中 | Hash 计算器 | MD5/SHA256/SHA512 | ⭐⭐ 中等 |
| 🔶 中 | 颜色转换器 | HEX/RGB/HSL转换 | ⭐⭐ 中等 |
| 🔵 低 | 正则测试器 | 正则表达式测试 | ⭐⭐⭐ 复杂 |
| 🔵 低 | JWT 解码器 | JWT Token解析 | ⭐⭐⭐ 复杂 |

## 📚 参考资料

- [Wails 官方文档](https://wails.io/)
- [Svelte 官方文档](https://svelte.dev/)
- [Go JSON 包文档](https://pkg.go.dev/encoding/json)

## 📄 许可证

遵循项目主许可证。

---

**🎉 恭喜！你已经成功实现了一个可扩展的开发工具集！**

有任何问题或建议，欢迎提Issue或PR。
