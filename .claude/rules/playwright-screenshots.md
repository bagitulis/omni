# Playwright Screenshot Rules

## Screenshot Directory (MANDATORY)

**ALL screenshots MUST be saved to `docs/Screenshots/` directory.**

```
CORRECT: docs/Screenshots/page-desktop.png
CORRECT: docs/Screenshots/task-1-sidebar-mobile.png
WRONG:   page-1234567890.png (root folder)
WRONG:   screenshot.png (root folder)
```

## When Using browser_take_screenshot

ALWAYS specify the `filename` parameter with `docs/Screenshots/` prefix:

```javascript
// CORRECT
browser_take_screenshot({
  type: "png",
  filename: "docs/Screenshots/page-name-viewport.png",
});

// WRONG - saves to root
browser_take_screenshot({
  type: "png",
});
```

## Naming Convention

```
docs/Screenshots/
├── {page}-desktop.png        # Desktop view (1440x900)
├── {page}-mobile.png         # Mobile view (375x667)
├── {page}-tablet.png         # Tablet view (768x1024)
├── {component}-hover.png     # Hover states
└── {component}-active.png    # Active states
```

## Examples

| Scenario           | Filename                                    |
| ------------------ | ------------------------------------------- |
| Dashboard desktop  | `docs/Screenshots/dashboard-desktop.png`    |
| Login mobile       | `docs/Screenshots/login-mobile.png`         |
| Sidebar collapsed  | `docs/Screenshots/sidebar-collapsed.png`    |
| Button hover state | `docs/Screenshots/button-primary-hover.png` |

## Anti-Patterns

| WRONG                 | CORRECT                                 |
| --------------------- | --------------------------------------- |
| `page-1234567890.png` | `docs/Screenshots/page-desktop.png`     |
| `screenshot.png`      | `docs/Screenshots/feature-viewport.png` |
| No filename param     | Always specify filename                 |
| Saving to root        | Always use `docs/Screenshots/` prefix   |
