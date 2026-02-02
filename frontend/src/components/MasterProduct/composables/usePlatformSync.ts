import { ref } from "vue";
import type { ProductSku } from "../ProductList.types";
import masterProductService from "@/services/masterProductService";

export function usePlatformSync() {
  const skusNeedingSync = ref<Map<number, Set<string>>>(new Map());
  const syncStatus = ref<Map<string, "syncing" | "success" | "error">>(
    new Map(),
  );
  const syncErrorMessages = ref<Map<string, string>>(new Map());
  const platformLoading = ref<Map<string, boolean>>(new Map());

  const isPlatformLinked = (sku: ProductSku, platform: string): boolean => {
    return (
      sku.platform_links?.some((link) => link.platform === platform) ?? false
    );
  };

  const needsSync = (skuId: number, platform: string): boolean => {
    return skusNeedingSync.value.get(skuId)?.has(platform) ?? false;
  };

  const getSyncState = (skuId: number, platform: string) => {
    return syncStatus.value.get(`${skuId}-${platform}`);
  };

  const isPlatformLoading = (skuId: number, platform: string): boolean => {
    return platformLoading.value.get(`${skuId}-${platform}`) ?? false;
  };

  const getPlatformTitle = (platform: string, sku: ProductSku): string => {
    const syncKey = `${sku.id}-${platform}`;
    const errorMsg = syncErrorMessages.value.get(syncKey);
    if (errorMsg) return errorMsg;

    const linked = isPlatformLinked(sku, platform);
    const platformName = platform.charAt(0).toUpperCase() + platform.slice(1);
    return linked
      ? `Terhubung ke ${platformName}`
      : `Belum terhubung ke ${platformName}`;
  };

  const markSkuAsNeedingSync = (sku: ProductSku) => {
    if (sku.platform_links?.length) {
      if (!skusNeedingSync.value.has(sku.id)) {
        skusNeedingSync.value.set(sku.id, new Set());
      }
      const platformSet = skusNeedingSync.value.get(sku.id)!;
      sku.platform_links.forEach((link) => {
        if (link.platform && link.platform_sku_id) {
          platformSet.add(link.platform);
        }
      });
    }
  };

  const syncToPlatform = async (
    productId: number,
    skuId: number,
    platform: string,
  ) => {
    const syncKey = `${skuId}-${platform}`;
    syncStatus.value.set(syncKey, "syncing");
    syncErrorMessages.value.delete(syncKey);

    try {
      await masterProductService.sync(productId, platform);
      syncStatus.value.set(syncKey, "success");

      // Remove from needing sync list
      const platformSet = skusNeedingSync.value.get(skuId);
      if (platformSet) {
        platformSet.delete(platform);
        if (platformSet.size === 0) {
          skusNeedingSync.value.delete(skuId);
        }
      }

      // Clear success status after 3 seconds
      setTimeout(() => {
        syncStatus.value.delete(syncKey);
      }, 3000);
    } catch (error: any) {
      console.error(`Sync to ${platform} failed:`, error);
      syncStatus.value.set(syncKey, "error");
      syncErrorMessages.value.set(
        syncKey,
        error.response?.data?.error || "Gagal sinkronisasi",
      );
    }
  };

  const handlePlatformClick = async (sku: ProductSku, platform: string) => {
    const loadingKey = `${sku.id}-${platform}`;
    if (platformLoading.value.get(loadingKey)) return;

    const isLinked = isPlatformLinked(sku, platform);

    if (isLinked) {
      if (confirm(`Putuskan koneksi dari ${platform}?`)) {
        platformLoading.value.set(loadingKey, true);
        try {
          await masterProductService.unlinkSku(sku.id, platform);
          if (sku.platform_links) {
            sku.platform_links = sku.platform_links.filter(
              (link) => link.platform !== platform,
            );
          }
        } catch (error: any) {
          console.error(`Failed to unlink ${platform}:`, error);
          alert(
            `Gagal memutuskan koneksi: ${
              error.response?.data?.error || error.message
            }`,
          );
        } finally {
          platformLoading.value.delete(loadingKey);
        }
      }
    } else {
      const platformItemId = prompt(
        `Masukkan ID Produk untuk ${platform} (Item ID):`,
      );
      if (!platformItemId) return;

      platformLoading.value.set(loadingKey, true);
      try {
        await masterProductService.manualLink(sku.id, platform, platformItemId);

        if (!sku.platform_links) {
          sku.platform_links = [];
        }

        sku.platform_links.push({
          id: Date.now(),
          master_product_id: sku.master_product_id,
          master_sku_id: sku.id,
          platform: platform,
          platform_product_id: platformItemId,
          sync_status: "linked",
        });

        if (!skusNeedingSync.value.has(sku.id)) {
          skusNeedingSync.value.set(sku.id, new Set());
        }
        skusNeedingSync.value.get(sku.id)!.add(platform);
      } catch (error: any) {
        console.error(`Failed to link ${platform}:`, error);
        alert(
          `Gagal menghubungkan: ${error.response?.data?.error || error.message}`,
        );
      } finally {
        platformLoading.value.delete(loadingKey);
      }
    }
  };

  return {
    skusNeedingSync,
    syncStatus,
    syncErrorMessages,
    platformLoading,
    isPlatformLinked,
    needsSync,
    getSyncState,
    isPlatformLoading,
    getPlatformTitle,
    syncToPlatform,
    handlePlatformClick,
    markSkuAsNeedingSync,
  };
}
