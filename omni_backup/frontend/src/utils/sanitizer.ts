/**
 * HTML Sanitizer Utilities
 * SECURITY: Prevents XSS attacks by sanitizing user-generated content
 */

// Allowed HTML tags for display (whitelist approach)
const ALLOWED_TAGS = new Set([
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
const ALLOWED_ATTRIBUTES: Record<string, Set<string>> = {
  a: new Set(["href", "title"]),
  span: new Set(["style"]),
  div: new Set(["style"]),
  td: new Set(["colspan", "rowspan"]),
  th: new Set(["colspan", "rowspan"]),
};

// Allowed style properties (to prevent CSS injection)
const ALLOWED_STYLES = new Set([
  "color",
  "background-color",
  "font-weight",
  "font-style",
  "text-align",
  "text-decoration",
]);

/**
 * Escape HTML entities to prevent XSS
 */
export function escapeHtml(text: string | null | undefined): string {
  if (!text) return "";
  return text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

/**
 * Strip all HTML tags and return plain text
 * SECURITY: Safe alternative to innerHTML parsing
 */
export function stripHtmlSafe(html: string | null | undefined): string {
  if (!html) return "";
  // Use regex to strip tags instead of DOM parsing (avoids script execution)
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
 * Validate and sanitize inline style
 */
function sanitizeStyle(style: string): string {
  const sanitizedParts: string[] = [];
  const parts = style
    .split(";")
    .map((p) => p.trim())
    .filter(Boolean);

  for (const part of parts) {
    const [prop, value] = part.split(":").map((s) => s.trim());
    if (prop && value && ALLOWED_STYLES.has(prop.toLowerCase())) {
      // Additional validation: no url(), expression(), etc.
      if (!/url\s*\(|expression\s*\(|javascript:/i.test(value)) {
        sanitizedParts.push(`${prop}: ${value}`);
      }
    }
  }

  return sanitizedParts.join("; ");
}

/**
 * Sanitize HTML content - whitelist approach
 * Only allows safe tags and attributes
 */
export function sanitizeHtml(html: string | null | undefined): string {
  if (!html) return "";

  // Create a temporary element for parsing
  const template = document.createElement("template");
  template.innerHTML = html;

  const sanitizeNode = (node: Node): void => {
    if (node.nodeType === Node.ELEMENT_NODE) {
      const element = node as Element;
      const tagName = element.tagName.toLowerCase();

      // Remove disallowed tags entirely
      if (!ALLOWED_TAGS.has(tagName)) {
        element.remove();
        return;
      }

      // Remove disallowed attributes
      const allowedAttrs = ALLOWED_ATTRIBUTES[tagName] || new Set();
      const attrsToRemove: string[] = [];

      for (const attr of Array.from(element.attributes)) {
        if (!allowedAttrs.has(attr.name.toLowerCase())) {
          attrsToRemove.push(attr.name);
        } else if (attr.name.toLowerCase() === "style") {
          // Sanitize style attribute
          const sanitized = sanitizeStyle(attr.value);
          if (sanitized) {
            element.setAttribute("style", sanitized);
          } else {
            attrsToRemove.push("style");
          }
        } else if (attr.name.toLowerCase() === "href") {
          // Validate href - only allow safe protocols
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

      for (const attr of attrsToRemove) {
        element.removeAttribute(attr);
      }

      // Remove event handlers (onclick, onmouseover, etc.)
      for (const attr of Array.from(element.attributes)) {
        if (attr.name.toLowerCase().startsWith("on")) {
          element.removeAttribute(attr.name);
        }
      }
    }

    // Recursively sanitize child nodes
    for (const child of Array.from(node.childNodes)) {
      sanitizeNode(child);
    }
  };

  sanitizeNode(template.content);
  return template.innerHTML;
}

/**
 * Format status text with safe HTML
 * Used for token status display
 */
export function formatSafeStatusHtml(text: string): string {
  // Only allow specific emoji replacements with safe HTML
  return escapeHtml(text)
    .replace(/\n/g, "<br>")
    .replace(/✅/g, '<span style="color: #27ae60;">✅</span>')
    .replace(/❌/g, '<span style="color: #e74c3c;">❌</span>')
    .replace(/⚠️/g, '<span style="color: #f39c12;">⚠️</span>');
}

export default {
  escapeHtml,
  stripHtmlSafe,
  sanitizeHtml,
  formatSafeStatusHtml,
};
