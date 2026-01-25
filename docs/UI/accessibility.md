# Accessibility Guidelines

> **Target:** WCAG 2.1 Level AA | Lighthouse Score ≥ 95%

---

## Core Requirements

| Requirement           | Standard   | Example                    |
| --------------------- | ---------- | -------------------------- |
| **Text Contrast**     | 4.5:1      | Gray-500 on White          |
| **Large Text**        | 3:1        | 18px+ text                 |
| **Touch Target**      | 44×44px    | Button minimum size        |
| **Focus Indicator**   | Visible    | 2px outline                |
| **Alt Text**          | Required   | Descriptive or empty       |
| **Form Labels**       | Required   | `for` + `id` connection    |
| **Heading Hierarchy** | Sequential | h1 → h2 → h3 (no skipping) |

---

## Semantic HTML

```html
<!-- ✅ Good -->
<button>Click me</button>
<nav><a href="/">Home</a></nav>

<!-- ❌ Bad -->
<div onclick="handleClick()">Click me</div>
<span class="link">Home</span>
```

**Heading Rule:** Never skip levels (h1 → h2 → h3, not h1 → h3)

---

## ARIA Labels

### Icon-Only Buttons

```html
<button aria-label="Close modal">
  <i class="pi pi-times" aria-hidden="true"></i>
</button>
```

### Buttons with Text

```html
<!-- No aria-label needed -->
<button>
  <i class="pi pi-save" aria-hidden="true"></i>
  Save
</button>
```

### Tables

```html
<table aria-label="Product data">
  <thead>
    <tr>
      <th scope="col">Name</th>
    </tr>
  </thead>
  <tbody>
    ...
  </tbody>
</table>
```

---

## Form Accessibility

```html
<!-- ✅ Always link label to input -->
<label for="email">Email <span aria-label="required">*</span></label>
<input id="email" type="email" required aria-required="true" />

<!-- Validation feedback -->
<input aria-invalid="true" aria-describedby="email-error" />
<p id="email-error" role="alert" class="text-red-600">Invalid email</p>
```

---

## Focus States

```css
/* Always visible focus ring */
:focus-visible {
  outline: 2px solid #3b82f6;
  outline-offset: 2px;
}

/* Remove default outline only if custom visible */
button:focus:not(:focus-visible) {
  outline: none;
}
```

---

## Color Contrast

| Color    | Hex       | Contrast | Use             |
| -------- | --------- | -------- | --------------- |
| Gray-500 | `#6b7280` | 4.9:1 ✅ | Body text       |
| Gray-600 | `#4b5563` | 7.0:1 ✅ | Strong text     |
| Gray-400 | `#9ca3af` | 2.8:1 ❌ | **Don't use**   |
| Blue-500 | `#3b82f6` | 3.1:1 ❌ | Large text only |
| Blue-600 | `#2563eb` | 4.5:1 ✅ | Links, buttons  |

---

## Keyboard Navigation

```css
/* Trap focus in modals */
.modal-backdrop {
  /* backdrop */
}
.modal-content {
  /* first focusable element gets focus */
}

/* Tab order */
.btn {
  tabindex: 0;
}
.btn:disabled {
  tabindex: -1;
}
```

---

## Screen Reader Only

```css
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border-width: 0;
}
```

---

## Pre-Commit Checklist

- [ ] All images have alt text (or alt="" for decorative)
- [ ] All buttons have accessible names (text or aria-label)
- [ ] All form inputs have associated labels
- [ ] Focus states are visible
- [ ] Color contrast ≥ 4.5:1
- [ ] Heading hierarchy correct (no skipping)
- [ ] Decorative icons have `aria-hidden="true"`
- [ ] Touch targets ≥ 44×44px

## Testing Tools

- Lighthouse (Chrome DevTools)
- axe DevTools extension
- WAVE (WebAIM)
- Keyboard testing (unplug mouse!)

---

**See Also:** [Colors](./colors.md) | [Forms](./components/forms.md)
