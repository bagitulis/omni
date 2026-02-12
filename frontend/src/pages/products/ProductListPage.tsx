import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import {
  Button,
  Card,
  Space,
  Typography,
  message,
  Modal,
  Form,
  InputNumber,
} from "antd";
import {
  PlusOutlined,
  UploadOutlined,
  CloudDownloadOutlined,
} from "@ant-design/icons";
import { ProductFilters } from "@/components/forms/ProductFilters";
import { ProductTable } from "@/components/tables/ProductTable";
import { CloneProductModal } from "@/components/clone/CloneProductModal";
import { CloneBatchModal } from "@/components/clone/CloneBatchModal";
import { useProducts, useDeleteProduct } from "@/hooks/useProducts";
import { useBatchSkuUpdate } from "@/hooks/useBatchSkuUpdate";
import { ProductGridView } from "./components/ProductGridView";
import { ProductBatchBar } from "./components/ProductBatchBar";
import type { Product } from "@/types/product";

export function ProductListPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const [viewMode, setViewMode] = useState<"grid" | "list">("list");
  const [filters, setFilters] = useState({
    search: "",
    status: "all",
    platform: "all",
    category: "all",
  });
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [cloneModalOpen, setCloneModalOpen] = useState(false);
  const [batchCloneModalOpen, setBatchCloneModalOpen] = useState(false);
  const [batchSkuModalOpen, setBatchSkuModalOpen] = useState(false);
  const [selectedProductForClone, setSelectedProductForClone] =
    useState<Product | null>(null);
  const [batchSkuForm] = Form.useForm();

  const { data, isLoading } = useProducts({
    page,
    limit: pageSize,
    status: filters.status,
    search: filters.search,
    platform: filters.platform,
  });

  const deleteMutation = useDeleteProduct();
  const { mutation: batchSkuUpdateMutation } = useBatchSkuUpdate();

  useEffect(() => {
    const platformFromUrl = searchParams.get("platform");
    if (!platformFromUrl) return;

    const normalized = platformFromUrl.toLowerCase();
    const allowedPlatforms = new Set(["all", "shopee", "lazada", "tiktok"]);
    if (!allowedPlatforms.has(normalized)) return;

    setFilters((prev) => {
      if (prev.platform === normalized) return prev;
      return { ...prev, platform: normalized };
    });
  }, [searchParams]);

  const handleFilterChange = (
    key: string,
    value: string | number | boolean | undefined,
  ) => {
    setFilters((prev) => ({ ...prev, [key]: value }));
    setPage(1); // Reset to first page

    if (key === "platform") {
      setSearchParams((prev) => {
        const next = new URLSearchParams(prev);

        if (value === "all") {
          next.delete("platform");
        } else {
          next.set("platform", String(value));
        }

        return next;
      });
    }
  };

  const handleDelete = (id: string) => {
    if (confirm("Are you sure you want to delete this product?")) {
      deleteMutation.mutate(id, {
        onSuccess: () => {
          message.success("Product deleted successfully");
        },
        onError: () => {
          message.error("Failed to delete product");
        },
      });
    }
  };

  const handleClone = (product: Product) => {
    setSelectedProductForClone(product);
    setCloneModalOpen(true);
  };

  const handleBatchSkuUpdate = () => {
    batchSkuForm.validateFields().then((values) => {
      const items = selectedRowKeys.map((key) => ({
        id: Number(key),
        ...(values.price !== undefined && { price: values.price }),
        ...(values.stock !== undefined && { stock: values.stock }),
      }));

      batchSkuUpdateMutation.mutate(items, {
        onSuccess: () => {
          setBatchSkuModalOpen(false);
          setSelectedRowKeys([]);
          batchSkuForm.resetFields();
        },
      });
    });
  };

  return (
    <div style={{ padding: 24 }}>
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          marginBottom: 24,
        }}
      >
        <Typography.Title level={2} style={{ margin: 0 }}>
          Products
        </Typography.Title>
        <Space>
          <Button icon={<CloudDownloadOutlined />}>Export</Button>
          <Button icon={<UploadOutlined />}>Import</Button>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => navigate("/master-products/add")}
          >
            Add Product
          </Button>
        </Space>
      </div>

      <Card>
        <ProductFilters
          filters={filters}
          onFilterChange={handleFilterChange}
          viewMode={viewMode}
          onViewModeChange={setViewMode}
        />

        <ProductBatchBar
          selectedRowKeys={selectedRowKeys}
          onBulkSync={() => message.info("Bulk Sync feature coming soon")}
          onBatchSkuUpdate={() => setBatchSkuModalOpen(true)}
          onBatchClone={() => setBatchCloneModalOpen(true)}
          onDeleteSelected={() => {
            if (confirm(`Delete ${selectedRowKeys.length} items?`)) {
              selectedRowKeys.forEach((key) => {
                deleteMutation.mutate(String(key));
              });
              setSelectedRowKeys([]);
            }
          }}
        />

        {viewMode === "list" ? (
          <ProductTable
            loading={isLoading}
            products={data?.products || []}
            total={data?.total || 0}
            page={page}
            pageSize={pageSize}
            selectedRowKeys={selectedRowKeys}
            onPageChange={(p, ps) => {
              setPage(p);
              setPageSize(ps);
            }}
            onSelectionChange={setSelectedRowKeys}
            onDelete={handleDelete}
            onClone={handleClone}
          />
        ) : (
          <ProductGridView
            products={data?.products || []}
            page={page}
            pageSize={pageSize}
            total={data?.total || 0}
            onPageChange={(p, ps) => {
              setPage(p);
              setPageSize(ps);
            }}
            onDelete={handleDelete}
          />
        )}

        <CloneProductModal
          open={cloneModalOpen}
          onClose={() => {
            setCloneModalOpen(false);
            setSelectedProductForClone(null);
          }}
          initialSku={selectedProductForClone?.item_sku || ""}
          initialPlatform={selectedProductForClone?.platform || "shopee"}
        />

        <CloneBatchModal
          open={batchCloneModalOpen}
          onClose={() => setBatchCloneModalOpen(false)}
          products={
            data?.products.filter((p) => selectedRowKeys.includes(p.item_id)) ||
            []
          }
        />

        <Modal
          title="Batch Update SKUs"
          open={batchSkuModalOpen}
          onOk={handleBatchSkuUpdate}
          onCancel={() => {
            setBatchSkuModalOpen(false);
            batchSkuForm.resetFields();
          }}
          confirmLoading={batchSkuUpdateMutation.isPending}
        >
          <p>
            Update price and/or stock for {selectedRowKeys.length} selected
            product(s).
          </p>
          <Form form={batchSkuForm} layout="vertical">
            <Form.Item
              label="New Price"
              name="price"
              help="Leave empty to keep current price"
            >
              <InputNumber
                style={{ width: "100%" }}
                min={0}
                placeholder="Enter new price"
              />
            </Form.Item>
            <Form.Item
              label="New Stock"
              name="stock"
              help="Leave empty to keep current stock"
            >
              <InputNumber
                style={{ width: "100%" }}
                min={0}
                placeholder="Enter new stock"
              />
            </Form.Item>
          </Form>
        </Modal>
      </Card>
    </div>
  );
}

export default ProductListPage;
