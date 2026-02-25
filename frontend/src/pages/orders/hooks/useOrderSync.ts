import { useCallback, useEffect, useState } from "react";
import { message } from "antd";
import {
  isSyncableOrderTab,
  syncOrdersByCategory,
  syncOrdersToday,
} from "@/api/orders";

export function useOrderSync(
  activeTab: string,
  platform: string,
  refetch: () => void,
  autoRefresh: boolean,
) {
  const [isSyncing, setIsSyncing] = useState(false);

  const syncActiveTab = useCallback(
    async (tabKey: string) => {
      setIsSyncing(true);
      try {
        if (isSyncableOrderTab(tabKey)) {
          await syncOrdersByCategory(tabKey, platform);
        } else if (tabKey === "today") {
          await syncOrdersToday();
        }
        // Note: "locked" tab does NOT need explicit sync here because
        // getOrders("locked") already does POST which syncs + returns data.
        // Adding lockOrdersToday() here would cause duplicate POST calls.
      } catch (error) {
        message.error(
          error instanceof Error ? error.message : "Failed to sync orders",
        );
      } finally {
        setIsSyncing(false);
        refetch();
      }
    },
    [platform, refetch],
  );

  useEffect(() => {
    void syncActiveTab(activeTab);
  }, [activeTab, syncActiveTab]);

  useEffect(() => {
    if (!autoRefresh) return;
    if (!isSyncableOrderTab(activeTab)) return;

    const intervalId = window.setInterval(() => {
      if (isSyncing) return;
      void syncActiveTab(activeTab);
    }, 30_000);

    return () => window.clearInterval(intervalId);
  }, [activeTab, autoRefresh, isSyncing, syncActiveTab]);

  return { isSyncing, syncActiveTab };
}
