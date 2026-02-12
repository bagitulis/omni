import { useState } from "react";
import { Steps, message, Card, Typography, theme, Flex } from "antd";
import type { UploadFile } from "antd/es/upload/interface";
import { useNavigate, Link } from "react-router-dom";
import { ArrowLeftOutlined } from "@ant-design/icons";
import {
  ProductBasicForm,
  BasicFormValues,
} from "../../components/forms/ProductBasicForm";
import {
  ProductCategoryForm,
  CategoryFormValues,
} from "../../components/forms/ProductCategoryForm";
import {
  ProductMediaForm,
  MediaFormValues,
} from "../../components/forms/ProductMediaForm";
import {
  ProductPricingForm,
  PricingFormValues,
} from "../../components/forms/ProductPricingForm";
import { createProduct } from "@/api/products";

// Combined type for the full product form
type ProductFormValues = BasicFormValues &
  CategoryFormValues &
  MediaFormValues &
  PricingFormValues;

const INITIAL_VALUES: ProductFormValues = {
  // Basic
  item_name: "",
  description: "",
  brand: "",
  // Category
  category: "",
  attributes: {},
  // Media
  images: [],
  skus: [{ seller_sku: "", variant_name: "Default", stock: 0, price: 0 }],
  // Pricing/Shipping
  weight: 0,
  length: 0,
  width: 0,
  height: 0,
};

export default function ProductAddPage() {
  const navigate = useNavigate();
  const { token } = theme.useToken();
  const [currentStep, setCurrentStep] = useState(0);
  const [formData, setFormData] = useState<ProductFormValues>(INITIAL_VALUES);
  const [submitting, setSubmitting] = useState(false);

  const handleNext = (values: Partial<ProductFormValues>) => {
    setFormData((prev) => ({ ...prev, ...values }));
    setCurrentStep((prev) => prev + 1);
    window.scrollTo(0, 0);
  };

  const handleBack = () => {
    setCurrentStep((prev) => prev - 1);
    window.scrollTo(0, 0);
  };

  const handleSubmit = async (values: Partial<ProductFormValues>) => {
    const finalData = { ...formData, ...values };
    setSubmitting(true);

    try {
      // Transform form data to API format
      const apiData = {
        ...finalData,
        title: finalData.item_name,
        // Convert UploadFile[] to string[] (use response.url or name as fallback)
        images: (finalData.images || []).map(
          (file: UploadFile) =>
            file.response?.url || file.url || file.name || "",
        ),
      };

      await createProduct(apiData);
      message.success("Product created successfully!");
      navigate("/master-products");
    } catch (error) {
      message.error((error as Error).message || "Failed to create product");
    } finally {
      setSubmitting(false);
    }
  };

  const steps = [
    {
      title: "Basic Info",
      content: (
        <ProductBasicForm initialValues={formData} onFinish={handleNext} />
      ),
    },
    {
      title: "Category",
      content: (
        <ProductCategoryForm
          initialValues={formData}
          onFinish={handleNext}
          onBack={handleBack}
        />
      ),
    },
    {
      title: "Media & SKU",
      content: (
        <ProductMediaForm
          initialValues={formData}
          onFinish={handleNext}
          onBack={handleBack}
        />
      ),
    },
    {
      title: "Shipping",
      content: (
        <ProductPricingForm
          initialValues={formData}
          onFinish={handleSubmit}
          onBack={handleBack}
          submitting={submitting}
        />
      ),
    },
  ];

  return (
    <div style={{ padding: 24, maxWidth: 1024, margin: "0 auto" }}>
      <Flex align="center" gap={16} style={{ marginBottom: 24 }}>
        <Link
          to="/master-products"
          style={{
            color: token.colorTextSecondary,
            fontSize: 18,
            lineHeight: 1,
          }}
        >
          <ArrowLeftOutlined />
        </Link>
        <Typography.Title level={2} style={{ margin: 0 }}>
          Add New Product
        </Typography.Title>
      </Flex>

      <Card>
        <Steps
          current={currentStep}
          items={steps.map((item) => ({ title: item.title }))}
          style={{ marginBottom: 32, maxWidth: 768, marginInline: "auto" }}
        />

        <div style={{ marginTop: 32 }}>{steps[currentStep].content}</div>
      </Card>
    </div>
  );
}
