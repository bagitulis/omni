import { useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Card, Flex, Tabs, message } from "antd";
import { useMutation } from "@tanstack/react-query";
import type { ProductManagerPlatform } from "@/types/product_manager";
import { useDbProducts } from "@/hooks/useProductManager";
import { syncPlatformProducts } from "@/api/productManager";
import { PLATFORMS, coercePlatform } from "./utils";
import { ProductManagerHeader } from "./components/ProductManagerHeader";
import { ProductManagerTable } from "./components/ProductManagerTable";

export function ProductManagerPage() {
  const navigate = useNavigate();
  const { platform: platformParam } = useParams<{ platform?: string }>();

  const [activeTab, setActiveTab] = useState<ProductManagerPlatform>(
    coercePlatform(platformParam),
  );
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [autoSync, setAutoSync] = useState(true);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(50);
  const lastAutoSyncedPlatformRef = useRef<ProductManagerPlatform | null>(null);

  useEffect(() => {
    setActiveTab(coercePlatform(platformParam));
  }, [platformParam]);

  const offset = (page - 1) * pageSize;

  const { data, isLoading, refetch } = useDbProducts(
    activeTab,
    { offset, limit: pageSize },
    { autoRefresh, refetchInterval: 15_000 },
  );

  const syncMutation = useMutation({
    mutationFn: () => syncPlatformProducts(activeTab),
    onSuccess: (result) => {
      const synced = result.data?.synced;
      message.success(
        typeof synced === "number"
          ? `Synced ${synced} products from ${activeTab}`
          : `Products synced from ${activeTab}`,
      );
      void refetch();
    },
    onError: (error: Error) => {
      message.error(error.message || "Failed to sync products");
    },
  });

  useEffect(() => {
    if (!autoSync) return;
    if (syncMutation.isPending) return;
    if (lastAutoSyncedPlatformRef.current === activeTab) return;

    lastAutoSyncedPlatformRef.current = activeTab;
    void syncMutation.mutateAsync();
  }, [activeTab, autoSync, syncMutation]);

  const handleTabChange = (key: string) => {
    const next = coercePlatform(key);
    setActiveTab(next);
    setPage(1);
    navigate(`/product-manager/${next}`);
  };

  const tabItems = PLATFORMS.map((p) => ({
    key: p.key,
    label: p.label,
  }));

  const handlePageChange = (p: number, ps: number) => {
    setPage(p);
    setPageSize(ps);
  };

  const handleSync = () => {
    syncMutation.mutate();
  };

  return (
    <div style={{ padding: 24 }}>
      <Flex vertical gap={16}>
        <ProductManagerHeader
          autoRefresh={autoRefresh}
          setAutoRefresh={setAutoRefresh}
          autoSync={autoSync}
          setAutoSync={setAutoSync}
          syncLoading={syncMutation.isPending}
          onSync={handleSync}
        />

        <Card style={{ borderRadius: 4 }}>
          <Tabs
            activeKey={activeTab}
            onChange={handleTabChange}
            items={tabItems}
          />
          <ProductManagerTable
            activeTab={activeTab}
            isLoading={isLoading}
            products={data?.products ?? []}
            total={data?.total ?? 0}
            page={page}
            pageSize={pageSize}
            onPageChange={handlePageChange}
          />
        </Card>
      </Flex>
    </div>
  );
}
