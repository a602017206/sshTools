# 变更：恢复文件管理右键菜单并替换文件图标

## 背景

文件管理列表右键菜单（重命名、删除、下载）看起来“没了”。原因是菜单用视口 `fixed` 坐标，被侧栏 `overflow: hidden` 与 `backdrop-filter` 形成的包含块裁切；同时列表 `click` 会在右键后立刻关闭菜单。文件/文件夹仍是扁平 SVG，缺少类型贴图。列表还把 `file.modified` 当成时间字段，实际后端是 `mod_time`，界面显示 `undefined`。

## 范围

- 右键菜单改为面板内 `absolute` 定位，避开裁切与误关闭。
- 按扩展名显示文件夹/脚本/压缩包/JAR 等贴图图标。
- 修正修改时间展示。

## 修改文件

- `frontend/src/components/FileManager.svelte`
- `frontend/src/components/icons/FileTypeIcon.svelte`
- `frontend/src/lib/fileType.js`
- `frontend/src/lib/fileType.test.js`

## 验证

- `node --test frontend/src/lib/fileType.test.js`
- 手工：右键文件出现 下载/重命名/复制路径/删除；右键文件夹出现 打开/重命名/删除；菜单不被裁切。

## 剩余风险

- 空白处右键仍无“新建文件夹”（本轮未加）。
