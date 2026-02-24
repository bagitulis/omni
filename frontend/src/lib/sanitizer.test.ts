import { describe, it, expect } from "vitest";
import {
  escapeHtml,
  stripHtmlSafe,
  sanitizeHtml,
  formatSafeStatusHtml,
  ALLOWED_TAGS,
  ALLOWED_ATTRIBUTES,
  ALLOWED_STYLES,
} from "./sanitizer";

describe("escapeHtml", () => {
  it("returns empty string for null", () => {
    expect(escapeHtml(null)).toBe("");
  });

  it("returns empty string for undefined", () => {
    expect(escapeHtml(undefined)).toBe("");
  });

  it("returns empty string for empty string", () => {
    expect(escapeHtml("")).toBe("");
  });

  it("escapes ampersand", () => {
    expect(escapeHtml("a & b")).toBe("a &amp; b");
  });

  it("escapes less-than sign", () => {
    expect(escapeHtml("<script>")).toBe("&lt;script&gt;");
  });

  it("escapes double quotes", () => {
    expect(escapeHtml('"hello"')).toBe("&quot;hello&quot;");
  });

  it("escapes single quotes", () => {
    expect(escapeHtml("it's")).toBe("it&#039;s");
  });

  it("escapes XSS payload", () => {
    const xss = '<script>alert("XSS")</script>';
    const result = escapeHtml(xss);
    expect(result).not.toContain("<script>");
    expect(result).toContain("&lt;script&gt;");
  });

  it("handles plain text without modification", () => {
    expect(escapeHtml("hello world")).toBe("hello world");
  });

  it("escapes all special chars in one string", () => {
    const result = escapeHtml(`<div class="x">&'text'</div>`);
    expect(result).not.toContain("<");
    expect(result).not.toContain(">");
    expect(result).not.toContain('"');
    expect(result).not.toContain("'");
    expect(result).not.toContain("&d");
  });
});

describe("stripHtmlSafe", () => {
  it("returns empty string for null", () => {
    expect(stripHtmlSafe(null)).toBe("");
  });

  it("returns empty string for undefined", () => {
    expect(stripHtmlSafe(undefined)).toBe("");
  });

  it("returns empty string for empty string", () => {
    expect(stripHtmlSafe("")).toBe("");
  });

  it("strips basic HTML tags", () => {
    expect(stripHtmlSafe("<b>bold</b>")).toBe("bold");
  });

  it("strips script tags and content", () => {
    const result = stripHtmlSafe('<script>alert("xss")</script>Hello');
    expect(result).not.toContain("script");
    expect(result).not.toContain("alert");
    expect(result).toContain("Hello");
  });

  it("strips style tags and content", () => {
    const result = stripHtmlSafe("<style>body { color: red; }</style>Text");
    expect(result).not.toContain("style");
    expect(result).not.toContain("color");
    expect(result).toContain("Text");
  });

  it("decodes HTML entities", () => {
    expect(stripHtmlSafe("&amp;")).toBe("&");
    expect(stripHtmlSafe("&lt;")).toBe("<");
    expect(stripHtmlSafe("&gt;")).toBe(">");
    expect(stripHtmlSafe("&quot;")).toBe('"');
    expect(stripHtmlSafe("&#039;")).toBe("'");
  });

  it("replaces &nbsp; with space", () => {
    const result = stripHtmlSafe("a&nbsp;b");
    expect(result).toBe("a b");
  });

  it("trims whitespace from result", () => {
    const result = stripHtmlSafe("  <p>text</p>  ");
    expect(result).toBe("text");
  });

  it("handles nested tags", () => {
    expect(stripHtmlSafe("<div><p><b>deep</b></p></div>")).toBe("deep");
  });
});

