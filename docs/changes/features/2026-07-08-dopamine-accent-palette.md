# Feature: Dopamine Accent Palette

## Background

The operations console visual refresh introduced a blue-gray control system. The next step is to make button and theme accent colors more comfortable and expressive without losing professional usability.

## Scope

This change updates frontend appearance tokens and the global settings theme-color selector.

## Modified Areas

- `frontend/src/settings/appearance.js`
- `frontend/src/styles/app.css`
- `frontend/src/App.svelte`
- `frontend/src/components/GlobalSettingsDialog.svelte`

## Behavior

- Accent presets now include `accentSecondary` for softer button gradients.
- Existing accent IDs remain compatible.
- Existing color presets are tuned to lower-saturation dopamine tones.
- New accent presets are available: 湖水蓝, 樱桃粉, 珊瑚红, 青柠绿.
- Settings preview and selected theme buttons use primary-to-secondary gradients.

## Verification

- `npm run build`
- `GOCACHE="$(pwd)/.cache/go-build" go test ./...`
- `GOCACHE="$(pwd)/.cache/go-build" go vet ./...`

## Notes

- Existing users with saved accent IDs keep their selected theme.
- The default accent remains the blue family for operational readability.
