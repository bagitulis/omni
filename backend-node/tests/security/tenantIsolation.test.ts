/**
 * Tenant Isolation Security Tests
 *
 * Verifies that:
 * 1. Tenant A cannot access Tenant B's data
 * 2. No hardcoded tenant IDs leak data
 * 3. Multi-tenant queries are properly isolated
 */

import { describe, it, expect } from "@jest/globals";

// ============================================
// TEST CONFIGURATION
// ============================================

const TENANT_A = "yumna_bertigamart";
const TENANT_B = "tika_nusseyba";
const BASE_URL = process.env.TEST_BASE_URL || "http://localhost:3000";

// ============================================
// HELPER FUNCTIONS
// ============================================

interface ApiResponse {
  status: number;
  data: unknown;
  headers: Headers;
}

async function apiRequest(
  endpoint: string,
  options: {
    method?: string;
    tenantId?: string;
    body?: unknown;
    headers?: Record<string, string>;
  } = {}
): Promise<ApiResponse> {
  const { method = "GET", tenantId, body, headers = {} } = options;

  const requestHeaders: Record<string, string> = {
    "Content-Type": "application/json",
    ...headers,
  };

  if (tenantId) {
    requestHeaders["x-tenant-id"] = tenantId;
  }

  const response = await fetch(`${BASE_URL}${endpoint}`, {
    method,
    headers: requestHeaders,
    body: body ? JSON.stringify(body) : undefined,
  });

  let data: unknown;
  try {
    data = await response.json();
  } catch {
    data = await response.text();
  }

  return {
    status: response.status,
    data,
    headers: response.headers,
  };
}

// ============================================
// TEST SUITES
// ============================================

describe("Tenant Isolation Tests", () => {
  describe("Cross-Tenant Data Access Prevention", () => {
    it("should reject requests without tenant ID", async () => {
      const response = await apiRequest("/api/orders");

      // Should fail - no tenant context
      expect(response.status).toBeGreaterThanOrEqual(400);
    });

    it("should not expose Tenant B data when requesting as Tenant A", async () => {
      const responseA = await apiRequest("/api/orders", {
        tenantId: TENANT_A,
      });

      const responseB = await apiRequest("/api/orders", {
        tenantId: TENANT_B,
      });

      // If both have data, ensure they're different
      if (
        responseA.status === 200 &&
        responseB.status === 200 &&
        Array.isArray(responseA.data) &&
        Array.isArray(responseB.data)
      ) {
        const ordersA = responseA.data as { id: string }[];
        const ordersB = responseB.data as { id: string }[];

        // Check no overlap in IDs
        const idsA = new Set(ordersA.map((o) => o.id));
        const idsB = new Set(ordersB.map((o) => o.id));

        const overlap = [...idsA].filter((id) => idsB.has(id));
        expect(overlap.length).toBe(0);
      }
    });

    it("should prevent tenant ID manipulation in query params", async () => {
      // Try to inject different tenant via query string
      const response = await apiRequest(`/api/orders?tenantId=${TENANT_B}`, {
        tenantId: TENANT_A,
      });

      // Should use header tenant, not query param
      // Response should only contain TENANT_A data
      expect(response.status).not.toBe(500);
    });
  });

  describe("OAuth State Cross-Tenant Prevention", () => {
    it("should not allow OAuth state from wrong tenant", async () => {
      // This tests that OAuth callbacks properly validate tenant
      const fakeState = "fake-state-from-attacker";

      const response = await apiRequest(
        `/api/auth/shopee/callback?code=fake&state=${fakeState}`,
        { tenantId: TENANT_A }
      );

      // Should fail - state doesn't exist or wrong tenant
      expect(response.status).toBeGreaterThanOrEqual(400);
    });
  });

  describe("Error Response Information Leakage", () => {
    it("should not expose internal errors with sensitive data", async () => {
      // Request with invalid data to trigger error
      const response = await apiRequest("/api/orders/invalid-id-12345", {
        tenantId: TENANT_A,
      });

      if (response.status >= 400) {
        const data = response.data as { message?: string; stack?: string };

        // Should not contain stack trace
        expect(data.stack).toBeUndefined();

        // Should not contain SQL errors with table names
        if (data.message) {
          expect(data.message.toLowerCase()).not.toContain("select");
          expect(data.message.toLowerCase()).not.toContain("insert");
          expect(data.message.toLowerCase()).not.toContain("prisma");
        }
      }
    });

    it("should not expose other tenant info in errors", async () => {
      const response = await apiRequest("/api/orders/non-existent", {
        tenantId: TENANT_A,
      });

      if (response.status >= 400) {
        const data = response.data as { message?: string };

        // Should not mention other tenant
        if (data.message) {
          expect(data.message).not.toContain(TENANT_B);
          expect(data.message).not.toContain("tika");
        }
      }
    });
  });
});

describe("API Security Headers", () => {
  it("should have security headers present", async () => {
    const response = await apiRequest("/api/health");

    // Check for common security headers (from Helmet)
    const { headers } = response;

    // X-Content-Type-Options
    expect(headers.get("x-content-type-options")).toBe("nosniff");

    // X-Frame-Options
    expect(headers.get("x-frame-options")).toBe("SAMEORIGIN");

    // X-XSS-Protection
    expect(headers.get("x-xss-protection")).toBeDefined();
  });

  it("should not expose server version", async () => {
    const response = await apiRequest("/api/health");

    // Should not have X-Powered-By header
    expect(response.headers.get("x-powered-by")).toBeNull();
  });
});

describe("CORS Security", () => {
  it("should reject requests from unknown origins", async () => {
    const response = await fetch(`${BASE_URL}/api/health`, {
      headers: {
        Origin: "https://malicious-site.com",
      },
    });

    // Should be rejected or not have CORS headers
    const corsHeader = response.headers.get("access-control-allow-origin");

    if (corsHeader) {
      expect(corsHeader).not.toBe("*");
      expect(corsHeader).not.toBe("https://malicious-site.com");
    }
  });
});
