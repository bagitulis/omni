# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: booking.spec.ts >> Booking Tab UI Flows >> Empty state shows correct copy when no bookings
- Location: e2e/booking.spec.ts:193:3

# Error details

```
Error: page.waitForURL: Target page, context or browser has been closed
=========================== logs ===========================
waiting for navigation until "load"
  navigated to "http://localhost:5174/login"
============================================================
```

```
Error: page.evaluate: Target page, context or browser has been closed
```

```
Error: browserContext.close: Target page, context or browser has been closed
```