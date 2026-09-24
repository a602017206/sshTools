# 文件管理收藏不再跟随访问历史

## 背景

右键「取消收藏当前路径」在没有手动收藏时也会出现。收藏和最近访问共用同一份 `history`，进入目录就会写入，当前路径因此总被当成已收藏。收藏路径也没有单独入口，时钟按钮只是最近访问。

## 范围

- 收藏写入独立的 `favorites`，只有右键收藏或取消才会改变
- 路径栏增加星标按钮，列出收藏并可以跳转或取消
- 最近访问仍由时钟按钮查看，进入目录不再改变收藏状态

## 修改文件

- `internal/config/config.go`
- `internal/config/config_test.go`
- `frontend/src/lib/fileManagerContextMenu.js`
- `frontend/src/components/FileManager.svelte`
- `frontend/src/components/FileManagerContextMenu.svelte`
- `frontend/src/stores.js`
- `frontend/wailsjs/go/models.ts`
- `frontend/test/fileManagerContextMenu.test.js`
- `.github/workflows/release.yml`
- `docs/changes/bugs/2026-09-23-file-manager-favorites.md`（本文）

## 验证

- `go test ./internal/config -count=1 -run 'TestDefaultFileManagerDirectoryTrackingEnabled|TestUpdateSettingsPersistsFileManagerFavorites'`
- `cd frontend && node --test test/fileManagerContextMenu.test.js`
- 未在桌面应用里点选星标菜单

## 剩余风险

- 以前自动记进历史的路径不会迁入收藏，需要重新右键收藏
- 收藏上限为 50 条
