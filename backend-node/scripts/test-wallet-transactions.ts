/**
 * Test Script: Check Wallet Transactions for Price Data
 * Purpose: Investigate if wallet transactions contain order prices
 */

import { ShopeeWalletService } from '../src/services/orders/shopeeWalletService';
import { TokenManager } from '../src/services/tokenManager';

async function main() {
  console.log('🔍 Checking Wallet Transactions for 1 month ago...\n');

  // Get Shopee API client
  const tokenManager = TokenManager.getInstance();
  const shopeeApiClient = tokenManager.getShopeeClient();

  if (!shopeeApiClient) {
    throw new Error('Shopee API client not initialized');
  }

  // Create wallet service
  const walletService = new ShopeeWalletService(shopeeApiClient);

  // Get transactions from 1 month ago (December 2025)
  const now = new Date();
  const month = now.getMonth(); // December = 11 (0-indexed), so 11+1 - 1 = 11 for Dec
  const year = now.getFullYear();
  
  console.log(`📅 Fetching transactions for: Month ${month}, Year ${year}\n`);

  try {
    // Get order income transactions (contains order prices)
    const rawTransactions = await walletService.getTransactions(
      month,
      year,
      'wallet_order_income' // Only order income
    );

    console.log(`✅ Found ${rawTransactions.length} wallet transactions\n`);

    if (rawTransactions.length === 0) {
      console.log('❌ No transactions found for this period');
      return;
    }

    // Process transactions
    const processedData = walletService.processTransactions(rawTransactions);

    console.log('📊 Sample Transactions (First 5):');
    console.log('─'.repeat(80));
    console.log('Date       | Order SN       | Amount    | Description');
    console.log('─'.repeat(80));

    processedData.transactions.slice(0, 5).forEach((tx: any) => {
      console.log(
        `${tx.Date} | ${tx['Order SN'].padEnd(14)} | ${tx.Amount.toString().padStart(9)} | ${tx.Description.substring(0, 30)}`
      );
    });

    console.log('─'.repeat(80));

    console.log(`\n📈 Summary:`);
    console.log(`   Total Transactions: ${processedData.count}`);
    console.log(`   Total Amount:       Rp ${processedData.totalAmount.toLocaleString('id-ID')}`);

    // Check raw transaction structure
    console.log(`\n🔍 Raw Transaction Structure (First item):`);
    const firstRaw = rawTransactions[0];
    console.log(JSON.stringify(firstRaw, null, 2));

    // Save sample data
    const fs = require('fs');
    const path = require('path');
    const outputPath = path.join(__dirname, 'wallet-sample-transactions.json');
    fs.writeFileSync(
      outputPath,
      JSON.stringify(
        {
          processed: processedData.transactions.slice(0, 10),
          raw: rawTransactions.slice(0, 3),
        },
        null,
        2
      )
    );
    console.log(`\n💾 Sample data saved to: ${outputPath}`);
  } catch (error) {
    console.error('❌ Error fetching wallet transactions:', error);
  }
}

main().catch(console.error);
