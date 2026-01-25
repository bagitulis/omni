/**
 * Data Leak Prevention Tests
 *
 * Tests to ensure no sensitive data is exposed:
 * 1. Passwords never returned in API responses
 * 2. Tokens never exposed in logs or errors
 * 3. Internal IDs/paths not leaked
 * 4. SQL queries not exposed in errors
 */

import { describe, it, expect } from "@jest/globals";

// ============================================
// TEST CONFIGURATION
// ============================================

const TENANT_ID = "yumna_bertigamart";
const BASE_URL = process.env.TEST_BASE_URL || "http://localhost:3000";

// ============================================
// SENSITIVE PATTERNS TO CHECK
// ============================================

const SENSITIVE_PATTERNS = [
  /password/i,
  /secret/i,
  /token[^i]/i, // token but not "tokenize"
  /api[_-]?key/i,
  /private[_-]?key/i,
  /access[_-]?token/i,
  /refresh[_-]?token/i,
  /authorization.*bearer/i,
  /SELECT.*FROM/i,
  /INSERT.*INTO/i,
  /UPDATE.*SET/i,
  /DELETE.*FROM/i,
  /\.prisma/i,
  /node_modules/i,
  /at\s+\w+\s+\(/i, // Stack trace pattern
];

// ============================================
// HELPER FUNCTIONS
// ============================================

function containsSensitiveData(obj: unknown): string[] {
  const violations: string[] = [];

  function check(value: unknown, path: string): void {
    if (value === null || value === undefined) return;

    if (typeof value === "string") {
      for (const pattern of SENSITIVE_PATTERNS) {
        if (pattern.test(value)) {
          violations.push(`${path}: matches pattern ${pattern}`);
        }
      }
    } else if (Array.isArray(value)) {
      value.forEach((item, index) => check(item, `${path}[${index}]`));
    } else if (typeof value === "object") {
      for (const [key, val] of Object.entries(value)) {
        // Check key names
        if (/password|secret|token|key/i.test(key)) {
          // Key is sensitive - value should be redacted or missing
          if (val && typeof val === "string" && val.length > 0) {
            violations.push(`${path}.${key}: sensitive field has value`);
          }
        }
        check(val, `${path}.${key}`);
      }
    }
  }

  check(obj, "response");
  return violations;
}

async function apiRequest(
  endpoint: string,
  options: {
    method?: string;
    tenantId?: string;
    body?: unknown;
  } = {}
): Promise<{ status: number; data: unknown }> {
  const { method = "GET", tenantId, body } = options;

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };

  if (tenantId) {
    headers["x-tenant-id"] = tenantId;
  }

  const response = await fetch(`${BASE_URL}${endpoint}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  let data: unknown;
  try {
    data = await response.json();
  } catch {
    data = await response.text();
  }

  return { status: response.status, data };
}

// ============================================
// TEST SUITES
// ============================================

describe("Data Leak Prevention", () => {
  describe("User API Responses", () => {
    it("should not expose password hash in user list", async () => {
      const response = await apiRequest("/api/users", {
        tenantId: TENANT_ID,
      });

      if (response.status === 200) {
        const violations = containsSensitiveData(response.data);
        expect(violations).toEqual([]);
      }
    });

    it("should not expose password in user details", async () => {
      const response = await apiRequest("/api/users/me", {
        tenantId: TENANT_ID,
      });

      if (response.status === 200) {
        const data = response.data as Record<string, unknown>;
        expect(data.password).toBeUndefined();
        expect(data.passwordHash).toBeUndefined();
      }
    });
  });

  describe("Error Response Sanitization", () => {
    it("should not expose SQL in error messages", async () => {
      // Try to trigger a database error
      const response = await apiRequest("/api/orders/'; DROP TABLE orders;--", {
        tenantId: TENANT_ID,
      });

      if (response.status >= 400) {
        const violations = containsSensitiveData(response.data);
        const sqlViolations = violations.filter(
          (v) => v.includes("SELECT") || v.includes("INSERT")
        );
        expect(sqlViolations).toEqual([]);
      }
    });

    it("should not expose stack traces in production errors", async () => {
      const response = await apiRequest("/api/non-existent-endpoint-xyz", {
        tenantId: TENANT_ID,
      });

      if (response.status >= 400) {
        const data = response.data as Record<string, unknown>;

        // No stack property
        expect(data.stack).toBeUndefined();

        // Message should not contain file paths
        if (typeof data.message === "string") {
          expect(data.message).not.toContain("at ");
          expect(data.message).not.toContain(".ts:");
          expect(data.message).not.toContain(".js:");
        }
      }
    });
  });

  describe("OAuth Token Protection", () => {
    it("should not expose access tokens in platform config", async () => {
      const response = await apiRequest("/api/platform/shopee/config", {
        tenantId: TENANT_ID,
      });

      if (response.status === 200) {
        const data = response.data as Record<string, unknown>;

        // Access token should be masked or absent
        if (data.accessToken) {
          expect(data.accessToken).toMatch(/^\*+$|^[*]+.*[*]+$/);
        }
      }
    });

    it("should not expose refresh tokens", async () => {
      const response = await apiRequest("/api/platform/shopee/config", {
        tenantId: TENANT_ID,
      });

      if (response.status === 200) {
        const data = response.data as Record<string, unknown>;

        // Refresh token should never be returned
        expect(data.refreshToken).toBeUndefined();
      }
    });
  });

  describe("Internal Path Protection", () => {
    it("should not expose internal file paths in errors", async () => {
      const response = await apiRequest("/api/trigger-error", {
        tenantId: TENANT_ID,
      });

      if (response.status >= 400) {
        const data = response.data as { message?: string };

        if (data.message) {
          // Should not contain internal paths
          expect(data.message).not.toMatch(/\/app\//);
          expect(data.message).not.toMatch(/\/src\//);
          expect(data.message).not.toMatch(/C:\\/);
          expect(data.message).not.toMatch(/node_modules/);
        }
      }
    });
  });
});

describe("XSS Prevention", () => {
  describe("Input Sanitization", () => {
    it("should escape HTML in product names", async () => {
      const maliciousName = '<script>alert("xss")</script>';

      const response = await apiRequest("/api/products", {
        method: "POST",
        tenantId: TENANT_ID,
        body: {
          name: maliciousName,
          price: 10000,
        },
      });

      if (response.status === 200 || response.status === 201) {
        const data = response.data as { name?: string };

        if (data.name) {
          // Should be escaped or stripped
          expect(data.name).not.toContain("<script>");
        }
      }
    });
  });
});

describe("SQL Injection Prevention", () => {
  describe("Query Parameter Injection", () => {
    it("should prevent SQL injection in order ID", async () => {
      const maliciousId = "1'; DROP TABLE orders; --";

      const response = await apiRequest(
        `/api/orders/${encodeURIComponent(maliciousId)}`,
        { tenantId: TENANT_ID }
      );

      // Should not crash the server
      expect(response.status).not.toBe(500);

      // Should return 404 or 400, not execute the SQL
      expect([400, 404]).toContain(response.status);
    });

    it("should prevent SQL injection in search params", async () => {
      const maliciousSearch = "'; SELECT * FROM users; --";

      const response = await apiRequest(
        `/api/orders?search=${encodeURIComponent(maliciousSearch)}`,
        { tenantId: TENANT_ID }
      );

      // Should not crash
      expect(response.status).not.toBe(500);
    });
  });
});
