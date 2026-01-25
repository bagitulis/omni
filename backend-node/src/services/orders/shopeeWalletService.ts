/**
 * Shopee Wallet Service
 * SRP: Fetch wallet transactions from Shopee API
 * Export/formatting delegated to ShopeeWalletExporter
 */

import { ShopeeAPIClient } from "../../api/clients/shopeeAPIClient";
import { getLogger } from "../../utils/logger";
import { shopeeWalletExporter, ShopeeWalletExporter } from "./shopeeWalletExporter";

export interface WalletTransaction {
  Date: string;
  "Order SN": string;
  Description: string;
  Amount: number;
  Status: string;
  "Transaction Type": string;
  "Tab Type": string;
  "Buyer Name": string;
}

export interface ProcessedWalletData {
  transactions: WalletTransaction[];
  totalAmount: number;
  count: number;
}

export class ShopeeWalletService {
  private api: ShopeeAPIClient;
  private logger: any;
  private exporter: ShopeeWalletExporter;

  private readonly TRANSACTION_TAB_TYPES = {
    all: null,
    order_income: "wallet_order_income",
    adjustment: "wallet_adjustment_filter",
    wallet_payment: "wallet_wallet_payment",
    refund: "wallet_refund_from_order",
    withdrawals: "wallet_withdrawals",
    fast_escrow: "fast_escrow_repayment",
    fast_pay: "fast_pay",
    seller_loan: "seller_loan",
    corporate_loan: "corporate_loan",
  };

  constructor(apiClient: ShopeeAPIClient) {
    this.api = apiClient;
    this.logger = getLogger("ShopeeWalletService");
    this.exporter = shopeeWalletExporter;
  }

  /**
   * Get wallet transactions from Shopee API
   */
  async getTransactions(month?: number, year?: number, transactionTabType?: string): Promise<any[]> {
    try {
      const now = new Date();
      const targetMonth = month || now.getMonth() + 1;
      const targetYear = year || now.getFullYear();

      const firstDay = new Date(targetYear, targetMonth - 1, 1);
      const lastDay = targetMonth === 12
        ? new Date(targetYear + 1, 0, 1)
        : new Date(targetYear, targetMonth, 1);
      lastDay.setSeconds(lastDay.getSeconds() - 1);

      this.logger.info(`Processing transactions from ${firstDay.toDateString()} to ${lastDay.toDateString()}`);
      return await this.getTransactionsChunked(firstDay, lastDay, transactionTabType);
    } catch (error) {
      this.logger.error(`Error getting transactions: ${error}`);
      throw error;
    }
  }

  private async getTransactionsChunked(firstDay: Date, lastDay: Date, transactionTabType?: string): Promise<any[]> {
    const allTransactions: any[] = [];
    const processedTransactions = new Set<string>();
    let currentStart = new Date(firstDay);

    while (currentStart <= lastDay) {
      let currentEnd = new Date(currentStart);
      currentEnd.setDate(currentEnd.getDate() + 14);

      if (currentEnd > lastDay) currentEnd = new Date(lastDay);

      this.logger.info(`Processing chunk: ${currentStart.toDateString()} to ${currentEnd.toDateString()}`);
      const chunkTransactions = await this.processDateRange(currentStart, currentEnd, transactionTabType);

      if (chunkTransactions && chunkTransactions.length > 0) {
        const newTransactions = [];
        for (const tx of chunkTransactions) {
          const txKey = this.getTransactionKey(tx);
          if (!processedTransactions.has(txKey)) {
            processedTransactions.add(txKey);
            newTransactions.push(tx);
          }
        }

        if (newTransactions.length > 0) {
          allTransactions.push(...newTransactions);
          this.logger.info(`Added ${newTransactions.length} unique transactions from this chunk`);
        }
      }

      currentStart = new Date(currentEnd);
      currentStart.setSeconds(currentStart.getSeconds() + 1);
    }

    return allTransactions;
  }

  private async processDateRange(startDate: Date, endDate: Date, transactionTabType?: string): Promise<any[]> {
    const transactions: any[] = [];
    let pageNo = 0;
    let hasMore = true;
    const processedInRange = new Set<string>();

    while (hasMore) {
      pageNo++;

      try {
        const response = await this.api.request("/api/v2/payment/get_wallet_transaction_list", "GET", {
          page_no: pageNo,
          page_size: 100,
          create_time_from: Math.floor(startDate.getTime() / 1000),
          create_time_to: Math.floor(endDate.getTime() / 1000),
          money_flow: "MONEY_IN",
          ...(transactionTabType && { transaction_tab_type: transactionTabType }),
        });

        if (!response?.response?.transaction_list) {
          this.logger.info("No more transactions in this date range");
          break;
        }

        const newTransactions = response.response.transaction_list;
        if (newTransactions.length === 0) {
          this.logger.info("No more transactions in this date range");
          break;
        }

        const uniqueTransactions = [];
        for (const tx of newTransactions) {
          const txKey = this.getTransactionKey(tx);
          if (!processedInRange.has(txKey)) {
            processedInRange.add(txKey);
            uniqueTransactions.push(tx);
          }
        }

        if (uniqueTransactions.length > 0) {
          transactions.push(...uniqueTransactions);
          this.logger.info(`Found ${uniqueTransactions.length} new transactions`);
        }

        if (newTransactions.length < 100) hasMore = false;
        await new Promise((resolve) => setTimeout(resolve, 500));
      } catch (error) {
        this.logger.error(`Error processing page: ${error}`);
        break;
      }
    }

    return transactions;
  }

  private getTransactionKey(tx: any): string {
    return `${tx.order_sn}_${tx.create_time}_${tx.amount}`;
  }

  // Delegate to exporter
  processTransactions(rawTransactions: any[]): ProcessedWalletData {
    return this.exporter.processTransactions(rawTransactions);
  }

  async exportToCsv(processedData: ProcessedWalletData, filename: string): Promise<boolean> {
    return this.exporter.exportToCsv(processedData, filename);
  }

  extractOrderNumbers(transactions: WalletTransaction[]): Set<string> {
    return this.exporter.extractOrderNumbers(transactions);
  }

  formatAmount(amount: number): string {
    return this.exporter.formatAmount(amount);
  }

  getTransactionTabType(key: keyof typeof this.TRANSACTION_TAB_TYPES): string | null {
    return this.TRANSACTION_TAB_TYPES[key] as string | null;
  }
}
