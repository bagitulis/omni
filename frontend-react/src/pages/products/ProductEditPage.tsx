import { useState, useEffect } from "react";
import { useParams, Link } from "react-router-dom";
import {
  Tabs,
  Button,
  message,
  Spin,
  Table,
  Input,
  InputNumber,
  Upload,
  Card,
  Badge,
  Alert,
} from "antd";
import {
  ArrowLeftOutlined,
  SaveOutlined,
  PlusOutlined,
  DeleteOutlined,
} from "@ant-design/icons";
import { ProductBasicForm } from "../../components/forms/ProductBasicForm";
import { getProductById } from "../../api/products";
import type { UploadFile } from "antd/es/upload/interface";

interface ProductData {
  id: number;
  title: string;
  description: string;
  images: string[];
  status: string;
  skus?: Array<{
    key: string;
    seller_sku: string;
    variant_name: string;
    stock: number;
    price: number;
  }>;
  platforms?: Array<{
    platform: string;
    status: string;
    last_sync: string;
  }>;
}

const VariantsTab = ({ initialValues }: { initialValues: any[] }) => {
  const [dataSource, setDataSource] = useState(initialValues);

  const columns = [
    {
      title: "Variant Name",
      dataIndex: "variant_name",
      render: (text: string) => <Input defaultValue={text} />,
    },
    {
      title: "Seller SKU",
      dataIndex: "seller_sku",
      render: (text: string) => <Input defaultValue={text} />,
    },
    {
      title: "Stock",
      dataIndex: "stock",
      render: (val: number) => <InputNumber defaultValue={val} />,
    },
    {
      title: "Price",
      dataIndex: "price",
      render: (val: number) => (
        <InputNumber
          defaultValue={val}
          formatter={(value) =>
            `Rp ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ",")
          }
          parser={(value) => {
            const v = value?.replace(/\D/g, "");
            return v ? parseInt(v, 10) : 0;
          }}
          style={{ width: "100%" }}
        />
      ),
    },
    {
      title: "Action",
      render: (_: any, record: any) => (
        <Button
          type="text"
          danger
          icon={<DeleteOutlined />}
          onClick={() =>
            setDataSource(dataSource.filter((item) => item.key !== record.key))
          }
        />
      ),
    },
  ];

  return (
    <div>
      <div className="mb-4 flex justify-end">
        <Button
          type="dashed"
          icon={<PlusOutlined />}
          onClick={() =>
            setDataSource([
              ...dataSource,
              {
                key: `${Date.now()}`,
                variant_name: "",
                seller_sku: "",
                stock: 0,
                price: 0,
              },
            ])
          }
        >
          Add Variant
        </Button>
      </div>
      <Table
        dataSource={dataSource}
        columns={columns}
        pagination={false}
        size="small"
      />
      <div className="mt-4 flex justify-end">
        <Button type="primary" icon={<SaveOutlined />}>
          Save Variants
        </Button>
      </div>
    </div>
  );
};

const ImagesTab = ({ initialValues }: { initialValues: UploadFile[] }) => {
  const [fileList, setFileList] = useState<UploadFile[]>(initialValues);
  return (
    <div>
      <Upload
        listType="picture-card"
        fileList={fileList}
        onChange={({ fileList: newFileList }) => setFileList(newFileList)}
        maxCount={8}
        beforeUpload={() => false}
      >
        {fileList.length < 8 && (
          <div>
            <PlusOutlined />
            <div style={{ marginTop: 8 }}>Upload</div>
          </div>
        )}
      </Upload>
      <div className="mt-4 flex justify-end">
        <Button type="primary" icon={<SaveOutlined />}>
          Save Images
        </Button>
      </div>
    </div>
  );
};

const PlatformSyncTab = ({ platforms }: { platforms: any[] }) => {
  const columns = [
    { title: "Platform", dataIndex: "platform", key: "platform" },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      render: (status: string) => {
        const color =
          status === "synced"
            ? "success"
            : status === "pending"
              ? "warning"
              : "error";
        return <Badge status={color as any} text={status.toUpperCase()} />;
      },
    },
    { title: "Last Sync", dataIndex: "last_sync", key: "last_sync" },
    {
      title: "Action",
      key: "action",
      render: () => <Button size="small">Sync Now</Button>,
    },
  ];
  return <Table dataSource={platforms} columns={columns} pagination={false} />;
};

export default function ProductEditPage() {
  const { id } = useParams();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [product, setProduct] = useState<ProductData | null>(null);

  useEffect(() => {
    if (!id) return;

    async function fetchProduct() {
      try {
        setLoading(true);
        setError(null);
        const data = await getProductById(id!);
        setProduct({
          id: data.id,
          title: data.title,
          description: data.description,
          images: data.images || [],
          status: data.status,
          skus: [],
          platforms: [],
        });
      } catch (err: any) {
        setError(err.message || "Failed to load product");
      } finally {
        setLoading(false);
      }
    }

    fetchProduct();
  }, [id]);

  if (loading) {
    return (
      <div className="p-12 text-center">
        <Spin size="large" />
        <div className="mt-4">Loading product...</div>
      </div>
    );
  }

  if (error || !product) {
    return (
      <div className="p-6">
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
      children: <VariantsTab initialValues={product.skus || []} />,
    },
    {
      key: "3",
      label: "Images",
      children: <ImagesTab initialValues={imageFiles} />,
    },
    {
      key: "4",
      label: "Platform Sync",
      children: <PlatformSyncTab platforms={product.platforms || []} />,
    },
  ];

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <div className="mb-6 flex items-center gap-4">
        <Link
          to="/master-products"
          className="text-gray-500 hover:text-blue-600"
        >
          <ArrowLeftOutlined style={{ fontSize: 18 }} />
        </Link>
        <h1 className="text-2xl font-bold m-0">
          Edit Product: {product.title}
        </h1>
      </div>
      <Card>
        <Tabs defaultActiveKey="1" items={items} type="card" />
      </Card>
      <div className="mt-4 text-xs text-gray-400">Product ID: {id}</div>
    </div>
  );
}
