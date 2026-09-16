import { beforeEach, describe, expect, it, vi } from "vitest";

const useQueryMock = vi.fn();
const useMutationMock = vi.fn((options: unknown) => options);
const invalidateQueriesMock = vi.fn();

vi.mock("@tanstack/react-query", () => ({
  useQuery: (options: unknown) => useQueryMock(options),
  useMutation: (options: unknown) => useMutationMock(options),
  useQueryClient: () => ({ invalidateQueries: invalidateQueriesMock }),
}));

vi.mock("@/api/scrapeJob", () => ({
  getScrapeJob: vi.fn(),
  cancelScrapeJob: vi.fn(),
  resumeScrape: vi.fn(),
  shouldStopPolling: (status: string | undefined) =>
    status === "completed" ||
    status === "failed" ||
    status === "cancelled" ||
    status === "blocked",
  extractBlocker: vi.fn(() => null),
}));

import { useScrapeJob } from "./useScrapeJob";
import { getScrapeJob } from "@/api/scrapeJob";

interface QueryOptions {
  queryKey: unknown[];
  queryFn: (context: { signal: AbortSignal }) => unknown;
  enabled?: boolean;
  refetchInterval?: (query: { state: { data?: { status?: string } } }) => unknown;
}

function capturedOptions(): QueryOptions {
  return useQueryMock.mock.calls[useQueryMock.mock.calls.length - 1][0] as QueryOptions;
}

describe("useScrapeJob", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryMock.mockReturnValue({ data: undefined, isError: false });
  });

  it("is disabled until there is a job id", () => {
    useScrapeJob(undefined);
    expect(capturedOptions().enabled).toBe(false);

    useScrapeJob("   ");
    expect(capturedOptions().enabled).toBe(false);

    useScrapeJob("job-1");
    expect(capturedOptions().enabled).toBe(true);
  });

  it("forwards the query signal so an abandoned poll actually aborts", () => {
    useScrapeJob("job-1");
    const controller = new AbortController();
    capturedOptions().queryFn({ signal: controller.signal });
    expect(getScrapeJob).toHaveBeenCalledWith("job-1", controller.signal);
  });

  it("keeps polling while the job is in flight", () => {
    useScrapeJob("job-1");
    const refetchInterval = capturedOptions().refetchInterval;
    expect(refetchInterval?.({ state: { data: { status: "running" } } })).toBe(3000);
    expect(refetchInterval?.({ state: { data: { status: "pending" } } })).toBe(3000);
  });

  it("stops polling once the job reaches a resting state", () => {
    useScrapeJob("job-1");
    const refetchInterval = capturedOptions().refetchInterval;
    expect(refetchInterval?.({ state: { data: { status: "completed" } } })).toBe(false);
    expect(refetchInterval?.({ state: { data: { status: "failed" } } })).toBe(false);
    expect(refetchInterval?.({ state: { data: { status: "cancelled" } } })).toBe(false);
    expect(refetchInterval?.({ state: { data: { status: "blocked" } } })).toBe(false);
  });

  it("keeps polling before the first payload arrives", () => {
    useScrapeJob("job-1");
    const refetchInterval = capturedOptions().refetchInterval;
    expect(refetchInterval?.({ state: { data: undefined } })).toBe(3000);
  });
});
