import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { OrderStatusTabs } from "./OrderStatusTabs";

const expectedLabels = [
  "To Ship",
  "Processed",
  "Shipped",
  "Completed",
  "Cancelled",
  "Locked Today",
  "Today's Orders",
];

describe("OrderStatusTabs", () => {
  it("renders one consistent tab set for all platform filters", () => {
    const { rerender } = render(
      <OrderStatusTabs
        activeTab="unprocess"
        onChange={() => {}}
        totalCount={12}
        platform="all"
      />,
    );

    for (const label of expectedLabels) {
      expect(screen.queryByText(label)).not.toBeNull();
    }
    expect(screen.queryByText("Unpaid")).toBeNull();

    rerender(
      <OrderStatusTabs
        activeTab="unprocess"
        onChange={() => {}}
        totalCount={12}
        platform="shopee"
      />,
    );

    for (const label of expectedLabels) {
      expect(screen.queryByText(label)).not.toBeNull();
    }
    expect(screen.queryByText("Unpaid")).toBeNull();
  });
});
