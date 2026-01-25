import { defineStore } from "pinia";
import apiService from "../services/api";

export const useOrdersStore = defineStore("orders", () => {
  async function executeOrderExport(
    platform: string,
    orderType: string,
    days: number = 7
  ): Promise<any> {
    try {
      return await apiService.post("/execute-order-export", {
        platform,
        order_type: orderType,
        days,
      });
    } catch (error) {
      console.error("Order export failed:", error);
      throw error;
    }
  }

  return {
    executeOrderExport,
  };
});