describe("sanitizeHtml", () => {
  it("returns empty string for null", () => {
    expect(sanitizeHtml(null)).toBe("");
  });

  it("returns empty string for undefined", () => {
    expect(sanitizeHtml(undefined)).toBe("");
  });

  it("returns empty string for empty string", () => {
    expect(sanitizeHtml("")).toBe("");
  });

  it("preserves allowed tags like <b> and <i>", () => {
    const result = sanitizeHtml("<b>bold</b> and <i>italic</i>");
    expect(result).toContain("<b>bold</b>");
    expect(result).toContain("<i>italic</i>");
  });

  it("removes disallowed tags like <script>", () => {
    const result = sanitizeHtml('<script>alert("xss")</script>');
    expect(result).not.toContain("<script>");
  });

  it("removes onclick and other on* event handlers", () => {
    const result = sanitizeHtml('<b onclick="evil()">text</b>');
    expect(result).not.toContain("onclick");
  });

  it("removes disallowed attributes", () => {
    const result = sanitizeHtml('<b class="danger">text</b>');
    expect(result).not.toContain("class=");
  });

  it("blocks javascript: href", () => {
    const result = sanitizeHtml('<a href="javascript:alert(1)">click</a>');
    expect(result).not.toContain("javascript:");
  });

  it("blocks data: href", () => {
    const result = sanitizeHtml(
      '<a href="data:text/html,<h1>xss</h1>">click</a>',
    );
    expect(result).not.toContain('href="data:');
  });

  it("allows valid href on anchor", () => {
    const result = sanitizeHtml('<a href="https://example.com">link</a>');
    expect(result).toContain("href=");
  });

  it("allows colspan on <td>", () => {
    const result = sanitizeHtml(
      '<table><tbody><tr><td colspan="2">cell</td></tr></tbody></table>',
    );
    expect(result).toContain('colspan="2"');
  });

  it("sanitizes CSS injection in style attribute", () => {
    const result = sanitizeHtml(
      '<span style="color: red; background-image: url(javascript:evil)">text</span>',
    );
    // url() value should be stripped
    expect(result).not.toContain("url(");
  });
});

describe("formatSafeStatusHtml", () => {
  it("escapes HTML in text before formatting", () => {
    const result = formatSafeStatusHtml("<script>evil</script>");
    expect(result).not.toContain("<script>");
    expect(result).toContain("&lt;script&gt;");
  });

  it("converts newlines to <br>", () => {
    const result = formatSafeStatusHtml("line1\nline2");
    expect(result).toContain("<br>");
  });

  it("wraps ✅ emoji in success span", () => {
    const result = formatSafeStatusHtml("✅ done");
    expect(result).toContain("color: var(--color-success)");
    expect(result).toContain("✅");
  });

  it("wraps ❌ emoji in error span", () => {
    const result = formatSafeStatusHtml("❌ failed");
    expect(result).toContain("color: var(--color-error)");
    expect(result).toContain("❌");
  });

  it("wraps ⚠️ emoji in warning span", () => {
    const result = formatSafeStatusHtml("⚠️ warning");
    expect(result).toContain("color: var(--color-warning)");
    expect(result).toContain("⚠️");
  });

  it("returns plain text unchanged when no special chars", () => {
    const result = formatSafeStatusHtml("all good");
    expect(result).toBe("all good");
  });
});

describe("ALLOWED_TAGS constant", () => {
  it("includes expected safe tags", () => {
    expect(ALLOWED_TAGS.has("b")).toBe(true);
    expect(ALLOWED_TAGS.has("a")).toBe(true);
    expect(ALLOWED_TAGS.has("table")).toBe(true);
    expect(ALLOWED_TAGS.has("span")).toBe(true);
  });

  it("does not include dangerous tags", () => {
    expect(ALLOWED_TAGS.has("script")).toBe(false);
    expect(ALLOWED_TAGS.has("iframe")).toBe(false);
    expect(ALLOWED_TAGS.has("style")).toBe(false);
  });
});

describe("ALLOWED_ATTRIBUTES constant", () => {
  it("allows href and title on anchor", () => {
    expect(ALLOWED_ATTRIBUTES["a"].has("href")).toBe(true);
    expect(ALLOWED_ATTRIBUTES["a"].has("title")).toBe(true);
  });

  it("allows style on span and div", () => {
    expect(ALLOWED_ATTRIBUTES["span"].has("style")).toBe(true);
    expect(ALLOWED_ATTRIBUTES["div"].has("style")).toBe(true);
  });
});

describe("ALLOWED_STYLES constant", () => {
  it("includes common safe style properties", () => {
    expect(ALLOWED_STYLES.has("color")).toBe(true);
    expect(ALLOWED_STYLES.has("font-weight")).toBe(true);
    expect(ALLOWED_STYLES.has("text-align")).toBe(true);
  });
});
