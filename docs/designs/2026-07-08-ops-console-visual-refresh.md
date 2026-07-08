# Design: Ops Console Visual Refresh

## Goal

Refresh the desktop SSH tool into a daily operations console. The interface should prioritize terminal work, keep connection management compact, support light and dark themes, and avoid cinematic hacker styling.

## Design Direction

- Product character: practical DevOps workstation.
- Primary workflow: terminal-first.
- Theme behavior: default follows system settings; light and dark themes share one blue-gray visual system.
- Right panel: collapsed by default, exposed through a narrow tool rail for SFTP and monitoring.
- Left panel: keep the existing full connection tree width, but reduce row density and visual noise.

## Visual System

- Light background: `#f5f7fb`, elevated surfaces `#ffffff`.
- Dark background: `#08111f`, elevated surfaces `#0d1728`.
- Accent: operations blue, light `#2563eb`, dark `#60a5fa`.
- Terminal: deep graphite in dark mode and off-white in light mode.
- Borders: low-contrast blue-gray lines to structure panes without heavy card styling.

## Layout

```text
+--------------------------------------------------------------+
| Top bar: product, upload, settings, devtools                 |
+-------------+--------------------------------------+---------+
| Connection  | Session tabs                         | Tool    |
| tree        +--------------------------------------+ rail    |
| compact     | Terminal header                      | SFTP    |
| rows        +--------------------------------------+ Monitor |
|             | Terminal canvas                      |         |
+-------------+--------------------------------------+---------+
```

## Component Decisions

- Main app shell uses tokenized `ops-*` CSS utilities so future views can migrate without copying Tailwind color classes.
- Asset rows show icon, connection name, and a status dot by default. Host, username, and port are still available as hover title text.
- SSH sessions no longer auto-open the right side panel.
- Collapsed right panel keeps a 44px rail so file and monitor tools remain discoverable.
- Terminal tabs and toolbar use smaller typography and tighter padding.
- Shared dialog styling is flatter, smaller, and aligned with the same token system.

## Out Of Scope

- No restructuring of backend, SSH protocol handling, SFTP, or database logic.
- No new UI routes or panels.
- No full accessibility cleanup for pre-existing warnings outside the touched flow.
