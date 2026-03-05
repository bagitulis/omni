import { describe, it, expect } from "vitest";
import { sanitizeForUser } from "./notificationSecurity";

const SAFE_FALLBACK =
  "An error occurred. Please try again or contact support.";

describe("sanitizeForUser", () => {
  it("returns fallback for null/undefined/empty", () => {
    expect(sanitizeForUser(null)).toBe(SAFE_FALLBACK);
    expect(sanitizeForUser(undefined)).toBe(SAFE_FALLBACK);
    expect(sanitizeForUser("")).toBe(SAFE_FALLBACK);
  });

  it("passes through safe messages", () => {
    expect(sanitizeForUser("Failed to sync orders")).toBe(
      "Failed to sync orders",
    );
    expect(sanitizeForUser("Permission denied")).toBe("Permission denied");
    expect(sanitizeForUser("Product not found")).toBe("Product not found");
  });

  it("blocks SQL-related errors", () => {
    expect(sanitizeForUser('column "user_id" does not exist')).toBe(
      SAFE_FALLBACK,
    );
    expect(sanitizeForUser("query failed: SELECT * FROM users")).toBe(
      SAFE_FALLBACK,
    );
    expect(sanitizeForUser("database connection refused")).toBe(SAFE_FALLBACK);
  });

  it("blocks stack traces", () => {
    expect(sanitizeForUser("Error at main.handleRequest (server.go:42)")).toBe(
      SAFE_FALLBACK,
    );
  });

  it("blocks internal paths", () => {
    expect(sanitizeForUser("error in /internal/services/sync.go")).toBe(
      SAFE_FALLBACK,
    );
    expect(sanitizeForUser("failed at /pkg/shopee/client.go")).toBe(
      SAFE_FALLBACK,
    );
  });

  it("blocks credential patterns", () => {
    expect(sanitizeForUser("token=abc123xyz")).toBe(SAFE_FALLBACK);
    expect(sanitizeForUser("jwt: invalid signature")).toBe(SAFE_FALLBACK);
    expect(sanitizeForUser("apikey is expired")).toBe(SAFE_FALLBACK);
  });

  it("blocks password patterns", () => {
    expect(sanitizeForUser("password: incorrect")).toBe(SAFE_FALLBACK);
  });

  it("blocks internal URLs", () => {
    expect(sanitizeForUser("connect to localhost:3000 failed")).toBe(
      SAFE_FALLBACK,
    );
    expect(sanitizeForUser("127.0.0.1:8080 unreachable")).toBe(SAFE_FALLBACK);
  });

  it("blocks Go panic info", () => {
    expect(sanitizeForUser("panic: runtime error")).toBe(SAFE_FALLBACK);
    expect(sanitizeForUser("goroutine 1 [running]")).toBe(SAFE_FALLBACK);
  });

  it("blocks low-level network errors", () => {
    expect(sanitizeForUser("ECONNREFUSED")).toBe(SAFE_FALLBACK);
    expect(sanitizeForUser("ECONNRESET by peer")).toBe(SAFE_FALLBACK);
  });

  it("truncates messages longer than 200 chars", () => {
    const long = "A".repeat(250);
    const result = sanitizeForUser(long);
    expect(result.length).toBe(200);
    expect(result.endsWith("...")).toBe(true);
  });
});
