# 修复营销站主题面板被终端遮挡导致看不到鼠标造型

## 背景

鼠标造型已上线，但主题下拉被右侧终端预览层挡住；「鼠标造型」标题能看到，下方选项点不到或像空白。

## 范围

- 提高 `site-header` / 主题弹层 z-index，避免被 hero 终端遮挡
- 把鼠标造型区块移到配色预设之上，打开即可看到
- 默认造型改为小猫，首次进入更易察觉自定义光标

## 修改文件

- `website/assets/css/site.css`
- `website/assets/js/nav.js`
- `website/assets/js/cursor.js`
- `website/test/cursor.test.js`
- `docs/changes/bugs/2026-10-28-marketing-cursor-panel-occlusion.md`

## 验证

```bash
cd website && node --test test/cursor.test.js test/nav.test.js
```

打开站点 → 点「主题」，应直接看到特效/萌宠/十二生肖按钮，并可点击切换。

## 剩余风险

- 触控设备仍不启用自定义光标（仅保留选择记忆）
