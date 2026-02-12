/**
 * HTML Sanitizer Utilities
 * SECURITY: Prevents XSS attacks by sanitizing user-generated content
 */

// Allowed HTML tags for display (whitelist approach)
export const ALLOWED_TAGS = new Set<string>([
  "b",
  "i",
  "u",
  "strong",
  "em",
  "br",
  "p",
  "span",
  "div",
  "ul",
  "ol",
  "li",
  "a",
  "table",
  "tr",
  "td",
  "th",
  "thead",
  "tbody",
]);

// Allowed attributes (very restrictive)
export const ALLOWED_ATTRIBUTES: Record<string, Set<string>> = {
  a: new Set<string>(["href", "title"]),
  span: new Set<string>(["style"]),
  div: new Set<string>(["style"]),
  td: new Set<string>(["colspan", "rowspan"]),
  th: new Set<string>(["colspan", "rowspan"]),
};

// Allowed style properties (to prevent CSS injection)
export const ALLOWED_STYLES = new Set<string>([
  "color",
  "background-color",
  "font-weight",
  "font-style",
  "text-align",
  "text-decoration",
]);

/**
 * Escape HTML entities to prevent XSS.
 */
export function escapeHtml(text: string | null | undefined): string {
  if (!text) {
    return "";
  }

  return text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

/**
 * Strip all HTML tags and return plain text.
 * SECURITY: Safe alternative to innerHTML parsing.
 */
export function stripHtmlSafe(html: string | null | undefined): string {
  if (!html) {
    return "";
  }

  return html
    .replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, "")
    .replace(/<style\b[^<]*(?:(?!<\/style>)<[^<]*)*<\/style>/gi, "")
    .replace(/<[^>]+>/g, "")
    .replace(/&nbsp;/g, " ")
    .replace(/&amp;/g, "&")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&quot;/g, '"')
    .replace(/&#039;/g, "'")
    .trim();
}

/**
 * Validate and sanitize inline style.
 */
function sanitizeStyle(style: string): string {
  const sanitizedParts: string[] = [];
  const parts = style
    .split(";")
    .map((part) => part.trim())
    .filter(Boolean);

  for (const part of parts) {
    const [prop, value] = part.split(":").map((segment) => segment.trim());
    if (prop && value && ALLOWED_STYLES.has(prop.toLowerCase())) {
      if (!/url\s*\(|expression\s*\(|javascript:/i.test(value)) {
        sanitizedParts.push(`${prop}: ${value}`);
      }
    }
  }

  return sanitizedParts.join("; ");
}

/**
 * Sanitize HTML content using whitelist approach.
 */
export function sanitizeHtml(html: string | null | undefined): string {
  if (!html) {
    return "";
  }

  const template = document.createElement("template");
  template.innerHTML = html;

  const sanitizeNode = (node: Node): void => {
    if (node.nodeType === Node.ELEMENT_NODE) {
      const element = node as Element;
      const tagName = element.tagName.toLowerCase();

      if (!ALLOWED_TAGS.has(tagName)) {
        element.remove();
        return;
      }

      const allowedAttrs = ALLOWED_ATTRIBUTES[tagName] || new Set<string>();
      const attrsToRemove: string[] = [];

      for (const attr of Array.from(element.attributes)) {
        const attrName = attr.name.toLowerCase();

        if (!allowedAttrs.has(attrName)) {
          attrsToRemove.push(attr.name);
          continue;
        }

        if (attrName === "style") {
          const sanitizedStyle = sanitizeStyle(attr.value);
          if (sanitizedStyle) {
            element.setAttribute("style", sanitizedStyle);
          } else {
            attrsToRemove.push("style");
          }
          continue;
        }

        if (attrName === "href") {
          const href = attr.value.toLowerCase().trim();
          if (
            href.startsWith("javascript:") ||
            href.startsWith("data:") ||
            href.startsWith("vbscript:")
          ) {
            attrsToRemove.push("href");
          }
        }
      }

      for (const attrName of attrsToRemove) {
        element.removeAttribute(attrName);
      }

      for (const attr of Array.from(element.attributes)) {
        if (attr.name.toLowerCase().startsWith("on")) {
          element.removeAttribute(attr.name);
        }
      }
    }

    for (const child of Array.from(node.childNodes)) {
      sanitizeNode(child);
    }
  };

  sanitizeNode(template.content);
  return template.innerHTML;
}

/**
 * Format status text with safe HTML.
 */
export function formatSafeStatusHtml(text: string): string {
  return escapeHtml(text)
    .replace(/\n/g, "<br>")
    .replace(/✅/g, '<span style="color: var(--color-success);">✅</span>')
    .replace(/❌/g, '<span style="color: var(--color-error);">❌</span>')
    .replace(/⚠️/g, '<span style="color: var(--color-warning);">⚠️</span>');
}

export default {
  escapeHtml,
  stripHtmlSafe,
  sanitizeHtml,
  formatSafeStatusHtml,
};
