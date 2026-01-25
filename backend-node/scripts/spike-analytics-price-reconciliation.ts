/**
 * SPIKE VALIDATION: Analytics - Order Price Reconciliation
 * 
 * Purpose: Validate technical feasibility untuk pricing reconciliation feature
 * Phase: POC / Quick Spike
 * Date: 2026-01-06
 * 
 * Objectives:
 * 1. ✅ Fetch order history (Shopee/TikTok) - 30 days
 * 2. ✅ Lookup inventory by SKU  
 * 3. ✅ Parse JSON dari InventoryRecord.data
 * 4. ✅ Test calculation logic
 * 5. ✅ Performance benchmark
 */

import { PrismaClient } from '@prisma/client';
import path from 'path';
import fs from 'fs';

// ============================================================
// TYPES & INTERFACES
// ============================================================

interface InventoryData {
  SKU?: string;
  'Nama Barang'?: string;
  'Nama Variasi'?: string;
  HARGA?: number;
  [key: string]: any; // Flexible untuk kolom lain dari Google Sheets
}

interface OrderItemAnalysis {
  platform: 'shopee' | 'tiktok';
  orderSn: string;
  sku: string;
  productName: string;
  orderPrice: number;
  expectedMinPrice: number | null;
  hargaMarketplace: number | null;
  status: 'MATCH' | 'MISMATCH' | 'MISSING_INVENTORY' | 'MISSING_PRICE';
  discrepancy: number;
  createdAt: Date;
}

interface SpikeFindings {
  summary: {
    totalOrders: number;
    totalItems: number;
    matchedItems: number;
    mismatchedItems: number;
    missingInventory: number;
    missingPrice: number;
    totalDiscrepancy: number;
  };
  performance: {
    queryTimeMs: number;
    itemsPerSecond: number;
  };
  constraints: string[];
  recommendations: string[];
  sampleMismatches: OrderItemAnalysis[];
}

// ============================================================
// CONFIGURATION
// ============================================================

const SPIKE_CONFIG = {
  // Period untuk test (default 30 days)
  periodDays: 30,
  
  // Kolom harga di inventory (configurable)
  marketplacePriceColumn: 'HARGA',
  
  // Formula constants
  formula: {
    deduction: 1500,
    multiplier: 0.84,
  },
  
  // Performance thresholds
  performanceTarget: {
    maxQueryTimeMs: 3000, // < 3 seconds
  },
};

// ============================================================
// PRICING CALCULATION
// ============================================================

/**
 * Calculate expected minimum selling price (Harga Jual)
 * Formula: Harga Jual = (Harga Marketplace - 1500) × 0.84
 */
function calculateExpectedMinPrice(hargaMarketplace: number): number {
  const { deduction, multiplier } = SPIKE_CONFIG.formula;
  return (hargaMarketplace - deduction) * multiplier;
}

/**
 * Compare order price vs expected minimum price
 */
function comparePrice(
  orderPrice: number,
  hargaMarketplace: number | null
): {
  status: OrderItemAnalysis['status'];
  expectedMinPrice: number | null;
  discrepancy: number;
} {
  if (hargaMarketplace === null) {
    return {
      status: 'MISSING_PRICE',
      expectedMinPrice: null,
      discrepancy: 0,
    };
  }

  const expectedMinPrice = calculateExpectedMinPrice(hargaMarketplace);
  const discrepancy = expectedMinPrice - orderPrice;

  const status = orderPrice >= expectedMinPrice ? 'MATCH' : 'MISMATCH';

  return {
    status,
    expectedMinPrice,
    discrepancy: status === 'MISMATCH' ? discrepancy : 0,
  };
}

// ============================================================
// INVENTORY LOOKUP
// ============================================================

/**
 * Lookup inventory record by SKU
 * Returns parsed inventory data or null if not found
 */
async function lookupInventory(
  prisma: PrismaClient,
  tenantId: string,
  sku: string
): Promise<InventoryData | null> {
  try {
    const record = await prisma.inventoryRecord.findFirst({
      where: {
        tenantId,
        keyValue: sku,
        keyColumnName: 'SKU', // Assuming SKU is the key column
      },
    });

    if (!record) {
      return null;
    }

    // Parse JSON string dari data column
    const inventoryData: InventoryData = JSON.parse(record.data);
    return inventoryData;
  } catch (error) {
    console.error(`Error looking up inventory for SKU ${sku}:`, error);
    return null;
  }
}

// ============================================================
// ORDER FETCHING & ANALYSIS
// ============================================================

/**
 * Fetch and analyze Shopee orders
 */
