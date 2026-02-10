---
description: Playwright layout evaluation rules for React migration - verifies UI against Design System specs
---

# Playwright Layout Evaluation Rules

> **Role:** Verify UI/Layout implementation matches Design System specifications.
> **When:** After implementing any layout, page, or component in React migration.
> **Tool:** Use `skill_mcp(mcp_name="playwright", ...)` for browser automation.

---

## Evaluation Checklist

### 1. Layout Dimensions

| Element           | Expected | Selector                      |
| ----------------- | -------- | ----------------------------- |
| Sidebar expanded  | 220px    | `.ant-layout-sider`           |
| Sidebar collapsed | 80px     | `.ant-layout-sider.collapsed` |
| Header height     | 48px     | `.ant-layout-header`          |
| Mobile nav height | 56px     | `[data-testid="mobile-nav"]`  |
| Content padding   | 24px     | `.ant-layout-content`         |

### 2. Color Verification

| Element          | Expected Color               | Selector                  |
| ---------------- | ---------------------------- | ------------------------- |
| Primary button   | `#0369a1` / `rgb(3,105,161)` | `.ant-btn-primary`        |
| Active menu item | `#0369a1` background         | `.ant-menu-item-selected` |
| Success status   | `#16a34a`                    | `.ant-tag-success`        |
| Warning status   | `#d97706`                    | `.ant-tag-warning`        |
| Error status     | `#dc2626`                    | `.ant-tag-error`          |

### 3. Typography Verification

| Property    | Expected                   | Selector     |
| ----------- | -------------------------- | ------------ |
| Base font   | 12px                       | `body`       |
| Font family | system-ui or -apple-system | `body`       |
| Table text  | 12px                       | `.ant-table` |
| Page titles | 20px (text-2xl)            | `h1`         |
| Body text   | 12px (text-sm)             | `p, span`    |

### Font Consistency Check (CRITICAL)

```javascript
// Verify ALL text uses consistent font family and sizes from the scale
() => {
  const allowedSizes = [10, 12, 14, 16, 18, 20, 24]; // Font scale from Design System
  const issues = [];
  const fontFamilies = new Set();
  const fontSizes = new Map();

  // Check all text elements
  document
    .querySelectorAll(
      "p, span, h1, h2, h3, h4, h5, h6, td, th, label, button, a, li",
    )
    .forEach((el) => {
      if (!el.textContent?.trim()) return;

      const styles = getComputedStyle(el);
      const fontFamily = styles.fontFamily;
      const fontSize = parseInt(styles.fontSize);

      // Track font families
      const primaryFont = fontFamily.split(",")[0].trim().replace(/['"]/g, "");
      fontFamilies.add(primaryFont);

      // Check font size against allowed scale
      if (!allowedSizes.includes(fontSize)) {
        const existing = fontSizes.get(fontSize) || [];
        existing.push(el.tagName);
        fontSizes.set(fontSize, existing);
      }
    });

  // Check for Google Fonts (NOT allowed)
  const hasGoogleFonts = Array.from(fontFamilies).some((f) =>
    [
      "Inter",
      "Poppins",
      "Open Sans",
      "Lato",
      "Montserrat",
      "Nunito",
      "Raleway",
    ].includes(f),
  );

  // Check network for Google Fonts requests
  const googleFontRequests = performance
    .getEntriesByType("resource")
    .filter(
      (r) =>
        r.name.includes("fonts.googleapis.com") ||
        r.name.includes("fonts.gstatic.com"),
    );

  // Build issues list
  if (hasGoogleFonts) {
    issues.push("CRITICAL: Google Fonts detected - use system fonts only");
  }
  if (googleFontRequests.length > 0) {
    issues.push("CRITICAL: Google Fonts being loaded via network");
  }

  fontSizes.forEach((elements, size) => {
    if (elements.length > 2) {
      // Only report if used more than twice
      issues.push(
        `Non-standard font size ${size}px used in: ${elements.slice(0, 3).join(", ")}`,
      );
    }
  });

  return {
    fontFamiliesUsed: Array.from(fontFamilies),
    invalidFontSizes: Object.fromEntries(fontSizes),
    hasGoogleFonts,
    googleFontRequests: googleFontRequests.length,
    issues,
    consistent: issues.length === 0,
  };
};
```

