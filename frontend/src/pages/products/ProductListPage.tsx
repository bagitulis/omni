import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import {
  Button,
  Card,
  Col,
  Row,
  Space,
  Typography,
  Pagination,
  Empty,
  message,
} from "antd";
import {
  PlusOutlined,
  UploadOutlined,
  CloudDownloadOutlined,
  EditOutlined,
  DeleteOutlined,
  CopyOutlined,
} from "@ant-design/icons";
import { ProductFilters } from "@/components/forms/ProductFilters";
import { ProductTable } from "@/components/tables/ProductTable";
import { CloneProductModal } from "@/components/clone/CloneProductModal";
import { CloneBatchModal } from "@/components/clone/CloneBatchModal";
import { useProducts, useDeleteProduct } from "@/hooks/useProducts";
import { PlatformBadge } from "@/components/ui/PlatformBadge";
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
  const [selectedProductForClone, setSelectedProductForClone] =
    useState<Product | null>(null);

  const { data, isLoading } = useProducts({
    page,
    limit: pageSize,
    status: filters.status,
    search: filters.search,
    platform: filters.platform,
  });

  const deleteMutation = useDeleteProduct();

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

  const renderGridView = () => {
    if (!data?.products.length)
      return <Empty description="No products found" />;

    return (
      <>
        <Row gutter={[16, 16]}>
          {data.products.map((product) => (
            <Col key={product.item_id} xs={24} sm={12} md={8} lg={6} xl={4}>
              <Card
                hoverable
                cover={
                  <img
                    alt={product.item_name}
                    src={product.image_url}
                    style={{ height: 200, objectFit: "cover" }}
                  />
                }
                actions={[
                  <Button type="text" icon={<EditOutlined />} key="edit" />,
                  <Button
                    type="text"
                    danger
                    icon={<DeleteOutlined />}
                    key="delete"
                    onClick={() => handleDelete(product.item_id)}
                  />,
                ]}
              >
                <Card.Meta
                  title={
                    <div
                      style={{
                        display: "flex",
                        justifyContent: "space-between",
                        alignItems: "center",
                      }}
                    >
                      <Typography.Text strong ellipsis>
                        {product.item_name}
                      </Typography.Text>
                    </div>
                  }
                  description={
                    <Space
                      direction="vertical"
                      size={4}
                      style={{ width: "100%" }}
                    >
                      <Space>
                        <PlatformBadge platform={product.platform} />
                        <Typography.Text type="secondary">
                          {product.item_sku}
                        </Typography.Text>
                      </Space>
                      <div
                        style={{
                          display: "flex",
                          justifyContent: "space-between",
                        }}
                      >
                        <Typography.Text strong type="warning">
                          {product.price === null || product.price === undefined
                            ? "—"
                            : new Intl.NumberFormat("id-ID", {
                                style: "currency",
                                currency: "IDR",
                                maximumFractionDigits: 0,
                              }).format(product.price)}
                        </Typography.Text>
                        <Typography.Text
                          type={
                            product.stock === null ||
                            product.stock === undefined
                              ? "secondary"
                              : product.stock === 0
                                ? "danger"
                                : product.stock <= 10
                                  ? "warning"
                                  : "secondary"
                          }
                        >
                          Stock:{" "}
                          {product.stock === null || product.stock === undefined
                            ? "—"
                            : product.stock}
                        </Typography.Text>
                      </div>
                    </Space>
                  }
                />
              </Card>
            </Col>
          ))}
        </Row>
        <div style={{ marginTop: 16, textAlign: "right" }}>
          <Pagination
            current={page}
            pageSize={pageSize}
            total={data.total}
            onChange={(p, ps) => {
              setPage(p);
              setPageSize(ps);
            }}
            showSizeChanger
          />
        </div>
      </>
    );
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

        {selectedRowKeys.length > 0 && (
          <div
            style={{
              marginBottom: 16,
              padding: "8px 16px",
              background: "#e6f7ff",
              border: "1px solid #91d5ff",
              borderRadius: 4,
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
            }}
          >
            <span>Selected {selectedRowKeys.length} items</span>
            <Space>
              <Button
                size="small"
                onClick={() => message.info("Bulk Sync feature coming soon")}
              >
                Bulk Sync
              </Button>
              <Button
                size="small"
                icon={<CopyOutlined />}
                onClick={() => setBatchCloneModalOpen(true)}
              >
                Batch Clone
              </Button>
              <Button
                size="small"
                danger
                onClick={() => {
                  if (confirm(`Delete ${selectedRowKeys.length} items?`)) {
                    selectedRowKeys.forEach((key) =>
                      deleteMutation.mutate(String(key)),
                    );
                    setSelectedRowKeys([]);
                  }
                }}
              >
                Delete Selected
              </Button>
            </Space>
          </div>
        )}

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
          renderGridView()
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
      </Card>
    </div>
  );
}
