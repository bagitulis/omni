#!/usr/bin/env ts-node

/**
 * TESTING GUIDE: Lazada Order Document API
 *
 * This file provides step-by-step instructions and examples
 * for testing the Lazada /order/document/get endpoint
 */

// ============================================================================
// STEP 1: Understand the Mechanism
// ============================================================================

/*
LAZADA ORDER DOCUMENT API MECHANISM:

1. AUTHENTICATION
   - Uses OAuth2 Access Token (7-day expiry)
   - Token stored in tenant database after OAuth flow
   - Needs refresh if expired (via refresh token)

2. SIGNATURE REQUIREMENT
   - Every request must be signed with HMAC-SHA256
   - Signature = HMAC-SHA256(appSecret, signString)
   - signString = apiPath + sortedParamsString
   
   Example:
   - Path: "/order/document/get"
   - Params: {app_key: "123", doc_type: "shippingLabel", order_item_ids: "[279709]"}
   - Sorted: app_key, doc_type, order_item_ids
   - SignString: "/order/document/getapp_key123doc_typeshippingLabelorder_item_ids[279709]"
   - Signature: HMAC-SHA256(appSecret, signString).toHex().toUpperCase()

3. REQUEST PARAMETERS
   - doc_type: One of "shippingLabel", "invoice", "waybill"
   - order_item_ids: JSON string array of item IDs like "[279709, 279710]"
   - app_key: Lazada app key
   - access_token: OAuth access token
   - timestamp: Current unix timestamp
   - sign_method: "sha256"
   - sign: Generated HMAC-SHA256 signature

4. RESPONSE
   - Code "0": Success - contains base64-encoded file
   - Non-zero code: Error with message
   - File format: Base64-encoded (decode to get actual content)
   - MIME type: Tells you file type (text/html, application/pdf, etc.)

5. COUNTRY-SPECIFIC GATEWAYS
   - ID (Indonesia): https://api.lazada.co.id/rest
   - TH (Thailand): https://api.lazada.co.th/rest
   - SG (Singapore): https://api.lazada.sg/rest
   - MY (Malaysia): https://api.lazada.com.my/rest
   - VN (Vietnam): https://api.lazada.vn/rest
   - PH (Philippines): https://api.lazada.com.ph/rest
*/

// ============================================================================
// STEP 2: Configuration Check
// ============================================================================

import { LazadaConfigManager } from "./src/config/managers/lazadaConfigManager";
import { LazadaAPIClient } from "./src/api/clients/lazadaAPIClient";

async function checkConfiguration() {
  console.log("\n=== STEP 1: Check Configuration ===\n");

  const configManager = new LazadaConfigManager();
  await configManager.loadConfig();

  console.log("✅ Configuration loaded from database:\n");
  console.log("   App Key:", configManager.appKey.substring(0, 5) + "***");
  console.log(
    "   App Secret:",
    configManager.appSecret.substring(0, 5) + "***"
  );
  console.log(
    "   Access Token:",
    configManager.accessToken
      ? configManager.accessToken.substring(0, 10) + "***"
      : "❌ NOT FOUND"
  );
  console.log("   Country:", configManager.country);
  console.log(
    "   Token Expires:",
    configManager.tokenExpiry
      ? new Date(configManager.tokenExpiry).toISOString()
      : "Unknown"
  );

  if (!configManager.accessToken) {
    console.log("\n❌ ERROR: No access token found!");
    console.log("   ACTION: Complete OAuth flow in settings first");
    return false;
  }

  return configManager;
}

// ============================================================================
// STEP 3: Test Signature Generation
// ============================================================================

