# UI/UX Analysis Report: OrderManager

## 1. General Issues

- **Hardcoded Dimensions**: several buttons have `height: 36px`. This restricts content scaling and localization.
- **Inconsistent Media Queries**: Breakpoints are somewhat scattered (768px, 1024px, 1200px).
- **Color Usage**: Hardcoded hex values are used instead of CSS variables in some places (`layout.css`, `OrderDetailModal.styles.css`), making theming difficult.

## 2. OrderManager.vue & Styles

- **Issue**: `.order-controls` hard switches to column layout on mobile.
- **Fix**: Use `flex-wrap: wrap` effectively before switching to column. Ensure full width on mobile.
- **Issue**: `.stat-card` padding reduces on mobile.
- **Fix**: Ensure touch targets remain accessible (>44px).

## 3. OrderFilterBar.vue

- **Issue**: `<select>` elements are styled differently from `<input>`.
- **Fix**: Standardize height/padding/border to match search input.
- **Issue**: Search input width logic (`flex: 1 1 320px`) might be too aggressive on small screens.

## 4. OrderActionsBar.vue

- **Issue**: Button spacing is small (`0.5rem`).
- **Fix**: Increase gap to `0.75rem` or `1rem` for better touch targets.
- **Issue**: Hardcoded button height.

## 5. OrderTable.vue & OrderRow.vue (layout.css)

- **Issue**: Grid columns in `layout.css` (`2.5fr 1fr ...`) are rigid.
- **Fix**: Use `minmax()` to prevent content collapse.
- **Issue**: Mobile card view (`.order-content` flex-col) needs careful spacing verification.
- **Issue**: Badge colors are hardcoded. Use CSS variables.

## 6. OrderDetailModal.vue

- **Issue**: `max-height: 90vh` is good, but mobile view should maximize screen real estate.
- **Fix**: On mobile, set `width: 100%`, `height: 100%`, `border-radius: 0`.

## 7. Accessibility

- **Issue**: Low contrast gray text on light gray backgrounds.
- **Fix**: Darken secondary text colors or use darker gray variables.
