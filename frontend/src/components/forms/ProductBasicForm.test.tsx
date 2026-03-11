import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { ProductBasicForm } from "@/components/forms/ProductBasicForm";

// Mock antd
vi.mock("antd", async () => {
  const formInstance = {
    setFieldsValue: vi.fn(),
  };

  const FormComponent = ({
    children,
    onFinish,
    initialValues,
  }: {
    children?: React.ReactNode;
    onFinish?: (values: Record<string, unknown>) => void;
    initialValues?: Record<string, unknown>;
  }) => (
    <form
      data-testid="form"
      onSubmit={(e) => {
        e.preventDefault();
        onFinish?.(initialValues ?? {});
      }}
    >
      {children}
    </form>
  );

  FormComponent.Item = ({
    children,
    label,
    name,
  }: {
    children?: React.ReactNode;
    label?: React.ReactNode;
    name?: string;
  }) => (
    <label>
      {label}
      <div data-testid={`form-item-${name}`}>{children}</div>
    </label>
  );

  FormComponent.useForm = () => [formInstance];

  return {
    Form: FormComponent,
    Input: Object.assign(
      ({ value, placeholder }: { value?: string; placeholder?: string }) => (
        <input value={value} placeholder={placeholder} readOnly />
      ),
      {
        TextArea: ({
          value,
          placeholder,
        }: {
          value?: string;
          placeholder?: string;
        }) => <textarea value={value} placeholder={placeholder} readOnly />,
      },
    ),
    Button: ({
      children,
      htmlType,
    }: {
      children?: React.ReactNode;
      htmlType?: "button" | "submit" | "reset";
    }) => <button type={htmlType ?? "button"}>{children}</button>,
    Typography: {
      Text: ({ children }: { children?: React.ReactNode }) => (
        <span>{children}</span>
      ),
    },
  };
});

describe("ProductBasicForm", () => {
  const mockOnFinish = vi.fn();
  const initialValues = {
    item_name: "Test Product",
    description: "Test Description",
    brand: "Test Brand",
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  const renderForm = (props = {}) => {
    return render(
      <ProductBasicForm
        initialValues={initialValues}
        onFinish={mockOnFinish}
        {...props}
      />,
    );
  };

  it("renders form fields correctly", () => {
    renderForm();
    expect(screen.getByText("Product Name")).toBeInTheDocument();
    expect(screen.getByText("Description")).toBeInTheDocument();
    expect(screen.getByText("Brand")).toBeInTheDocument();
    expect(
      screen.getByPlaceholderText("Ex: Samsung Galaxy S24 Ultra"),
    ).toBeInTheDocument();
    expect(
      screen.getByPlaceholderText("Product details, specifications, etc."),
    ).toBeInTheDocument();
  });

  it("calls onFinish when submitted", async () => {
    renderForm();
    const submitBtn = screen.getByText("Next Step");
    fireEvent.click(submitBtn);

    // In our simplified mock, clicking button submits form which calls onFinish
    // But button type="submit" needs to trigger form onSubmit
    // FireEvent.submit might be safer for our mock
    fireEvent.submit(screen.getByTestId("form"));

    expect(mockOnFinish).toHaveBeenCalledWith(initialValues);
  });

  it("hides submit button when hideSubmit is true", () => {
    renderForm({ hideSubmit: true });
    expect(screen.queryByText("Next Step")).not.toBeInTheDocument();
  });

  it("renders custom submit label", () => {
    renderForm({ submitLabel: "Save" });
    expect(screen.getByText("Save")).toBeInTheDocument();
  });
});