async function testSignatureGeneration(configManager: any) {
  console.log("\n=== STEP 2: Test Signature Generation ===\n");

  const testParams: Record<string, string> = {
    app_key: configManager.appKey,
    doc_type: "shippingLabel",
    order_item_ids: "[279709]",
    timestamp: String(Math.floor(Date.now() / 1000)),
    access_token: configManager.accessToken,
    sign_method: "sha256",
  };

  console.log("Test parameters:");
  console.log(JSON.stringify(testParams, null, 2));

  // Manually calculate signature to verify
  const crypto = require("crypto");
  const sortedKeys = Object.keys(testParams)
    .filter((k) => k !== "sign")
    .sort();
  const paramString = sortedKeys.map((k) => `${k}${testParams[k]}`).join("");
  const signString = "/order/document/get" + paramString;

  const signature = crypto
    .createHmac("sha256", configManager.appSecret)
    .update(signString)
    .digest("hex")
    .toUpperCase();

  console.log("\nGenerated signature:");
  console.log("   ", signature.substring(0, 20) + "...");
  console.log("\nSignature generation: ✅ SUCCESS");
}

// ============================================================================
// STEP 4: Make API Request
// ============================================================================

async function testAPIRequest(configManager: any, orderItemIds: number[]) {
  console.log("\n=== STEP 3: Make API Request ===\n");

  const apiClient = new LazadaAPIClient(configManager);

  console.log(
    `📋 Requesting document for item IDs: ${orderItemIds.join(", ")}`
  );

  try {
    const response = await apiClient.request("/order/document/get", "GET", {
      doc_type: "shippingLabel",
      order_item_ids: JSON.stringify(orderItemIds),
    });

    if (response.code !== "0") {
      console.log(`\n❌ API Error: ${response.code}`);
      console.log(`   Message: ${response.message}`);

      // Handle specific errors
      if (response.code === "20001") {
        console.log(
          "   ACTION: Access token expired, need to refresh via OAuth"
        );
      } else if (response.code === "40001") {
        console.log("   ACTION: Invalid order_item_ids format");
      } else if (response.code === "40002") {
        console.log("   ACTION: Document not available for this order");
      }

      return null;
    }

    const { document } = response.data;

    console.log("\n✅ Document Retrieved Successfully!");
    console.log(`   Response ID: ${response.request_id}`);
    console.log(`   Document Type: ${document.document_type}`);
    console.log(`   MIME Type: ${document.mime_type}`);
    console.log(`   File Size: ${document.file.length} bytes (base64 encoded)`);
    console.log(
      `   Decoded Size: ${Math.round(document.file.length * 0.75)} bytes`
    );

    return document;
  } catch (error: any) {
    console.log(`\n❌ Request Failed: ${error.message}`);

    if (error.message.includes("404")) {
      console.log("   ACTION: Check if item IDs are valid");
    } else if (error.message.includes("403")) {
      console.log("   ACTION: Check if access token has permission");
    } else if (error.message.includes("401")) {
      console.log("   ACTION: Re-authenticate with Lazada OAuth");
    }

    return null;
  }
}

// ============================================================================
// STEP 5: Decode and Save Document
// ============================================================================

import fs from "fs";
import path from "path";

async function saveDocument(document: any, outputDir: string) {
  console.log("\n=== STEP 4: Decode and Save Document ===\n");

  if (!document) {
    console.log("❌ No document to save");
    return;
  }

  // Create output directory if needed
  if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir, { recursive: true });
  }

  // Determine file extension
  const extensionMap: Record<string, string> = {
    "text/html": ".html",
    "application/pdf": ".pdf",
    "image/png": ".png",
    "image/jpeg": ".jpg",
  };

  const ext = extensionMap[document.mime_type] || ".bin";
  const fileName = path.join(
    outputDir,
    `order-document-${document.document_type}-${Date.now()}${ext}`
  );

  try {
    // Decode base64 to binary
    const buffer = Buffer.from(document.file, "base64");

    // Save to file
    fs.writeFileSync(fileName, buffer as any);

    console.log(`✅ Document Saved!`);
    console.log(`   Path: ${fileName}`);
    console.log(`   Size: ${buffer.length} bytes`);
    console.log(`   Type: ${document.mime_type}`);

    // For HTML files, show preview hint
    if (document.mime_type === "text/html") {
      console.log(`\n   💡 To view: open "${fileName}" in your browser`);
    }

    return fileName;
  } catch (error) {
    console.log(`❌ Failed to save: ${error}`);
    return null;
  }
}

