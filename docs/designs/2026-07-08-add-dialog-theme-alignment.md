# 添加/编辑连接弹窗主题对齐设计

## 背景

添加和编辑连接弹窗外层已经使用全局 Dialog 主题，但内部表单仍保留固定的灰色、紫色和深色 Tailwind 类，导致切换深色/浅色主题或主题色后，输入区背景和交互态不一致。

## 设计约束

- 弹窗内部控件必须继承全局主题 token，不再写死 `gray`、`purple`、`dark:*` 背景色。
- 表单输入、只读输入、选择按钮、分组下拉、次级按钮和状态提示使用语义化样式类。
- 主操作按钮继续使用当前多巴胺主题渐变，即 `--accent-primary` 到 `--accent-secondary`。
- 保持现有弹窗尺寸、字段顺序和连接编辑流程不变，只调整视觉表现。

## 主题映射

- 输入框：`--bg-input`、`--bg-input-focus`、`--text-primary`、`--border-primary`。
- 选择项：默认使用 `--bg-secondary`，激活使用主题渐变。
- 下拉菜单：使用 `--bg-elevated`、`--border-primary`、`--shadow-lg`。
- 弱按钮：使用 `--bg-tertiary`，悬停态使用 `--bg-hover`。

