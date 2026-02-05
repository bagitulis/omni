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
} from "antd";
import {
  ArrowLeftOutlined,
  SaveOutlined,
  PlusOutlined,
  DeleteOutlined,
} from "@ant-design/icons";
import { ProductBasicForm } from "../../components/forms/ProductBasicForm";
import type { UploadFile } from "antd/es/upload/interface";

// Mock Data
const MOCK_PRODUCT = {
  item_id: "1",
  item_name: "Samsung Galaxy S24 Ultra",
  description: "Experience the new era of mobile AI.",
  brand: "Samsung",
  category: "Electronics",
  price: 18999000,
  stock: 50,
  skus: [
    {
      key: "1",
      seller_sku: "S24U-BLK-256",
      variant_name: "Phantom Black, 256GB",
      stock: 20,
      price: 18999000,
    },
    {
      key: "2",
      seller_sku: "S24U-GRY-512",
      variant_name: "Titanium Gray, 512GB",
      stock: 15,
      price: 21999000,
    },
  ],
  images: [
    {
      uid: "-1",
      name: "image.png",
      status: "done",
      url: "https://zos.alipayobjects.com/rmsportal/jkjgkEfvpUPVyRjUImniVslZfWPnJuuZ.png",
    },
  ],
  platforms: [
    { platform: "Shopee", status: "synced", last_sync: "2023-10-25 10:00" },
    { platform: "Lazada", status: "pending", last_sync: "2023-10-24 15:30" },
    { platform: "TikTok", status: "failed", last_sync: "2023-10-25 09:00" },
  ],
};

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
            const val = value?.replace(/\D/g, "");
            return val ? parseInt(val, 10) : 0;
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

  useEffect(() => {
    // Mock fetch
    setTimeout(() => setLoading(false), 800);
  }, []);

  if (loading)
    return (
      <div className="p-12 text-center">
        <Spin size="large" />
      </div>
    );

  const items = [
    {
      key: "1",
      label: "Basic Info",
      children: (
        <ProductBasicForm
          initialValues={MOCK_PRODUCT}
          onFinish={() => message.success("Saved")}
          submitLabel="Save Basic Info"
        />
      ),
    },
    {
      key: "2",
      label: "Variants",
      children: <VariantsTab initialValues={MOCK_PRODUCT.skus} />,
    },
    {
      key: "3",
      label: "Images",
      children: (
        <ImagesTab initialValues={MOCK_PRODUCT.images as UploadFile[]} />
      ),
    },
    {
      key: "4",
      label: "Platform Sync",
      children: <PlatformSyncTab platforms={MOCK_PRODUCT.platforms} />,
    },
  ];

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <div className="mb-6 flex items-center gap-4">
        <Link to="/products" className="text-gray-500 hover:text-blue-600">
          <ArrowLeftOutlined style={{ fontSize: 18 }} />
        </Link>
        <h1 className="text-2xl font-bold m-0">
          Edit Product: {MOCK_PRODUCT.item_name}
        </h1>
      </div>
      <Card>
        <Tabs defaultActiveKey="1" items={items} type="card" />
      </Card>
      <div className="mt-4 text-xs text-gray-400">
        Product ID: {id} (Mock Data)
      </div>
    </div>
  );
}
