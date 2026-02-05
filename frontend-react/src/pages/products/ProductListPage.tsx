import { useState } from "react";
import { useNavigate } from "react-router-dom";
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
} from "@ant-design/icons";
import { ProductFilters } from "@/components/forms/ProductFilters";
import { ProductTable } from "@/components/tables/ProductTable";
import { useProducts, useDeleteProduct } from "@/hooks/useProducts";
import { PlatformBadge } from "@/components/ui/PlatformBadge";

export function ProductListPage() {
  const navigate = useNavigate();
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

  const { data, isLoading } = useProducts({
    page,
    limit: pageSize,
    status: filters.status,
    search: filters.search,
    platform: filters.platform,
  });

  const deleteMutation = useDeleteProduct();

  const handleFilterChange = (key: string, value: any) => {
    setFilters((prev) => ({ ...prev, [key]: value }));
    setPage(1); // Reset to first page
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
                          {new Intl.NumberFormat("id-ID", {
                            style: "currency",
                            currency: "IDR",
                            maximumFractionDigits: 0,
                          }).format(product.price)}
                        </Typography.Text>
                        <Typography.Text
                          type={
                            product.stock === 0
                              ? "danger"
                              : product.stock <= 10
                                ? "warning"
                                : "secondary"
                          }
                        >
                          Stock: {product.stock}
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
            onClick={() => navigate("/products/add")}
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
              <Button size="small">Bulk Edit Price</Button>
              <Button size="small">Bulk Edit Stock</Button>
              <Button size="small" danger>
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
          />
        ) : (
          renderGridView()
        )}
      </Card>
    </div>
  );
}
