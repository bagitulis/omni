import {
  ref,
  computed,
  watch,
  onMounted,
  type Ref,
  type ComputedRef,
} from "vue";
import wholesaleService from "@/services/wholesaleService";
import type {
  WholesaleSettings,
  BatchUpdateBySkusResult,
  WholesaleTierCalculated,
} from "@/types/wholesale";

export interface UpdateItem {
  sku: string;
  price: number;
}

export interface PreviewItem extends UpdateItem {
  tiers: WholesaleTierCalculated[];
}

export interface UseWholesaleUpdateReturn {
  activeTab: Ref<"preview" | "settings">;
  processing: Ref<boolean>;
  result: Ref<BatchUpdateBySkusResult | null>;
  error: Ref<string>;
  settingsSaved: Ref<boolean>;
  settings: Ref<WholesaleSettings>;
  tier2Min: ComputedRef<number>;
  tier2Max: ComputedRef<number>;
  tier3Min: ComputedRef<number>;
  previewItems: ComputedRef<PreviewItem[]>;
  resultClass: ComputedRef<Record<string, boolean>>;
  loadSettings: () => Promise<void>;
  saveSettings: () => Promise<void>;
  handleUpdate: (items: UpdateItem[]) => Promise<void>;
  formatPrice: (price: number | undefined) => string;
  resetState: () => void;
}

export function useWholesaleUpdate(
  showRef: Ref<boolean>,
  itemsRef: Ref<UpdateItem[]>,
  onCompleted: (result: BatchUpdateBySkusResult) => void,
): UseWholesaleUpdateReturn {
  const activeTab = ref<"preview" | "settings">("preview");
  const processing = ref(false);
  const result = ref<BatchUpdateBySkusResult | null>(null);
  const error = ref("");
  const settingsSaved = ref(false);

  const settings = ref<WholesaleSettings>({
    tenant_id: "",
    platform: "shopee",
    admin_fee: 1500,
    min_order_1: 2,
    max_order_1: 3,
    max_order_tier_3: 1000,
  });

  const tier2Min = computed(() => settings.value.max_order_1 + 1);
  const tier2Max = computed(() => tier2Min.value + 1);
  const tier3Min = computed(() => tier2Max.value + 1);

  const previewItems = computed<PreviewItem[]>(() => {
    return itemsRef.value.slice(0, 5).map((item) => ({
      ...item,
      tiers: wholesaleService.calculateTiersLocal(item.price, settings.value),
    }));
  });

  const resultClass = computed(() => ({
    success: result.value?.success === true,
    partial:
      result.value?.success === false &&
      (result.value?.data?.processed ?? 0) > 0,
    failed:
      result.value?.success === false && result.value?.data?.processed === 0,
  }));

  function resetState() {
    result.value = null;
    error.value = "";
    settingsSaved.value = false;
    activeTab.value = "preview";
  }

  async function loadSettings() {
    try {
      const loaded = await wholesaleService.getSettings();
      if (loaded) {
        settings.value = loaded;
      }
    } catch (err) {
      console.error("Failed to load settings:", err);
    }
  }

  async function saveSettings() {
    try {
      const res = await wholesaleService.updateSettings({
        admin_fee: settings.value.admin_fee,
        min_order_1: settings.value.min_order_1,
        max_order_1: settings.value.max_order_1,
        max_order_tier_3: settings.value.max_order_tier_3,
      });
      if (res.success) {
        settingsSaved.value = true;
        setTimeout(() => (settingsSaved.value = false), 2000);
      }
    } catch (err) {
      console.error("Failed to save settings:", err);
    }
  }

  async function handleUpdate(items: UpdateItem[]) {
    if (items.length === 0) return;

    processing.value = true;
    error.value = "";
    result.value = null;

    try {
      const updateResult = await wholesaleService.batchUpdateBySkus(items);
      result.value = updateResult;
      onCompleted(updateResult);
    } catch (err: unknown) {
      const errMsg =
        err instanceof Error ? err.message : "Gagal update wholesale";
      error.value = errMsg;
    } finally {
      processing.value = false;
    }
  }

  function formatPrice(price: number | undefined): string {
    if (price === undefined) return "-";
    return new Intl.NumberFormat("id-ID").format(price);
  }

  watch(showRef, async (newVal) => {
    if (newVal) {
      resetState();
      await loadSettings();
    }
  });

  onMounted(async () => {
    await loadSettings();
  });

  return {
    activeTab,
    processing,
    result,
    error,
    settingsSaved,
    settings,
    tier2Min,
    tier2Max,
    tier3Min,
    previewItems,
    resultClass,
    loadSettings,
    saveSettings,
    handleUpdate,
    formatPrice,
    resetState,
  };
}
