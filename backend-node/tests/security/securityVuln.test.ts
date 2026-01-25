/**
 * Security Vulnerability Testing Suite
 *
 * Comprehensive tests for common vulnerabilities:
 * - SQL Injection
 * - XSS (Cross-Site Scripting)
 * - CSRF (Cross-Site Request Forgery)
 * - Authentication bypass
 * - Authorization bypass
 * - Tenant isolation
 * - Error message leakage
 * - Sensitive data exposure
 */

/* eslint-disable @typescript-eslint/no-unused-vars */
import { describe, it, expect } from "@jest/globals";

// ============================================
// TEST CONFIGURATION
// ============================================

const BASE_URL = process.env.TEST_BASE_URL || "http://localhost:3000";
const TENANT_A = "yumna_bertigamart";
const TENANT_B = "tika_nusseyba";

// Test credentials (for authenticated tests) - kept for future expansion
export const TEST_USER = {
  username: process.env.TEST_USERNAME || "test_user",
  password: process.env.TEST_PASSWORD || "TestPassword123!",
};

// ============================================
// HELPER FUNCTIONS
// ============================================

interface ApiResponse {
  status: number;
  data: unknown;
  headers: Headers;
}

async function request(
  endpoint: string,
  options: {
    method?: string;
    headers?: Record<string, string>;
    body?: unknown;
  } = {}
): Promise<ApiResponse> {
  const { method = "GET", headers = {}, body } = options;

  const response = await fetch(`${BASE_URL}${endpoint}`, {
    method,
    headers: {
      "Content-Type": "application/json",
      ...headers,
    },
    body: body ? JSON.stringify(body) : undefined,
  });

  let data: unknown;
  try {
    data = await response.json();
  } catch {
    data = await response.text();
  }

  return { status: response.status, data, headers: response.headers };
}

// Helper for authenticated requests - exported for use in other test files
export async function authenticatedRequest(
  endpoint: string,
  token: string,
  options: {
    method?: string;
    tenantId?: string;
    body?: unknown;
  } = {}
): Promise<ApiResponse> {
  const { method = "GET", tenantId = TENANT_A, body } = options;

  return request(endpoint, {
    method,
    headers: {
      Authorization: `Bearer ${token}`,
      "x-tenant-id": tenantId,
    },
    body,
  });
}

// ============================================
// SQL INJECTION TESTS
// ============================================

describe("SQL Injection Prevention", () => {
  const SQL_PAYLOADS = [
    "1'; DROP TABLE users;--",
    "1 OR 1=1",
    "1' OR '1'='1",
    "1; SELECT * FROM users--",
    "'; DELETE FROM orders WHERE '1'='1",
    "1 UNION SELECT * FROM users--",
    "1' AND (SELECT COUNT(*) FROM users) > 0--",
    "'; EXEC xp_cmdshell('dir')--",
  ];

  describe("URL Parameter Injection", () => {
    it.each(SQL_PAYLOADS)(
      "should prevent SQL injection in ID: %s",
      async (payload) => {
        const response = await request(
          `/api/orders/${encodeURIComponent(payload)}`,
          {
            headers: { "x-tenant-id": TENANT_A },
          }
        );

        // Should not crash with 500
        expect(response.status).not.toBe(500);

        // Should return 400 or 404
        expect([400, 401, 403, 404]).toContain(response.status);
      }
    );
  });

  describe("Query Parameter Injection", () => {
    it.each(SQL_PAYLOADS)(
      "should prevent SQL injection in search: %s",
      async (payload) => {
        const response = await request(
          `/api/products?search=${encodeURIComponent(payload)}`,
          {
            headers: { "x-tenant-id": TENANT_A },
          }
        );

        expect(response.status).not.toBe(500);
      }
    );
  });

  describe("Body Parameter Injection", () => {
    it.each(SQL_PAYLOADS)(
      "should prevent SQL injection in body: %s",
      async (payload) => {
        const response = await request("/api/auth/login", {
          method: "POST",
          body: {
            username: payload,
            password: payload,
          },
        });

        // Should not expose SQL errors
        const data = response.data as Record<string, unknown>;
        if (data.error && typeof data.error === "string") {
          expect(data.error.toLowerCase()).not.toContain("sql");
          expect(data.error.toLowerCase()).not.toContain("syntax");
          expect(data.error.toLowerCase()).not.toContain("prisma");
        }
      }
    );
  });
});

// ============================================
// XSS PREVENTION TESTS
// ============================================

