import {
	Button,
	Card,
	Descriptions,
	Empty,
	Form,
	message,
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

const { Text, Title } = Typography;

const linkConfigs: LinkConfig[] = [
	{
		type: "inventory",
		field: "inventory_url",
		label: "Inventory Spreadsheet URL",
		sheetLabel: "Inventory Sheet",
	},
	{
		type: "wallet",
		field: "wallet_url",
		label: "Wallet Spreadsheet URL",
		sheetLabel: "Wallet Sheet",
	},
	{
		type: "shipping",
		field: "shipping_url",
		label: "Shipping Spreadsheet URL",
		sheetLabel: "Shipping Sheet",
	},
	{
		type: "order",
		field: "order_url",
		label: "Order Spreadsheet URL",
		sheetLabel: "Order Sheet",
	},
];

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

	const { data: savedLinks, isLoading: isLinksLoading } =
		useGoogleSheetsLinks();
	const { data: details, isLoading: isDetailsLoading } =
		useGoogleSheetsDetails();
	const validateLinkMutation = useValidateLink();
	const saveLinksMutation = useSaveLinks();
	const updateSettingsMutation = useUpdateSettings();

	useEffect(() => {
		if (!savedLinks) {
			return;
		}

		form.setFieldsValue(savedLinks);
	}, [form, savedLinks]);

	useEffect(() => {
		if (!details) {
			return;
		}

		setSelectedSheets((previous) => ({
			...previous,
			inventory: details.inventory_sheet_name || previous.inventory,
			wallet: details.wallet_sheet_name || previous.wallet,
			shipping: details.shipping_sheet_name || previous.shipping,
			order: details.order_sheet_name || previous.order,
		}));
	}, [details]);

	const hasMetadata = useMemo(() => {
		if (!details?.sheets_metadata) {
			return false;
		}

		return Object.values(details.sheets_metadata).some(
			(metadataList) => metadataList.length > 0,
		);
	}, [details]);

	const sheetOptionsByType = useMemo(() => {
		const buildOptions = (type: LinkType) => {
			const validationSheets = validationResults[type]?.sheets ?? [];
			const metadataSheets = details?.sheets_metadata?.[type] ?? [];
			const names = Array.from(
				new Set(
					[...validationSheets, ...metadataSheets]
						.map((sheet) => sheet.name)
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
		const rawValue = form.getFieldValue(config.field);
		const spreadsheetUrl = rawValue?.trim();

		if (!spreadsheetUrl) {
			message.warning(`Please enter ${config.label.toLowerCase()}`);
			return;
		}

		setValidatingType(config.type);

		validateLinkMutation.mutate(
			{
				spreadsheet_url: spreadsheetUrl,
				type: config.type,
			},
			{
				onSuccess: (result) => {
					setValidationResults((previous) => ({
						...previous,
						[config.type]: result,
					}));

					setSelectedSheets((previous) => {
						const currentSelection = previous[config.type];
						if (!currentSelection) {
							return previous;
						}

						const hasCurrentSelection = result.sheets.some(
							(sheet) => sheet.name === currentSelection,
						);

						if (hasCurrentSelection) {
							return previous;
						}

						return {
							...previous,
							[config.type]: undefined,
						};
					});
				},
				onSettled: () => {
					setValidatingType(null);
				},
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
		const selectedSheetName = sheetName || "";
		const validatedSpreadsheetID = validationResults[type]?.spreadsheet_id;

		setSelectedSheets((previous) => ({
			...previous,
			[type]: selectedSheetName || undefined,
		}));

		const payloadByType: Record<LinkType, UpdateSettingsPayload> = {
			inventory: {
				inventory_sheet_name: selectedSheetName,
				...(validatedSpreadsheetID
					? { inventory_spreadsheet_id: validatedSpreadsheetID }
					: {}),
			},
			wallet: {
				wallet_sheet_name: selectedSheetName,
				...(validatedSpreadsheetID
					? { wallet_spreadsheet_id: validatedSpreadsheetID }
					: {}),
			},
			shipping: {
				shipping_sheet_name: selectedSheetName,
				...(validatedSpreadsheetID
					? { shipping_spreadsheet_id: validatedSpreadsheetID }
					: {}),
			},
			order: {
				order_sheet_name: selectedSheetName,
				...(validatedSpreadsheetID
					? { order_spreadsheet_id: validatedSpreadsheetID }
					: {}),
			},
		};

		updateSettingsMutation.mutate(payloadByType[type]);
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
						<Text type="secondary" style={{ fontSize: 12 }}>
							Locked
						</Text>
						<Switch checked={isLocked} onChange={setIsLocked} />
					</Space>
				}
			>
				<Spin spinning={isLinksLoading}>
					<Form
						form={form}
						layout="vertical"
						initialValues={{
							inventory_url: "",
							wallet_url: "",
							shipping_url: "",
							order_url: "",
						}}
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
								onSheetSelect={(value) => handleSheetSelect(config.type, value)}
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

			<Card title="Sheet Metadata" style={{ marginTop: 16 }}>
				<Spin spinning={isDetailsLoading}>
					{!hasMetadata && <Empty description="No sheet metadata available" />}

					{hasMetadata && (
						<Space direction="vertical" size={12} style={{ width: "100%" }}>
							{linkConfigs.map((config) => {
								const metadataList =
									details?.sheets_metadata?.[config.type] ?? [];

								return (
									<Card
										key={`${config.type}-metadata`}
										size="small"
										title={config.label}
									>
										{metadataList.length === 0 ? (
											<Text type="secondary" style={{ fontSize: 12 }}>
												No metadata found.
											</Text>
										) : (
											<Space
												direction="vertical"
												size={8}
												style={{ width: "100%" }}
											>
												{metadataList.map((sheet) => (
													<Descriptions
														key={`${config.type}-${sheet.sheet_id}`}
														size="small"
														column={2}
														bordered
													>
														<Descriptions.Item label="Sheet Name">
															{sheet.name}
														</Descriptions.Item>
														<Descriptions.Item label="Sheet ID">
															{sheet.sheet_id}
														</Descriptions.Item>
														<Descriptions.Item label="Column Count">
															{sheet.column_count}
														</Descriptions.Item>
														<Descriptions.Item label="Row Count">
															{sheet.row_count}
														</Descriptions.Item>
													</Descriptions>
												))}
											</Space>
										)}
									</Card>
								);
							})}

							{details?.last_updated && (
								<Text type="secondary" style={{ fontSize: 12 }}>
									Last updated:{" "}
									{new Date(details.last_updated).toLocaleString()}
								</Text>
							)}
						</Space>
					)}
				</Spin>
			</Card>
		</div>
	);
}
