# 营销站鼠标造型：特效 / 萌宠 / 十二生肖

## 背景

自定义光标已随主题变色，但仍只有单一几何造型。需要支持多种造型选择（含小猫与十二生肖），并按造型切换更炫的拖尾与点击效果。

## 范围

- 新增 `cursor.js` 造型目录与本地存储 `sshtools-site-cursor`
- 主题面板增加「鼠标造型」分区：特效、萌宠、十二生肖
- 光标跟随方向旋转；彗星/星芒/emoji 拖尾；点击弹出对应火花或 emoji 爆开
- 移动时爱心/星芒可掉落微粒；尊重减弱动态与粗指针回退

## 修改文件

- `website/assets/js/cursor.js`
- `website/assets/js/fx.js`
- `website/assets/js/nav.js`
- `website/assets/js/i18n.js`
- `website/assets/css/site.css`
- `website/test/cursor.test.js`
- `website/test/nav.test.js`
- `docs/changes/features/2026-10-28-marketing-cursor-shapes.md`

## 验证

```bash
cd website && node --test test/cursor.test.js test/fx.test.js test/nav.test.js test/theme.test.js
```

## 剩余风险

- emoji 在部分系统字体下视觉尺寸不一
- 造型较多时主题面板需要滚动；移动端仍不启用自定义光标
