import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { ProductBasicForm } from "@/components/forms/ProductBasicForm";

// Mock antd
vi.mock("antd", async () => {
  const actual = await vi.importActual<typeof import("antd")>("antd");
  return {
    ...actual,
    Form: ({
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
    ),
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

    // Check input values (antd Input renders real inputs usually, but our mock might affect context)
    // Since we didn't mock Input fully, it might render real input.
    // However, antd Form.Item manages value.
    // Let's just check rendering for now.
    expect(screen.getByDisplayValue("Test Product")).toBeInTheDocument();
    expect(screen.getByDisplayValue("Test Description")).toBeInTheDocument();
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