async function analyzeShopeeOrders(
  prisma: PrismaClient,
  tenantId: string,
  fromDate: Date
): Promise<OrderItemAnalysis[]> {
  const results: OrderItemAnalysis[] = [];

  // Fetch Shopee orders within period
  const orders = await prisma.shopeeOrder.findMany({
    where: {
      tenantId,
      createdAt: {
        gte: fromDate,
      },
      orderStatus: {
        in: ['COMPLETED', 'SHIPPED', 'READY_TO_SHIP'], // Only completed orders
      },
    },
    include: {
      items: true,
    },
  });

  console.log(`📦 Found ${orders.length} Shopee orders`);

  // Analyze each order item
  for (const order of orders) {
    for (const item of order.items) {
      const sku = item.itemSku || item.modelSku || '';
      
      if (!sku) {
        console.warn(`⚠️ No SKU for Shopee order ${order.orderSn}, item ${item.itemId}`);
        continue;
      }

      // Lookup inventory
      const inventoryData = await lookupInventory(prisma, tenantId, sku);
      
      if (!inventoryData) {
        results.push({
          platform: 'shopee',
          orderSn: order.orderSn,
          sku,
          productName: item.itemName || 'Unknown',
          orderPrice: item.price || 0,
          expectedMinPrice: null,
          hargaMarketplace: null,
          status: 'MISSING_INVENTORY',
          discrepancy: 0,
          createdAt: order.createdAt,
        });
        continue;
      }

      // Get marketplace price from inventory
      const hargaMarketplace = inventoryData[SPIKE_CONFIG.marketplacePriceColumn] || null;
      
      // Compare prices
      const comparison = comparePrice(item.price || 0, hargaMarketplace);

      results.push({
        platform: 'shopee',
        orderSn: order.orderSn,
        sku,
        productName: item.itemName || inventoryData['Nama Barang'] || 'Unknown',
        orderPrice: item.price || 0,
        expectedMinPrice: comparison.expectedMinPrice,
        hargaMarketplace,
        status: comparison.status,
        discrepancy: comparison.discrepancy,
        createdAt: order.createdAt,
      });
    }
  }

  return results;
}

/**
 * Fetch and analyze TikTok orders
 */
async function analyzeTiktokOrders(
  prisma: PrismaClient,
  tenantId: string,
  fromDate: Date
): Promise<OrderItemAnalysis[]> {
  const results: OrderItemAnalysis[] = [];

  // Fetch TikTok orders within period
  const orders = await prisma.tiktokOrder.findMany({
    where: {
      tenantId,
      createdAt: {
        gte: fromDate,
      },
      orderStatus: {
        in: ['COMPLETED', 'DELIVERED'], // Only completed orders
      },
    },
    include: {
      items: true,
    },
  });

  console.log(`📦 Found ${orders.length} TikTok orders`);

  // Analyze each order item
  for (const order of orders) {
    for (const item of order.items) {
      const sku = item.sellerSku || item.skuId || '';
      
      if (!sku) {
        console.warn(`⚠️ No SKU for TikTok order ${order.orderSn}, item ${item.productId}`);
        continue;
      }

      // Lookup inventory
      const inventoryData = await lookupInventory(prisma, tenantId, sku);
      
      if (!inventoryData) {
        results.push({
          platform: 'tiktok',
          orderSn: order.orderSn,
          sku,
          productName: item.productName || 'Unknown',
          orderPrice: item.price || 0,
          expectedMinPrice: null,
          hargaMarketplace: null,
          status: 'MISSING_INVENTORY',
          discrepancy: 0,
          createdAt: order.createdAt,
        });
        continue;
      }

      // Get marketplace price from inventory
      const hargaMarketplace = inventoryData[SPIKE_CONFIG.marketplacePriceColumn] || null;
      
      // Compare prices
      const comparison = comparePrice(item.price || 0, hargaMarketplace);

      results.push({
        platform: 'tiktok',
        orderSn: order.orderSn,
        sku,
        productName: item.productName || inventoryData['Nama Barang'] || 'Unknown',
        orderPrice: item.price || 0,
        expectedMinPrice: comparison.expectedMinPrice,
        hargaMarketplace,
        status: comparison.status,
        discrepancy: comparison.discrepancy,
        createdAt: order.createdAt,
      });
    }
  }

  return results;
}

// ============================================================
// MAIN SPIKE EXECUTION
// ============================================================

