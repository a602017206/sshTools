# 变更：窄侧栏下文件管理「更多」子菜单被裁切

## 背景

文件管理右键菜单的「更多」子菜单原先只在右侧空间不足时翻到左侧。侧栏较窄时主菜单几乎占满面板宽度，左侧同样放不下，子菜单被 `.file-manager` 的 `overflow: hidden` 裁掉，快捷键只露出一截。

## 范围

- 子菜单位置改为：右侧够则向右，左侧够则向左，两边都不够则在菜单内向下展开。
- 向下展开时把主菜单上移，避免底部再被裁切。

## 修改文件

- `frontend/src/lib/fileManagerContextMenu.js`
- `frontend/test/fileManagerContextMenu.test.js`
- `frontend/src/components/FileManagerContextMenu.svelte`
- `frontend/src/components/FileManager.svelte`

## 验证

- `node --test frontend/test/fileManagerContextMenu.test.js`
- 手工：把右侧文件管理拖窄后右键文件，悬停「更多」应能完整看到复制路径、新建和修改权限。

## 剩余风险

- 面板极矮时，即使上移，向下展开仍可能接近底部；菜单仍在面板内滚动容器之外，不会被文件列表滚动裁切。
