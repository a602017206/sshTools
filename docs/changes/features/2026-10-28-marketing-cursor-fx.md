# 营销站主题光标与额外特效

## 背景

主题切换已支持背景/字体/强调色，但系统默认箭头不会跟随主题；页面特效也偏克制。需要主题色自定义光标，并再加一些氛围向的杂项特效。

## 范围

- 精细指针设备启用自定义光标：强调色圆环 + 实心点，悬停可点击元素放大，按下收缩
- 光标拖尾点、点击火花与涟漪，颜色跟随 CSS 变量（随主题即时变化）
- 环境层补充噪点、扫描线、上浮火花粒子
- `prefers-reduced-motion` 与粗指针设备自动回退系统光标并关闭特效

## 修改文件

- `website/assets/js/fx.js`
- `website/assets/css/site.css`
- `website/test/fx.test.js`
- `docs/changes/features/2026-10-28-marketing-cursor-fx.md`

## 验证

```bash
cd website && node --test test/fx.test.js test/theme.test.js test/nav.test.js
```

## 剩余风险

- 部分浏览器对 `cursor: none` 与系统可用性光标策略不一致；触控设备不启用自定义光标
- 额外 DOM 粒子在低端机可能略增开销，已限制数量并尊重减弱动态偏好
