---
alwaysApply: true
description: Proxy configuration - disable vision features when not supported
---

# Proxy Mode (No Vision)

> **Context**: Using local proxy (localhost:20128) that may not support vision APIs.

## Visual Features - DISABLED

When proxy doesn't support vision, **SKIP** these features:

- ❌ `browser_take_screenshot` - Don't capture screenshots
- ❌ `Read` image files - Can't analyze images
- ❌ Screenshot-based UI verification - Use text/DOM verification instead
- ❌ Visual diff tools - Use text diff instead

## Alternative Approaches

**Instead of screenshots:**
```bash
# Use DOM inspection
browser_snapshot
browser_evaluate "document.body.innerText"

# Use accessibility tree
browser_snapshot

# Use text output
cat file.txt
```

**Instead of image analysis:**
```bash
# Describe in text
"File contains: [text description]"

# Use structured data
browser_evaluate "JSON.stringify({...})"
```

## Check Vision Support

If unsure, test once:
```bash
curl -X POST http://localhost:20128/v1/messages \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":[{"type":"image","source":{"type":"base64","media_type":"image/png","data":"test"}}]}]}'
```

If returns error → vision not supported → follow rules above.
If returns 200 → vision works → can use screenshots normally.

## Override

User can override with: "enable vision" or "proxy supports vision now"
