/**
 * TikTok Ads Analytics MCP Server
 *
 * Standalone MCP server for AI agents to query TikTok Ads data
 * Run: npx ts-node backend/src/mcp/tiktokAdsMcpServer.ts
 */

import * as readline from "readline";
import {
  handleDashboard,
  handleTopProducts,
  handleCreativeComparison,
} from "./tiktokAdsMcpHandlers";

// Tool definitions
const TOOLS = [
  {
    name: "tiktok_ads_dashboard",
    description: "Get TikTok Ads dashboard summary with metrics",
    inputSchema: {
      type: "object",
      properties: {
        tenantId: { type: "string", description: "Tenant ID" },
      },
      required: ["tenantId"],
    },
  },
  {
    name: "tiktok_ads_top_products",
    description: "Get top performing products by revenue or ROI",
    inputSchema: {
      type: "object",
      properties: {
        tenantId: { type: "string", description: "Tenant ID" },
        sortBy: { type: "string", enum: ["revenue", "roi", "orders"] },
        limit: { type: "number", description: "Number of products" },
      },
      required: ["tenantId"],
    },
  },
  {
    name: "tiktok_ads_creative_comparison",
    description: "Compare Video vs Kartu Produk creatives",
    inputSchema: {
      type: "object",
      properties: { tenantId: { type: "string" } },
      required: ["tenantId"],
    },
  },
];

// MCP message types
interface McpRequest {
  jsonrpc: "2.0";
  id: number | string;
  method: string;
  params?: Record<string, unknown>;
}

interface McpResponse {
  jsonrpc: "2.0";
  id: number | string;
  result?: unknown;
  error?: { code: number; message: string };
}

// Handle MCP requests
async function handleRequest(request: McpRequest): Promise<McpResponse> {
  const { id, method, params } = request;

  try {
    switch (method) {
      case "initialize":
        return {
          jsonrpc: "2.0",
          id,
          result: {
            protocolVersion: "2024-11-05",
            serverInfo: { name: "tiktok-ads-analytics", version: "1.0.0" },
            capabilities: { tools: {} },
          },
        };

      case "tools/list":
        return { jsonrpc: "2.0", id, result: { tools: TOOLS } };

      case "tools/call": {
        const toolName = (params?.name as string) || "";
        const toolArgs = (params?.arguments as Record<string, unknown>) || {};
        let result: unknown;

        switch (toolName) {
          case "tiktok_ads_dashboard":
            result = await handleDashboard(toolArgs as { tenantId: string });
            break;
          case "tiktok_ads_top_products":
            result = await handleTopProducts(
              toolArgs as { tenantId: string; sortBy?: string; limit?: number }
            );
            break;
          case "tiktok_ads_creative_comparison":
            result = await handleCreativeComparison(
              toolArgs as { tenantId: string }
            );
            break;
          default:
            throw new Error(`Unknown tool: ${toolName}`);
        }

        return {
          jsonrpc: "2.0",
          id,
          result: {
            content: [{ type: "text", text: JSON.stringify(result, null, 2) }],
          },
        };
      }

      default:
        return {
          jsonrpc: "2.0",
          id,
          error: { code: -32601, message: `Method not found: ${method}` },
        };
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    return { jsonrpc: "2.0", id, error: { code: -32000, message } };
  }
}

// Main - stdio transport
async function main() {
  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
    terminal: false,
  });

  rl.on("line", async (line) => {
    try {
      const request = JSON.parse(line) as McpRequest;
      const response = await handleRequest(request);
      console.log(JSON.stringify(response));
    } catch {
      console.log(
        JSON.stringify({
          jsonrpc: "2.0",
          id: null,
          error: { code: -32700, message: "Parse error" },
        })
      );
    }
  });

  process.stderr.write("TikTok Ads MCP Server started\n");
}

main().catch((err) => process.stderr.write(`Error: ${err}\n`));
