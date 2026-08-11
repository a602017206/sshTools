# 开发说明：设置背景图片

## 实现要点

1. **选择与持久化分离**：`InstallBackgroundImage` 只拷贝文件并返回 `path` + `data_url`；`UpdateSettings` 在用户点保存时写入 `background_image_*`。
2. **唯一文件名**：拷贝为 `wallpaper_<nano>.<ext>`，避免覆盖已保存壁纸，使「取消」仍能指向旧路径。
3. **清理**：`SettingsService.UpdateSettings` 在路径变更或关闭背景时删除旧文件；`ClearBackgroundImage` 供显式清空。
4. **前端应用**：`applyBackgroundImage` 设置 `:root[data-bg-image]` 与 CSS 变量；`.ops-shell::before` 以可调透明度绘制壁纸，壳层渐变仍可透出。

## 配置字段

| 字段 | 含义 |
|------|------|
| `background_image_enabled` | 是否启用 |
| `background_image_path` | 应用数据目录内路径 |
| `background_image_fit` | `cover` / `contain` |
| `background_image_opacity` | 0–100，默认 35 |

`background_image_data_url` 仅内存/预览使用，不落盘。
