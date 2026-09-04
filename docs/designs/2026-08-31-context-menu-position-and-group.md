# 右键菜单视口定位与分组预填

## 背景

侧栏、文件面板使用毛玻璃：`backdrop-filter` + `overflow` + `border-radius`。在 WebKit（Wails / macOS）里，这类祖先会变成 `position: fixed` 的包含块。菜单仍用 `clientX/clientY`（相对窗口），实际却相对侧栏偏移，看起来离指针很远。

Mac 触控板双指点按会同时触发 `contextmenu` 和选区，文件夹名出现蓝色高亮。这与菜单位移是两件事。

新建弹窗用 `{#key dialogRequestVersion}` 重建实例，分组预填只在 `resetForm` 里写一次，重建时可能先落到空表单。

## 决策

1. 菜单挂到 `document.body`，坐标一律用视口 `clientX/clientY`，窗口边缘夹紧。
2. 可右键的行设 `user-select: none`，打开菜单时 `preventDefault` 并清掉 `window.getSelection()`。
3. 弹窗按 `dialogRequestVersion` 再写一次 `preferredGroup`，保证右键文件夹路径进分组字段。

不改分组存储，不引入独立菜单组件库。

## 风险

无法在无头环境点 Wails 窗口。若以后给 `body` 加 `transform`，固定定位会再次偏移。
