import {
  AppstoreOutlined,
  BarsOutlined,
  CloudDownloadOutlined,
  PlusOutlined,
  UploadOutlined,
} from "@ant-design/icons";
import {
  Button,
  Card,
  Divider,
  Grid,
  Space,
  Table,
  type TableColumnsType,
  Typography,
} from "antd";
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
import { ColumnManager } from "@/components/shared/ColumnManager";
import { ProductFilters } from "@/components/shared/ProductFilters";
import { UnifiedBatchBar } from "@/components/shared/UnifiedBatchBar";
import { useColumnManager } from "@/hooks/useColumnManager";
import { useMarketplaceSyncHistory } from "@/hooks/useMarketplaceSyncHistory";
import { useUnifiedProducts } from "@/hooks/useUnifiedProducts";
import { ProductGridView } from "@/pages/products/components/ProductGridView";
import { UnifiedProductsModals } from "@/pages/products/components/UnifiedProductsModals";
import { useUnifiedProductsActions } from "@/pages/products/hooks/useUnifiedProductsActions";
import { buildProductColumns } from "@/pages/products/utils/productColumns";
import {
  areFiltersEqual,
  buildFilterSearchParams,
  DEFAULT_PRODUCT_PAGE_COLUMNS,
  readFiltersFromUrl,
  toLegacyProduct,
} from "@/pages/products/utils/unifiedProductUtils";
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
    cloneModalOpen,
    setCloneModalOpen,
    batchCloneOpen,
    setBatchCloneOpen,
    skuMappingOpen,
    setSkuMappingOpen,
    wholesaleMpqOpen,
    setWholesaleMpqOpen,
    wholesaleMpqDefaultTab,
    batchPriceOpen,
    setBatchPriceOpen,
    clonePreviewOpen,
    setClonePreviewOpen,
    batchPriceValue,
    setBatchPriceValue,
    selectedProduct,
    setSelectedProduct,
    skuMappingProduct,
    setSkuMappingProduct,
    skuMappingLoading,
    selectedSkus,
    handleDeleteProduct,
    handleInlinePriceSave,
    handleInlineStockSave,
    handleBatchPriceUpdate,
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
        onInlinePriceSave: handleInlinePriceSave,
        onInlineStockSave: handleInlineStockSave,
        onRowAction: handleRowAction,
      }),
    [handleInlinePriceSave, handleInlineStockSave, handleRowAction],
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
    <div style={{ padding: 24 }}>
      <Space direction="vertical" size={16} style={{ width: "100%" }}>
        <Space style={{ width: "100%", justifyContent: "space-between" }} wrap>
          <Typography.Title level={2} style={{ margin: 0 }}>
            Products
          </Typography.Title>
          <Space wrap>
            <Button
              icon={<PlusOutlined />}
              type="primary"
              onClick={() => navigate("/master-products/add")}
            >
              Add Product
            </Button>
            <Button
              icon={<UploadOutlined />}
              onClick={() => navigate("/master-products/import")}
            >
              Import
            </Button>
            <Button
              icon={<CloudDownloadOutlined />}
              onClick={() => navigate("/products/sync-history")}
            >
              Sync History ({syncHistoryData?.total ?? 0})
            </Button>
          </Space>
        </Space>

        <Suspense
          fallback={<div style={{ padding: 8 }}>Loading sync panel...</div>}
        >
          <PlatformSyncPanel />
        </Suspense>

        <Card>
          <ProductFilters
            values={filters}
            onChange={handleFilterChange}
            showCategory={true}
            showStatus={true}
          />
          <Divider style={{ margin: "8px 0" }} />

          <Space
            style={{
              width: "100%",
              justifyContent: "space-between",
              marginBottom: 12,
            }}
          >
            <Space className="view-toggle">
              <Button
                data-view="list"
                type={viewMode === "list" ? "primary" : "default"}
                icon={<BarsOutlined />}
                onClick={() => setViewMode("list")}
              >
                List
              </Button>
              <Button
                data-view="grid"
                type={viewMode === "grid" ? "primary" : "default"}
                icon={<AppstoreOutlined />}
                onClick={() => setViewMode("grid")}
              >
                Grid
              </Button>
            </Space>

            {!isMobile ? (
              <ColumnManager
                columns={columns}
                onChange={updateOrder}
                onReset={reset}
              />
            ) : null}
          </Space>

          {viewMode === "list" ? (
            <Table<UnifiedProductRow>
              rowKey="id"
              loading={isLoading}
              columns={activeColumns}
              dataSource={products}
              rowSelection={{
                selectedRowKeys,
                onChange: (keys, rows) => {
                  setSelectedRowKeys(keys);
                  setSelectedRecords(rows);
                },
              }}
              scroll={{ x: "max-content" }}
              pagination={{
                current: page,
                pageSize,
                total,
                showSizeChanger: true,
                showTotal: (count) => `Total ${count} products`,
                onChange: (nextPage, nextPageSize) => {
                  setPage(nextPage);
                  setPageSize(nextPageSize);
                },
              }}
            />
          ) : (
            <div className="product-grid-view">
              <ProductGridView
                products={legacyProducts}
                page={page}
                pageSize={pageSize}
                total={total}
                onPageChange={(nextPage, nextPageSize) => {
                  setPage(nextPage);
                  setPageSize(nextPageSize);
                }}
                onDelete={(id) => {
                  void handleDeleteProduct(id);
                }}
              />
            </div>
          )}
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
        selectedLegacyProducts={selectedLegacyProducts}
        selectedProduct={selectedProduct}
        selectedSkus={selectedSkus}
        wholesaleMpqDefaultTab={wholesaleMpqDefaultTab}
        batchPriceValue={batchPriceValue}
        stockSyncOpen={stockSyncOpen}
        wholesaleMpqOpen={wholesaleMpqOpen}
        batchPriceOpen={batchPriceOpen}
        clonePreviewOpen={clonePreviewOpen}
        cloneModalOpen={cloneModalOpen}
        batchCloneOpen={batchCloneOpen}
        skuMappingOpen={skuMappingOpen}
        skuMappingLoading={skuMappingLoading}
        skuMappingProduct={skuMappingProduct}
        onStockSyncClose={() => setStockSyncOpen(false)}
        onWholesaleMpqClose={() => setWholesaleMpqOpen(false)}
        onBatchPriceClose={() => setBatchPriceOpen(false)}
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
        onBatchPriceValueChange={setBatchPriceValue}
        onBatchPriceUpdate={handleBatchPriceUpdate}
        onStockSync={handleStockSync}
        onSkuMappingUpdate={() => {
          void refreshProducts();
        }}
      />
    </div>
  );
}
