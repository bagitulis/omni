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
            color="default"
            style={{ marginLeft: 4 }}
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
            color="red"
            style={{ marginLeft: 4 }}
          />
        </span>
      ),
    },
  ];
}
