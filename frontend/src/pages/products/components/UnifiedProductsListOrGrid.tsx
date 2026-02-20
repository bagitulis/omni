import { Table, type TableColumnsType } from "antd";
import type { Key } from "react";
import { ProductGridView } from "@/pages/products/components/ProductGridView";
import { ProductVariantExpandedRow } from "@/pages/products/components/ProductVariantExpandedRow";
import type { Product } from "@/types/product";
import type { UnifiedProductRow } from "@/types/shared";

interface UnifiedProductsListOrGridProps {
  viewMode: "grid" | "list";
  isLoading: boolean;
  activeColumns: TableColumnsType<UnifiedProductRow>;
  products: UnifiedProductRow[];
  legacyProducts: Product[];
  selectedRowKeys: Key[];
  onSelectionChange: (keys: Key[], rows: UnifiedProductRow[]) => void;
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (nextPage: number, nextPageSize: number) => void;
  onDeleteProduct: (productId: number | string) => Promise<void>;
}

export function UnifiedProductsListOrGrid({
  viewMode,
  isLoading,
  activeColumns,
  products,
  legacyProducts,
  selectedRowKeys,
  onSelectionChange,
  page,
  pageSize,
  total,
  onPageChange,
  onDeleteProduct,
}: UnifiedProductsListOrGridProps) {
  if (viewMode === "list") {
    return (
      <Table<UnifiedProductRow>
        rowKey="id"
        loading={isLoading}
        columns={activeColumns}
        dataSource={products}
        size="small"
        bordered
        expandable={{
          rowExpandable: (record) => record.skus.length > 1,
          expandIconColumnIndex: 1,
          expandedRowRender: (record) => (
            <ProductVariantExpandedRow product={record} />
          ),
        }}
        rowSelection={{
          selectedRowKeys,
          onChange: onSelectionChange,
        }}
        scroll={{ x: "max-content" }}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          showTotal: (count) => `Total ${count} products`,
          onChange: onPageChange,
        }}
      />
    );
  }

  return (
    <div className="product-grid-view">
      <ProductGridView
        products={legacyProducts}
        page={page}
        pageSize={pageSize}
        total={total}
        onPageChange={onPageChange}
        onDelete={(id) => {
          void onDeleteProduct(id);
        }}
      />
    </div>
  );
}
