import { useState } from "react";
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
      render: (_: any, record: ProductSku) => (
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
      <Flex justify="flex-end" style={{ marginBottom: 16 }}>
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
