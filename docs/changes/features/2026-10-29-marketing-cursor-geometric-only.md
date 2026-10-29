# 营销站鼠标造型精简为圆圈 / 彗星 / 锥形

## 背景

动物、生肖、爱心等造型不再需要；仅保留几何类：圆圈、彗星、类锥形。

## 范围

- 光标目录缩减为 `ring` / `comet` / `cone`
- 主题面板去掉萌宠与十二生肖分区
- 旧存储中的动物 id 自动回落到圆圈

## 修改文件

- `website/assets/js/cursor.js`
- `website/assets/js/fx.js`
- `website/assets/js/nav.js`
- `website/assets/js/i18n.js`
- `website/test/cursor.test.js`
- `website/test/nav.test.js`
- `docs/changes/features/2026-10-29-marketing-cursor-geometric-only.md`

## 验证

```bash
cd website && node --test test/cursor.test.js test/fx.test.js test/nav.test.js
```

## 剩余风险

- 用户本地若已选动物造型，刷新后会变成圆圈
