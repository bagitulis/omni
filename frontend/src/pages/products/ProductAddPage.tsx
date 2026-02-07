import { useState } from "react";
import { Steps, message, Card } from "antd";
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
    <div className="p-6 max-w-5xl mx-auto">
      <div className="mb-6 flex items-center gap-4">
        <Link
          to="/master-products"
          className="text-gray-500 hover:text-blue-600"
        >
          <ArrowLeftOutlined style={{ fontSize: 18 }} />
        </Link>
        <h1 className="text-2xl font-bold m-0">Add New Product</h1>
      </div>

      <Card>
        <Steps
          current={currentStep}
          items={steps.map((item) => ({ title: item.title }))}
          className="mb-8 max-w-3xl mx-auto"
        />

        <div className="mt-8">{steps[currentStep].content}</div>
      </Card>
    </div>
  );
}
