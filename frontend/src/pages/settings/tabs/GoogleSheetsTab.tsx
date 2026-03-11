import {
	Button,
	Card,
	Form,
	Space,
	Spin,
	Switch,
	Typography,
} from "antd";
import { useEffect, useMemo, useState } from "react";
import {
	useGoogleSheetsDetails,
	useGoogleSheetsLinks,
	useSaveLinks,
	useUpdateSettings,
	useValidateLink,
} from "@/hooks/useGoogleSheets";
import type {
	SpreadsheetLinks,
	UpdateSettingsPayload,
	ValidationResult,
} from "@/types/googleSheets";
import {
	GoogleSheetsLinkRow,
	type LinkConfig,
	type LinkType,
} from "./GoogleSheetsLinkRow";
import { SheetMetadataCard } from "./SheetMetadataCard";
import { message } from "@/components/AntStaticApi";

const { Text, Title } = Typography;

const linkConfigs: LinkConfig[] = [
	{ type: "inventory", field: "inventory_url", label: "Inventory Spreadsheet URL", sheetLabel: "Inventory Sheet" },
	{ type: "wallet", field: "wallet_url", label: "Wallet Spreadsheet URL", sheetLabel: "Wallet Sheet" },
	{ type: "shipping", field: "shipping_url", label: "Shipping Spreadsheet URL", sheetLabel: "Shipping Sheet" },
	{ type: "order", field: "order_url", label: "Order Spreadsheet URL", sheetLabel: "Order Sheet" },
];

/** Build an UpdateSettingsPayload from type, sheet name, and optional spreadsheet ID */
function buildSheetPayload(
	type: LinkType,
	sheetName: string,
	spreadsheetId?: string,
): UpdateSettingsPayload {
	const payload: UpdateSettingsPayload = {};
	payload[`${type}_sheet_name`] = sheetName;
	if (spreadsheetId) {
		payload[`${type}_spreadsheet_id`] = spreadsheetId;
	}
	return payload;
}

export default function GoogleSheetsTab() {
	const [form] = Form.useForm<SpreadsheetLinks>();
	const [isLocked, setIsLocked] = useState(false);
	const [validationResults, setValidationResults] = useState<
		Partial<Record<LinkType, ValidationResult>>
	>({});
	const [selectedSheets, setSelectedSheets] = useState<
		Partial<Record<LinkType, string>>
	>({});
	const [validatingType, setValidatingType] = useState<LinkType | null>(null);

	const { data: savedLinks, isLoading: isLinksLoading } = useGoogleSheetsLinks();
	const { data: details, isLoading: isDetailsLoading } = useGoogleSheetsDetails();
	const validateLinkMutation = useValidateLink();
	const saveLinksMutation = useSaveLinks();
	const updateSettingsMutation = useUpdateSettings();

	useEffect(() => {
		if (savedLinks) form.setFieldsValue(savedLinks);
	}, [form, savedLinks]);

	useEffect(() => {
		if (!details) return;
		setSelectedSheets((prev) => ({
			...prev,
			inventory: details.inventory_sheet_name || prev.inventory,
			wallet: details.wallet_sheet_name || prev.wallet,
			shipping: details.shipping_sheet_name || prev.shipping,
			order: details.order_sheet_name || prev.order,
		}));
	}, [details]);

	const sheetOptionsByType = useMemo(() => {
		const buildOptions = (type: LinkType) => {
			const validationSheets = validationResults[type]?.sheets ?? [];
			const metadataSheets = details?.sheets_metadata?.[type] ?? [];
			const names = Array.from(
				new Set(
					[...validationSheets, ...metadataSheets]
						.map((s) => s.name)
						.filter(Boolean),
				),
			);
			return names.map((name) => ({ label: name, value: name }));
		};
		return {
			inventory: buildOptions("inventory"),
			wallet: buildOptions("wallet"),
			shipping: buildOptions("shipping"),
			order: buildOptions("order"),
		};
	}, [details, validationResults]);

	const handleValidate = (config: LinkConfig) => {
		const spreadsheetUrl = form.getFieldValue(config.field)?.trim();
		if (!spreadsheetUrl) {
			message.warning(`Please enter ${config.label.toLowerCase()}`);
			return;
		}
		setValidatingType(config.type);
		validateLinkMutation.mutate(
			{ spreadsheet_url: spreadsheetUrl, type: config.type },
			{
				onSuccess: (result) => {
					setValidationResults((prev) => ({ ...prev, [config.type]: result }));
					setSelectedSheets((prev) => {
						const cur = prev[config.type];
						if (!cur || result.sheets.some((s) => s.name === cur)) return prev;
						return { ...prev, [config.type]: undefined };
					});
				},
				onSettled: () => setValidatingType(null),
			},
		);
	};

	const handleSaveAll = () => {
		const values = form.getFieldsValue();
		saveLinksMutation.mutate({
			inventory_url: values.inventory_url?.trim() || null,
			wallet_url: values.wallet_url?.trim() || null,
			shipping_url: values.shipping_url?.trim() || null,
			order_url: values.order_url?.trim() || null,
		});
	};

	const handleSheetSelect = (type: LinkType, sheetName?: string) => {
		const name = sheetName || "";
		setSelectedSheets((prev) => ({ ...prev, [type]: name || undefined }));
		const payload = buildSheetPayload(
			type,
			name,
			validationResults[type]?.spreadsheet_id,
		);
		updateSettingsMutation.mutate(payload);
	};

	return (
		<div>
			<div style={{ marginBottom: 16 }}>
				<Title level={5} style={{ margin: 0, fontSize: 14 }}>
					Google Sheets Settings
				</Title>
				<Text type="secondary" style={{ fontSize: 12 }}>
					Configure spreadsheet links and review discovered sheet metadata.
				</Text>
			</div>

			<Card
				title="Spreadsheet Links"
				extra={
					<Space size={8}>
						<Text type="secondary" style={{ fontSize: 12 }}>Locked</Text>
						<Switch checked={isLocked} onChange={setIsLocked} />
					</Space>
				}
			>
				<Spin spinning={isLinksLoading}>
					<Form
						form={form}
						layout="vertical"
						initialValues={{ inventory_url: "", wallet_url: "", shipping_url: "", order_url: "" }}
					>
						{linkConfigs.map((config) => (
							<GoogleSheetsLinkRow
								key={config.field}
								config={config}
								selectedSheet={selectedSheets[config.type]}
								sheetOptions={sheetOptionsByType[config.type]}
								validationResult={validationResults[config.type]}
								isLocked={isLocked}
								isValidating={
									validateLinkMutation.isPending &&
									validatingType === config.type
								}
								onValidate={() => handleValidate(config)}
								onSheetSelect={(v) => handleSheetSelect(config.type, v)}
							/>
						))}
						<Button
							type="primary"
							onClick={handleSaveAll}
							loading={saveLinksMutation.isPending}
							disabled={isLocked}
						>
							Save All Links
						</Button>
					</Form>
				</Spin>
			</Card>

			<SheetMetadataCard
				linkConfigs={linkConfigs}
				sheetsMetadata={details?.sheets_metadata}
				lastUpdated={details?.last_updated}
				isLoading={isDetailsLoading}
			/>
		</div>
	);
}
