import { useEffect } from "react";
import { App, Form, Input, Modal, Select } from "antd";
import { usersApi, type UpdateUserPayload, type UserResponse } from "@/api/users";

const ROLE_LEVEL: Record<string, number> = {
  developer: 4,
  owner: 3,
  admin: 2,
  user: 1,
};

const ALL_ROLES = ["developer", "owner", "admin", "user"] as const;

interface EditUserModalProps {
  open: boolean;
  user: UserResponse | null;
  onClose: () => void;
  onSuccess: () => void;
  actorRole: string;
}

export function EditUserModal({
  open,
  user,
  onClose,
  onSuccess,
  actorRole,
}: EditUserModalProps) {
  const { message } = App.useApp();
  const [form] = Form.useForm<UpdateUserPayload>();

  const actorLevel = ROLE_LEVEL[actorRole] ?? 0;
  const assignableRoles = ALL_ROLES.filter(
    (role) => ROLE_LEVEL[role] < actorLevel,
  );

  useEffect(() => {
    if (open && user) {
      form.setFieldsValue({
        username: user.username,
        email: user.email,
        role: user.role,
      });
    } else {
      form.resetFields();
    }
  }, [open, user, form]);

  const handleSubmit = async (values: UpdateUserPayload) => {
    if (!user) return;
    const response = await usersApi.update(user.id, values);
    if (!response.success) {
      message.error(response.error ?? "Failed to update user");
      return;
    }
    message.success("User updated successfully");
    onSuccess();
    onClose();
  };

  return (
    <Modal
      title="Edit User"
      open={open}
      onCancel={onClose}
      footer={null}
      destroyOnClose
      width={480}
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
        style={{ marginTop: 16 }}
      >
        <Form.Item
          name="username"
          label="Username"
          rules={[{ required: true, message: "Username is required" }]}
        >
          <Input placeholder="Enter username" />
        </Form.Item>

        <Form.Item
          name="email"
          label="Email"
          rules={[
            { required: true, message: "Email is required" },
            { type: "email", message: "Please enter a valid email" },
          ]}
        >
          <Input placeholder="Enter email" />
        </Form.Item>

        <Form.Item
          name="role"
          label="Role"
          rules={[{ required: true, message: "Role is required" }]}
        >
          <Select placeholder="Select role">
            {assignableRoles.map((role) => (
              <Select.Option key={role} value={role}>
                {role.charAt(0).toUpperCase() + role.slice(1)}
              </Select.Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item style={{ marginBottom: 0, textAlign: "right" }}>
          <button
            type="button"
            onClick={onClose}
            style={{
              marginRight: 8,
              padding: "4px 15px",
              border: "1px solid #d9d9d9",
              borderRadius: 3,
              background: "transparent",
              cursor: "pointer",
            }}
          >
            Cancel
          </button>
          <button
            type="submit"
            style={{
              padding: "4px 15px",
              border: "none",
              borderRadius: 3,
              background: "#0369a1",
              color: "#fff",
              cursor: "pointer",
            }}
          >
            Save
          </button>
        </Form.Item>
      </Form>
    </Modal>
  );
}