### Font Scale Reference (from Design System)

| Token       | Size | Usage                    |
| ----------- | ---- | ------------------------ |
| `text-xs`   | 10px | Badges, captions         |
| `text-sm`   | 12px | **DEFAULT** body, tables |
| `text-base` | 14px | Emphasized text          |
| `text-lg`   | 16px | Subheadings              |
| `text-xl`   | 18px | Section titles           |
| `text-2xl`  | 20px | Page titles              |
| `text-3xl`  | 24px | Large headings           |

### 4. Responsive Breakpoints

| Viewport | Width  | Expected Behavior                  |
| -------- | ------ | ---------------------------------- |
| Mobile   | 375px  | Bottom nav visible, sidebar hidden |
| Tablet   | 768px  | Sidebar collapsible                |
| Desktop  | 1200px | Sidebar expanded                   |

---

## Standard Verification Scenarios

### Layout Verification

```
Scenario: Verify layout dimensions match Design System
  Tool: skill_mcp(mcp_name="playwright")
  Steps:
    1. browser_navigate to page URL
    2. browser_snapshot to capture accessibility tree
    3. browser_evaluate to check dimensions:
       - Sidebar width (expanded: 220px, collapsed: 80px)
       - Header height (48px)
       - Content padding (24px)
    4. browser_take_screenshot for evidence
  Expected: All dimensions match Design System
  Evidence: Screenshot saved to .sisyphus/evidence/
```

### Mobile Layout Verification

```
Scenario: Verify mobile responsive behavior
  Tool: skill_mcp(mcp_name="playwright")
  Steps:
    1. browser_resize(width=375, height=667)
    2. browser_navigate to page URL
    3. browser_snapshot
    4. browser_evaluate to check:
       - Sidebar is hidden (display:none or width:0)
       - Bottom nav is visible
       - Bottom nav height is 56px
    5. browser_take_screenshot
  Expected: Bottom nav visible, sidebar hidden
  Evidence: Screenshot with mobile viewport
```

### Color Verification

```
Scenario: Verify primary color application
  Tool: skill_mcp(mcp_name="playwright")
  Steps:
    1. browser_navigate to page URL
    2. browser_evaluate with function:
       () => {
         const btn = document.querySelector('.ant-btn-primary');
         return getComputedStyle(btn).backgroundColor;
       }
    3. Assert: Color is rgb(3, 105, 161) or contains #0369a1
  Expected: Primary button uses sky-700 color
```

### Navigation Structure Verification

```
Scenario: Verify flat navigation (no nested menus)
  Tool: skill_mcp(mcp_name="playwright")
  Steps:
    1. browser_navigate to page URL
    2. browser_evaluate:
       () => document.querySelectorAll('.ant-menu-submenu').length
    3. Assert: Count is 0 (flat navigation)
  Expected: No nested submenus (Ginee-style flat nav)
```

---

## JavaScript Evaluation Templates

### Check Element Dimensions

```javascript
(page) => {
  const element = document.querySelector(".ant-layout-sider");
  const rect = element.getBoundingClientRect();
  return {
    width: rect.width,
    height: rect.height,
    visible: rect.width > 0 && rect.height > 0,
  };
};
```

### Check Computed Styles

```javascript
(page) => {
  const element = document.querySelector(".ant-btn-primary");
  const styles = getComputedStyle(element);
  return {
    backgroundColor: styles.backgroundColor,
    borderRadius: styles.borderRadius,
    fontSize: styles.fontSize,
    fontFamily: styles.fontFamily,
  };
};
```

### Check Mobile Layout

```javascript
(page) => {
  const sidebar = document.querySelector(".ant-layout-sider");
  const mobileNav = document.querySelector('[data-testid="mobile-nav"]');
  return {
    sidebarHidden: !sidebar || getComputedStyle(sidebar).display === "none",
    mobileNavVisible:
      mobileNav && getComputedStyle(mobileNav).display !== "none",
    mobileNavHeight: mobileNav ? mobileNav.getBoundingClientRect().height : 0,
  };
};
```

