/**
 * Shipping Store
 * SRP: Manages shipping files and fee processing
 * Refactored to use shared actions (no circular dependency)
 */
import { defineStore } from "pinia";
import { ref, Ref } from "vue";
import apiService from "../services/api";
import { addLog, executeOperation } from "@/composables/useSharedAppActions";

export const useShippingStore = defineStore("shipping", () => {
  const shippingFiles: Ref<any[]> = ref([]);
  const shippingParams = ref({
    month: new Date().getMonth() + 1,
    year: new Date().getFullYear(),
    option: 1,
  });

  async function getShippingFiles(): Promise<any> {
    try {
      const response = await apiService.getShippingFiles();
      if ((response as any).success) {
        shippingFiles.value = (response as any).files || [];
        return response;
      }
      return response;
    } catch (error: any) {
      addLog(`❌ Error getting shipping files: ${error.message}`);
      throw error;
    }
  }

  async function processShippingFile(filename: string): Promise<any> {
    try {
      const response = await apiService.processShippingFile(filename);
      if ((response as any).success) {
        addLog(`✅ Shipping file ${filename} processed successfully`);
      } else {
        addLog(
          `❌ Failed to process shipping file: ${(response as any).message || (response as any).error}`
        );
      }
      return response;
    } catch (error: any) {
      addLog(`❌ Error processing shipping file: ${error.message}`);
      throw error;
    }
  }

  async function processShippingFee(params: Record<string, any>): Promise<any> {
    return await executeOperation("process_shipping_fee", params);
  }

  function updateShippingParams(newParams: Record<string, any>) {
    shippingParams.value = { ...shippingParams.value, ...newParams };
  }

  return {
    shippingFiles,
    shippingParams,
    getShippingFiles,
    processShippingFile,
    processShippingFee,
    updateShippingParams,
  };
});