async function runSpike(tenantId: string): Promise<SpikeFindings> {
  console.log('🔬 Starting Spike Validation: Analytics - Price Reconciliation\n');
  console.log(`📅 Period: Last ${SPIKE_CONFIG.periodDays} days`);
  console.log(`🏢 Tenant: ${tenantId}`);
  console.log(`💰 Price Column: ${SPIKE_CONFIG.marketplacePriceColumn}\n`);

  // Setup database URL for tenant
  // Database naming: {tenant}_{shopname}.db (e.g., yumna_bertigamart.db)
  const databasesDir = path.resolve(__dirname, '../config/databases');
  const dbPath = path.join(databasesDir, `${tenantId}_bertigamart.db`);
  
  console.log(`📁 Looking for database: ${dbPath}`);
  
  if (!fs.existsSync(dbPath)) {
    // List available databases
    const files = fs.readdirSync(databasesDir);
    const dbFiles = files.filter(f => f.endsWith('.db') && !f.includes('jobs'));
    throw new Error(
      `Database not found for tenant ${tenantId}: ${dbPath}\n` +
      `Available databases: ${dbFiles.join(', ')}`
    );
  }
  
  // Set DATABASE_URL for this tenant
  process.env.DATABASE_URL = `file:${dbPath}`;
  
  console.log(`✅ Connected to database: ${path.basename(dbPath)}\n`);
  
  // Create Prisma client
  const prisma = new PrismaClient();

  // Calculate date range
  const fromDate = new Date();
  fromDate.setDate(fromDate.getDate() - SPIKE_CONFIG.periodDays);

  // Start performance timer
  const startTime = Date.now();

  // Fetch and analyze orders from both platforms
  console.log('📊 Analyzing Shopee orders...');
  const shopeeResults = await analyzeShopeeOrders(prisma, tenantId, fromDate);
  
  console.log('📊 Analyzing TikTok orders...');
  const tiktokResults = await analyzeTiktokOrders(prisma, tenantId, fromDate);

  // Combine results
  const allResults = [...shopeeResults, ...tiktokResults];

  // Stop timer
  const queryTimeMs = Date.now() - startTime;

  // Calculate summary statistics
  const summary = {
    totalOrders: new Set(allResults.map((r) => r.orderSn)).size,
    totalItems: allResults.length,
    matchedItems: allResults.filter((r) => r.status === 'MATCH').length,
    mismatchedItems: allResults.filter((r) => r.status === 'MISMATCH').length,
    missingInventory: allResults.filter((r) => r.status === 'MISSING_INVENTORY').length,
    missingPrice: allResults.filter((r) => r.status === 'MISSING_PRICE').length,
    totalDiscrepancy: allResults
      .filter((r) => r.status === 'MISMATCH')
      .reduce((sum, r) => sum + r.discrepancy, 0),
  };

  // Performance metrics
  const performance = {
    queryTimeMs,
    itemsPerSecond: allResults.length / (queryTimeMs / 1000),
  };

  // Identify constraints
  const constraints: string[] = [];
  
  if (queryTimeMs > SPIKE_CONFIG.performanceTarget.maxQueryTimeMs) {
    constraints.push(`⚠️ Query time (${queryTimeMs}ms) exceeds target (${SPIKE_CONFIG.performanceTarget.maxQueryTimeMs}ms)`);
  }
  
  if (summary.missingInventory > summary.totalItems * 0.1) {
    constraints.push(`⚠️ High missing inventory rate: ${((summary.missingInventory / summary.totalItems) * 100).toFixed(1)}%`);
  }

  if (summary.missingPrice > summary.totalItems * 0.1) {
    constraints.push(`⚠️ High missing price data rate: ${((summary.missingPrice / summary.totalItems) * 100).toFixed(1)}%`);
  }

  // Generate recommendations
  const recommendations: string[] = [];
  
  if (queryTimeMs > 2000) {
    recommendations.push('💡 Consider adding pagination for large datasets');
    recommendations.push('💡 Implement caching for inventory lookups');
  }

  if (summary.missingInventory > 0) {
    recommendations.push('💡 Add inventory sync validation to prevent missing data');
    recommendations.push('💡 Show clear warnings when inventory data is missing');
  }

  recommendations.push('💡 Add configurable price column selection in settings');
  recommendations.push('💡 Implement export functionality (CSV/Excel)');
  recommendations.push('💡 Add filters for date range, platform, status');

  // Get sample mismatches (top 10 by discrepancy)
  const sampleMismatches = allResults
    .filter((r) => r.status === 'MISMATCH')
    .sort((a, b) => b.discrepancy - a.discrepancy)
    .slice(0, 10);

  return {
    summary,
    performance,
    constraints,
    recommendations,
    sampleMismatches,
  };
}

