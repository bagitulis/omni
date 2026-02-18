import { useState, useEffect, useCallback } from "react";
import { useParams, Link } from "react-router-dom";
import {
  Tabs,
  Button,
  message,
  Spin,
  Card,
  Alert,
  Flex,
  Typography,
} from "antd";
import { ArrowLeftOutlined } from "@ant-design/icons";

import { ProductBasicForm } from "../../components/forms/ProductBasicForm";
import { getProductById, updateProduct, syncProduct } from "../../api/products";
import type { UploadFile } from "antd/es/upload/interface";
import { ProductVariantsTab } from "./components/ProductVariantsTab";
import { ProductImagesTab } from "./components/ProductImagesTab";
import { ProductSyncTab } from "./components/ProductSyncTab";
import { SkuMappingPanel } from "../../components/shared/SkuMappingPanel";
import type { MasterProduct } from "../../types/product";
import type { ProductData, ProductSku } from "./types";

const { Title, Text } = Typography;

export default function ProductEditPage() {
  const { id } = useParams();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [product, setProduct] = useState<ProductData | null>(null);
  const [rawProduct, setRawProduct] = useState<MasterProduct | null>(null);
  const [saveVariantsLoading, setSaveVariantsLoading] = useState(false);
  const [saveImagesLoading, setSaveImagesLoading] = useState(false);
  const [syncLoading, setSyncLoading] = useState(false);

  const fetchProduct = useCallback(async () => {
    if (!id) return;

    try {
      setLoading(true);
      setError(null);
      const data = await getProductById(id);
      setRawProduct(data);
      setProduct({
        id: data.id,
        title: data.title,
        description: data.description,
        images: data.images || [],
        status: data.status,
        skus: [],
        platforms: [],
      });
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to load product");
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    fetchProduct();
  }, [fetchProduct]);

  const handleSaveVariants = async (variants: ProductSku[]) => {
    if (!id) return;

    setSaveVariantsLoading(true);
    try {
      // Prepare update data - include variants if backend supports it
      const updateData = {
        title: product?.title,
        description: product?.description,
        images: product?.images,
      };
      // Spread variants data for future backend compatibility
      Object.assign(updateData, variants.length > 0 ? { variants } : {});

      await updateProduct(id, updateData);
      message.success("Variants saved successfully!");
    } catch (err) {
      message.error((err as Error).message || "Failed to save variants");
    } finally {
      setSaveVariantsLoading(false);
    }
  };

  const handleSaveImages = async (files: UploadFile[]) => {
    if (!id) return;

    setSaveImagesLoading(true);
    try {
      const imageUrls = files
        .map((file) => file.url || file.response?.url)
        .filter(Boolean);
      await updateProduct(id, { images: imageUrls });
      message.success("Images saved successfully!");
    } catch (err) {
      message.error((err as Error).message || "Failed to save images");
    } finally {
      setSaveImagesLoading(false);
    }
  };

  const handleSyncProduct = async (platform: string) => {
    if (!id) return;

    setSyncLoading(true);
    try {
      await syncProduct(id);
      message.success(`Product synced to ${platform} successfully!`);
    } catch (err) {
      message.error((err as Error).message || "Failed to sync product");
    } finally {
      setSyncLoading(false);
    }
  };

  if (loading) {
    return (
      <Flex vertical align="center" justify="center" style={{ padding: 48 }}>
        <Spin size="large" />
        <Text style={{ marginTop: 16 }}>Loading product...</Text>
      </Flex>
    );
  }

  if (error || !product) {
    return (
      <div style={{ padding: 24 }}>
        <Alert
          type="error"
          message="Failed to load product"
          description={error || "Product not found"}
          action={
            <Link to="/master-products">
              <Button>Back to Products</Button>
            </Link>
          }
        />
      </div>
    );
  }

  const imageFiles: UploadFile[] = product.images.map((url, idx) => ({
    uid: `-${idx}`,
    name: `image-${idx}.png`,
    status: "done",
    url,
  }));

  const items = [
    {
      key: "1",
      label: "Basic Info",
      children: (
        <ProductBasicForm
          initialValues={{
            item_name: product.title,
            description: product.description,
          }}
          onFinish={() => message.success("Saved")}
          submitLabel="Save Basic Info"
        />
      ),
    },
    {
      key: "2",
      label: "Variants",
      children: (
        <ProductVariantsTab
          initialValues={product.skus || []}
          onSave={handleSaveVariants}
          loading={saveVariantsLoading}
        />
      ),
    },
    {
      key: "3",
      label: "Images",
      children: (
        <ProductImagesTab
          initialValues={imageFiles}
          onSave={handleSaveImages}
          loading={saveImagesLoading}
        />
      ),
    },
    {
      key: "4",
      label: "SKU Mapping",
      children: rawProduct ? (
        <SkuMappingPanel masterProduct={rawProduct} onUpdate={fetchProduct} />
      ) : null,
    },
    {
      key: "5",
      label: "Platform Sync",
      children: (
        <ProductSyncTab
          platforms={product.platforms || []}
          onSync={handleSyncProduct}
          loading={syncLoading}
        />
      ),
    },
  ];

  return (
    <div style={{ padding: 24, maxWidth: 1024, margin: "0 auto" }}>
      <Flex align="center" gap={16} style={{ marginBottom: 24 }}>
        <Link to="/master-products">
          <Button type="text" icon={<ArrowLeftOutlined />} />
        </Link>
        <Title level={4} style={{ margin: 0 }}>
          Edit Product: {product.title}
        </Title>
      </Flex>
      <Card>
        <Tabs defaultActiveKey="1" items={items} type="card" />
      </Card>
      <Text
        type="secondary"
        style={{ fontSize: 12, marginTop: 16, display: "block" }}
      >
        Product ID: {id}
      </Text>
    </div>
  );
}
