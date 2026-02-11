import { useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Card, Flex, Tabs, message, Modal, InputNumber } from "antd";
import { useMutation } from "@tanstack/react-query";
import type { ProductManagerPlatform } from "@/types/product_manager";
import { useDbProducts } from "@/hooks/useProductManager";
import { syncPlatformProducts } from "@/api/productManager";
import { useBatchSkuCheck } from "./hooks/useBatchSkuCheck";
import { PLATFORMS, coercePlatform, extractSku, getRowKey } from "./utils";
import { ProductManagerHeader } from "./components/ProductManagerHeader";
import { ProductManagerTable } from "./components/ProductManagerTable";
import { usePriceUpdate } from "./hooks/usePriceUpdate";
import { useFilterPreferences } from "./hooks/useFilterPreferences";

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
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [isBatchModalOpen, setIsBatchModalOpen] = useState(false);
  const [batchPrice, setBatchPrice] = useState<number>(0);

  const lastAutoSyncedPlatformRef = useRef<ProductManagerPlatform | null>(null);

  const {
    preferences,
    updatePreference,
    isLoading: preferencesLoading,
  } = useFilterPreferences(activeTab);

  const {
    results,
    getSkuResult,
    clearResults,
    mutation: skuCheckMutation,
  } = useBatchSkuCheck();

  const { singleMutation, batchMutation } = usePriceUpdate(() => {
    void refetch();
  });

  useEffect(() => {
    setActiveTab(coercePlatform(platformParam));
    setSelectedRowKeys([]);
  }, [platformParam]);

  // Apply loaded preferences to page size
  useEffect(() => {
    if (!preferencesLoading && preferences.page_size) {
      setPageSize(preferences.page_size);
    }
  }, [preferencesLoading, preferences.page_size]);

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
    setSelectedRowKeys([]);
    clearResults();
    navigate(`/product-manager/${next}`);
  };

  const tabItems = PLATFORMS.map((p) => ({
    key: p.key,
    label: p.label,
  }));

  const handlePageChange = (p: number, ps: number) => {
    setPage(p);
    setPageSize(ps);

    // Save page size preference
    if (ps !== preferences.page_size) {
      updatePreference({ page_size: ps });
    }
  };

  const handleSync = () => {
    syncMutation.mutate();
  };

  const handleCheckSku = () => {
    if (selectedRowKeys.length === 0) {
      message.warning("Please select products first");
      return;
    }

    const products = data?.products || [];
    const selectedProducts = products.filter((p) =>
      selectedRowKeys.includes(getRowKey(activeTab, p)),
    );

    const skus = selectedProducts
      .map((row) => extractSku(row))
      .filter((s) => s !== "");

    if (skus.length === 0) {
      message.warning("No valid SKUs found in selection");
      return;
    }

    skuCheckMutation.mutate(skus);
  };

  const handleSelectionChange = (keys: React.Key[]) => {
    setSelectedRowKeys(keys);
  };

  const handleUpdateSinglePrice = async (sku: string, price: number) => {
    await singleMutation.mutateAsync({ sku, price });
  };

  const handleOpenBatchModal = () => {
    if (selectedRowKeys.length === 0) return;
    setBatchPrice(0);
    setIsBatchModalOpen(true);
  };

  const handleBatchUpdate = () => {
    if (batchPrice <= 0) {
      message.error("Price must be greater than 0");
      return;
    }

    const products = data?.products || [];
    const selectedProducts = products.filter((p) =>
      selectedRowKeys.includes(getRowKey(activeTab, p)),
    );

    const items = selectedProducts
      .map((p) => {
        const sku = extractSku(p);
        if (!sku) return null;
        return { sku, price: batchPrice };
      })
      .filter((item) => item !== null) as { sku: string; price: number }[];

    if (items.length === 0) {
      message.error("No valid SKUs found in selection");
      return;
    }

    batchMutation.mutate(items, {
      onSuccess: () => {
        setIsBatchModalOpen(false);
        setSelectedRowKeys([]);
      },
    });
  };

  const handleColumnVisibilityChange = (
    columnVisibility: Record<string, boolean>,
  ) => {
    updatePreference({ column_visibility: columnVisibility });
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
          onCheckSku={handleCheckSku}
          skuCheckLoading={skuCheckMutation.isPending}
          selectedCount={selectedRowKeys.length}
          onUpdateBatch={handleOpenBatchModal}
          hasSelection={selectedRowKeys.length > 0}
          columnVisibility={preferences.column_visibility}
          onColumnVisibilityChange={handleColumnVisibilityChange}
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
            onUpdatePrice={handleUpdateSinglePrice}
            selectedRowKeys={selectedRowKeys}
            onSelectionChange={handleSelectionChange}
            skuCheckResults={results}
            getSkuResult={getSkuResult}
            columnVisibility={preferences.column_visibility}
          />
        </Card>
      </Flex>

      <Modal
        title={`Update Price for ${selectedRowKeys.length} Products`}
        open={isBatchModalOpen}
        onOk={handleBatchUpdate}
        onCancel={() => setIsBatchModalOpen(false)}
        confirmLoading={batchMutation.isPending}
      >
        <Flex vertical gap={8}>
          <div>New Price:</div>
          <InputNumber
            style={{ width: "100%" }}
            min={1}
            value={batchPrice}
            onChange={(v) => setBatchPrice(v ?? 0)}
            formatter={(value) =>
              `Rp ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ",")
            }
            parser={(value) => Number(value?.replace(/Rp\s?|(,*)/g, "") || 0)}
          />
        </Flex>
      </Modal>
    </div>
  );
}