---

## Evidence Collection

### Screenshot Naming Convention

**CRITICAL: ALL screenshots MUST be saved to `docs/Screenshots/` directory. NEVER save to root folder.**

```
docs/Screenshots/
├── task-{N}-{component}-desktop.png     # Desktop view
├── task-{N}-{component}-mobile.png      # Mobile view (375x667)
├── task-{N}-{component}-tablet.png      # Tablet view (768x1024)
└── task-{N}-{component}-dark.png        # Dark mode (if applicable)
```

### Screenshot Tool Usage

ALWAYS specify filename with `docs/Screenshots/` prefix:

```javascript
// CORRECT
browser_take_screenshot({
  type: "png",
  filename: "docs/Screenshots/task-1-sidebar-desktop.png",
});

// WRONG - saves to root, creates mess
browser_take_screenshot({ type: "png" });
```

### Required Screenshots Per Page

1. **Desktop full page** - Default viewport
2. **Mobile responsive** - 375x667 viewport
3. **Key interactions** - Hover states, active states

---

## Pass/Fail Criteria

### PASS Conditions

- All layout dimensions match within 2px tolerance
- Primary color is #0369a1 (rgb(3, 105, 161))
- Font size is 12px for body/tables
- System fonts used (no Google Fonts requests)
- Flat navigation (0 nested submenus)
- Mobile bottom nav visible at < 768px

### FAIL Conditions

- Layout dimensions off by > 5px
- Wrong primary color used
- Google Fonts loaded (check network requests)
- Nested submenus present
- Mobile nav missing on mobile viewport
- Border radius > 6px (should be 3px sharp corners)

---

## Integration with Task Workflow

1. **After implementing layout/page:**
   - Load this skill
   - Run verification scenarios
   - Capture screenshots

2. **Before marking task complete:**
   - All verification scenarios pass
   - Evidence screenshots saved

3. **If verification fails:**
   - Fix the specific issue
   - Re-run verification
   - Update evidence

---

## Responsive Transition Verification (CRITICAL)

### Smooth Resize Test (MANDATORY)

Every page MUST be tested for smooth responsive transitions:

```
Scenario: Verify layout transitions smoothly during resize
  Tool: skill_mcp(mcp_name="playwright")
  Steps:
    1. browser_navigate to page URL
    2. browser_resize(width=1400, height=900) - Start at desktop
    3. browser_take_screenshot(filename="resize-1400.png")
    4. browser_resize(width=1024, height=768) - Tablet landscape
    5. browser_wait_for(time=0.5) - Wait for CSS transition
    6. browser_take_screenshot(filename="resize-1024.png")
    7. browser_resize(width=768, height=1024) - Tablet portrait
    8. browser_wait_for(time=0.5)
    9. browser_take_screenshot(filename="resize-768.png")
    10. browser_resize(width=375, height=667) - Mobile
    11. browser_wait_for(time=0.5)
    12. browser_take_screenshot(filename="resize-375.png")
  Expected: No overlapping elements, no broken layout, smooth transitions
```

### Layout Integrity Check

```javascript
// Check for layout issues at each viewport
() => {
  const issues = [];

  // Check for overlapping elements
  const allElements = document.querySelectorAll("*");
  const rects = [];
  allElements.forEach((el) => {
    const rect = el.getBoundingClientRect();
    if (rect.width > 0 && rect.height > 0) {
      rects.push({ el, rect });
    }
  });

  // Check if any element overflows viewport
  const viewportWidth = window.innerWidth;
  const overflowing = document.querySelectorAll("*");
  overflowing.forEach((el) => {
    const rect = el.getBoundingClientRect();
    if (rect.right > viewportWidth + 5) {
      // 5px tolerance
      issues.push(`Element overflows: ${el.className || el.tagName}`);
    }
  });

  // Check for horizontal scroll (bad!)
  const hasHorizontalScroll =
    document.documentElement.scrollWidth > document.documentElement.clientWidth;
  if (hasHorizontalScroll) {
    issues.push("Page has unwanted horizontal scroll");
  }

  // Check content is not cut off
  const mainContent = document.querySelector(
    '.ant-layout-content, main, [role="main"]',
  );
  if (mainContent) {
    const contentRect = mainContent.getBoundingClientRect();
    if (contentRect.width < 300) {
      issues.push("Content area too narrow");
    }
  }

  return {
    hasIssues: issues.length > 0,
    issues,
    viewportWidth,
    hasHorizontalScroll,
  };
};
```

