import { describe, it, expect, vi } from "vitest";

const { mockMessageError, mockQueryClientConstructor, capturedHolder } =
  vi.hoisted(() => {
    // Use an object holder so assignment inside vi.mock closure works by reference
    const capturedHolder: { options: Record<string, unknown> } = {
      options: {},
    };
    return {
      mockMessageError: vi.fn(),
      mockQueryClientConstructor: vi.fn(),
      capturedHolder,
    };
  });

vi.mock("antd", () => ({
  message: { error: mockMessageError },
}));

vi.mock("@tanstack/react-query", () => {
  class MockQueryClient {
    options: Record<string, unknown>;
    constructor(opts: Record<string, unknown> = {}) {
      capturedHolder.options = opts;
      this.options = opts;
      mockQueryClientConstructor(opts);
    }
  }
  return { QueryClient: MockQueryClient };
});

import { queryClient } from "./queryClient";

describe("queryClient", () => {
  it("is defined and exported", () => {
    expect(queryClient).toBeDefined();
  });

  it("QueryClient constructor is called once", () => {
    expect(mockQueryClientConstructor).toHaveBeenCalledOnce();
  });

  it("is called with defaultOptions", () => {
    expect(mockQueryClientConstructor).toHaveBeenCalledWith(
      expect.objectContaining({ defaultOptions: expect.any(Object) }),
    );
  });
});

describe("queryClient defaultOptions.queries", () => {
  it("has staleTime of 5 minutes (300000 ms)", () => {
    const queries = capturedHolder.options.defaultOptions as {
      queries?: { staleTime?: number };
    };
    expect(queries.queries?.staleTime).toBe(5 * 60 * 1000);
  });

  it("has retry: 1", () => {
    const queries = capturedHolder.options.defaultOptions as {
      queries?: { retry?: number };
    };
    expect(queries.queries?.retry).toBe(1);
  });

  it("has refetchOnWindowFocus: false", () => {
    const queries = capturedHolder.options.defaultOptions as {
      queries?: { refetchOnWindowFocus?: boolean };
    };
    expect(queries.queries?.refetchOnWindowFocus).toBe(false);
  });
});

describe("queryClient defaultOptions.mutations", () => {
  it("has NO global onError (prevents double notifications)", () => {
    const opts = capturedHolder.options.defaultOptions as {
      mutations?: { onError?: unknown };
    };
    // Each mutation hook defines its own onError with contextual messages.
    // A global handler here would cause every error to appear twice.
    expect(opts.mutations?.onError).toBeUndefined();
  });
});
