import { beforeEach, describe, expect, it, vi } from "vitest";

const useQueryMock = vi.fn();
const useMutationMock = vi.fn((options: unknown) => options);
const invalidateQueriesMock = vi.fn();

vi.mock("@tanstack/react-query", () => ({
  useQuery: (...args: unknown[]) => useQueryMock(...args),
  useMutation: (options: unknown) => useMutationMock(options),
  useQueryClient: () => ({
    invalidateQueries: invalidateQueriesMock,
    setQueryData: vi.fn(),
  }),
  keepPreviousData: undefined,
}));

vi.mock("@/api/products", () => ({
  getProducts: vi.fn().mockResolvedValue({ data: [], meta: { total: 0 } }),
  getProductById: vi.fn(),
  createProduct: vi.fn(),
  updateProduct: vi.fn(),
  deleteProduct: vi.fn(),
  syncProduct: vi.fn(),
  batchUpdateSkus: vi.fn(),
}));

import {
  useProducts,
  useProduct,
  useCreateProduct,
  useUpdateProduct,
  useDeleteProduct,
  useSyncProduct,
} from "./useProducts";
import * as productsApi from "@/api/products";

describe("useProducts", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useQuery with correct queryKey", () => {
    const params = { page: 1, limit: 20 };
    useProducts(params);
    expect(useQueryMock).toHaveBeenCalledOnce();
    const options = useQueryMock.mock.calls[0][0] as { queryKey: unknown[] };
    expect(options.queryKey).toEqual(["products", params]);
  });

  it("calls useQuery with a queryFn that invokes getProducts with params", async () => {
    const params = { page: 1, limit: 20 };
    useProducts(params);
    const options = useQueryMock.mock.calls[0][0] as {
      queryFn: () => Promise<unknown>;
    };
    await options.queryFn();
    expect(productsApi.getProducts).toHaveBeenCalledWith(params);
  });
});

describe("useProduct", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useQuery with correct queryKey when id is provided", () => {
    useProduct("123");
    const options = useQueryMock.mock.calls[0][0] as {
      queryKey: unknown[];
      enabled: boolean;
    };
    expect(options.queryKey).toEqual(["product", "123"]);
    expect(options.enabled).toBe(true);
  });

  it("sets enabled to false when id is undefined", () => {
    useProduct(undefined);
    const options = useQueryMock.mock.calls[0][0] as { enabled: boolean };
    expect(options.enabled).toBe(false);
  });
});

describe("useCreateProduct", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useMutation with correct mutationFn", () => {
    useCreateProduct();
    const options = useMutationMock.mock.calls[0][0] as {
      mutationFn: (data: unknown) => unknown;
    };
    const data = { title: "Test Product" };
    options.mutationFn(data);
    expect(productsApi.createProduct).toHaveBeenCalledWith(data);
  });

  it("invalidates products queries on success", () => {
    useCreateProduct();
    const options = useMutationMock.mock.calls[0][0] as {
      onSuccess: () => void;
    };
    options.onSuccess();
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["products"],
    });
  });
});

describe("useUpdateProduct", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useMutation with correct mutationFn", () => {
    useUpdateProduct();
    const options = useMutationMock.mock.calls[0][0] as {
      mutationFn: (args: { id: string; data: unknown }) => unknown;
    };
    options.mutationFn({ id: "42", data: { title: "Updated" } });
    expect(productsApi.updateProduct).toHaveBeenCalledWith("42", {
      title: "Updated",
    });
  });

  it("invalidates both products and specific product query on success", () => {
    useUpdateProduct();
    const options = useMutationMock.mock.calls[0][0] as {
      onSuccess: (data: { id: number }) => void;
    };
    options.onSuccess({ id: 42 });
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["products"],
    });
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["product", "42"],
    });
  });
});

describe("useDeleteProduct", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useMutation with deleteProduct as mutationFn", () => {
    useDeleteProduct();
    const options = useMutationMock.mock.calls[0][0] as {
      mutationFn: (id: string) => unknown;
    };
    options.mutationFn("99");
    expect(productsApi.deleteProduct).toHaveBeenCalledWith("99");
  });
});

describe("useSyncProduct", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useMutation with correct mutationFn", () => {
    useSyncProduct();
    const options = useMutationMock.mock.calls[0][0] as {
      mutationFn: (args: { id: string; platform?: string }) => unknown;
    };
    options.mutationFn({ id: "10", platform: "shopee" });
    expect(productsApi.syncProduct).toHaveBeenCalledWith("10", "shopee");
  });
});
