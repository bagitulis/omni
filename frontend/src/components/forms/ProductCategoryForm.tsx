import { Form, Select, Input, Button, Typography, theme } from "antd";
import { useState, useEffect } from "react";

export interface CategoryFormValues {
  category: string;
  attributes: Record<string, string>;
}

interface Attribute {
  id: string;
  name: string;
  required: boolean;
  options?: string[];
}

const CATEGORIES = [
  { label: "Electronics", value: "electronics" },
  { label: "Clothing", value: "clothing" },
  { label: "Home & Garden", value: "home_garden" },
];

const CATEGORY_ATTRIBUTES: Record<string, Attribute[]> = {
  electronics: [
    { id: "screen_size", name: "Screen Size", required: true },
    { id: "processor", name: "Processor", required: false },
    {
      id: "warranty",
      name: "Warranty Period",
      required: true,
      options: ["1 Year", "2 Years", "Lifetime"],
    },
  ],
  clothing: [
    {
      id: "material",
      name: "Material",
      required: true,
      options: ["Cotton", "Polyester", "Silk", "Wool"],
    },
    {
      id: "season",
      name: "Season",
      required: false,
      options: ["Summer", "Winter", "All Season"],
    },
    {
      id: "gender",
      name: "Gender",
      required: true,
      options: ["Men", "Women", "Unisex", "Kids"],
    },
  ],
  home_garden: [
    { id: "dimensions", name: "Dimensions", required: false },
    { id: "material", name: "Material", required: true },
  ],
};

interface Props {
  initialValues: CategoryFormValues;
  onFinish: (values: CategoryFormValues) => void;
  onBack: () => void;
}

export function ProductCategoryForm({
  initialValues,
  onFinish,
  onBack,
}: Props) {
  const { token } = theme.useToken();
  const [form] = Form.useForm<CategoryFormValues>();
  const [selectedCategory, setSelectedCategory] = useState<string | undefined>(
    initialValues.category,
  );

  useEffect(() => {
    form.setFieldsValue(initialValues);
    setSelectedCategory(initialValues.category);
  }, [initialValues, form]);

  const handleCategoryChange = (value: string) => {
    setSelectedCategory(value);
    // Reset attributes when category changes but keep other fields if any
    const currentValues = form.getFieldsValue();
    form.setFieldsValue({
      ...currentValues,
      attributes: {},
    });
  };

  const currentAttributes = selectedCategory
    ? CATEGORY_ATTRIBUTES[selectedCategory] || []
    : [];

  return (
    <Form
      form={form}
      layout="vertical"
      onFinish={onFinish}
      initialValues={initialValues}
      style={{ maxWidth: 600, margin: "0 auto" }}
    >
      <Form.Item
        label={<Typography.Text strong>Category</Typography.Text>}
        name="category"
        rules={[{ required: true, message: "Please select a category" }]}
      >
        <Select
          options={CATEGORIES}
          onChange={handleCategoryChange}
          placeholder="Select category"
        />
      </Form.Item>

      {selectedCategory && (
        <div
          style={{
            background: token.colorBgLayout,
            padding: 24,
            borderRadius: token.borderRadiusLG,
            marginBottom: 24,
            border: `1px solid ${token.colorBorderSecondary}`,
          }}
        >
          <h4
            style={{
              margin: "0 0 16px",
              fontSize: 14,
              fontWeight: 600,
              color: token.colorText,
            }}
          >
            Category Attributes
          </h4>

          {currentAttributes.length === 0 ? (
            <div
              style={{
                color: token.colorTextSecondary,
                fontStyle: "italic",
              }}
            >
              No specific attributes for this category.
            </div>
          ) : (
            currentAttributes.map((attr) => (
              <Form.Item
                key={attr.id}
                label={attr.name}
                name={["attributes", attr.id]}
                rules={[
                  {
                    required: attr.required,
                    message: `${attr.name} is required`,
                  },
                ]}
              >
                {attr.options ? (
                  <Select placeholder={`Select ${attr.name}`}>
                    {attr.options.map((opt) => (
                      <Select.Option key={opt} value={opt}>
                        {opt}
                      </Select.Option>
                    ))}
                  </Select>
                ) : (
                  <Input placeholder={`Enter ${attr.name}`} />
                )}
              </Form.Item>
            ))
          )}
        </div>
      )}

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
