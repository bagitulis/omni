// Composable untuk platform-specific logic
export function useOperationPlatform(platform: string) {
  const isShopee = () => platform === "shopee";
  const isLazada = () => platform === "lazada";
  const isTiktok = () => platform === "tiktok";

  const getOrderActionCount = () => {
    return isShopee() ? 3 : 1;
  };

  return {
    isShopee,
    isLazada,
    isTiktok,
    getOrderActionCount,
  };
}
