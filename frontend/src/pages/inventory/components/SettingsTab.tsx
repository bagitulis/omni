import { useEffect, useState } from "react";
import {
	Alert,
	Button,
	Form,
	InputNumber,
	Spin,
	Typography,
} from "antd";
import { message } from "@/components/AntStaticApi";
import {
	getSettings,
	updateSettings,
	type WholesaleSettings,
} from "@/api/wholesale";

const DEFAULT_SETTINGS: WholesaleSettings = {
	admin_fee: 1500,
	min_order_1: 2,
	max_order_1: 3,
	max_order_tier_3: 1000,
};

export function SettingsTab() {
	const [form] = Form.useForm<WholesaleSettings>();
	const [loading, setLoading] = useState(true);
	const [saving, setSaving] = useState(false);
	const [resultMessage, setResultMessage] = useState<string | null>(null);
	const [resultType, setResultType] = useState<
		"success" | "error" | null
	>(null);

	useEffect(() => {
		let cancelled = false;

		async function loadSettings() {
			setLoading(true);
			try {
				const loaded = await getSettings();
				if (!cancelled && loaded) {
					form.setFieldsValue(loaded);
				}
			} catch (err) { logger.warn("Operation failed:", { err: err });
				if (!cancelled) {
					form.setFieldsValue(DEFAULT_SETTINGS);
				}
			} finally {
				if (!cancelled) setLoading(false);
			}
		}

		void loadSettings();
		return () => {
			cancelled = true;
		};
	}, [form]);

	const handleSave = async (values: WholesaleSettings) => {
		setSaving(true);
		setResultMessage(null);
		setResultType(null);

		try {
			const result = await updateSettings(values);
			if (result.success) {
				setResultType("success");
				setResultMessage("Settings saved successfully");
			} else {
				setResultType("error");
				setResultMessage(result.error ?? "Failed to save settings");
			}
		} catch (error) {
			const msg =
				error instanceof Error ? error.message : "Failed to save settings";
			setResultType("error");
			setResultMessage(msg);
			message.error(msg);
		} finally {
			setSaving(false);
		}
	};

	if (loading) {
		return (
			<div style={{ textAlign: "center", padding: 40 }}>
				<Spin tip="Loading settings..." />
			</div>
		);
	}

	return (
		<div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
			<Alert
				type="info"
				showIcon
				message="Wholesale pricing settings"
				description="These settings control how wholesale tier prices are calculated for Shopee products."
			/>

			{resultMessage && resultType ? (
				<Alert
					type={resultType}
					showIcon
					message={resultMessage}
					closable
					onClose={() => {
						setResultMessage(null);
						setResultType(null);
					}}
				/>
			) : null}

			<Form
				form={form}
				layout="vertical"
				onFinish={handleSave}
				initialValues={DEFAULT_SETTINGS}
				style={{ maxWidth: 400 }}
			>
				<Form.Item
					label="Admin Fee (Rp)"
					name="admin_fee"
					rules={[{ required: true, message: "Required" }]}
					tooltip="Fixed fee subtracted from base price for tier calculations"
				>
					<InputNumber
						min={0}
						step={100}
						style={{ width: "100%" }}
						formatter={(v) =>
							`${v}`.replace(/\B(?=(\d{3})+(?!\d))/g, ",")
						}
					/>
				</Form.Item>

				<Form.Item
					label="Min Order Tier 1"
					name="min_order_1"
					rules={[{ required: true, message: "Required" }]}
					tooltip="Minimum quantity for wholesale tier 1"
				>
					<InputNumber min={2} max={100} style={{ width: "100%" }} />
				</Form.Item>

				<Form.Item
					label="Max Order Tier 1"
					name="max_order_1"
					rules={[{ required: true, message: "Required" }]}
					tooltip="Maximum quantity for tier 1. Tier 2 starts at max+1"
				>
					<InputNumber min={2} max={100} style={{ width: "100%" }} />
				</Form.Item>

				<Form.Item
					label="Max Order Tier 3"
					name="max_order_tier_3"
					rules={[{ required: true, message: "Required" }]}
					tooltip="Maximum quantity for the highest tier"
				>
					<InputNumber min={10} max={10000} style={{ width: "100%" }} />
				</Form.Item>

				<Typography.Text type="secondary" style={{ display: "block", marginBottom: 12 }}>
					Tier ranges: Tier 1 = min_order_1 to max_order_1, Tier 2 =
					max_order_1+1 to max_order_1+2, Tier 3 = max_order_1+3 to
					max_order_tier_3
				</Typography.Text>

				<Button type="primary" htmlType="submit" loading={saving}>
					Save Settings
				</Button>
			</Form>
		</div>
	);
}
