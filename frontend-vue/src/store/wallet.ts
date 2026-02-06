/**
 * Wallet Store
 * SRP: Manages wallet transactions
 * Refactored to use shared actions (no circular dependency)
 */
import { defineStore } from "pinia";
import { ref } from "vue";
import { executeOperation } from "@/composables/useSharedAppActions";

export const useWalletStore = defineStore("wallet", () => {
  const walletParams = ref({
    month: new Date().getMonth() + 1,
    year: new Date().getFullYear(),
    transaction_type: "wallet_order_income",
  });

  async function getWalletTransactions(
    params: Record<string, any>
  ): Promise<any> {
    return await executeOperation("get_wallet_transactions", params);
  }

  function updateWalletParams(newParams: Record<string, any>) {
    walletParams.value = { ...walletParams.value, ...newParams };
  }

  return {
    walletParams,
    getWalletTransactions,
    updateWalletParams,
  };
});
