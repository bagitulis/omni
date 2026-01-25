/**
 * Shopee Wallet Exporter
 * SRP: Export and format wallet transaction data
 */

import * as fs from "fs";
import * as path from "path";
import { getLogger } from "../../utils/logger";
import { WalletTransaction, ProcessedWalletData } from "./shopeeWalletService";

const logger = getLogger("ShopeeWalletExporter");

export class ShopeeWalletExporter {
  /**
   * Process raw transaction data into structured format
   */
  processTransactions(rawTransactions: any[]): ProcessedWalletData {
    const transactions: WalletTransaction[] = [];
    let totalAmount = 0;

    for (const tx of rawTransactions) {
      try {
        const date = new Date(tx.create_time * 1000);
        const amount = parseFloat(tx.amount || "0");
        totalAmount += amount;

        transactions.push({
          Date: date.toISOString().split("T")[0],
          "Order SN": tx.order_sn || "",
          Description: tx.description || "",
          Amount: amount,
          Status: tx.status || "",
          "Transaction Type": tx.transaction_type || "",
          "Tab Type": tx.transaction_tab_type || "",
          "Buyer Name": tx.buyer_name || "Unknown",
        });
      } catch {
        continue;
      }
    }

    transactions.sort(
      (a, b) => new Date(a.Date).getTime() - new Date(b.Date).getTime()
    );

    return { transactions, totalAmount, count: transactions.length };
  }

  /**
   * Export transactions to CSV
   */
  async exportToCsv(
    processedData: ProcessedWalletData,
    filename: string
  ): Promise<boolean> {
    try {
      if (
        !processedData.transactions ||
        processedData.transactions.length === 0
      ) {
        logger.warn("No transactions to export");
        return false;
      }

      const dir = path.dirname(filename);
      if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir, { recursive: true });
      }

      const headers = [
        "Date",
        "Order SN",
        "Description",
        "Amount",
        "Status",
        "Transaction Type",
        "Tab Type",
        "Buyer Name",
      ];
      const rows = processedData.transactions.map((tx) => [
        tx.Date,
        tx["Order SN"],
        tx.Description,
        tx.Amount.toString(),
        tx.Status,
        tx["Transaction Type"],
        tx["Tab Type"],
        tx["Buyer Name"],
      ]);

      const csvContent = [
        headers.join(","),
        ...rows.map((row) => row.map((cell) => `"${cell}"`).join(",")),
      ].join("\n");

      fs.writeFileSync(filename, csvContent, "utf-8");
      logger.info(`✅ Wallet transactions exported`);

      return true;
    } catch (error) {
      logger.error(`Error exporting to CSV: ${error}`);
      return false;
    }
  }

  /**
   * Extract unique order numbers from transactions
   */
  extractOrderNumbers(transactions: WalletTransaction[]): Set<string> {
    const orderNumbers = new Set<string>();

    for (const transaction of transactions) {
      const orderSN = transaction["Order SN"];
      if (orderSN && orderSN.trim()) {
        orderNumbers.add(orderSN.trim());
      }
    }

    return orderNumbers;
  }

  /**
   * Format amount with currency
   */
  formatAmount(amount: number): string {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 2,
    }).format(amount);
  }
}

export const shopeeWalletExporter = new ShopeeWalletExporter();
