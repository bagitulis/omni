/**
 * Test Analytics Full Flow
 * Login -> Check Sync Status -> Analyze -> Show Results
 * 
 * Tenant: yumna | Password: password123
 */

import axios from "axios";
import { getLogger } from "./src/utils/logger";

const logger = getLogger("AnalyticsTest");

const API_BASE = "http://localhost:3000/api";
const USERNAME = "tester";
const PASSWORD = "tester@123";
const TENANT_ID = "yumna";

interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

async function login(): Promise<string> {
  logger.info("🔐 Logging in...");
  try {
    const response = await axios.post(`${API_BASE}/auth/login`, {
      username: USERNAME,
      password: PASSWORD,
    });
    logger.info("✅ Login successful");
    
    // Switch tenant to yumna
    logger.info("🔄 Switching to tenant: yumna...");
    const switchResponse = await axios.post(
      `${API_BASE}/auth/switch-tenant`,
      { tenantId: TENANT_ID },
      { headers: { Authorization: `Bearer ${response.data.token}` } }
    );
    logger.info("✅ Tenant switched successfully");
    
    return switchResponse.data.token || response.data.token;
  } catch (error: any) {
    logger.error(`❌ Login failed: ${error.response?.data?.error || error.message}`);
    throw error;
  }
}

async function getSyncStatus(token: string, month: number, year: number): Promise<any> {
  logger.info(`\n📊 Getting sync status for ${year}-${month}...`);
  try {
    const response = await axios.get<ApiResponse>(
      `${API_BASE}/analytics/shopee/sync-status`,
      {
        params: { month, year },
        headers: {
          Authorization: `Bearer ${token}`,
          "x-tenant-id": TENANT_ID,
        },
      }
    );
    logger.info(`   Synced: ${response.data.data?.synced}`);
    logger.info(`   Total Orders: ${response.data.data?.totalOrders || 0}`);
    logger.info(`   Synced At: ${response.data.data?.syncedAt || 'Never'}`);
    return response.data.data;
  } catch (error: any) {
    logger.error(`❌ Failed: ${error.response?.data?.error || error.message}`);
    throw error;
  }
}

async function getSettings(token: string): Promise<any> {
  logger.info("\n⚙️ Getting analytics settings...");
  try {
    const response = await axios.get<ApiResponse>(
      `${API_BASE}/analytics/shopee/settings`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
          "x-tenant-id": TENANT_ID,
        },
      }
    );
    logger.info(`   Price Column: ${response.data.data?.priceColumn}`);
    logger.info(`   Formula Deduction: ${response.data.data?.formulaDeduction}`);
    logger.info(`   Formula Multiplier: ${response.data.data?.formulaMultiplier}`);
    return response.data.data;
  } catch (error: any) {
    logger.error(`❌ Failed: ${error.response?.data?.error || error.message}`);
    throw error;
  }
}

async function syncEscrowData(token: string, month: number, year: number): Promise<any> {
  logger.info(`\n🔄 Syncing escrow data for ${year}-${month}...`);
  logger.info(`   This may take a while (fetching from Shopee API)...`);
  try {
    const response = await axios.post<ApiResponse>(
      `${API_BASE}/analytics/shopee/sync`,
      { month, year, forceResync: false },
      {
        headers: {
          Authorization: `Bearer ${token}`,
          "x-tenant-id": TENANT_ID,
        },
        timeout: 300000, // 5 minutes timeout
      }
    );
    logger.info(`   ✅ ${response.data.message}`);
    logger.info(`   Total Orders: ${response.data.data?.totalOrders || 0}`);
    logger.info(`   Total Items: ${response.data.data?.totalItems || 0}`);
    return response.data;
  } catch (error: any) {
    logger.error(`❌ Sync failed: ${error.response?.data?.error || error.message}`);
    throw error;
  }
}

