export function useInventoryUtils(state: any) {
  function inputType(columnType: string) {
    if (columnType.includes("INTEGER")) return "number";
    if (columnType.includes("REAL")) return "number";
    if (columnType.includes("DATE")) return "date";
    return "text";
  }

  function formatDate(dateStr: string) {
    try {
      return new Date(dateStr).toLocaleString("id-ID");
    } catch {
      return dateStr;
    }
  }

  function getItemValue(item: any, columnName: string) {
    if (!item) return "";

    // Data dari API - supports both string (Node.js) and object (Go) formats
    let itemData = item;

    // Handle nested data field
    if (item.data !== undefined && item.data !== null) {
      if (typeof item.data === "string") {
        try {
          itemData = JSON.parse(item.data);
        } catch {
          itemData = item;
        }
      } else if (typeof item.data === "object") {
        itemData = item.data;
      }
    }

    const value = itemData[columnName];
    return value !== null && value !== undefined ? value : "";
  }

  function getColumnWidthClass(columnName: string) {
    const longColumns = [
      "variant_name",
      "item_name",
      "sku_name",
      "Nama Variasi",
      "Nama Barang",
    ];
    return longColumns.includes(columnName) ? "wide-column" : "normal-column";
  }

  function nextPage() {
    state.currentOffset += state.pageSize;
    scrollToTop();
  }

  function previousPage() {
    state.currentOffset = Math.max(0, state.currentOffset - state.pageSize);
    scrollToTop();
  }

  function scrollToTop() {
    const element = document.querySelector(".inventory-content");
    if (element) {
      element.scrollIntoView({ behavior: "smooth" });
    }
  }

  function toggleSyncHistory() {
    state.showSyncHistory = !state.showSyncHistory;
  }

  return {
    inputType,
    formatDate,
    getItemValue,
    getColumnWidthClass,
    nextPage,
    previousPage,
    scrollToTop,
    toggleSyncHistory,
  };
}
