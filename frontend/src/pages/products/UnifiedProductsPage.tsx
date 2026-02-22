import { Card, Grid, Space, Tabs, type TableColumnsType } from "antd";
import type { Key } from "react";
import {
  lazy,
  Suspense,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { UnifiedBatchBar } from "@/components/shared/UnifiedBatchBar";
import { useColumnManager } from "@/hooks/useColumnManager";
import { useMarketplaceSyncHistory } from "@/hooks/useMarketplaceSyncHistory";
import { useUnifiedProducts } from "@/hooks/useUnifiedProducts";
import { UnifiedProductsHeaderActions } from "@/pages/products/components/UnifiedProductsHeaderActions";
import { UnifiedProductsListOrGrid } from "@/pages/products/components/UnifiedProductsListOrGrid";
import { UnifiedProductsModals } from "@/pages/products/components/UnifiedProductsModals";
import { UnifiedProductsControls } from "@/pages/products/components/UnifiedProductsControls";
import { useUnifiedProductsActions } from "@/pages/products/hooks/useUnifiedProductsActions";
import { buildProductColumns } from "@/pages/products/utils/productColumns";
import {
  areFiltersEqual,
  buildFilterSearchParams,
  DEFAULT_PRODUCT_PAGE_COLUMNS,
  readFiltersFromUrl,
  toLegacyProduct,
} from "@/pages/products/utils/unifiedProductUtils";
import { buildProductTabItems } from "@/pages/products/utils/productTabItems";
import type {
  BatchActionType,
  ProductFilterValues,
  UnifiedProductRow,
} from "@/types/shared";

const PlatformSyncPanel = lazy(() =>
  import("@/components/shared/PlatformSyncPanel").then((m) => ({
    default: m.PlatformSyncPanel,
  })),
);

const MOBILE_ESSENTIAL_COLUMN_KEYS = ["image", "name", "price", "actions"];

export default function UnifiedProductsPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const screens = Grid.useBreakpoint();
  const isMobile = !screens.md;

  const [viewMode, setViewMode] = useState<"grid" | "list">("list");
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [selectedRowKeys, setSelectedRowKeys] = useState<Key[]>([]);
  const [selectedRecords, setSelectedRecords] = useState<UnifiedProductRow[]>(
    [],
  );
  const [filters, setFilters] = useState<ProductFilterValues>(() =>
    readFiltersFromUrl(searchParams),
  );

  const { data, isLoading, refetch } = useUnifiedProducts(
    filters,
    page,
    pageSize,
  );
  const { data: syncHistoryData } = useMarketplaceSyncHistory({
    page: 1,
    page_size: 1,
  });
  const { columns, visibleColumns, updateOrder, reset } = useColumnManager(
    DEFAULT_PRODUCT_PAGE_COLUMNS,
    "products-page-columns",
  );

  const allProducts = useMemo(() => data?.products ?? [], [data?.products]);
  const products = allProducts;
  const total = data?.total ?? 0;
  const legacyProducts = useMemo(
    () => products.map(toLegacyProduct),
    [products],
  );
  const selectedLegacyProducts = useMemo(
    () => selectedRecords.map(toLegacyProduct),
    [selectedRecords],
  );

  useEffect(() => {
    const nextFilters = readFiltersFromUrl(searchParams);
    setFilters((current) =>
      areFiltersEqual(current, nextFilters) ? current : nextFilters,
    );
  }, [searchParams]);

  const handleFilterChange = useCallback(
    (nextFilters: ProductFilterValues) => {
      setFilters(nextFilters);
      setPage(1);
      setSearchParams(buildFilterSearchParams(nextFilters));
    },
    [setSearchParams],
  );

  const clearSelection = useCallback(() => {
    setSelectedRowKeys([]);
    setSelectedRecords([]);
  }, []);

  const refreshProducts = useCallback(async () => {
    await refetch();
  }, [refetch]);

  const {
    stockSyncOpen,
    setStockSyncOpen,
    priceSyncOpen,
    setPriceSyncOpen,
    cloneModalOpen,
    setCloneModalOpen,
    batchCloneOpen,
    setBatchCloneOpen,
    skuMappingOpen,
    setSkuMappingOpen,
    wholesaleMpqOpen,
    setWholesaleMpqOpen,
    wholesaleMpqDefaultTab,
    clonePreviewOpen,
    setClonePreviewOpen,
    stockSyncProducts,
    priceSyncProducts,
    selectedProduct,
    setSelectedProduct,
    skuMappingProduct,
    setSkuMappingProduct,
    skuMappingLoading,
    handleDeleteProduct,
    handlePriceSync,
    handleStockSync,
    handleRowAction,
    handleBatchAction,
  } = useUnifiedProductsActions({
    navigate,
    selectedRecords,
    clearSelection,
    refreshProducts,
  });

  const tableColumnMap = useMemo(
    () =>
      buildProductColumns({
        onRowAction: handleRowAction,
      }),
    [handleRowAction],
  );

  const activeColumns = useMemo(() => {
    const sourceColumns = isMobile
      ? visibleColumns.filter((column) =>
          MOBILE_ESSENTIAL_COLUMN_KEYS.includes(column.key),
        )
      : visibleColumns;

    return sourceColumns.reduce<TableColumnsType<UnifiedProductRow>>(
      (acc, column) => {
        const resolvedColumn = tableColumnMap[column.key];
        if (resolvedColumn) acc.push(resolvedColumn);
        return acc;
      },
      [],
    );
  }, [isMobile, tableColumnMap, visibleColumns]);

  return (
    <div style={{ padding: isMobile ? 12 : 24 }}>
      <Space direction="vertical" size={16} style={{ width: "100%" }}>
        <UnifiedProductsHeaderActions
          navigate={navigate}
          syncHistoryTotal={syncHistoryData?.total ?? 0}
        />

        <Card styles={{ body: { padding: isMobile ? 16 : 24 } }}>
          <Suspense
            fallback={<div style={{ padding: 8 }}>Loading sync panel...</div>}
          >
            <PlatformSyncPanel onImportCompleted={refreshProducts} />
          </Suspense>
        </Card>

        <Card styles={{ body: { padding: isMobile ? 16 : 24 } }}>
          <Tabs
            activeKey={filters.mapping}
            onChange={(key) =>
              handleFilterChange({
                ...filters,
                mapping: key as "all" | "mapped" | "unmapped",
              })
            }
            items={buildProductTabItems(total, filters.mapping)}
            style={{ marginBottom: 16 }}
          />

          <UnifiedProductsControls
            filters={filters}
            onFilterChange={handleFilterChange}
            viewMode={viewMode}
            onViewModeChange={setViewMode}
            isMobile={isMobile}
            columns={columns}
            onColumnsChange={updateOrder}
            onResetColumns={reset}
          />

          <UnifiedProductsListOrGrid
            viewMode={viewMode}
            isLoading={isLoading}
            activeColumns={activeColumns}
            products={products}
            legacyProducts={legacyProducts}
            selectedRowKeys={selectedRowKeys}
            onSelectionChange={(keys, rows) => {
              setSelectedRowKeys(keys);
              setSelectedRecords(rows);
            }}
            page={page}
            pageSize={pageSize}
            total={total}
            onPageChange={(nextPage, nextPageSize) => {
              setPage(nextPage);
              setPageSize(nextPageSize);
            }}
            onDeleteProduct={handleDeleteProduct}
          />
        </Card>
      </Space>

      {selectedRowKeys.length > 0 ? (
        <div className="unified-batch-bar">
          <UnifiedBatchBar
            selectedCount={selectedRowKeys.length}
            onAction={(action: BatchActionType) => {
              void handleBatchAction(action);
            }}
            onClearSelection={clearSelection}
          />
        </div>
      ) : null}

      <UnifiedProductsModals
        selectedRecords={selectedRecords}
        stockSyncProducts={stockSyncProducts}
        priceSyncProducts={priceSyncProducts}
        selectedLegacyProducts={selectedLegacyProducts}
        selectedProduct={selectedProduct}
        wholesaleMpqDefaultTab={wholesaleMpqDefaultTab}
        stockSyncOpen={stockSyncOpen}
        priceSyncOpen={priceSyncOpen}
        wholesaleMpqOpen={wholesaleMpqOpen}
        clonePreviewOpen={clonePreviewOpen}
        cloneModalOpen={cloneModalOpen}
        batchCloneOpen={batchCloneOpen}
        skuMappingOpen={skuMappingOpen}
        skuMappingLoading={skuMappingLoading}
        skuMappingProduct={skuMappingProduct}
        onStockSyncClose={() => setStockSyncOpen(false)}
        onPriceSyncClose={() => setPriceSyncOpen(false)}
        onWholesaleMpqClose={() => setWholesaleMpqOpen(false)}
        onClonePreviewClose={() => setClonePreviewOpen(false)}
        onClonePreviewContinue={() => {
          setClonePreviewOpen(false);
          setCloneModalOpen(true);
        }}
        onCloneModalClose={() => {
          setCloneModalOpen(false);
          setSelectedProduct(null);
        }}
        onBatchCloneClose={() => setBatchCloneOpen(false)}
        onSkuMappingClose={() => {
          setSkuMappingOpen(false);
          setSkuMappingProduct(null);
        }}
        onPriceSync={handlePriceSync}
        onStockSync={handleStockSync}
        onSkuMappingUpdate={() => {
          void refreshProducts();
        }}
      />
    </div>
  );
}