### Visual Alignment Verification

```javascript
// Check that cards/grid items are properly aligned
() => {
  const cards = document.querySelectorAll(".ant-card, .ant-col");
  const rows = {};

  cards.forEach((card) => {
    const rect = card.getBoundingClientRect();
    const rowKey = Math.round(rect.top / 10) * 10; // Group by approximate row
    if (!rows[rowKey]) rows[rowKey] = [];
    rows[rowKey].push({
      left: rect.left,
      width: rect.width,
      element: card.className,
    });
  });

  // Check alignment within each row
  const alignmentIssues = [];
  Object.entries(rows).forEach(([row, items]) => {
    if (items.length > 1) {
      // Check if items in same row have consistent spacing
      const gaps = [];
      for (let i = 1; i < items.length; i++) {
        gaps.push(items[i].left - (items[i - 1].left + items[i - 1].width));
      }
      const avgGap = gaps.reduce((a, b) => a + b, 0) / gaps.length;
      gaps.forEach((gap, i) => {
        if (Math.abs(gap - avgGap) > 10) {
          // 10px tolerance
          alignmentIssues.push(`Inconsistent spacing in row ${row}`);
        }
      });
    }
  });

  return {
    totalCards: cards.length,
    rowCount: Object.keys(rows).length,
    alignmentIssues,
    isAligned: alignmentIssues.length === 0,
  };
};
```

---

## Required Viewport Tests (MANDATORY)

| Viewport    | Width x Height | What to Check                                 |
| ----------- | -------------- | --------------------------------------------- |
| Desktop XL  | 1920 x 1080    | Full sidebar, wide tables                     |
| Desktop     | 1440 x 900     | Standard desktop layout                       |
| Laptop      | 1280 x 800     | Slightly condensed                            |
| Tablet Land | 1024 x 768     | Sidebar may collapse                          |
| Tablet Port | 768 x 1024     | Transition point - sidebar hidden, bottom nav |
| Mobile L    | 425 x 896      | Mobile layout with bottom nav                 |
| Mobile M    | 375 x 667      | iPhone SE size - tight but no overflow        |
| Mobile S    | 320 x 568      | Smallest supported - must still work          |

### Minimum Viable Test Set

At minimum, test these 4 viewports:

1. **1440x900** - Desktop
2. **768x1024** - Tablet (transition point)
3. **375x667** - Mobile standard
4. **320x568** - Mobile minimum (catch edge cases)

---

## Layout Messiness Detection

### What Counts as "Messy/Broken" Layout

| Issue                       | Detection Method                       |
| --------------------------- | -------------------------------------- |
| Overlapping text/elements   | Check bounding box intersections       |
| Horizontal scroll           | `scrollWidth > clientWidth`            |
| Cut-off content             | Element `right` > viewport width       |
| Misaligned cards/grid       | Inconsistent top positions in same row |
| Missing responsive behavior | Sidebar visible on mobile              |
| Broken flex/grid            | Items with width=0 or negative margins |
| Truncated without ellipsis  | Text overflow without `text-overflow`  |
| Broken images               | Images with naturalWidth=0             |

### Automatic Layout Health Check

