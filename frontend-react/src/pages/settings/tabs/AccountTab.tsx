import {
  Card,
  Form,
  Input,
  Button,
  Avatar,
  Typography,
  Divider,
  message,
  theme,
} from "antd";
import { UserOutlined, LockOutlined } from "@ant-design/icons";

const { Text, Title } = Typography;
const { useToken } = theme;

interface ProfileForm {
  name: string;
  email: string;
  phone?: string;
}

interface PasswordForm {
  current_password: string;
  new_password: string;
  confirm_password: string;
}

export default function AccountTab() {
  const { token } = useToken();
  const [profileForm] = Form.useForm<ProfileForm>();
  const [passwordForm] = Form.useForm<PasswordForm>();

  // Mock user data
  const user = {
    name: "Admin User",
    email: "admin@example.com",
    phone: "+62 812 3456 7890",
    avatar: null,
  };

  const handleProfileSave = (values: ProfileForm) => {
    console.log("Saving profile:", values);
    message.success("Profile updated successfully");
  };

  const handlePasswordChange = (values: PasswordForm) => {
    console.log("Changing password:", values);
    if (values.new_password !== values.confirm_password) {
      message.error("Passwords do not match");
      return;
    }
    message.success("Password changed successfully");
    passwordForm.resetFields();
  };

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Title level={5} style={{ margin: 0, fontSize: 14 }}>
          Account Settings
        </Title>
        <Text type="secondary" style={{ fontSize: 12 }}>
          Manage your account profile and security
        </Text>
      </div>

      <Card
        title="Profile Information"
        size="small"
        style={{ marginBottom: 16, borderRadius: token.borderRadius }}
      >
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 16,
            marginBottom: 24,
          }}
        >
          <Avatar
            size={64}
            icon={<UserOutlined />}
            style={{ backgroundColor: token.colorPrimary }}
          />
          <div>
            <Text strong style={{ fontSize: 14, display: "block" }}>
              {user.name}
            </Text>
            <Text type="secondary" style={{ fontSize: 12 }}>
              {user.email}
            </Text>
          </div>
        </div>

        <Form
          form={profileForm}
          layout="vertical"
          initialValues={{
            name: user.name,
            email: user.email,
            phone: user.phone,
          }}
          onFinish={handleProfileSave}
          style={{ maxWidth: 400 }}
        >
          <Form.Item
            label="Full Name"
            name="name"
            rules={[{ required: true, message: "Name is required" }]}
          >
            <Input prefix={<UserOutlined />} />
          </Form.Item>
          <Form.Item
            label="Email"
            name="email"
            rules={[
              { required: true, message: "Email is required" },
              { type: "email", message: "Invalid email format" },
            ]}
          >
            <Input disabled />
          </Form.Item>
          <Form.Item label="Phone" name="phone">
            <Input placeholder="+62 xxx xxxx xxxx" />
          </Form.Item>
          <Form.Item style={{ marginBottom: 0 }}>
            <Button type="primary" htmlType="submit">
              Update Profile
            </Button>
          </Form.Item>
        </Form>
      </Card>

      <Card
        title="Change Password"
        size="small"
        style={{ borderRadius: token.borderRadius }}
      >
        <Form
          form={passwordForm}
          layout="vertical"
          onFinish={handlePasswordChange}
          style={{ maxWidth: 400 }}
        >
          <Form.Item
            label="Current Password"
            name="current_password"
            rules={[
              { required: true, message: "Current password is required" },
            ]}
          >
            <Input.Password prefix={<LockOutlined />} />
          </Form.Item>
          <Divider style={{ margin: "16px 0" }} />
          <Form.Item
            label="New Password"
            name="new_password"
            rules={[
              { required: true, message: "New password is required" },
              { min: 8, message: "Password must be at least 8 characters" },
            ]}
          >
            <Input.Password prefix={<LockOutlined />} />
          </Form.Item>
          <Form.Item
            label="Confirm New Password"
            name="confirm_password"
            rules={[
              { required: true, message: "Please confirm your password" },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue("new_password") === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error("Passwords do not match"));
                },
              }),
            ]}
          >
            <Input.Password prefix={<LockOutlined />} />
          </Form.Item>
          <Form.Item style={{ marginBottom: 0 }}>
            <Button type="primary" htmlType="submit">
              Change Password
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
