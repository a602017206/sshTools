# Design: Dopamine Accent Palette

## Goal

Make button and theme accent colors more comfortable while keeping the daily operations console professional. The palette should feel lighter and more pleasant than the previous single-hue technical colors, but avoid loud neon or toy-like saturation.

## Design Direction

- Keep the current blue-gray operations console as the base.
- Use low-saturation dopamine colors only as accents for buttons, focus rings, status highlights, and selected controls.
- Use two-color button gradients: a reliable primary color plus a warmer or brighter secondary color.
- Keep hover colors darker and calmer so interactions remain clear.

## Token Model

Each accent preset provides:

- `accentPrimary`: main button and active state color.
- `accentSecondary`: companion gradient endpoint.
- `accentHover`: hover and pressed state.
- `accentSoft`: quiet background highlight.
- `focusRing`: keyboard focus outline.
- `glow1` / `glow2`: subtle app shell ambience.

## Palette

- 晴空蓝: blue to cyan, default professional dopamine accent.
- 海盐青: teal to sky, soft and reliable.
- 薄荷绿: green to lime, fresh but not fluorescent.
- 柚子橙: amber to rose, warm but controlled.
- 莓果紫: purple to pink, expressive but still usable.
- 湖水蓝: sky to cyan, lighter cloud-console feeling.
- 樱桃粉: rose to pink, friendly high-energy option.
- 珊瑚红: orange to rose, warm action-oriented option.
- 青柠绿: lime to green, bright but softened for dark mode.

## Guardrails

- Do not use accent colors as large page backgrounds.
- Do not increase glow strength beyond subtle ambient hints.
- Keep critical/error states red and success states green independent from user accent choices.