```javascript
// Run this after every resize to check layout health
() => {
  const report = {
    viewport: { width: window.innerWidth, height: window.innerHeight },
    issues: [],
    warnings: [],
    score: 100,
  };

  // 1. Horizontal overflow (CRITICAL - fail)
  if (
    document.documentElement.scrollWidth > document.documentElement.clientWidth
  ) {
    report.issues.push("CRITICAL: Horizontal scroll detected");
    report.score -= 30;
  }

  // 2. Elements overflowing viewport
  document
    .querySelectorAll(".ant-card, .ant-table, .ant-form, main, section")
    .forEach((el) => {
      const rect = el.getBoundingClientRect();
      if (rect.right > window.innerWidth + 5) {
        report.issues.push(`Element overflows: ${el.className.split(" ")[0]}`);
        report.score -= 10;
      }
    });

  // 3. Broken images
  document.querySelectorAll("img").forEach((img) => {
    if (img.naturalWidth === 0 && img.complete) {
      report.warnings.push(`Broken image: ${img.src.slice(-30)}`);
      report.score -= 5;
    }
  });

  // 4. Empty containers that shouldn't be empty
  document
    .querySelectorAll(".ant-layout-content, .ant-card-body")
    .forEach((el) => {
      if (el.children.length === 0 && el.textContent.trim() === "") {
        report.warnings.push(`Empty container: ${el.className.split(" ")[0]}`);
      }
    });

  // 5. Z-index stacking issues (modal/overlay visible when shouldn't be)
  const modals = document.querySelectorAll(".ant-modal-wrap, .ant-drawer");
  modals.forEach((modal) => {
    if (
      getComputedStyle(modal).display !== "none" &&
      !modal.classList.contains("visible")
    ) {
      report.warnings.push("Possible stray modal/drawer visible");
    }
  });

  report.passed = report.score >= 80;
  return report;
};
```

---

## Anti-Patterns

| Don't                    | Do Instead                     |
| ------------------------ | ------------------------------ |
| Skip layout verification | ALWAYS verify dimensions       |
| Eyeball color matching   | Use browser_evaluate for exact |
| Assume responsive works  | Test at 375px, 768px, 1200px   |
| Skip mobile verification | ALWAYS test mobile viewport    |
| Manual measurement       | Use automated evaluation       |
| Test only one viewport   | Test ALL 4 minimum viewports   |
| Skip resize transitions  | Animate resize and check each  |
| Ignore horizontal scroll | CRITICAL fail if present       |

---

## ⚠️ MANDATORY: UI Bug Reporting (Even When Not Your Task)

> **This rule applies to ALL agents using Playwright/browser, regardless of their current task.**

When you open a page in the browser (for ANY reason — testing, screenshots, verification, debugging), and you notice layout issues, **you MUST report them to the main agent (Sisyphus)**.

### What Counts as Reportable

| Issue                 | Example                                                         |
| --------------------- | --------------------------------------------------------------- |
| Broken layout         | Elements overlapping, content cut off, page structure collapsed |
| Missing table columns | Table exists but columns are invisible or not rendering         |
| Horizontal scroll     | Page has unwanted horizontal scrollbar                          |
| Invisible UI elements | Buttons, forms, or sections that should be visible but aren't   |
| Misaligned components | Cards, grids, or elements clearly out of alignment              |
| Empty containers      | Sections that should have content but render as blank           |
| Mobile layout broken  | Sidebar visible on mobile, bottom nav missing                   |

### How to Report

When you find a UI bug that is NOT part of your current task:

1. **Take a screenshot** → save to `docs/Screenshots/bug-{page-name}-{issue}.png`
2. **Include in your response** to the main agent:

```
🐛 UI BUG FOUND (not my current task):
- Page: [URL or page name]
- Issue: [brief description]
- Screenshot: docs/Screenshots/bug-{name}.png
- Severity: [CRITICAL / WARNING]
```

### Severity Guide

| Severity     | When                                                          |
| ------------ | ------------------------------------------------------------- |
| **CRITICAL** | Page is unusable, data not visible, core functionality broken |
| **WARNING**  | Layout is messy but functional, minor visual issues           |

### Rules

- **NEVER silently ignore** a layout bug you discover
- **ALWAYS report** even if fixing it is not your task
- **DO NOT fix it yourself** unless explicitly asked — just report
- This applies to ALL pages you visit, not just the one you're working on
