/**
 * Test Lazada Order Document API
 * Mechanism: GET /order/document/get
 *
 * Purpose: Retrieve order-related documents (invoices, shipping labels)
 *
 * API Mechanism:
 * - Endpoint: GET /order/document/get
 * - Required Auth: Access Token (OAuth2)
 * - Parameters:
 *   - doc_type: Type of document (shippingLabel, invoice, waybill)
 *   - order_item_ids: JSON array of order item IDs [id1, id2, ...]
 * - Response: Base64-encoded file data with mime_type
 */

import { LazadaConfigManager } from "./src/config/managers/lazadaConfigManager";
import { LazadaAPIClient } from "./src/api/clients/lazadaAPIClient";
import fs from "fs";
import path from "path";

interface DocumentResponse {
  code: string;
  data: {
    document: {
      file: string; // Base64-encoded file content
      mime_type: string; // e.g., "text/html", "application/pdf"
      document_type: string; // e.g., "shippingLabel"
    };
  };
  request_id: string;
}

interface DocumentTestRequest {
  orderItemIds: number[];
  docType: "shippingLabel" | "invoice" | "waybill";
  outputPath?: string;
}

class LazadaOrderDocumentTester {
  private configManager: LazadaConfigManager;
  private apiClient: LazadaAPIClient;

  constructor() {
    this.configManager = new LazadaConfigManager();
    this.apiClient = new LazadaAPIClient(this.configManager);
  }

  /**
   * Initialize and load config from database
   */
  async initialize(): Promise<void> {
    console.log("🔧 Initializing Lazada Order Document Tester...");
    await this.configManager.loadConfig();

    if (!this.configManager.accessToken) {
      throw new Error("❌ No access token found. Please authenticate first.");
    }

    console.log("✅ Config loaded successfully");
    console.log(`   Country: ${this.configManager.country}`);
    console.log(
      `   Access Token: ${this.configManager.accessToken.substring(0, 10)}...`
    );
  }

  /**
   * Fetch order document from Lazada
   *
   * API Mechanism Details:
   * - Method: GET
   * - Path: /order/document/get
   * - Signature: Required (HMAC-SHA256)
   * - Auth: accessToken (OAuth2)
   *
   * Parameters:
   * - doc_type: Document type identifier
   * - order_item_ids: JSON array string of item IDs
   *
   * Response:
   * {
   *   "code": "0",
   *   "data": {
   *     "document": {
   *       "file": "base64EncodedContent",
   *       "mime_type": "text/html",
   *       "document_type": "shippingLabel"
   *     }
   *   },
   *   "request_id": "xxxxx"
   * }
   */
  async fetchDocument(request: DocumentTestRequest): Promise<DocumentResponse> {
    console.log("\n📋 Fetching Order Document...");
    console.log(`   Document Type: ${request.docType}`);
    console.log(`   Order Item IDs: ${request.orderItemIds.join(", ")}`);

    try {
      // Prepare parameters according to Lazada API spec
      const params = {
        doc_type: request.docType,
        order_item_ids: JSON.stringify(request.orderItemIds), // Must be JSON string
      };

      console.log(`\n🔍 Request Parameters:`);
      console.log(`   ${JSON.stringify(params, null, 2)}`);

      // Make API request
      const response = await this.apiClient.request(
        "/order/document/get",
        "GET",
        params
      );

      if (response.code !== "0") {
        throw new Error(
          `API Error: ${response.code} - ${response.message || "Unknown error"}`
        );
      }

      console.log("\n✅ Document retrieved successfully");
      console.log(`   Response ID: ${response.request_id}`);
      console.log(`   MIME Type: ${response.data.document.mime_type}`);
      console.log(`   File Size: ${response.data.document.file.length} bytes`);

      return response;
    } catch (error) {
      console.error(`\n❌ Failed to fetch document:`, error);
      throw error;
    }
  }

  /**
   * Save document to file (decode base64)
   */
  async saveDocument(
    response: DocumentResponse,
    outputPath?: string
  ): Promise<string> {
    const { file, mime_type, document_type } = response.data.document;

    // Determine file extension from MIME type
    const extensionMap: Record<string, string> = {
      "text/html": ".html",
      "application/pdf": ".pdf",
      "image/png": ".png",
      "image/jpeg": ".jpg",
    };

    const ext = extensionMap[mime_type] || ".bin";
    const fileName =
      outputPath ||
      path.join(
        __dirname,
        `order-document-${document_type}-${Date.now()}${ext}`
      );

    try {
      const buffer = Buffer.from(file, "base64");
      fs.writeFileSync(fileName, buffer);

      console.log(`\n💾 Document saved to: ${fileName}`);
      console.log(`   File Size: ${buffer.length} bytes`);

      return fileName;
    } catch (error) {
      console.error(`\n❌ Failed to save document:`, error);
      throw error;
    }
  }

  /**
   * Test full workflow
   */
  async runTests(): Promise<void> {
    try {
      // 1. Initialize
      await this.initialize();

      // 2. Example test cases (you'll need real order item IDs from your Lazada shop)
      const testCases: DocumentTestRequest[] = [
        {
          orderItemIds: [279709], // Replace with real order item ID
          docType: "shippingLabel",
          outputPath: path.join(
            __dirname,
            "../_bmad-output/lazada-shipping-label.html"
          ),
        },
        // Uncomment to test invoice
        // {
        //   orderItemIds: [279709],
        //   docType: "invoice",
        //   outputPath: path.join(__dirname, "../_bmad-output/lazada-invoice.html"),
        // },
      ];

      for (const testCase of testCases) {
        try {
          const response = await this.fetchDocument(testCase);
          await this.saveDocument(response, testCase.outputPath);
        } catch (error) {
          console.error(
            `\n⚠️  Test case failed for ${testCase.docType}:`,
            error
          );
          // Continue with next test
        }
      }

      console.log("\n✅ All tests completed!");
    } catch (error) {
      console.error("\n❌ Fatal error:", error);
      process.exit(1);
    }
  }
}

/**
 * API Mechanism Summary
 *
 * Endpoint: /order/document/get
 * Method: GET
 * Auth: OAuth2 (accessToken required)
 *
 * Request Flow:
 * 1. Client sends GET request with:
 *    - doc_type: Type of document needed
 *    - order_item_ids: JSON array of item IDs
 *    - Signature: HMAC-SHA256 signed with appSecret
 *
 * 2. Lazada validates:
 *    - Access token validity
 *    - Signature correctness
 *    - Document availability
 *
 * 3. Response contains:
 *    - Base64-encoded file content
 *    - MIME type for file type detection
 *    - Document type identifier
 *
 * Supported Document Types:
 * - shippingLabel: Shipping label for order
 * - invoice: Tax invoice for order
 * - waybill: Waybill for logistics
 *
 * Response Format:
 * {
 *   "code": "0",                    // Success code
 *   "data": {
 *     "document": {
 *       "file": "base64string",    // Base64-encoded content
 *       "mime_type": "text/html",  // File MIME type
 *       "document_type": "shippingLabel"
 *     }
 *   },
 *   "request_id": "xxxxx"          // Request tracking ID
 * }
 */

// Run tests if executed directly
if (require.main === module) {
  const tester = new LazadaOrderDocumentTester();
  tester.runTests().catch((error) => {
    console.error("Fatal error:", error);
    process.exit(1);
  });
}

export { LazadaOrderDocumentTester, DocumentTestRequest, DocumentResponse };
