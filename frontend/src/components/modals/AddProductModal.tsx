import { useState } from "react";
import {
  Modal,
  Steps,
  Button,
  Form,
  Input,
  Select,
  message,
  Typography,
  Upload,
  Space,
} from "antd";
import { PlusOutlined } from "@ant-design/icons";

const { TextArea } = Input;
const { Text } = Typography;

interface AddProductModalProps {
  open: boolean;
  onClose: () => void;
  platform: "shopee" | "tiktok" | "lazada";
}

export function AddProductModal({
  open,
  onClose,
  platform,
}: AddProductModalProps) {
  const [currentStep, setCurrentStep] = useState(0);
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);

  // Platform specific styles
  const getPlatformColor = () => {
    switch (platform) {
      case "shopee":
        return "#ee4d2d";
      case "tiktok":
        return "#000000";
      case "lazada":
        return "#0f146d";
      default:
        return "#1890ff";
    }
  };

  const steps = [
    { title: "Basic Info", content: "basic" },
    { title: "Attributes", content: "attributes" },
    { title: "Images & SKU", content: "images" },
    { title: "Shipping", content: "shipping" },
  ];

  const next = async () => {
    try {
      await form.validateFields();
      setCurrentStep(currentStep + 1);
    } catch (error) {
      // Form validation failed
    }
  };

  const prev = () => {
    setCurrentStep(currentStep - 1);
  };

  const handleSubmit = async () => {
    try {
      setLoading(true);
      const values = await form.validateFields();
      console.log("Submitting:", values);
      // TODO: Implement actual submission logic using platform-specific API
      message.success("Product created successfully (Mock)");
      onClose();
    } catch (error) {
      console.error("Submission failed:", error);
    } finally {
      setLoading(false);
    }
  };

  const renderBasicInfo = () => (
    <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
      <Form.Item
        name="title"
        label="Product Title"
        rules={[{ required: true, message: "Please enter product title" }]}
      >
        <Input placeholder="Enter product title" />
      </Form.Item>
      <Form.Item
        name="description"
        label="Description"
        rules={[{ required: true, message: "Please enter description" }]}
      >
        <TextArea rows={4} placeholder="Enter product description" />
      </Form.Item>
      <Form.Item
        name="categoryId"
        label="Category"
        rules={[{ required: true, message: "Please select category" }]}
      >
        <Select placeholder="Select category">
          <Select.Option value="1">Category 1</Select.Option>
          <Select.Option value="2">Category 2</Select.Option>
        </Select>
      </Form.Item>
      <Form.Item name="brandId" label="Brand">
        <Select placeholder="Select brand (Optional)">
          <Select.Option value="1">Brand A</Select.Option>
          <Select.Option value="2">Brand B</Select.Option>
        </Select>
      </Form.Item>
    </div>
  );

  const renderAttributes = () => (
    <div>
      <Text type="secondary">
        Category attributes will appear here after selecting a category.
      </Text>
      {/* Placeholder for dynamic attributes */}
      <Form.Item name={["attributes", "material"]} label="Material">
        <Input placeholder="e.g. Cotton" />
      </Form.Item>
    </div>
  );

  const renderImagesSku = () => (
    <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
      <Form.Item label="Product Images">
        <Upload listType="picture-card">
          <div>
            <PlusOutlined />
            <div style={{ marginTop: 8 }}>Upload</div>
          </div>
        </Upload>
      </Form.Item>
      <Form.Item name="skus" label="SKUs">
        {/* Placeholder for SKU builder */}
        <Button>Manage Variations</Button>
      </Form.Item>
    </div>
  );

  const renderShipping = () => (
    <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
      <Form.Item
        name="weight"
        label="Weight (kg)"
        rules={[{ required: true, message: "Please enter weight" }]}
      >
        <Input type="number" suffix="kg" />
      </Form.Item>
      <Space>
        <Form.Item name="length" label="Length (cm)">
          <Input type="number" />
        </Form.Item>
        <Form.Item name="width" label="Width (cm)">
          <Input type="number" />
        </Form.Item>
        <Form.Item name="height" label="Height (cm)">
          <Input type="number" />
        </Form.Item>
      </Space>
    </div>
  );

  const renderContent = () => {
    switch (steps[currentStep].content) {
      case "basic":
        return renderBasicInfo();
      case "attributes":
        return renderAttributes();
      case "images":
        return renderImagesSku();
      case "shipping":
        return renderShipping();
      default:
        return null;
    }
  };

  return (
    <Modal
      title={`Add Product to ${platform.charAt(0).toUpperCase() + platform.slice(1)}`}
      open={open}
      onCancel={onClose}
      width={800}
      footer={
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            marginTop: 24,
          }}
        >
          {currentStep > 0 && <Button onClick={prev}>Previous</Button>}
          <div style={{ flex: 1 }} />
          {currentStep < steps.length - 1 && (
            <Button type="primary" onClick={next}>
              Next
            </Button>
          )}
          {currentStep === steps.length - 1 && (
            <Button
              type="primary"
              onClick={handleSubmit}
              loading={loading}
              style={{ backgroundColor: getPlatformColor() }}
            >
              Create Product
            </Button>
          )}
        </div>
      }
    >
      <Steps current={currentStep} items={steps} style={{ marginBottom: 24 }} />
      <Form form={form} layout="vertical">
        <div style={{ minHeight: 300 }}>{renderContent()}</div>
      </Form>
    </Modal>
  );
}
