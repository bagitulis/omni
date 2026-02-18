import { useEffect, useState } from "react";
import { Table, Input, InputNumber, Button, Flex } from "antd";
import { DeleteOutlined, PlusOutlined, SaveOutlined } from "@ant-design/icons";
import type { ProductSku } from "../types";

interface ProductVariantsTabProps {
  initialValues: ProductSku[];
  onSave: (data: ProductSku[]) => void;
  loading: boolean;
}

export const ProductVariantsTab = ({
  initialValues,
  onSave,
  loading,
}: ProductVariantsTabProps) => {
  const [dataSource, setDataSource] = useState<ProductSku[]>(initialValues);

  useEffect(() => {
    setDataSource(initialValues);
  }, [initialValues]);

  const updateSkuField = <K extends keyof ProductSku>(
    key: string,
    field: K,
    value: ProductSku[K],
  ) => {
    setDataSource((prev) =>
      prev.map((item) =>
        item.key === key ? { ...item, [field]: value } : item,
      ),
    );
  };

  const columns = [
    {
      title: "Variant Name",
      dataIndex: "variant_name",
      render: (_: string, record: ProductSku) => (
        <Input
          value={record.variant_name}
          onChange={(event) =>
            updateSkuField(record.key, "variant_name", event.target.value)
          }
        />
      ),
    },
    {
      title: "Seller SKU",
      dataIndex: "seller_sku",
      render: (_: string, record: ProductSku) => (
        <Input
          value={record.seller_sku}
          onChange={(event) =>
            updateSkuField(record.key, "seller_sku", event.target.value)
          }
        />
      ),
    },
    {
      title: "Stock",
      dataIndex: "stock",
      render: (_: number, record: ProductSku) => (
        <InputNumber
          value={record.stock}
          onChange={(value) =>
            updateSkuField(record.key, "stock", Number(value) || 0)
          }
        />
      ),
    },
    {
      title: "Price",
      dataIndex: "price",
      render: (_: number, record: ProductSku) => (
        <InputNumber
          value={record.price}
          formatter={(value) =>
            `Rp ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ",")
          }
          parser={(value) => {
            const v = value?.replace(/\D/g, "");
            return v ? parseInt(v, 10) : 0;
          }}
          onChange={(value) =>
            updateSkuField(record.key, "price", Number(value) || 0)
          }
          style={{ width: "100%" }}
        />
      ),
    },
    {
      title: "Action",
      render: (_: unknown, record: ProductSku) => (
        <Button
          type="text"
          danger
          icon={<DeleteOutlined />}
          onClick={() =>
            setDataSource((prev) =>
              prev.filter((item) => item.key !== record.key),
            )
          }
        />
      ),
    },
  ];

  return (
    <div>
      <Flex justify="flex-end" style={{ marginBottom: 16 }}>
        <Button
          type="dashed"
          icon={<PlusOutlined />}
          onClick={() =>
            setDataSource((prev) => [
              ...prev,
              {
                key: `new-${Date.now()}`,
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
      </Flex>
      <Table
        dataSource={dataSource}
        columns={columns}
        pagination={false}
        size="small"
      />
      <Flex justify="flex-end" style={{ marginTop: 16 }}>
        <Button
          type="primary"
          icon={<SaveOutlined />}
          onClick={() => onSave(dataSource)}
          loading={loading}
        >
          Save Variants
        </Button>
      </Flex>
    </div>
  );
};
