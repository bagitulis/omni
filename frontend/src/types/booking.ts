/**
 * Booking Types
 * API types use snake_case to match backend JSON response
 */

export interface Booking {
  booking_sn: string;
  order_sn: string;
  booking_status: string;
  match_status: string;
  region: string;
  shipping_carrier: string;
  recipient_name: string;
  recipient_phone: string;
  recipient_address_json: string;
  fulfillment_flag: string;
  item_count: number;
  has_parent_order: boolean;
  create_time: number;
  update_time: number;
  pickup_done_time: number;
  synced_at: string;
}

export interface BookingItem {
  booking_sn: string;
  item_id: string;
  model_id: string;
  item_name: string;
  model_name: string;
  item_sku: string;
  model_sku: string;
  sku: string;
  quantity: number;
  weight: number;
  image_url: string;
}

export interface BookingListResponse {
  success: boolean;
  data: Booking[];
  count: number;
  page: number;
  page_size: number;
  error?: string;
}

export interface BookingDetailResponse {
  success: boolean;
  data: {
    booking: Booking;
    items: BookingItem[];
  };
  error?: string;
}
