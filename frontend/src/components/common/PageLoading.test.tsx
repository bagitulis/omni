import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { PageLoading } from "@/components/common/PageLoading";

// Mock antd
vi.mock("antd", async () => {
  const actual = await vi.importActual<typeof import("antd")>("antd");
  return {
    ...actual,
    Spin: () => <div data-testid="spin">Loading...</div>,
  };
});

describe("PageLoading", () => {
  it("renders spin component", () => {
    render(<PageLoading />);
    expect(screen.getByTestId("spin")).toBeInTheDocument();
  });
});
