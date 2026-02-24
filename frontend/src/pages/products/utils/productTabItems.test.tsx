import "@testing-library/jest-dom";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { buildProductTabItems } from "./productTabItems";

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

describe("buildProductTabItems", () => {
  it("returns 3 items", () => {
    const items = buildProductTabItems(10, "all");
    expect(items).toHaveLength(3);
  });

  it("has keys: all, mapped, unmapped", () => {
    const items = buildProductTabItems(10, "all");
    const keys = items.map((i) => i.key);
    expect(keys).toEqual(["all", "mapped", "unmapped"]);
  });

  it("renders All Products label with total count badge", () => {
    const items = buildProductTabItems(42, "all");
    const allItem = items.find((i) => i.key === "all");
    expect(allItem).toBeDefined();
    render(<>{allItem!.label}</>);
    expect(screen.getByText(/All Products/)).toBeInTheDocument();
    expect(screen.getByTitle("42")).toBeInTheDocument();
  });

  it("renders Mapped label as plain string", () => {
    const items = buildProductTabItems(5, "mapped");
    const mappedItem = items.find((i) => i.key === "mapped");
    expect(mappedItem?.label).toBe("Mapped");
  });

  it("renders Unmapped label with total when activeMapping is 'unmapped'", () => {
    const items = buildProductTabItems(7, "unmapped");
    const unmappedItem = items.find((i) => i.key === "unmapped");
    expect(unmappedItem).toBeDefined();
    render(<>{unmappedItem!.label}</>);
    expect(screen.getByText(/Unmapped/)).toBeInTheDocument();
    expect(screen.getByText("7")).toBeInTheDocument();
  });

  it("renders '?' badge on Unmapped when activeMapping is not 'unmapped'", () => {
    const items = buildProductTabItems(100, "all");
    const unmappedItem = items.find((i) => i.key === "unmapped");
    expect(unmappedItem).toBeDefined();
    render(<>{unmappedItem!.label}</>);
    // When activeMapping is not "unmapped", Badge shows "?"
    expect(screen.getByText("?")).toBeInTheDocument();
  });

  it("works with zero total", () => {
    const items = buildProductTabItems(0, "all");
    expect(items).toHaveLength(3);
    render(<>{items[0].label}</>);
    expect(screen.getByText(/All Products/)).toBeInTheDocument();
  });
});
