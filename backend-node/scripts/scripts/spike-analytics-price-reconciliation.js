"use strict";
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
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __generator = (this && this.__generator) || function (thisArg, body) {
    var _ = { label: 0, sent: function() { if (t[0] & 1) throw t[1]; return t[1]; }, trys: [], ops: [] }, f, y, t, g = Object.create((typeof Iterator === "function" ? Iterator : Object).prototype);
    return g.next = verb(0), g["throw"] = verb(1), g["return"] = verb(2), typeof Symbol === "function" && (g[Symbol.iterator] = function() { return this; }), g;
    function verb(n) { return function (v) { return step([n, v]); }; }
    function step(op) {
        if (f) throw new TypeError("Generator is already executing.");
        while (g && (g = 0, op[0] && (_ = 0)), _) try {
            if (f = 1, y && (t = op[0] & 2 ? y["return"] : op[0] ? y["throw"] || ((t = y["return"]) && t.call(y), 0) : y.next) && !(t = t.call(y, op[1])).done) return t;
            if (y = 0, t) op = [op[0] & 2, t.value];
            switch (op[0]) {
                case 0: case 1: t = op; break;
                case 4: _.label++; return { value: op[1], done: false };
                case 5: _.label++; y = op[1]; op = [0]; continue;
                case 7: op = _.ops.pop(); _.trys.pop(); continue;
                default:
                    if (!(t = _.trys, t = t.length > 0 && t[t.length - 1]) && (op[0] === 6 || op[0] === 2)) { _ = 0; continue; }
                    if (op[0] === 3 && (!t || (op[1] > t[0] && op[1] < t[3]))) { _.label = op[1]; break; }
                    if (op[0] === 6 && _.label < t[1]) { _.label = t[1]; t = op; break; }
                    if (t && _.label < t[2]) { _.label = t[2]; _.ops.push(op); break; }
                    if (t[2]) _.ops.pop();
                    _.trys.pop(); continue;
            }
            op = body.call(thisArg, _);
        } catch (e) { op = [6, e]; y = 0; } finally { f = t = 0; }
        if (op[0] & 5) throw op[1]; return { value: op[0] ? op[1] : void 0, done: true };
    }
};
var __spreadArray = (this && this.__spreadArray) || function (to, from, pack) {
    if (pack || arguments.length === 2) for (var i = 0, l = from.length, ar; i < l; i++) {
        if (ar || !(i in from)) {
            if (!ar) ar = Array.prototype.slice.call(from, 0, i);
            ar[i] = from[i];
        }
    }
    return to.concat(ar || Array.prototype.slice.call(from));
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.runSpike = runSpike;
var tenantManagementService_1 = require("../src/services/tenantManagementService");
// ============================================================
// CONFIGURATION
// ============================================================
var SPIKE_CONFIG = {
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
function calculateExpectedMinPrice(hargaMarketplace) {
    var _a = SPIKE_CONFIG.formula, deduction = _a.deduction, multiplier = _a.multiplier;
    return (hargaMarketplace - deduction) * multiplier;
}
/**
 * Compare order price vs expected minimum price
 */
function comparePrice(orderPrice, hargaMarketplace) {
    if (hargaMarketplace === null) {
        return {
            status: 'MISSING_PRICE',
            expectedMinPrice: null,
            discrepancy: 0,
        };
    }
    var expectedMinPrice = calculateExpectedMinPrice(hargaMarketplace);
    var discrepancy = expectedMinPrice - orderPrice;
    var status = orderPrice >= expectedMinPrice ? 'MATCH' : 'MISMATCH';
    return {
        status: status,
        expectedMinPrice: expectedMinPrice,
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
function lookupInventory(prisma, tenantId, sku) {
    return __awaiter(this, void 0, void 0, function () {
        var record, inventoryData, error_1;
        return __generator(this, function (_a) {
            switch (_a.label) {
                case 0:
                    _a.trys.push([0, 2, , 3]);
                    return [4 /*yield*/, prisma.inventoryRecord.findFirst({
                            where: {
                                tenantId: tenantId,
                                keyValue: sku,
                                keyColumnName: 'SKU', // Assuming SKU is the key column
                            },
                        })];
                case 1:
                    record = _a.sent();
                    if (!record) {
                        return [2 /*return*/, null];
                    }
                    inventoryData = JSON.parse(record.data);
                    return [2 /*return*/, inventoryData];
                case 2:
                    error_1 = _a.sent();
                    console.error("Error looking up inventory for SKU ".concat(sku, ":"), error_1);
                    return [2 /*return*/, null];
                case 3: return [2 /*return*/];
            }
        });
    });
}
// ============================================================
// ORDER FETCHING & ANALYSIS
// ============================================================
/**
 * Fetch and analyze Shopee orders
 */
function analyzeShopeeOrders(prisma, tenantId, fromDate) {
    return __awaiter(this, void 0, void 0, function () {
        var results, orders, _i, orders_1, order, _a, _b, item, sku, inventoryData, hargaMarketplace, comparison;
        return __generator(this, function (_c) {
            switch (_c.label) {
                case 0:
                    results = [];
                    return [4 /*yield*/, prisma.shopeeOrder.findMany({
                            where: {
                                tenantId: tenantId,
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
                        })];
                case 1:
                    orders = _c.sent();
                    console.log("\uD83D\uDCE6 Found ".concat(orders.length, " Shopee orders"));
                    _i = 0, orders_1 = orders;
                    _c.label = 2;
                case 2:
                    if (!(_i < orders_1.length)) return [3 /*break*/, 7];
                    order = orders_1[_i];
                    _a = 0, _b = order.items;
                    _c.label = 3;
                case 3:
                    if (!(_a < _b.length)) return [3 /*break*/, 6];
                    item = _b[_a];
                    sku = item.itemSku || item.modelSku || '';
                    if (!sku) {
                        console.warn("\u26A0\uFE0F No SKU for Shopee order ".concat(order.orderSn, ", item ").concat(item.itemId));
                        return [3 /*break*/, 5];
                    }
                    return [4 /*yield*/, lookupInventory(prisma, tenantId, sku)];
                case 4:
                    inventoryData = _c.sent();
                    if (!inventoryData) {
                        results.push({
                            platform: 'shopee',
                            orderSn: order.orderSn,
                            sku: sku,
                            productName: item.itemName || 'Unknown',
                            orderPrice: item.price || 0,
                            expectedMinPrice: null,
                            hargaMarketplace: null,
                            status: 'MISSING_INVENTORY',
                            discrepancy: 0,
                            createdAt: order.createdAt,
                        });
                        return [3 /*break*/, 5];
                    }
                    hargaMarketplace = inventoryData[SPIKE_CONFIG.marketplacePriceColumn] || null;
                    comparison = comparePrice(item.price || 0, hargaMarketplace);
                    results.push({
                        platform: 'shopee',
                        orderSn: order.orderSn,
                        sku: sku,
                        productName: item.itemName || inventoryData['Nama Barang'] || 'Unknown',
                        orderPrice: item.price || 0,
                        expectedMinPrice: comparison.expectedMinPrice,
                        hargaMarketplace: hargaMarketplace,
                        status: comparison.status,
                        discrepancy: comparison.discrepancy,
                        createdAt: order.createdAt,
                    });
                    _c.label = 5;
                case 5:
                    _a++;
                    return [3 /*break*/, 3];
                case 6:
                    _i++;
                    return [3 /*break*/, 2];
                case 7: return [2 /*return*/, results];
            }
        });
    });
}
/**
 * Fetch and analyze TikTok orders
 */
function analyzeTiktokOrders(prisma, tenantId, fromDate) {
    return __awaiter(this, void 0, void 0, function () {
        var results, orders, _i, orders_2, order, _a, _b, item, sku, inventoryData, hargaMarketplace, comparison;
        return __generator(this, function (_c) {
            switch (_c.label) {
                case 0:
                    results = [];
                    return [4 /*yield*/, prisma.tiktokOrder.findMany({
                            where: {
                                tenantId: tenantId,
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
                        })];
                case 1:
                    orders = _c.sent();
                    console.log("\uD83D\uDCE6 Found ".concat(orders.length, " TikTok orders"));
                    _i = 0, orders_2 = orders;
                    _c.label = 2;
                case 2:
                    if (!(_i < orders_2.length)) return [3 /*break*/, 7];
                    order = orders_2[_i];
                    _a = 0, _b = order.items;
                    _c.label = 3;
                case 3:
                    if (!(_a < _b.length)) return [3 /*break*/, 6];
                    item = _b[_a];
                    sku = item.sellerSku || item.skuId || '';
                    if (!sku) {
                        console.warn("\u26A0\uFE0F No SKU for TikTok order ".concat(order.orderSn, ", item ").concat(item.productId));
                        return [3 /*break*/, 5];
                    }
                    return [4 /*yield*/, lookupInventory(prisma, tenantId, sku)];
                case 4:
                    inventoryData = _c.sent();
                    if (!inventoryData) {
                        results.push({
                            platform: 'tiktok',
                            orderSn: order.orderSn,
                            sku: sku,
                            productName: item.productName || 'Unknown',
                            orderPrice: item.price || 0,
                            expectedMinPrice: null,
                            hargaMarketplace: null,
                            status: 'MISSING_INVENTORY',
                            discrepancy: 0,
                            createdAt: order.createdAt,
                        });
                        return [3 /*break*/, 5];
                    }
                    hargaMarketplace = inventoryData[SPIKE_CONFIG.marketplacePriceColumn] || null;
                    comparison = comparePrice(item.price || 0, hargaMarketplace);
                    results.push({
                        platform: 'tiktok',
                        orderSn: order.orderSn,
                        sku: sku,
                        productName: item.productName || inventoryData['Nama Barang'] || 'Unknown',
                        orderPrice: item.price || 0,
                        expectedMinPrice: comparison.expectedMinPrice,
                        hargaMarketplace: hargaMarketplace,
                        status: comparison.status,
                        discrepancy: comparison.discrepancy,
                        createdAt: order.createdAt,
                    });
                    _c.label = 5;
                case 5:
                    _a++;
                    return [3 /*break*/, 3];
                case 6:
                    _i++;
                    return [3 /*break*/, 2];
                case 7: return [2 /*return*/, results];
            }
        });
    });
}
// ============================================================
// MAIN SPIKE EXECUTION
// ============================================================
function runSpike(tenantId) {
    return __awaiter(this, void 0, void 0, function () {
        var prisma, fromDate, startTime, shopeeResults, tiktokResults, allResults, queryTimeMs, summary, performance, constraints, recommendations, sampleMismatches;
        return __generator(this, function (_a) {
            switch (_a.label) {
                case 0:
                    console.log('🔬 Starting Spike Validation: Analytics - Price Reconciliation\n');
                    console.log("\uD83D\uDCC5 Period: Last ".concat(SPIKE_CONFIG.periodDays, " days"));
                    console.log("\uD83C\uDFE2 Tenant: ".concat(tenantId));
                    console.log("\uD83D\uDCB0 Price Column: ".concat(SPIKE_CONFIG.marketplacePriceColumn, "\n"));
                    prisma = tenantManagementService_1.tenantManagement.getPrismaClient(tenantId);
                    if (!prisma) {
                        throw new Error("Cannot get Prisma client for tenant: ".concat(tenantId));
                    }
                    fromDate = new Date();
                    fromDate.setDate(fromDate.getDate() - SPIKE_CONFIG.periodDays);
                    startTime = Date.now();
                    // Fetch and analyze orders from both platforms
                    console.log('📊 Analyzing Shopee orders...');
                    return [4 /*yield*/, analyzeShopeeOrders(prisma, tenantId, fromDate)];
                case 1:
                    shopeeResults = _a.sent();
                    console.log('📊 Analyzing TikTok orders...');
                    return [4 /*yield*/, analyzeTiktokOrders(prisma, tenantId, fromDate)];
                case 2:
                    tiktokResults = _a.sent();
                    allResults = __spreadArray(__spreadArray([], shopeeResults, true), tiktokResults, true);
                    queryTimeMs = Date.now() - startTime;
                    summary = {
                        totalOrders: new Set(allResults.map(function (r) { return r.orderSn; })).size,
                        totalItems: allResults.length,
                        matchedItems: allResults.filter(function (r) { return r.status === 'MATCH'; }).length,
                        mismatchedItems: allResults.filter(function (r) { return r.status === 'MISMATCH'; }).length,
                        missingInventory: allResults.filter(function (r) { return r.status === 'MISSING_INVENTORY'; }).length,
                        missingPrice: allResults.filter(function (r) { return r.status === 'MISSING_PRICE'; }).length,
                        totalDiscrepancy: allResults
                            .filter(function (r) { return r.status === 'MISMATCH'; })
                            .reduce(function (sum, r) { return sum + r.discrepancy; }, 0),
                    };
                    performance = {
                        queryTimeMs: queryTimeMs,
                        itemsPerSecond: allResults.length / (queryTimeMs / 1000),
                    };
                    constraints = [];
                    if (queryTimeMs > SPIKE_CONFIG.performanceTarget.maxQueryTimeMs) {
                        constraints.push("\u26A0\uFE0F Query time (".concat(queryTimeMs, "ms) exceeds target (").concat(SPIKE_CONFIG.performanceTarget.maxQueryTimeMs, "ms)"));
                    }
                    if (summary.missingInventory > summary.totalItems * 0.1) {
                        constraints.push("\u26A0\uFE0F High missing inventory rate: ".concat(((summary.missingInventory / summary.totalItems) * 100).toFixed(1), "%"));
                    }
                    if (summary.missingPrice > summary.totalItems * 0.1) {
                        constraints.push("\u26A0\uFE0F High missing price data rate: ".concat(((summary.missingPrice / summary.totalItems) * 100).toFixed(1), "%"));
                    }
                    recommendations = [];
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
                    sampleMismatches = allResults
                        .filter(function (r) { return r.status === 'MISMATCH'; })
                        .sort(function (a, b) { return b.discrepancy - a.discrepancy; })
                        .slice(0, 10);
                    return [2 /*return*/, {
                            summary: summary,
                            performance: performance,
                            constraints: constraints,
                            recommendations: recommendations,
                            sampleMismatches: sampleMismatches,
                        }];
            }
        });
    });
}
// ============================================================
// REPORT GENERATION
// ============================================================
function printReport(findings) {
    console.log('\n' + '='.repeat(80));
    console.log('📊 SPIKE VALIDATION FINDINGS');
    console.log('='.repeat(80) + '\n');
    // Summary
    console.log('📈 SUMMARY STATISTICS:');
    console.log("   Total Orders:        ".concat(findings.summary.totalOrders));
    console.log("   Total Items:         ".concat(findings.summary.totalItems));
    console.log("   \u2705 Matched:          ".concat(findings.summary.matchedItems, " (").concat(((findings.summary.matchedItems / findings.summary.totalItems) * 100).toFixed(1), "%)"));
    console.log("   \u274C Mismatched:       ".concat(findings.summary.mismatchedItems, " (").concat(((findings.summary.mismatchedItems / findings.summary.totalItems) * 100).toFixed(1), "%)"));
    console.log("   \uD83D\uDD0D Missing Inventory: ".concat(findings.summary.missingInventory, " (").concat(((findings.summary.missingInventory / findings.summary.totalItems) * 100).toFixed(1), "%)"));
    console.log("   \uD83D\uDCB0 Missing Price:    ".concat(findings.summary.missingPrice, " (").concat(((findings.summary.missingPrice / findings.summary.totalItems) * 100).toFixed(1), "%)"));
    console.log("   \uD83D\uDCB8 Total Discrepancy: Rp ".concat(findings.summary.totalDiscrepancy.toLocaleString('id-ID'), "\n"));
    // Performance
    console.log('⚡ PERFORMANCE METRICS:');
    console.log("   Query Time:       ".concat(findings.performance.queryTimeMs.toLocaleString('id-ID'), "ms"));
    console.log("   Items/Second:     ".concat(findings.performance.itemsPerSecond.toFixed(2)));
    console.log("   Target:           < ".concat(SPIKE_CONFIG.performanceTarget.maxQueryTimeMs, "ms"));
    console.log("   Status:           ".concat(findings.performance.queryTimeMs < SPIKE_CONFIG.performanceTarget.maxQueryTimeMs ? '✅ PASS' : '❌ FAIL', "\n"));
    // Constraints
    if (findings.constraints.length > 0) {
        console.log('⚠️ CONSTRAINTS IDENTIFIED:');
        findings.constraints.forEach(function (c) { return console.log("   ".concat(c)); });
        console.log('');
    }
    // Recommendations
    console.log('💡 RECOMMENDATIONS:');
    findings.recommendations.forEach(function (r) { return console.log("   ".concat(r)); });
    console.log('');
    // Sample mismatches
    if (findings.sampleMismatches.length > 0) {
        console.log('❌ SAMPLE MISMATCHES (Top 10 by Discrepancy):');
        console.log('   ' + '-'.repeat(76));
        console.log('   Platform | Order SN      | SKU        | Order Price | Expected | Discrepancy');
        console.log('   ' + '-'.repeat(76));
        findings.sampleMismatches.forEach(function (item) {
            console.log("   ".concat(item.platform.padEnd(8), " | ").concat(item.orderSn.padEnd(13), " | ").concat(item.sku.padEnd(10), " | ") +
                "".concat(item.orderPrice.toLocaleString('id-ID').padStart(11), " | ") +
                "".concat((item.expectedMinPrice || 0).toLocaleString('id-ID').padStart(8), " | ") +
                "".concat(item.discrepancy.toLocaleString('id-ID').padStart(11)));
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
function main() {
    return __awaiter(this, void 0, void 0, function () {
        var tenantId, findings, fs, outputPath, error_2;
        return __generator(this, function (_a) {
            switch (_a.label) {
                case 0:
                    _a.trys.push([0, 2, , 3]);
                    tenantId = process.env.TENANT_ID || 'bertigamart';
                    console.log('🚀 Initializing spike validation...\n');
                    return [4 /*yield*/, runSpike(tenantId)];
                case 1:
                    findings = _a.sent();
                    // Print report
                    printReport(findings);
                    fs = require('fs');
                    outputPath = './scripts/spike-findings-price-reconciliation.json';
                    fs.writeFileSync(outputPath, JSON.stringify(findings, null, 2));
                    console.log("\uD83D\uDCBE Findings saved to: ".concat(outputPath, "\n"));
                    process.exit(0);
                    return [3 /*break*/, 3];
                case 2:
                    error_2 = _a.sent();
                    console.error('❌ Spike validation failed:', error_2);
                    process.exit(1);
                    return [3 /*break*/, 3];
                case 3: return [2 /*return*/];
            }
        });
    });
}
// Run if executed directly
if (require.main === module) {
    main();
}
