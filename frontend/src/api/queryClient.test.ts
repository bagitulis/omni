import { describe, it, expect, vi, beforeEach } from "vitest";

const { mockMessageError, mockQueryClientConstructor } = vi.hoisted(() => ({
  mockMessageError: vi.fn(),
  mockQueryClientConstructor: vi.fn(),
}));

vi.mock("antd", () => ({
  message: { error: mockMessageError },
}));

// Capture the options passed to QueryClient constructor
let capturedOptions: Record<string, unknown> = {};
vi.mock("@tanstack/react-query", () => {
  class MockQueryClient {
    options: Record<string, unknown>;
    constructor(opts: Record<string, unknown> = {}) {
      capturedOptions = opts;
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
    const queries = capturedOptions.defaultOptions as {
      queries?: { staleTime?: number };
    };
    expect(queries.queries?.staleTime).toBe(5 * 60 * 1000);
  });

  it("has retry: 1", () => {
    const queries = capturedOptions.defaultOptions as {
      queries?: { retry?: number };
    };
    expect(queries.queries?.retry).toBe(1);
  });

  it("has refetchOnWindowFocus: false", () => {
    const queries = capturedOptions.defaultOptions as {
      queries?: { refetchOnWindowFocus?: boolean };
    };
    expect(queries.queries?.refetchOnWindowFocus).toBe(false);
  });
});

describe("queryClient defaultOptions.mutations.onError", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls message.error with the error message", () => {
    const mutations = capturedOptions.defaultOptions as {
      mutations?: { onError?: (error: Error) => void };
    };
    mutations.mutations?.onError?.(new Error("Something went wrong"));
    expect(mockMessageError).toHaveBeenCalledWith("Something went wrong");
  });

  it("falls back to 'Operation failed' when error has no message", () => {
    const mutations = capturedOptions.defaultOptions as {
      mutations?: { onError?: (error: Error) => void };
    };
    mutations.mutations?.onError?.(new Error(""));
    expect(mockMessageError).toHaveBeenCalledWith("Operation failed");
  });
});
