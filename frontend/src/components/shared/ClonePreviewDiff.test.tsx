import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import "@testing-library/jest-dom";
import { ClonePreviewDiff } from "./ClonePreviewDiff";
import type { CloneDifference } from "./CloneDiffSummary";

vi.mock("@/contexts/ThemeContext.hooks", () => ({
  useTheme: () => ({ isDark: false }),
}));

// Mock DiffSummarySection to isolate ClonePreviewDiff rendering
vi.mock("./CloneDiffSummary", () => ({
  DiffSummarySection: ({ differences }: { differences: CloneDifference[] }) => (
    <div data-testid="diff-summary-section">
      {differences.length > 0
        ? `${differences.length} differences`
        : "no differences"}
    </div>
  ),
}));

// Ant Design components use matchMedia internally — mock it for jsdom
Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

const sampleSource = {
  title: "Awesome T-Shirt",
  description: "Great quality tee",
  price: 150_000,
  stock: 50,
  images: ["https://example.com/img1.jpg"],
  sku: "TS-001",
};

const sampleShopeeTarget = {
  platform: "shopee" as const,
  title: "Awesome T-Shirt Shopee",
  description: "Great quality tee",
  price: 135_000,
  stock: 50,
  images: ["https://example.com/img1.jpg"],
};

const titleDiff: CloneDifference = {
  field: "title",
  source_value: "Awesome T-Shirt",
  target_value: "Awesome T-Shirt Shopee",
  reason: "Title truncated to 120 chars for Shopee",
};

const priceDiff: CloneDifference = {
  field: "price",
  source_value: "Rp 150.000",
  target_value: "Rp 135.000",
  reason: "Price adjusted by platform fee",
};

describe("ClonePreviewDiff", () => {
  it("renders source panel with Source tag and source SKU", () => {
    render(
      <ClonePreviewDiff
        source={sampleSource}
        target={sampleShopeeTarget}
        differences={[]}
        warnings={[]}
      />,
    );
    expect(screen.getByText("Source")).toBeInTheDocument();
    expect(screen.getByText("TS-001")).toBeInTheDocument();
  });

  it("renders target panel with platform label and Clone Target heading", () => {
    render(
      <ClonePreviewDiff
        source={sampleSource}
        target={sampleShopeeTarget}
        differences={[]}
        warnings={[]}
      />,
    );
    expect(screen.getByText("Shopee")).toBeInTheDocument();
    expect(screen.getByText("Clone Target")).toBeInTheDocument();
  });

  it("renders DiffSummarySection with correct differences count", () => {
    render(
      <ClonePreviewDiff
        source={sampleSource}
        target={sampleShopeeTarget}
        differences={[titleDiff, priceDiff]}
        warnings={[]}
      />,
    );
    expect(screen.getByTestId("diff-summary-section")).toBeInTheDocument();
    expect(screen.getByText("2 differences")).toBeInTheDocument();
  });

  it("renders DiffSummarySection with 'no differences' when differences is empty", () => {
    render(
      <ClonePreviewDiff
        source={sampleSource}
        target={sampleShopeeTarget}
        differences={[]}
        warnings={[]}
      />,
    );
    expect(screen.getByTestId("diff-summary-section")).toBeInTheDocument();
    expect(screen.getByText("no differences")).toBeInTheDocument();
  });

  it("shows warning alerts when warnings prop is non-empty", () => {
    render(
      <ClonePreviewDiff
        source={sampleSource}
        target={sampleShopeeTarget}
        differences={[]}
        warnings={["Stock may be insufficient", "Images exceed platform limit"]}
      />,
    );
    expect(screen.getByText("Stock may be insufficient")).toBeInTheDocument();
    expect(
      screen.getByText("Images exceed platform limit"),
    ).toBeInTheDocument();
  });

  it("shows no warning alerts when warnings prop is empty", () => {
    render(
      <ClonePreviewDiff
        source={sampleSource}
        target={sampleShopeeTarget}
        differences={[]}
        warnings={[]}
      />,
    );
    // No alerts rendered when warnings is empty
    expect(screen.queryByRole("alert")).toBeNull();
  });

  it("displays diff reason text in target panel for fields with differences", () => {
    render(
      <ClonePreviewDiff
        source={sampleSource}
        target={sampleShopeeTarget}
        differences={[titleDiff]}
        warnings={[]}
      />,
    );
    // Reason text appears below the differing field in the target panel
    expect(
      screen.getByText("Title truncated to 120 chars for Shopee"),
    ).toBeInTheDocument();
  });
});