describe("XSS Prevention", () => {
  const XSS_PAYLOADS = [
    '<script>alert("XSS")</script>',
    '<img src="x" onerror="alert(\'XSS\')">',
    "<svg onload=\"alert('XSS')\">",
    "javascript:alert('XSS')",
    "<a href=\"javascript:alert('XSS')\">Click</a>",
    "<div onmouseover=\"alert('XSS')\">Hover</div>",
    "{{constructor.constructor('alert(1)')()}}",
    "${alert('XSS')}",
  ];

  describe("Input Sanitization", () => {
    it.each(XSS_PAYLOADS)(
      "should sanitize XSS payload: %s",
      async (payload) => {
        const response = await request("/api/products", {
          method: "POST",
          headers: { "x-tenant-id": TENANT_A },
          body: {
            name: payload,
            description: payload,
          },
        });

        // If created, check the response doesn't contain raw script
        if (response.status === 200 || response.status === 201) {
          const data = response.data as Record<string, unknown>;
          const responseStr = JSON.stringify(data);

          expect(responseStr).not.toContain("<script>");
          expect(responseStr).not.toContain("javascript:");
          expect(responseStr).not.toContain("onerror=");
          expect(responseStr).not.toContain("onload=");
        }
      }
    );
  });
});

// ============================================
// AUTHENTICATION TESTS
// ============================================

describe("Authentication Security", () => {
  describe("No Token", () => {
    it("should reject requests without token", async () => {
      const response = await request("/api/orders", {
        headers: { "x-tenant-id": TENANT_A },
      });

      expect(response.status).toBe(401);
    });
  });

  describe("Invalid Token", () => {
    const INVALID_TOKENS = [
      "invalid-token",
      "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.payload",
      "Bearer invalid",
      "null",
      "undefined",
      "",
    ];

    it.each(INVALID_TOKENS)(
      "should reject invalid token: %s",
      async (token) => {
        const response = await request("/api/orders", {
          headers: {
            Authorization: `Bearer ${token}`,
            "x-tenant-id": TENANT_A,
          },
        });

        expect(response.status).toBe(401);
      }
    );
  });

  describe("Expired Token", () => {
    it("should reject expired tokens", async () => {
      // This is a manually expired token (set exp to past)
      const expiredToken =
        "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIxIiwidXNlcm5hbWUiOiJ0ZXN0IiwiaWF0IjoxNjAwMDAwMDAwLCJleHAiOjE2MDAwMDAwMDF9.invalid";

      const response = await request("/api/orders", {
        headers: {
          Authorization: `Bearer ${expiredToken}`,
          "x-tenant-id": TENANT_A,
        },
      });

      expect(response.status).toBe(401);
    });
  });

  describe("Brute Force Protection", () => {
    it("should handle multiple failed login attempts", async () => {
      const attempts = 10;
      const results: number[] = [];

      for (let i = 0; i < attempts; i++) {
        const response = await request("/api/auth/login", {
          method: "POST",
          body: {
            username: "attacker",
            password: `wrong_password_${i}`,
          },
        });
        results.push(response.status);
      }

      // All should fail with 401, not crash with 500
      // 400 = bad request, 401 = unauthorized, 403 = forbidden, 429 = rate limited
      results.forEach((status) => {
        expect([400, 401, 403, 429]).toContain(status);
      });
    });
  });
});

// ============================================
// AUTHORIZATION TESTS
// ============================================

describe("Authorization Security", () => {
  describe("Role-Based Access", () => {
    it("should reject access to admin routes without admin role", async () => {
      // Try accessing admin route with regular user token
      const response = await request("/api/admin/users", {
        headers: { "x-tenant-id": TENANT_A },
      });

      expect([401, 403]).toContain(response.status);
    });

    it("should reject access to security routes without proper role", async () => {
      const response = await request("/api/security/events", {
        headers: { "x-tenant-id": TENANT_A },
      });

      expect([401, 403]).toContain(response.status);
    });
  });
});

// ============================================
// TENANT ISOLATION TESTS
// ============================================

describe("Tenant Isolation", () => {
  describe("Cross-Tenant Access", () => {
    it("should not expose tenant B data when requesting as tenant A", async () => {
      const responseA = await request("/api/orders", {
        headers: { "x-tenant-id": TENANT_A },
      });

      const responseB = await request("/api/orders", {
        headers: { "x-tenant-id": TENANT_B },
      });

      // Both should fail (no auth) but not leak data
      expect([401, 403]).toContain(responseA.status);
      expect([401, 403]).toContain(responseB.status);
    });

    it("should reject requests without tenant ID", async () => {
      const response = await request("/api/orders");

      expect([400, 401, 403]).toContain(response.status);
    });

    it("should prevent tenant ID injection via query params", async () => {
      const response = await request(`/api/orders?tenantId=${TENANT_B}`, {
        headers: { "x-tenant-id": TENANT_A },
      });

      // Should use header tenant, not query param
      expect(response.status).not.toBe(500);
    });
  });
});

