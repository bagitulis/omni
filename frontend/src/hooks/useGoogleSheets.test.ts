import { beforeEach, describe, expect, it, vi } from "vitest";

const useMutationMock = vi.fn((options: unknown) => options);
const useQueryMock = vi.fn();
const invalidateQueriesMock = vi.fn();

vi.mock("@tanstack/react-query", () => ({
  useQuery: (...args: unknown[]) => useQueryMock(...args),
  useMutation: (options: unknown) => useMutationMock(options),
  useQueryClient: () => ({
    invalidateQueries: invalidateQueriesMock,
  }),
}));

vi.mock("antd", () => ({
  message: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

vi.mock("@/api/googleSheets", () => ({
  getSavedLinks: vi.fn(),
  getDetailedSettings: vi.fn(),
  validateLink: vi.fn(),
  saveLinks: vi.fn(),
  updateSettings: vi.fn(),
}));

import {
  useGoogleSheetsLinks,
  useGoogleSheetsDetails,
  useValidateLink,
  useSaveLinks,
  useUpdateSettings,
} from "./useGoogleSheets";
import { message } from "antd";
import {
  getSavedLinks,
  getDetailedSettings,
  validateLink,
  saveLinks,
  updateSettings,
} from "@/api/googleSheets";

describe("useGoogleSheetsLinks", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useQuery with google-sheets-links queryKey", () => {
    useGoogleSheetsLinks();
    const args = useQueryMock.mock.calls[0]?.[0] as { queryKey: unknown[] };
    expect(args.queryKey).toEqual(["google-sheets-links"]);
  });

  it("uses getSavedLinks as queryFn", () => {
    useGoogleSheetsLinks();
    const args = useQueryMock.mock.calls[0]?.[0] as { queryFn: () => unknown };
    args.queryFn();
    expect(getSavedLinks).toHaveBeenCalled();
  });
});

describe("useGoogleSheetsDetails", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls useQuery with google-sheets-details queryKey", () => {
    useGoogleSheetsDetails();
    const args = useQueryMock.mock.calls[0]?.[0] as { queryKey: unknown[] };
    expect(args.queryKey).toEqual(["google-sheets-details"]);
  });

  it("uses getDetailedSettings as queryFn", () => {
    useGoogleSheetsDetails();
    const args = useQueryMock.mock.calls[0]?.[0] as { queryFn: () => unknown };
    args.queryFn();
    expect(getDetailedSettings).toHaveBeenCalled();
  });
});

describe("useValidateLink", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls validateLink in mutationFn", () => {
    useValidateLink();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (arg: unknown) => void;
    };
    opts.mutationFn("https://docs.google.com/spreadsheets/test");
    expect(validateLink).toHaveBeenCalledWith(
      "https://docs.google.com/spreadsheets/test",
    );
  });

  it("shows success message on onSuccess", () => {
    useValidateLink();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess: () => void;
    };
    opts.onSuccess();
    expect(message.success).toHaveBeenCalledWith("Link validated successfully");
  });

  it("shows error message on onError", () => {
    useValidateLink();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onError: (e: Error) => void;
    };
    opts.onError(new Error("invalid link"));
    expect(message.error).toHaveBeenCalledWith("invalid link");
  });
});

describe("useSaveLinks", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls saveLinks in mutationFn", () => {
    useSaveLinks();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (arg: unknown) => void;
    };
    opts.mutationFn({ links: [] });
    expect(saveLinks).toHaveBeenCalledWith({ links: [] });
  });

  it("shows success message and invalidates on onSuccess", () => {
    useSaveLinks();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess: () => void;
    };
    opts.onSuccess();
    expect(message.success).toHaveBeenCalledWith("Links saved successfully");
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["google-sheets-links"],
    });
  });
});

describe("useUpdateSettings", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls updateSettings in mutationFn", () => {
    useUpdateSettings();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      mutationFn: (arg: unknown) => void;
    };
    opts.mutationFn({ sync_enabled: true });
    expect(updateSettings).toHaveBeenCalledWith({ sync_enabled: true });
  });

  it("shows success message and invalidates both queries on onSuccess", () => {
    useUpdateSettings();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onSuccess: () => void;
    };
    opts.onSuccess();
    expect(message.success).toHaveBeenCalledWith(
      "Settings updated successfully",
    );
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["google-sheets-links"],
    });
    expect(invalidateQueriesMock).toHaveBeenCalledWith({
      queryKey: ["google-sheets-details"],
    });
  });

  it("shows error message on onError", () => {
    useUpdateSettings();
    const opts = useMutationMock.mock.calls[0]?.[0] as {
      onError: (e: Error) => void;
    };
    opts.onError(new Error("settings error"));
    expect(message.error).toHaveBeenCalledWith("settings error");
  });
});