// ============================================================================
// STEP 6: Full Test Workflow
// ============================================================================

async function runFullTest() {
  console.log(
    "\n╔════════════════════════════════════════════════════════════╗"
  );
  console.log("║  LAZADA ORDER DOCUMENT API - FULL TEST WORKFLOW           ║");
  console.log("║  Testing: GET /order/document/get                         ║");
  console.log("╚════════════════════════════════════════════════════════════╝");

  try {
    // Step 1: Check configuration
    const configManager = await checkConfiguration();
    if (!configManager) {
      console.log(
        "\n❌ Configuration check failed. Cannot proceed with tests."
      );
      process.exit(1);
    }

    // Step 2: Test signature generation
    await testSignatureGeneration(configManager);

    // Step 3: Make API request
    // NOTE: Replace with real order item ID from your Lazada shop
    const testOrderItemId = 279709; // CHANGE THIS to real ID
    console.log(`\n⚠️  USING TEST ITEM ID: ${testOrderItemId}`);
    console.log("   ℹ️  Replace with real order item ID for actual testing\n");

    const document = await testAPIRequest(configManager, [testOrderItemId]);

    if (document) {
      // Step 4: Save document
      const outputDir = path.join(__dirname, "_bmad-output", "lazada-docs");
      await saveDocument(document, outputDir);
    }

    console.log(
      "\n╔════════════════════════════════════════════════════════════╗"
    );
    console.log(
      "║  ✅ TEST WORKFLOW COMPLETED                               ║"
    );
    console.log(
      "╚════════════════════════════════════════════════════════════╝\n"
    );
  } catch (error) {
    console.log(`\n❌ Test failed with error: ${error}`);
    process.exit(1);
  }
}

// ============================================================================
// USAGE EXAMPLES
// ============================================================================

/*
QUICK START EXAMPLES:

1. RUN FULL TEST:
   $ npx ts-node test-lazada-order-document.ts

2. MANUAL TEST IN CODE:
   const configManager = new LazadaConfigManager();
   await configManager.loadConfig();
   
   const apiClient = new LazadaAPIClient(configManager);
   const response = await apiClient.request("/order/document/get", "GET", {
     doc_type: "shippingLabel",
     order_item_ids: "[279709]"
   });

3. FROM FRONTEND (via backend endpoint):
   GET /api/lazada/order/document?orderItemIds=279709&docType=shippingLabel
   
   Response: Binary file download with Content-Disposition attachment

4. CHECK AVAILABLE DOCUMENTS:
   GET /api/lazada/order/documents/metadata?orderItemIds=279709
   
   Response: JSON array showing which document types are available

GETTING REAL ORDER ITEM IDS:
1. Fetch orders: await lazadaOrderManager.getOrderList("processed")
2. Get items: await lazadaOrderManager.getOrderItems(orderIds)
3. Use order_item_id field from response

COMMON TEST SCENARIOS:

Scenario 1: Shipping Label
- Doc Type: shippingLabel
- Use Case: Print and stick on package
- Response: Usually HTML format

Scenario 2: Invoice
- Doc Type: invoice
- Use Case: Tax records, customer documentation
- Response: Usually HTML or PDF

Scenario 3: Waybill
- Doc Type: waybill
- Use Case: Logistics tracking
- Response: Usually HTML

ERROR TROUBLESHOOTING:

Code 0: ✅ Success
Code 1: Signature error - check appSecret
Code 20001: Token expired - re-authenticate
Code 40001: Invalid item IDs format
Code 40002: Document not available - order not ready

TOKEN EXPIRATION:
- Access tokens valid for 7 days
- Check configManager.tokenExpiresAt
- Use refresh token to get new token if expired
- Implement auto-refresh in config manager
*/

// Run tests if executed directly
if (require.main === module) {
  runFullTest().catch((error) => {
    console.error("Fatal error:", error);
    process.exit(1);
  });
}

export {
  checkConfiguration,
  testSignatureGeneration,
  testAPIRequest,
  saveDocument,
  runFullTest,
};
