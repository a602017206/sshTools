# 移动端四项增强：分组搜索、克隆、终端字号、命令建议

## 背景

实机反馈修复后，用户点名要四项 PC 侧已有或呼声高的能力：终端命令建议/历史、连接分组 + 搜索、连接克隆、终端字号调节。本次一并落地。

## 实现

- **连接分组 + 搜索**：沿用桌面端约定——连接 `tags` 的第一个即分组（`frontend/src/App.svelte` 里 `group = tags[0] || '默认分组'`），后端零改动。
  - 表单新增「分组」输入框（下方带已有分组 ActionChip 快选），原标签输入改名「其它标签」；保存时按 `[分组, ...其余标签]` 组装并剔除与分组同名项。
  - 资产列表顶部加搜索框（名称/主机/用户/任意标签过滤，可清空），列表按分组分段展示（组头含名称与台数），「默认分组」垫底；无匹配时给出搜索空状态。
- **连接克隆**：卡片新增「克隆」按钮（按钮行改 Wrap 后仍可换行）。以原连接预填表单走新增：ID 清空、名称加「副本」后缀；**密码不复制**（服务端从不下发密码），表单校验要求重新输入。
- **终端字号**：新增 `terminalFontSizeProvider`（10~20 磅，默认 13），SharedPreferences 持久化；设置页外观卡在终端主题下方加滑杆，终端页实时套用 `TerminalStyle`。
- **终端命令建议/历史**：
  - 新增 `commandHistoryProvider`：命令新到旧存储，上限 200 条、去重置顶，持久化到 SharedPreferences；设置页有「命令历史」卡显示条数并可一键清除。
  - 终端逐字跟踪当前行（回车记录、退格删尾、^C/^U 清行、ESC 序列忽略；只做本地跟踪，不改变发往远端的数据），输入时在特殊按键条上方出现建议条（前缀/包含匹配，最多 6 条，等宽字体）；点选以 `^U` 清行后填入候选，不自动回车，用户可改或直接执行。
  - 历史不分服务器（同类机器间通用性优先），太短（<2 字符）的输入不记录。

## 修改文件

- 新增：`flutter_ui/lib/presentation/providers/command_history_provider.dart`
- 修改：`flutter_ui/lib/core/theme/theme_provider.dart`（字号设置）、`flutter_ui/lib/presentation/screens/account/account_screen.dart`（字号滑杆、命令历史卡）、`flutter_ui/lib/presentation/screens/connections/connection_list_screen.dart`（Stateful 化、搜索、分组、克隆入口）、`flutter_ui/lib/presentation/screens/connections/widgets/connection_form_page.dart`（分组字段、克隆模式）、`flutter_ui/lib/presentation/screens/terminal/terminal_tab_view.dart`（行跟踪、建议条、字号接入）

## 验证

- GitHub Actions 移动端流水线（39714816652）成功（Dart 编译即检查）。
- 待真机验证：同一分组下多台机器展示与搜索过滤；克隆后新连接可独立连接（密码重填后）；字号滑杆拖动终端实时变大变小且重启保持；反复执行同一命令只记一条、输入前两个字母出现建议、点选填入后可正常回车执行。

## 剩余风险

- 命令行跟踪是本地推断：粘贴多行文本、shell 的复杂行编辑（如 ^R 反查、补全菜单）可能让当前行与建议短暂不一致，下一次回车/清行即自愈，不影响发往远端的真实数据。
- 分组约定与桌面共用 `tags[0]`：在桌面端把第一个标签当普通标签改掉时，移动端分组会随之变化，属同一数据约定而非缺陷。
