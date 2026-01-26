# SSH Tools Frontend Design Specification

## 1. Overview
This document outlines the design specifications for the File Management module in SSH Tools. The goal is to provide a modern, efficient, and consistent user experience.

## 2. Design Philosophy
- **Modern & Clean**: Minimalist interface with focus on content.
- **Information Density**: Balanced density to show enough information without clutter.
- **Visual Feedback**: Clear states for hover, selection, loading, and errors.
- **Consistency**: Unified color scheme, spacing, and typography.

## 3. Color System
The application uses a CSS variable-based theming system supporting Dark (default) and Light modes.

### Key Variables
- **Backgrounds**:
  - `bg-primary`: Main content area (Lists).
  - `bg-secondary`: Sidebars, Headers, Modals.
  - `bg-tertiary`: Toolbars, Inputs.
  - `bg-hover`: Interactive element hover state.
- **Text**:
  - `text-primary`: Main text.
  - `text-secondary`: Metadata, labels.
- **Accents**:
  - `accent-primary`: Primary actions, selection, focus rings.
  - `accent-error`: Destructive actions, errors.
  - `accent-success`: Success states.

## 4. Components

### 4.1. File Manager Container
- **Layout**: Flexbox-based with a collapsible sidebar.
- **Header**: Contains Breadcrumbs (navigation) and Window controls.
- **Toolbar**: Grouped action buttons with icons and tooltips.
- **File List**: Grid-based list with sortable headers.

### 4.2. File List Item
- **Grid Layout**:
  - Icon: 32px (Visual type indicator)
  - Name: Flexible (Primary info)
  - Size: 100px (Right aligned)
  - Date: 140px (Right aligned)
- **Interactions**:
  - Single Click: Select.
  - Double Click: Open/Navigate.
  - Ctrl/Cmd+Click: Multi-select.
  - Right Click: Context menu (reserved).

### 4.3. Icons
- Custom SVG icons are used instead of emojis for a professional look.
- Implementation: `Icon.svelte` component wrapping SVG paths.
- Consistent sizing: 16px for actions, 18px for list items.

### 4.4. Modals
- **Backdrop**: Blurred overlay for focus.
- **Animation**: Scale and Fade entrance.
- **Input**: Clear focus states with accent color.

## 5. Animations
- **Transitions**: `fly` and `fade` (Svelte transitions) used for smooth state changes.
- **Duration**: 150ms - 200ms for snappy feel.

## 6. Responsive Design
- Layout adapts to available width.
- Breadcrumbs scroll horizontally.
- Toolbar scrolls horizontally on small screens.