// ============================================
// ERROR MESSAGE LEAKAGE TESTS
// ============================================

describe("Error Message Leakage", () => {
  // Note: "token" is excluded because error messages like "No token provided" are acceptable
  const SENSITIVE_PATTERNS = [
    /password/i,
    /secret/i,
    /access_token|refresh_token|bearer\s+[a-z0-9]/i, // Actual tokens, not the word "token"
    /api_key|apikey/i, // API keys, not generic "key" word
    /SELECT.*FROM/i,
    /INSERT.*INTO/i,
    /\.prisma/i,
    /node_modules/i,
    /at\s+\w+\s+\(/i, // Stack trace
    /Error:.*at/i,
  ];

  it("should not expose sensitive data in 404 errors", async () => {
    const response = await request("/api/non-existent-endpoint-xyz-123");

    const data = response.data as Record<string, unknown>;
    const responseStr = JSON.stringify(data);

    SENSITIVE_PATTERNS.forEach((pattern) => {
      expect(responseStr).not.toMatch(pattern);
    });
  });

  it("should not expose sensitive data in 500 errors", async () => {
    // Try to trigger a server error
    const response = await request("/api/orders/trigger-error-test", {
      headers: { "x-tenant-id": TENANT_A },
    });

    if (response.status >= 500) {
      const data = response.data as Record<string, unknown>;
      const responseStr = JSON.stringify(data);

      SENSITIVE_PATTERNS.forEach((pattern) => {
        expect(responseStr).not.toMatch(pattern);
      });
    }
  });

  it("should not expose stack traces in production", async () => {
    const response = await request("/api/orders/invalid", {
      headers: { "x-tenant-id": TENANT_A },
    });

    const data = response.data as Record<string, unknown>;

    expect(data.stack).toBeUndefined();
    expect(data.trace).toBeUndefined();
  });
});

// ============================================
// SENSITIVE DATA EXPOSURE TESTS
// ============================================

describe("Sensitive Data Exposure", () => {
  describe("Password Protection", () => {
    it("should never expose password hashes in API responses", async () => {
      const response = await request("/api/users", {
        headers: { "x-tenant-id": TENANT_A },
      });

      if (response.status === 200) {
        const responseStr = JSON.stringify(response.data);

        expect(responseStr).not.toMatch(/password/i);
        expect(responseStr).not.toMatch(/\$2[ayb]\$.{56}/); // bcrypt hash pattern
      }
    });
  });

  describe("Token Protection", () => {
    it("should not expose access tokens in API responses", async () => {
      const response = await request("/api/platform/shopee/config", {
        headers: { "x-tenant-id": TENANT_A },
      });

      if (response.status === 200) {
        const data = response.data as Record<string, unknown>;

        // Access tokens should be masked or absent
        if (data.accessToken) {
          expect(data.accessToken).toMatch(/^\*+$/);
        }

        // Refresh tokens should never be returned
        expect(data.refreshToken).toBeUndefined();
      }
    });
  });

  describe("Internal Path Protection", () => {
    it("should not expose internal file paths", async () => {
      const response = await request("/api/health");

      const responseStr = JSON.stringify(response.data);

      expect(responseStr).not.toMatch(/\/app\//);
      expect(responseStr).not.toMatch(/C:\\/);
      expect(responseStr).not.toMatch(/node_modules/);
    });
  });
});

// ============================================
// SECURITY HEADERS TESTS
// ============================================

describe("Security Headers", () => {
  it("should have required security headers", async () => {
    const response = await request("/api/health");
    const { headers } = response;

    // Check for security headers (from Helmet)
    expect(headers.get("x-content-type-options")).toBe("nosniff");
    expect(headers.get("x-frame-options")).toBe("SAMEORIGIN");
  });

  it("should not expose server version", async () => {
    const response = await request("/api/health");

    // X-Powered-By should be removed
    expect(response.headers.get("x-powered-by")).toBeNull();
  });
});

// ============================================
// CORS SECURITY TESTS
// ============================================

describe("CORS Security", () => {
  it("should allow requests from whitelisted origins", async () => {
    const response = await fetch(`${BASE_URL}/api/health`, {
      headers: {
        Origin: "https://yndigital.my.id",
        Referer: "https://yndigital.my.id/",
        "x-forwarded-for": "1.2.3.4",
      },
    });

    expect(response.ok).toBe(true);
  });

  it("should reject requests from unknown origins", async () => {
    const response = await fetch(`${BASE_URL}/api/orders`, {
      headers: {
        Origin: "https://malicious-site.com",
        "x-tenant-id": TENANT_A,
      },
    });

    // Should be rejected
    expect([403, 500]).toContain(response.status);
  });
});
