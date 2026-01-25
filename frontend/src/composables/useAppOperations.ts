import apiService from "../services/api";

export function useAppOperations(addLog: (msg: string) => void) {
  const updatePrice = async (platform: string, priceType: string) => {
    return await executeOperation(`${platform}_operation`, {
      operation: "update_price",
      price_type: priceType,
    });
  };

  const exportOrders = async (platform: string, exportType: string, days: number = 7) => {
    return await executeOperation(`${platform}_operation`, {
      operation: "export_orders",
      export_type: exportType,
      days,
    });
  };

  const handleTokenOperation = async (
    platform: string,
    operation: string,
    code: string | null = null
  ) => {
    const params: Record<string, any> = { operation };
    if (operation === "update_code" && code) params.code = code;
    return await executeOperation(`${platform}_operation`, params);
  };

  const executeOperation = async (operation: string, params: Record<string, any> = {}) => {
    addLog(`🔄 Executing: ${operation}`);
    try {
      const response = await apiService.executeOperation(operation, params);
      if ((response as any).success) {
        addLog(`✅ ${operation} completed`);
      } else {
        addLog(`❌ ${operation} failed: ${(response as any).error}`);
      }
      return response;
    } catch (error: any) {
      addLog(`❌ ${operation} error: ${error.message}`);
      throw error;
    }
  };

  return {
    updatePrice,
    exportOrders,
    handleTokenOperation,
  };
}