// ============================================================
// REPORT GENERATION
// ============================================================

function printReport(findings: SpikeFindings): void {
  console.log('\n' + '='.repeat(80));
  console.log('📊 SPIKE VALIDATION FINDINGS');
  console.log('='.repeat(80) + '\n');

  // Summary
  console.log('📈 SUMMARY STATISTICS:');
  console.log(`   Total Orders:        ${findings.summary.totalOrders}`);
  console.log(`   Total Items:         ${findings.summary.totalItems}`);
  console.log(`   ✅ Matched:          ${findings.summary.matchedItems} (${((findings.summary.matchedItems / findings.summary.totalItems) * 100).toFixed(1)}%)`);
  console.log(`   ❌ Mismatched:       ${findings.summary.mismatchedItems} (${((findings.summary.mismatchedItems / findings.summary.totalItems) * 100).toFixed(1)}%)`);
  console.log(`   🔍 Missing Inventory: ${findings.summary.missingInventory} (${((findings.summary.missingInventory / findings.summary.totalItems) * 100).toFixed(1)}%)`);
  console.log(`   💰 Missing Price:    ${findings.summary.missingPrice} (${((findings.summary.missingPrice / findings.summary.totalItems) * 100).toFixed(1)}%)`);
  console.log(`   💸 Total Discrepancy: Rp ${findings.summary.totalDiscrepancy.toLocaleString('id-ID')}\n`);

  // Performance
  console.log('⚡ PERFORMANCE METRICS:');
  console.log(`   Query Time:       ${findings.performance.queryTimeMs.toLocaleString('id-ID')}ms`);
  console.log(`   Items/Second:     ${findings.performance.itemsPerSecond.toFixed(2)}`);
  console.log(`   Target:           < ${SPIKE_CONFIG.performanceTarget.maxQueryTimeMs}ms`);
  console.log(`   Status:           ${findings.performance.queryTimeMs < SPIKE_CONFIG.performanceTarget.maxQueryTimeMs ? '✅ PASS' : '❌ FAIL'}\n`);

  // Constraints
  if (findings.constraints.length > 0) {
    console.log('⚠️ CONSTRAINTS IDENTIFIED:');
    findings.constraints.forEach((c) => console.log(`   ${c}`));
    console.log('');
  }

  // Recommendations
  console.log('💡 RECOMMENDATIONS:');
  findings.recommendations.forEach((r) => console.log(`   ${r}`));
  console.log('');

  // Sample mismatches
  if (findings.sampleMismatches.length > 0) {
    console.log('❌ SAMPLE MISMATCHES (Top 10 by Discrepancy):');
    console.log('   ' + '-'.repeat(76));
    console.log('   Platform | Order SN      | SKU        | Order Price | Expected | Discrepancy');
    console.log('   ' + '-'.repeat(76));
    
    findings.sampleMismatches.forEach((item) => {
      console.log(
        `   ${item.platform.padEnd(8)} | ${item.orderSn.padEnd(13)} | ${item.sku.padEnd(10)} | ` +
        `${item.orderPrice.toLocaleString('id-ID').padStart(11)} | ` +
        `${(item.expectedMinPrice || 0).toLocaleString('id-ID').padStart(8)} | ` +
        `${item.discrepancy.toLocaleString('id-ID').padStart(11)}`
      );
    });
    console.log('   ' + '-'.repeat(76) + '\n');
  }

  console.log('='.repeat(80));
  console.log('✅ FEASIBILITY ASSESSMENT: ' + (findings.constraints.length === 0 ? 'APPROVED ✅' : 'APPROVED WITH CONSTRAINTS ⚠️'));
  console.log('='.repeat(80) + '\n');
}

// ============================================================
// ENTRY POINT
// ============================================================

async function main() {
  try {
    // Get tenant ID from environment or use default
    const tenantId = process.env.TENANT_ID || 'yumna'; // Default to yumna tenant
    
    console.log('🚀 Initializing spike validation...\n');
    
    // Run spike
    const findings = await runSpike(tenantId);
    
    // Print report
    printReport(findings);
    
    // Save findings to file for reference
    const outputPath = path.join(__dirname, 'spike-findings-price-reconciliation.json');
    fs.writeFileSync(outputPath, JSON.stringify(findings, null, 2));
    console.log(`💾 Findings saved to: ${outputPath}\n`);
    
    process.exit(0);
  } catch (error) {
    console.error('❌ Spike validation failed:', error);
    process.exit(1);
  }
}

// Run if executed directly
if (require.main === module) {
  main();
}

export { runSpike, SpikeFindings, OrderItemAnalysis };
