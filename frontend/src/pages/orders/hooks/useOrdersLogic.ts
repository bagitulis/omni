import { useEffect, useState } from "react";
import { useSearchParams, useParams, useNavigate } from "react-router-dom";
import { message } from "@/components/AntStaticApi";
import { useQueryClient } from "@tanstack/react-query";
import { useOrders } from "@/hooks/useOrders";
import { useOrderSync } from "./useOrderSync";
import { normalizeOrderTabKey } from "@/api/orderTabMapping";
import { useOrderBulkActions } from "./useOrderBulkActions";
import { useOrderSingleActions } from "./useOrderSingleActions";
import type { OrderDetail, Order } from "@/types/order";
import type { Dayjs } from "dayjs";
import { generateOrdersCSV, downloadCSV } from "../utils/csv";

const ORDER_MANAGER_VISIBLE_TABS = new Set([
  "unprocess",
  "processed",
  "shipped",
  "completed",
  "cancelled",
  "locked",
  "today",
]);

function getVisibleTab(tabKey: string): string {
  if (ORDER_MANAGER_VISIBLE_TABS.has(tabKey)) {
    return tabKey;
  }

  return "unprocess";
}

export function useOrdersLogic() {
  // State
  const [searchParams, setSearchParams] = useSearchParams();
  const { platform: routePlatform } = useParams();
  const navigate = useNavigate();
  const rawType = searchParams.get("type");
  const normalizedType = normalizeOrderTabKey(rawType || "unprocess");
  const activeTab = getVisibleTab(normalizedType);
  const platform = routePlatform || "all";

  useEffect(() => {
    if (rawType === activeTab) {
      return;
    }

    const nextParams = new URLSearchParams(searchParams);
    nextParams.set("type", activeTab);
    setSearchParams(nextParams, { replace: true });
  }, [activeTab, rawType, searchParams, setSearchParams]);

  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [search, setSearch] = useState("");
  // platform state is now derived from URL
  const [dateRange, setDateRange] = useState<[string, string] | null>(null);
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [selectedOrder, setSelectedOrder] = useState<
    OrderDetail | Order | null
  >(null);
  const [isDetailModalOpen, setIsDetailModalOpen] = useState(false);
  const [isShipModalOpen, setIsShipModalOpen] = useState(false);
  const [isCancelModalOpen, setIsCancelModalOpen] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(false);

  // Query client for cache invalidation
  const queryClient = useQueryClient();

  // Hooks
  const { data, isLoading, refetch } = useOrders(
    {
      page,
      pageSize,
      status: activeTab,
      platform,
      search,
      startDate: dateRange?.[0],
      endDate: dateRange?.[1],
    },
    { autoRefresh },
  );

  const { isSyncing, syncActiveTab } = useOrderSync(
    activeTab,
    platform,
    refetch,
    autoRefresh,
  );

  // Sub-hooks
  const bulkActions = useOrderBulkActions({
    selectedRowKeys,
    setSelectedRowKeys,
    data,
    refetch,
    platform,
  });

  const singleActions = useOrderSingleActions({
    selectedOrder,
    setSelectedOrder,
    setIsShipModalOpen,
    setIsCancelModalOpen,
    setIsDetailModalOpen,
  });

  // Handlers
  const handleTabChange = (key: string) => {
    const normalizedKey = normalizeOrderTabKey(key);
    queryClient.cancelQueries({ queryKey: ["orders"] });
    setSearchParams({ type: normalizedKey });
    setPage(1);
    setSelectedRowKeys([]);
  };

  const handleRefresh = async () => {
    await syncActiveTab(activeTab);
  };

  const handleSearch = (value: string) => {
    setSearch(value);
    setPage(1);
  };

  const handlePlatformChange = (value: string) => {
    const normalizedTab = normalizeOrderTabKey(activeTab);
    if (value === "all") {
      navigate(`/order-manager?type=${normalizedTab}`);
    } else {
      navigate(`/order-manager/${value}?type=${normalizedTab}`);
    }
    setPage(1);
  };

  const handleDateChange = (dates: [Dayjs | null, Dayjs | null] | null) => {
    if (dates?.[0] && dates[1]) {
      setDateRange([
        dates[0].format("YYYY-MM-DD"),
        dates[1].format("YYYY-MM-DD"),
      ]);
    } else {
      setDateRange(null);
    }
    setPage(1);
  };

  const handleSelectionChange = (keys: React.Key[]) => {
    setSelectedRowKeys(keys);
  };

  const handleExport = () => {
    if (!data?.orders || data.orders.length === 0) {
      message.warning("No orders to export");
      return;
    }

    try {
      const csv = generateOrdersCSV(data.orders);
      if (!csv) {
        message.error("Failed to generate CSV");
        return;
      }

      const dateStr = new Date().toISOString().split("T")[0];
      const filename = `orders-${activeTab}-${dateStr}.csv`;

      downloadCSV(csv, filename);
      message.success("Orders exported successfully");
    } catch (err) { console.warn("Operation failed:", err);
      message.error("Failed to export orders");
    }
  };

  return {
    state: {
      activeTab,
      page,
      pageSize,
      search,
      platform,
      dateRange,
      selectedRowKeys,
      selectedOrder,
      isDetailModalOpen,
      isShipModalOpen,
      isCancelModalOpen,
      autoRefresh,
      data,
      isLoading,
      isSyncing,
      isShipping: bulkActions.isShipping,
      isPrinting: bulkActions.isPrinting,
      isCancelling: bulkActions.isCancelling,
      isSingleShipping: singleActions.isSingleShipping,
      shipProgress: bulkActions.shipProgress,
      printProgress: bulkActions.printProgress,
      cancelProgress: bulkActions.cancelProgress,
      shipResult: bulkActions.shipResult,
      printResult: bulkActions.printResult,
      cancelResult: bulkActions.cancelResult,
    },
    setters: {
      setPage,
      setPageSize,
      setSelectedRowKeys,
      setSelectedOrder,
      setIsDetailModalOpen,
      setIsShipModalOpen,
      setIsCancelModalOpen,
      setAutoRefresh,
    },
    handlers: {
      handleTabChange,
      handleRefresh,
      handleSearch,
      handlePlatformChange,
      handleDateChange,
      handleSelectionChange,
      handleBulkShip: bulkActions.handleBulkShip,
      handleBulkPrint: bulkActions.handleBulkPrint,
      handleRetryFailedPrint: bulkActions.handleRetryFailedPrint,
      handleBulkCancel: bulkActions.handleBulkCancel,
      handleSingleShip: singleActions.handleSingleShip,
      handleShipConfirm: singleActions.handleShipConfirm,
      handleSingleCancel: singleActions.handleSingleCancel,
      handleCancelConfirm: singleActions.handleCancelConfirm,
      handleSinglePrint: singleActions.handleSinglePrint,
      handleViewDetails: singleActions.handleViewDetails,
      handleExport,
    },
  };
}
