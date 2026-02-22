import { Badge } from "antd";

/** Builds tab items for the products page mapping tabs. */
export function buildProductTabItems(total: number, activeMapping: string) {
  return [
    {
      key: "all",
      label: (
        <span>
          All Products{" "}
          <Badge
            count={total}
            showZero
            style={{ backgroundColor: "#94a3b8", marginLeft: 4 }}
          />
        </span>
      ),
    },
    {
      key: "mapped",
      label: "Mapped",
    },
    {
      key: "unmapped",
      label: (
        <span>
          Unmapped{" "}
          <Badge
            count={activeMapping === "unmapped" ? total : "?"}
            style={{ backgroundColor: "#ef4444", marginLeft: 4 }}
          />
        </span>
      ),
    },
  ];
}
