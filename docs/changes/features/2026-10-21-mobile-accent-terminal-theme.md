# 移动端对齐桌面版主题色与终端主题

## 背景

上一轮移动端只补了明暗三态，用户指出桌面（PC）版「主题很多、可以切换不同的颜色」。桌面外观体系实际由三部分组成：明暗三态、主题色（`frontend/src/settings/appearance.js` 的 `ACCENT_PRESETS`，9 种）、终端主题（`frontend/src/lib/terminalTheme.js`，深/浅/跟随界面三态，各带一套 16 色 xterm 配色）。本次将后两者全量移植到移动端，做到与桌面同名同色。

## 范围

- 主题色 9 种：海盐青 teal、晴空蓝 blue、薄荷绿 emerald、柚子橙 amber、莓果紫 purple、湖水蓝 sky、樱桃粉 rose、珊瑚红 coral、青柠绿 lime。浅/深各取桌面预设的 `accentPrimary` 作为 Material 3 种子色生成整体配色，辅色用于选择器渐变展示。
- 默认强调色由移动端原先的蓝色改为与桌面一致的海盐青（升级后默认观感会变色，属预期）。
- 终端主题三态：深色 / 浅色 / 跟随界面；两套 16 色配色与桌面 `DARK_XTERM_THEME`、`LIGHT_XTERM_THEME` 一一对应（含背景、前景、光标、选区色）。
- 账户页外观卡扩展：主题色渐变色点网格（选中带描边与对勾，下方显示当前主题名、长按有名称提示）、终端主题三段选择；文案同步更新。
- 三项设置（明暗、主题色、终端主题）均经 SharedPreferences 持久化，重启保持。

## 修改文件

- 修改：`flutter_ui/lib/core/theme/theme_provider.dart`（新增 `AppAccent`/`appAccents`/`accentProvider`、`TerminalThemeMode`/`terminalThemeModeProvider`、`resolveXtermTheme` 与两套 xterm 配色常量）、`flutter_ui/lib/main.dart`（按强调色生成浅/深配色）、`flutter_ui/lib/presentation/screens/account/account_screen.dart`（主题色与终端主题选择 UI）、`flutter_ui/lib/presentation/screens/terminal/terminal_tab_view.dart`（`TerminalView` 接入解析出的终端配色）

## 验证

- 实现前从 pub.dev 拉取 `xterm` 4.0.0 源码核对 `TerminalView(theme:)` 参数与 `TerminalTheme` 必填字段，避免流水线试错。
- GitHub Actions 移动端流水线本轮一次成功，Android release 签名包构建通过。
- 尚待真机验证：逐个切换 9 种主题色看导航栏/按钮/选中态染色是否协调；深浅明暗 × 终端主题三态的组合观感（尤其浅色终端下 `ls --color` 的蓝色目录是否可读）；重启后三项设置保持。

## 剩余风险

- Material 3 由种子色自动生成的配色与桌面手工调校的 CSS 变量不完全同色，只能做到主色一致、整体相近。
- 柚子橙/青柠绿等浅色种子在浅色界面下生成的深色文字对比度由框架保证，但个别自定义硬编码颜色（如标签状态点红/绿）不跟随主题色，属刻意保留的语义色。
- 桌面还有字体/字号/背景图/紧凑模式等外观项，本次未移植，需要时另行立项。
