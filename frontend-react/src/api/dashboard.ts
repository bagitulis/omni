// import api from './client';
import { DashboardData } from "../types/dashboard";

// Mock data for now
const MOCK_DATA: DashboardData = {
  metrics: {
    orders_pending: 12,
    stock_warning: 3,
    ready_to_ship: 8,
    platform_status: 98,
  },
  recent_orders: [
    {
      order_sn: "SH12345",
      status: "PAID",
      platform: "shopee",
      amount: 250000,
      buyer_username: "user123",
      created_at: new Date().toISOString(),
    },
    {
      order_sn: "TT67890",
      status: "READY_TO_SHIP",
      platform: "tiktok",
      amount: 180000,
      buyer_username: "tiktok_fan",
      created_at: new Date(Date.now() - 3600000).toISOString(),
    },
    {
      order_sn: "LZ33445",
      status: "SHIPPING",
      platform: "lazada",
      amount: 450000,
      buyer_username: "laz_buyer",
      created_at: new Date(Date.now() - 7200000).toISOString(),
    },
    {
      order_sn: "TK99887",
      status: "COMPLETED",
      platform: "tokopedia",
      amount: 125000,
      buyer_username: "toped_user",
      created_at: new Date(Date.now() - 86400000).toISOString(),
    },
  ],
};

export async function getDashboardData(): Promise<DashboardData> {
  // Simulate API delay
  await new Promise((resolve) => setTimeout(resolve, 800));

  // In real implementation:
  // const response = await api.get<{ success: boolean; data: DashboardData }>('/dashboard');
  // if (!response.data.success) throw new Error(response.data.error);
  // return response.data.data;

  return MOCK_DATA;
}
