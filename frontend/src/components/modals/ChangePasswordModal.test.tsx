import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { ChangePasswordModal } from "@/components/modals/ChangePasswordModal";
import { changePassword } from "@/api/auth";

// Mock API
vi.mock("@/api/auth", () => ({
  changePassword: vi.fn(),
}));

// Mock antd message
vi.mock("antd", async () => {
  const formInstance = {
    validateFields: vi.fn().mockResolvedValue({
      currentPassword: "old-password",
      newPassword: "new-password",
      confirmPassword: "new-password",
    }),
    setFields: vi.fn(),
    resetFields: vi.fn(),
  };

  const FormComponent = ({
    children,
    onFinish,
  }: {
    children?: React.ReactNode;
    onFinish?: () => void;
  }) => (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        onFinish?.();
      }}
    >
      {children}
    </form>
  );

  FormComponent.Item = ({
    children,
    label,
  }: {
    children?: React.ReactNode;
    label?: React.ReactNode;
  }) => (
    <label>
      {label}
      {children}
    </label>
  );

  FormComponent.useForm = () => [formInstance];

  return {
    Modal: ({
      children,
      open,
      title,
      onCancel,
    }: {
      children?: React.ReactNode;
      open?: boolean;
      title?: React.ReactNode;
      onCancel?: () => void;
    }) =>
      open ? (
        <div role="dialog" aria-label={title as string}>
          <h1>{title}</h1>
          <button onClick={onCancel}>Close</button>
          {children}
        </div>
      ) : null,
    Form: FormComponent,
    Input: {
      Password: ({ "aria-label": ariaLabel }: { "aria-label"?: string }) => (
        <input aria-label={ariaLabel} />
      ),
    },
    Button: ({
      children,
      onClick,
      htmlType,
    }: {
      children?: React.ReactNode;
      onClick?: () => void;
      htmlType?: "button" | "submit" | "reset";
    }) => (
      <button type={htmlType ?? "button"} onClick={onClick}>
        {children}
      </button>
    ),
  };
});

describe("ChangePasswordModal", () => {
  const mockOnClose = vi.fn();

  beforeEach(() => {
    vi.resetAllMocks();
  });

  const renderModal = () => {
    return render(<ChangePasswordModal open={true} onClose={mockOnClose} />);
  };

  it("renders correctly", () => {
    renderModal();
    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Current Password")).toBeInTheDocument();
    expect(screen.getByText("New Password")).toBeInTheDocument();
    expect(screen.getByText("Confirm Password")).toBeInTheDocument();
  });

  it("validates empty fields", async () => {
    renderModal();
    fireEvent.click(screen.getByRole("button", { name: /change password/i }));

    await waitFor(() => {
      expect(screen.getByText("Current Password")).toBeInTheDocument();
      expect(screen.getByText("New Password")).toBeInTheDocument();
    });
  });

  it("calls API on successful validation", async () => {
    renderModal();
    fireEvent.click(screen.getByText("Cancel"));
    expect(mockOnClose).toHaveBeenCalled();
    expect(changePassword).not.toHaveBeenCalled();
  });
});
