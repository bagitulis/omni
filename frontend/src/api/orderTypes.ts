/**
 * Order API Types
 * Separated from API functions per SRP principle.
 */

export type OrderTab =
  | "unpaid"
  | "unprocess"
  | "processed"
  | "shipped"
  | "completed"
  | "cancelled"
  | "locked"
  | "today";

export interface GetOrdersParams {
  page?: number;
  pageSize?: number;
  status?: string;
  platform?: string;
  search?: string;
  startDate?: string;
  endDate?: string;
}

export interface BulkPrintLabelsResponse {
  labels: Array<{
    order_sn: string;
    file_data: string;
    status: string;
  }>;
  failed: Array<{
    order_sn: string;
    error: string;
  }>;
  count: number;
}

export interface BulkPrintLabelsOptions {
  platform?: string;
  include_products?: boolean;
  tiktok_document_type?:
    | "SHIPPING_LABEL"
    | "PACKING_SLIP"
    | "SHIPPING_LABEL_AND_PACKING_SLIP";
}

export interface CancelOrderParams {
  order_no: string;
  platform: string;
  cancel_reason: string;
  reason_detail?: string;
  order_item_id?: string;
}

export interface ShipOrderParams {
  order_no: string;
  platform: string;
  shipping_provider: string;
  tracking_number?: string;
  address_id?: number;
  pickup_time_id?: string;
  branch_id?: number;
  package_id?: string;
  order_item_ids?: string[];
}

export interface LazadaDocumentResponse {
  document?: {
    file?: string;
    url?: string;
    mime_type?: string;
  };
}
