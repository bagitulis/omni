import { Segmented } from "antd";
import { useInventoryFilterStore } from "@/stores/inventoryFilterStore";

export function InventoryQuickView() {
  const { stockStatus, setStockStatusFilter } = useInventoryFilterStore();

  return (
    <Segmented
      value={stockStatus || "all"}
      onChange={(value) => {
        setStockStatusFilter(
          value === "all"
            ? ""
            : (value as "in_stock" | "low_stock" | "out_of_stock"),
        );
      }}
      options={[
        { label: "All", value: "all" },
        { label: "Low Stock", value: "low_stock" },
        { label: "Out of Stock", value: "out_of_stock" },
      ]}
      style={{ fontSize: 12 }}
    />
  );
}
