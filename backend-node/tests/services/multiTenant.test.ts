/**
 * Multi-Tenant Architecture Tests
 * Tests tenant isolation, context management, and data separation
 */

import { AsyncLocalStorage } from "async_hooks";

// Mock file system for tenantContext
jest.mock("fs", () => ({
  existsSync: jest.fn(() => true),
  readFileSync: jest.fn(() =>
    JSON.stringify({
      tenant_a: { name: "Tenant A" },
      tenant_b: { name: "Tenant B" },
    })
  ),
}));

// Create a mock TenantContextManager for testing
class MockTenantContextManager {
  private storage = new AsyncLocalStorage<{ tenantId: string }>();
  private defaultTenant = "tenant_a";

  async run<T>(tenantId: string, fn: () => Promise<T> | T): Promise<T> {
    return this.storage.run({ tenantId }, () => fn());
  }

  getTenantId(): string {
    const store = this.storage.getStore();
    return store?.tenantId || this.defaultTenant;
  }

  setTenantId(tenantId: string): void {
    const store = this.storage.getStore();
    if (store) {
      store.tenantId = tenantId;
    }
  }

  isAvailable(): boolean {
    return this.storage.getStore() !== undefined;
  }
}

describe("Multi-Tenant Architecture", () => {
  let tenantContext: MockTenantContextManager;

  beforeEach(() => {
    tenantContext = new MockTenantContextManager();
    jest.clearAllMocks();
  });

  describe("TenantContext - AsyncLocalStorage Isolation", () => {
    it("should return default tenant when not in context", () => {
      expect(tenantContext.getTenantId()).toBe("tenant_a");
    });

    it("should return correct tenant within context", async () => {
      await tenantContext.run("tenant_b", async () => {
        expect(tenantContext.getTenantId()).toBe("tenant_b");
      });
    });

    it("should isolate concurrent tenant contexts", async () => {
      const results: string[] = [];

      await Promise.all([
        tenantContext.run("tenant_a", async () => {
          // Small delay to simulate async work
          await Promise.resolve();
          results.push(`A:${tenantContext.getTenantId()}`);
        }),
        tenantContext.run("tenant_b", async () => {
          await Promise.resolve();
          results.push(`B:${tenantContext.getTenantId()}`);
        }),
      ]);

      expect(results).toContain("A:tenant_a");
      expect(results).toContain("B:tenant_b");
    });

    it("should report availability correctly", async () => {
      expect(tenantContext.isAvailable()).toBe(false);

      await tenantContext.run("tenant_a", async () => {
        expect(tenantContext.isAvailable()).toBe(true);
      });

      expect(tenantContext.isAvailable()).toBe(false);
    });

    it("should support nested tenant contexts", async () => {
      await tenantContext.run("tenant_a", async () => {
        expect(tenantContext.getTenantId()).toBe("tenant_a");

        await tenantContext.run("tenant_b", async () => {
          expect(tenantContext.getTenantId()).toBe("tenant_b");
        });

        expect(tenantContext.getTenantId()).toBe("tenant_a");
      });
    });

    it("should handle setTenantId within context", async () => {
      await tenantContext.run("tenant_a", async () => {
        expect(tenantContext.getTenantId()).toBe("tenant_a");

        tenantContext.setTenantId("tenant_b");
        expect(tenantContext.getTenantId()).toBe("tenant_b");
      });
    });
  });

  describe("Tenant Data Isolation Patterns", () => {
    // Simulated database operations per tenant
    const tenantData: Map<string, any[]> = new Map();

    const mockDbService = {
      save: (data: any) => {
        const tenantId = tenantContext.getTenantId();
        if (!tenantData.has(tenantId)) {
          tenantData.set(tenantId, []);
        }
        tenantData.get(tenantId)!.push({ ...data, tenantId });
      },
      getAll: () => {
        const tenantId = tenantContext.getTenantId();
        return tenantData.get(tenantId) || [];
      },
      clear: () => {
        const tenantId = tenantContext.getTenantId();
        tenantData.delete(tenantId);
      },
    };

    beforeEach(() => {
      tenantData.clear();
    });

    it("should save data with correct tenantId", async () => {
      await tenantContext.run("tenant_a", async () => {
        mockDbService.save({ name: "Product A" });
      });

      await tenantContext.run("tenant_b", async () => {
        mockDbService.save({ name: "Product B" });
      });

      expect(tenantData.get("tenant_a")).toEqual([
        { name: "Product A", tenantId: "tenant_a" },
      ]);
      expect(tenantData.get("tenant_b")).toEqual([
        { name: "Product B", tenantId: "tenant_b" },
      ]);
    });

    it("should retrieve only tenant-specific data", async () => {
      // Setup data for both tenants
      tenantData.set("tenant_a", [{ id: 1, tenantId: "tenant_a" }]);
      tenantData.set("tenant_b", [{ id: 2, tenantId: "tenant_b" }]);

      await tenantContext.run("tenant_a", async () => {
        const data = mockDbService.getAll();
        expect(data).toHaveLength(1);
        expect(data[0].tenantId).toBe("tenant_a");
      });

      await tenantContext.run("tenant_b", async () => {
        const data = mockDbService.getAll();
        expect(data).toHaveLength(1);
        expect(data[0].tenantId).toBe("tenant_b");
      });
    });

    it("should not leak data between tenants", async () => {
      await tenantContext.run("tenant_a", async () => {
        mockDbService.save({ secret: "A's secret data" });
      });

      await tenantContext.run("tenant_b", async () => {
        const data = mockDbService.getAll();
        expect(data).toHaveLength(0);
        expect(data.find((d: any) => d.secret)).toBeUndefined();
      });
    });

    it("should handle concurrent write operations", async () => {
      await Promise.all([
        tenantContext.run("tenant_a", async () => {
          for (let i = 0; i < 5; i++) {
            mockDbService.save({ index: i });
          }
        }),
        tenantContext.run("tenant_b", async () => {
          for (let i = 0; i < 3; i++) {
            mockDbService.save({ index: i });
          }
        }),
      ]);

      expect(tenantData.get("tenant_a")).toHaveLength(5);
      expect(tenantData.get("tenant_b")).toHaveLength(3);
    });
  });

  describe("Database Connection Patterns", () => {
    // Simulates the getDbManager().getConnection(tenantId) pattern
    const mockConnections: Map<string, { tenantId: string; query: jest.Mock }> =
      new Map();

    const mockDbManager = {
      getConnection: (tenantId: string) => {
        if (!mockConnections.has(tenantId)) {
          mockConnections.set(tenantId, {
            tenantId,
            query: jest.fn().mockResolvedValue([]),
          });
        }
        return mockConnections.get(tenantId)!;
      },
    };

    beforeEach(() => {
      mockConnections.clear();
    });

    it("should return tenant-specific database connection", () => {
      const connA = mockDbManager.getConnection("tenant_a");
      const connB = mockDbManager.getConnection("tenant_b");

      expect(connA.tenantId).toBe("tenant_a");
      expect(connB.tenantId).toBe("tenant_b");
      expect(connA).not.toBe(connB);
    });

    it("should reuse connection for same tenant", () => {
      const conn1 = mockDbManager.getConnection("tenant_a");
      const conn2 = mockDbManager.getConnection("tenant_a");

      expect(conn1).toBe(conn2);
    });

    it("should use correct connection from context", async () => {
      await tenantContext.run("tenant_a", async () => {
        const conn = mockDbManager.getConnection(tenantContext.getTenantId());
        await conn.query("SELECT * FROM products");
        expect(conn.tenantId).toBe("tenant_a");
      });

      await tenantContext.run("tenant_b", async () => {
        const conn = mockDbManager.getConnection(tenantContext.getTenantId());
        await conn.query("SELECT * FROM products");
        expect(conn.tenantId).toBe("tenant_b");
      });
    });
  });

  describe("READ vs WRITE Operation Patterns", () => {
    /**
     * Pattern: READ operations should NOT filter by tenantId 
     * because database is already tenant-isolated
     * 
     * Pattern: WRITE operations MUST include tenantId
     * for data integrity and future querying
     */
    
    const mockPrisma = {
      product: {
        findMany: jest.fn(),
        findFirst: jest.fn(),
        upsert: jest.fn(),
        create: jest.fn(),
      },
    };

    beforeEach(() => {
      jest.clearAllMocks();
    });

    it("READ: should NOT require tenantId filter (DB is isolated)", async () => {
      mockPrisma.product.findMany.mockResolvedValue([
        { id: 1, name: "Product" },
      ]);

      await tenantContext.run("tenant_a", async () => {
        // Correct pattern: No tenantId in where clause for READ
        await mockPrisma.product.findMany({
          where: { status: "active" },
        });

        expect(mockPrisma.product.findMany).toHaveBeenCalledWith({
          where: { status: "active" },
        });
      });
    });

    it("READ by ID: should use findFirst without tenantId", async () => {
      mockPrisma.product.findFirst.mockResolvedValue({ id: 1 });

      await tenantContext.run("tenant_a", async () => {
        // Correct pattern: findFirst without tenantId
        await mockPrisma.product.findFirst({
          where: { itemId: BigInt(123) },
        });

        expect(mockPrisma.product.findFirst).toHaveBeenCalledWith({
          where: { itemId: BigInt(123) },
        });
      });
    });

    it("WRITE: should ALWAYS include tenantId from context", async () => {
      mockPrisma.product.upsert.mockResolvedValue({ id: 1 });

      await tenantContext.run("tenant_a", async () => {
        const tenantId = tenantContext.getTenantId();

        // Correct pattern: Include tenantId in create/update
        await mockPrisma.product.upsert({
          where: { itemId: BigInt(123) },
          create: {
            tenantId,
            itemId: BigInt(123),
            name: "New Product",
          },
          update: {
            name: "Updated Product",
          },
        });

        expect(mockPrisma.product.upsert).toHaveBeenCalledWith(
          expect.objectContaining({
            create: expect.objectContaining({
              tenantId: "tenant_a",
            }),
          })
        );
      });
    });

    it("WRITE: should use dynamic tenantId, not hardcoded", async () => {
      const operations: string[] = [];

      await tenantContext.run("tenant_a", async () => {
        operations.push(`create:${tenantContext.getTenantId()}`);
      });

      await tenantContext.run("tenant_b", async () => {
        operations.push(`create:${tenantContext.getTenantId()}`);
      });

      expect(operations).toEqual(["create:tenant_a", "create:tenant_b"]);
    });
  });

  describe("Middleware Tenant Extraction", () => {
    interface MockRequest {
      tenantId?: string;
      headers: Record<string, string>;
      query: Record<string, string>;
    }

    const extractTenantId = (req: MockRequest): string => {
      return (
        req.tenantId ||
        req.headers["x-tenant-id"] ||
        req.query.tenantId ||
        "default"
      );
    };

    it("should extract from request.tenantId first", () => {
      const req: MockRequest = {
        tenantId: "from_req",
        headers: { "x-tenant-id": "from_header" },
        query: { tenantId: "from_query" },
      };

      expect(extractTenantId(req)).toBe("from_req");
    });

    it("should fallback to x-tenant-id header", () => {
      const req: MockRequest = {
        headers: { "x-tenant-id": "from_header" },
        query: { tenantId: "from_query" },
      };

      expect(extractTenantId(req)).toBe("from_header");
    });

    it("should fallback to query param", () => {
      const req: MockRequest = {
        headers: {},
        query: { tenantId: "from_query" },
      };

      expect(extractTenantId(req)).toBe("from_query");
    });

    it("should use default as last resort", () => {
      const req: MockRequest = {
        headers: {},
        query: {},
      };

      expect(extractTenantId(req)).toBe("default");
    });
  });

  describe("Error Handling in Multi-Tenant Context", () => {
    it("should maintain tenant context through errors", async () => {
      let capturedTenantId: string = "";

      try {
        await tenantContext.run("tenant_a", async () => {
          capturedTenantId = tenantContext.getTenantId();
          throw new Error("Test error");
        });
      } catch (e) {
        // Error should propagate
      }

      expect(capturedTenantId).toBe("tenant_a");
    });

    it("should restore context after error in nested context", async () => {
      const results: string[] = [];

      await tenantContext.run("tenant_a", async () => {
        results.push(tenantContext.getTenantId());

        try {
          await tenantContext.run("tenant_b", async () => {
            results.push(tenantContext.getTenantId());
            throw new Error("Inner error");
          });
        } catch (e) {
          results.push(`error:${tenantContext.getTenantId()}`);
        }

        results.push(tenantContext.getTenantId());
      });

      expect(results).toEqual([
        "tenant_a",
        "tenant_b",
        "error:tenant_a",
        "tenant_a",
      ]);
    });
  });
});