async function getReconciliation(token: string, month: number, year: number): Promise<any> {
  logger.info(`\n📈 Getting reconciliation for ${year}-${month}...`);
  try {
    const response = await axios.get<ApiResponse>(
      `${API_BASE}/analytics/shopee/reconciliation`,
      {
        params: { month, year },
        headers: {
          Authorization: `Bearer ${token}`,
          "x-tenant-id": TENANT_ID,
        },
      }
    );
    
    const result = response.data.data;
    if (result?.summary) {
      logger.info("\n📊 SUMMARY:");
      logger.info(`   Total SKUs: ${result.summary.totalSku}`);
      logger.info(`   Total Transactions: ${result.summary.totalTransactions}`);
      logger.info(`   SKU OK: ${result.summary.skuOk}`);
      logger.info(`   SKU with Price Diff: ${result.summary.skuWithPriceDiff}`);
      logger.info(`   SKU No Inventory: ${result.summary.skuNoInventory}`);
      
      if (result.skuGroups?.length > 0) {
        logger.info("\n📦 SAMPLE SKUs (first 5):");
        for (const sku of result.skuGroups.slice(0, 5)) {
          logger.info(`   ${sku.status} | ${sku.modelSku || sku.sku} | ${sku.itemName?.substring(0, 30)}...`);
          logger.info(`          Inventory: ${sku.inventoryPrice} | Variants: ${sku.priceVariants?.length || 0}`);
        }
      }
    } else {
      logger.info("   No data found (need to sync first)");
    }
    
    return result;
  } catch (error: any) {
    logger.error(`❌ Failed: ${error.response?.data?.error || error.message}`);
    throw error;
  }
}

async function getShippingFeeAnalysis(token: string, month: number, year: number): Promise<any> {
  logger.info(`\n🚚 Getting shipping fee analysis for ${year}-${month}...`);
  try {
    const response = await axios.get<ApiResponse>(
      `${API_BASE}/analytics/shopee/shipping-fee`,
      {
        params: { month, year },
        headers: {
          Authorization: `Bearer ${token}`,
          "x-tenant-id": TENANT_ID,
        },
      }
    );
    
    const result = response.data.data;
    if (result?.summary) {
      logger.info("\n🚚 SHIPPING FEE SUMMARY:");
      logger.info(`   Total Orders: ${result.summary.totalOrders}`);
      logger.info(`   Orders with Difference: ${result.summary.ordersWithDifference}`);
      logger.info(`   Total Profit: Rp ${result.summary.totalProfit.toLocaleString()}`);
      logger.info(`   Total Loss: Rp ${result.summary.totalLoss.toLocaleString()}`);
      logger.info(`   Net Impact: Rp ${result.summary.netImpact.toLocaleString()}`);
      
      if (result.orders?.length > 0) {
        logger.info("\n📦 SAMPLE ORDERS (first 5):");
        for (const order of result.orders.slice(0, 5)) {
          const diff = order.difference > 0 ? `+${order.difference}` : order.difference;
          logger.info(`   ${order.orderSn} | Buyer: ${order.buyerPaid} | Actual: ${order.actualFee} | Diff: ${diff}`);
        }
      }
    } else {
      logger.info("   No shipping fee data found");
    }
    
    return result;
  } catch (error: any) {
    logger.error(`❌ Failed: ${error.response?.data?.error || error.message}`);
    throw error;
  }
}

async function main() {
  logger.info("\n╔════════════════════════════════════════════════════════════╗");
  logger.info("║  SHOPEE ANALYTICS FULL TEST                                 ║");
  logger.info("╚════════════════════════════════════════════════════════════╝\n");

  try {
    // 1. Login
    const token = await login();

    // 2. Check Sync Status for December 2025
    const syncStatus = await getSyncStatus(token, 12, 2025);

    // 3. Get Settings  
    await getSettings(token);

    // 4. If not synced, sync first
    if (!syncStatus?.synced) {
      logger.info("\n⚠️ Data not synced. Starting sync...");
      await syncEscrowData(token, 12, 2025);
      
      // Re-check sync status
      const newStatus = await getSyncStatus(token, 12, 2025);
      if (newStatus?.synced) {
        await getReconciliation(token, 12, 2025);
      }
    } else {
      // 5. Get Reconciliation
      await getReconciliation(token, 12, 2025);
    }

    // 6. Get Shipping Fee Analysis
    await getShippingFeeAnalysis(token, 12, 2025);

    logger.info("\n✅ All tests completed successfully!");
    
  } catch (error: any) {
    logger.error(`\n❌ Test failed: ${error.message}`);
    process.exit(1);
  }
}

main();
