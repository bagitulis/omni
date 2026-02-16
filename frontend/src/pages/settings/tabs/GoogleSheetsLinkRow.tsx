import { Button, Form, Input, Select, Space, Typography } from "antd";
import type { SpreadsheetLinks, ValidationResult } from "@/types/googleSheets";

const { Text } = Typography;

export type LinkType = "inventory" | "wallet" | "shipping" | "order";
type LinkField = keyof SpreadsheetLinks;

export interface LinkConfig {
	type: LinkType;
	field: LinkField;
	label: string;
	sheetLabel: string;
}

interface GoogleSheetsLinkRowProps {
	config: LinkConfig;
	selectedSheet?: string;
	sheetOptions: Array<{ label: string; value: string }>;
	validationResult?: ValidationResult;
	isLocked: boolean;
	isValidating: boolean;
	onValidate: () => void;
	onSheetSelect: (value?: string) => void;
}

export function GoogleSheetsLinkRow({
	config,
	selectedSheet,
	sheetOptions,
	validationResult,
	isLocked,
	isValidating,
	onValidate,
	onSheetSelect,
}: GoogleSheetsLinkRowProps) {
	return (
		<Form.Item
			key={config.field}
			label={config.label}
			style={{ marginBottom: 16 }}
		>
			<Space direction="vertical" size={8} style={{ width: "100%" }}>
				<Space.Compact style={{ width: "100%" }}>
					<Form.Item name={config.field} noStyle>
						<Input
							placeholder="https://docs.google.com/spreadsheets/d/..."
							disabled={isLocked}
						/>
					</Form.Item>
					<Select
						value={selectedSheet}
						options={sheetOptions}
						placeholder={config.sheetLabel}
						disabled={isLocked || sheetOptions.length === 0}
						allowClear
						style={{ width: 220 }}
						onChange={onSheetSelect}
					/>
					<Button
						onClick={onValidate}
						loading={isValidating}
						disabled={isLocked}
					>
						Validate
					</Button>
				</Space.Compact>

				{(validationResult || selectedSheet) && (
					<Space direction="vertical" size={4} style={{ width: "100%" }}>
						{validationResult && (
							<>
								<Text style={{ fontSize: 12 }}>
									Spreadsheet: <Text strong>{validationResult.name}</Text>
								</Text>
								<Text type="secondary" style={{ fontSize: 12 }}>
									Sheets found: {validationResult.sheets.length} - select one
									from the dropdown beside validate.
								</Text>
							</>
						)}
						{selectedSheet && (
							<Text type="secondary" style={{ fontSize: 12 }}>
								Selected sheet: <Text strong>{selectedSheet}</Text>
							</Text>
						)}
					</Space>
				)}
			</Space>
		</Form.Item>
	);
}
