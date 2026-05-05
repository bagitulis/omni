import { Form, Input, Button, Upload, InputNumber, Table, theme } from "antd";
import { PlusOutlined, DeleteOutlined } from "@ant-design/icons";
import { useEffect, useState } from "react";
import type { UploadFile } from "antd/es/upload/interface";

export interface MediaFormValues {
  images: UploadFile[];
  skus: Array<{
    seller_sku: string;
    variant_name: string;
    stock: number;
    price: number;
  }>;
}

interface Props {
  initialValues: MediaFormValues;
  onFinish: (values: MediaFormValues) => void;
  onBack: () => void;
}

export function ProductMediaForm({ initialValues, onFinish, onBack }: Props) {
  const { token } = theme.useToken();
  const [form] = Form.useForm<MediaFormValues>();
  const [fileList, setFileList] = useState<UploadFile[]>(
    initialValues.images || [],
  );

  useEffect(() => {
    form.setFieldsValue(initialValues);
    if (initialValues.images) {
      setFileList(initialValues.images);
    }
  }, [initialValues, form]);

  const handleUploadChange = ({
    fileList: newFileList,
  }: {
    fileList: UploadFile[];
  }) => {
    setFileList(newFileList);
    // Determine which files are valid to keep in form state
    form.setFieldValue("images", newFileList);
  };

  const normFile = (e: { fileList: UploadFile[] } | UploadFile[]) => {
    if (Array.isArray(e)) {
      return e;
    }
    return e?.fileList;
  };

  return (
    <Form
      form={form}
      layout="vertical"
      onFinish={onFinish}
      initialValues={initialValues}
      style={{ maxWidth: 800, margin: "0 auto" }}
    >
      <div
        style={{
          background: token.colorBgContainer,
          padding: 24,
          borderRadius: 6,
          marginBottom: 24,
          border: `1px solid ${token.colorBorderSecondary}`,
        }}
      >
        <h3 style={{ fontSize: 16, fontWeight: 600, marginBottom: 16 }}>
          Product Images
        </h3>
        <Form.Item
          name="images"
          valuePropName="fileList"
          getValueFromEvent={normFile}
          rules={[
            { required: true, message: "Please upload at least one image" },
          ]}
        >
          <Upload
            listType="picture-card"
            fileList={fileList}
            onChange={handleUploadChange}
            beforeUpload={() => false} // Prevent auto upload, just mock
            maxCount={8}
            accept="image/*"
          >
            {fileList.length < 8 && (
              <div>
                <PlusOutlined />
                <div style={{ marginTop: 8 }}>Upload</div>
              </div>
            )}
          </Upload>
        </Form.Item>
      </div>

      <div
        style={{
          background: token.colorBgContainer,
          padding: 24,
          borderRadius: 6,
          marginBottom: 24,
          border: `1px solid ${token.colorBorderSecondary}`,
        }}
      >
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: 16,
          }}
        >
          <h3 style={{ fontSize: 16, fontWeight: 600, margin: 0 }}>
            SKU Variants
          </h3>
        </div>

        <Form.List
          name="skus"
          initialValue={[
            { seller_sku: "", variant_name: "Default", stock: 0, price: 0 },
          ]}
        >
          {(fields, { add, remove }) => (
            <>
              <Table
                dataSource={fields}
                pagination={false}
                rowKey="key"
                size="small"
                columns={[
                  {
                    title: "Variant Name",
                    dataIndex: "variant_name",
                    key: "variant_name",
                    render: (_, field) => (
                      <Form.Item
                        {...field}
                        name={[field.name, "variant_name"]}
                        noStyle
                        rules={[{ required: true, message: "Required" }]}
                      >
                        <Input placeholder="e.g. Red, XL" />
                      </Form.Item>
                    ),
                  },
                  {
                    title: "Seller SKU",
                    dataIndex: "seller_sku",
                    key: "seller_sku",
                    render: (_, field) => (
                      <Form.Item
                        {...field}
                        name={[field.name, "seller_sku"]}
                        noStyle
                        rules={[{ required: true, message: "Required" }]}
                      >
                        <Input placeholder="SKU-123" />
                      </Form.Item>
                    ),
                  },
                  {
                    title: "Stock",
                    dataIndex: "stock",
                    key: "stock",
                    width: 120,
                    render: (_, field) => (
                      <Form.Item
                        {...field}
                        name={[field.name, "stock"]}
                        noStyle
                        rules={[{ required: true, message: "Required" }]}
                      >
                        <InputNumber
                          min={0}
                          placeholder="0"
                          style={{ width: "100%" }}
                        />
                      </Form.Item>
                    ),
                  },
                  {
                    title: "Price",
                    dataIndex: "price",
                    key: "price",
                    width: 150,
                    render: (_, field) => (
                      <Form.Item
                        {...field}
                        name={[field.name, "price"]}
                        noStyle
                        rules={[{ required: true, message: "Required" }]}
                      >
                        <InputNumber<number>
                          min={0}
                          placeholder="Rp"
                          style={{ width: "100%" }}
                          formatter={(value) =>
                            `Rp ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ",")
                          }
                          parser={(value) => {
                            const val = value?.replace(/\D/g, "");
                            return val ? parseInt(val, 10) : 0;
                          }}
                        />
                      </Form.Item>
                    ),
                  },
                  {
                    title: "",
                    key: "action",
                    width: 50,
                    render: (_, field) =>
                      fields.length > 1 ? (
                        <Button
                          type="text"
                          danger
                          icon={<DeleteOutlined />}
                          onClick={() => remove(field.name)}
                        />
                      ) : null,
                  },
                ]}
              />
              <Button
                type="dashed"
                onClick={() => add()}
                block
                icon={<PlusOutlined />}
                style={{ marginTop: 16 }}
              >
                Add Variant
              </Button>
            </>
          )}
        </Form.List>
      </div>

      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          marginTop: 24,
        }}
      >
        <Button onClick={onBack}>Back</Button>
        <Button type="primary" htmlType="submit">
          Next Step
        </Button>
      </div>
    </Form>
  );
}
