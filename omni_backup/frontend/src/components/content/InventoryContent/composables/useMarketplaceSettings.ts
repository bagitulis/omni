import { ref, computed } from "vue";

/**
 * useMarketplaceSettings Composable
 * RESPONSIBILITY: Manage marketplace allocation settings
 * - Store and persist total/auto column selections
 * - Provide dropdown options from schema columns
 */

interface MarketplaceSettings {
  totalColumn: string;
  autoColumn: string;
  shopeeRatio: number;
  tiktokRatio: number;
}

const STORAGE_KEY = "inventory_marketplace_settings";

// Singleton state
const settings = ref<MarketplaceSettings>({
  totalColumn: "Total",
  autoColumn: "Auto",
  shopeeRatio: 0.6,
  tiktokRatio: 0.3,
});

const isLoaded = ref(false);

function loadSettings(): MarketplaceSettings {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved) {
      const parsed = JSON.parse(saved);
      return { ...settings.value, ...parsed };
    }
  } catch (error) {
    console.warn("⚠️ Failed to load marketplace settings:", error);
  }
  return settings.value;
}

function saveSettings(newSettings: MarketplaceSettings): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(newSettings));
    settings.value = newSettings;
  } catch (error) {
    console.warn("⚠️ Failed to save marketplace settings:", error);
  }
}

export function useMarketplaceSettings() {
  // Load on first use
  if (!isLoaded.value) {
    settings.value = loadSettings();
    isLoaded.value = true;
  }

  const totalColumn = computed({
    get: () => settings.value.totalColumn,
    set: (val: string) => {
      settings.value.totalColumn = val;
      saveSettings(settings.value);
    },
  });

  const autoColumn = computed({
    get: () => settings.value.autoColumn,
    set: (val: string) => {
      settings.value.autoColumn = val;
      saveSettings(settings.value);
    },
  });

  const shopeeRatio = computed({
    get: () => settings.value.shopeeRatio,
    set: (val: number) => {
      settings.value.shopeeRatio = val;
      saveSettings(settings.value);
    },
  });

  const tiktokRatio = computed({
    get: () => settings.value.tiktokRatio,
    set: (val: number) => {
      settings.value.tiktokRatio = val;
      saveSettings(settings.value);
    },
  });

  function updateSettings(partial: Partial<MarketplaceSettings>): void {
    const newSettings = { ...settings.value, ...partial };
    saveSettings(newSettings);
  }

  function getSettings(): MarketplaceSettings {
    return { ...settings.value };
  }

  function resetToDefaults(): void {
    const defaults: MarketplaceSettings = {
      totalColumn: "Total",
      autoColumn: "Auto",
      shopeeRatio: 0.6,
      tiktokRatio: 0.3,
    };
    saveSettings(defaults);
  }

  return {
    totalColumn,
    autoColumn,
    shopeeRatio,
    tiktokRatio,
    settings: computed(() => settings.value),
    updateSettings,
    getSettings,
    resetToDefaults,
  };
}
